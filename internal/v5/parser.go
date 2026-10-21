package v5

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"math"

	"github.com/scigolib/matlab/types"
)

// Parser handles parsing of v5 MAT-files.
type Parser struct {
	r      io.Reader
	Header *Header
	pos    int64
	limit  int64 // total stream length when known; 0 = unknown
}

// Mat5File represents a parsed v5 MAT-file.
type Mat5File struct {
	Header    *Header
	Variables []*types.Variable
}

// NewParser creates a new v5 parser, deriving the stream bound from the reader
// when it can report its size (see ReaderSize).
func NewParser(r io.Reader) (*Parser, error) {
	limit, _ := ReaderSize(r)
	return newParser(r, limit)
}

// NewParserSized creates a v5 parser with an explicit bound on the stream's
// total length. Use it when the reader given to the parser hides the size of
// the underlying input — matlab.Open wraps its reader in an io.MultiReader,
// which does exactly that and passes the real length here. A limit of zero or
// less means unknown, leaving only the absolute cap as a bound.
func NewParserSized(r io.Reader, limit int64) (*Parser, error) {
	return newParser(r, limit)
}

func newParser(r io.Reader, limit int64) (*Parser, error) {
	if limit < 0 {
		limit = 0
	}

	p := &Parser{r: r, limit: limit}
	if err := p.parseHeader(); err != nil {
		return nil, err
	}
	return p, nil
}

// parseHeader reads and parses the MAT-file header (128 bytes).
func (p *Parser) parseHeader() error {
	header := make([]byte, 128)
	if _, err := io.ReadFull(p.r, header); err != nil {
		return err
	}
	p.pos += 128

	hdr, err := parseHeader(header)
	if err != nil {
		return err
	}
	p.Header = hdr
	return nil
}

// Parse reads the entire MAT-file.
//
//nolint:gocognit // Parser complexity is acceptable for handling multiple element types.
func (p *Parser) Parse() (*Mat5File, error) {
	file := &Mat5File{
		Header: p.Header,
	}

	for {
		tag, err := p.readTag()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, err
		}

		switch tag.DataType {
		case miMATRIX:
			variable, err := p.parseMatrix(tag)
			if err != nil {
				return nil, err
			}
			file.Variables = append(file.Variables, variable)
		case miCOMPRESSED:
			// Decompress the data
			decompressed, err := p.decompress(tag.Size)
			if err != nil {
				return nil, err
			}
			p.pos += int64(tag.Size)

			// Note: Compressed elements do NOT have padding after the data.
			// The next element starts immediately after the compressed bytes.

			// Parse the decompressed content (should contain a miMATRIX element).
			// The buffer length is the bound every nested tag is checked against.
			sub := &Parser{
				r:      bytes.NewReader(decompressed),
				Header: p.Header,
				limit:  int64(len(decompressed)),
				pos:    0,
			}

			// Read the tag from decompressed data
			subTag, err := sub.readTag()
			if err != nil {
				return nil, err
			}

			// The decompressed data should contain a matrix
			if subTag.DataType == miMATRIX {
				variable, err := sub.parseMatrix(subTag)
				if err != nil {
					return nil, err
				}
				file.Variables = append(file.Variables, variable)
			}
		default:
			p.skipData(tag)
		}
	}

	return file, nil
}

// parseMatrix parses a matrix element.
func (p *Parser) parseMatrix(tag *DataTag) (*types.Variable, error) {
	data, err := p.readPayload(tag.Size)
	if err != nil {
		return nil, err
	}
	p.pos += int64(tag.Size)

	// The matrix payload is the bound every sub-element tag is checked against:
	// a sub-element cannot legitimately extend past the element that contains it.
	sub := &Parser{
		r:      bytes.NewReader(data),
		Header: p.Header,
		limit:  int64(len(data)),
		pos:    0,
	}
	return sub.parseMatrixContent()
}

// parseMatrixContent parses the components of a matrix.
func (p *Parser) parseMatrixContent() (*types.Variable, error) {
	// Read array flags
	flagsTag, err := p.readTag()
	if err != nil {
		return nil, err
	}
	flagsData, err := p.readData(flagsTag)
	if err != nil {
		return nil, err
	}

	if len(flagsData) != 8 {
		return nil, errors.New("invalid array flags size")
	}

	// Per MAT-file v5 spec: word #1 contains class (bits 0-7) and flags (bits 8-11).
	// Word #2 is nzmax and is not used for non-sparse arrays.
	combined := p.Header.Order.Uint32(flagsData[:4])
	class := combined & 0xFF              // bits 0-7
	isComplex := (combined & 0x0800) != 0 // bit 11

	// Read dimensions
	dimsTag, err := p.readTag()
	if err != nil {
		return nil, err
	}
	dimsData, err := p.readData(dimsTag)
	if err != nil {
		return nil, err
	}

	dimensions, err := p.parseDimensions(dimsData)
	if err != nil {
		return nil, err
	}

	// Read variable name
	nameTag, err := p.readTag()
	if err != nil {
		return nil, err
	}
	nameData, err := p.readData(nameTag)
	if err != nil {
		return nil, err
	}
	name := string(nameData)

	// Read real data
	realTag, err := p.readTag()
	if err != nil {
		return nil, err
	}
	realData, err := p.readData(realTag)
	if err != nil {
		return nil, err
	}
	realValue := p.convertData(realData, realTag.DataType, class)

	// Read imaginary data if complex
	var imagValue interface{}
	if isComplex {
		imagTag, err := p.readTag()
		if err != nil {
			return nil, err
		}
		imagData, err := p.readData(imagTag)
		if err != nil {
			return nil, err
		}
		imagValue = p.convertData(imagData, imagTag.DataType, class)
	}

	// Create variable
	variable := &types.Variable{
		Name:       name,
		Dimensions: dimensions,
		DataType:   classToDataType(class),
		Data:       realValue,
		IsComplex:  isComplex,
	}

	// For complex numbers, create a complex array
	if isComplex {
		variable.Data = &types.NumericArray{
			Real:       realValue,
			Imag:       imagValue,
			Dimensions: dimensions,
			Type:       classToDataType(class),
		}
	}

	return variable, nil
}

// parseDimensions decodes the dimensions sub-element payload.
//
// Dimensions are a signed int32 array (MAT v5 spec). A payload that is not a
// whole number of entries, or an entry with the sign bit set, is corrupt:
// converting it to int would describe a different array per platform (negative
// on 32-bit, a huge positive on 64-bit).
func (p *Parser) parseDimensions(dimsData []byte) ([]int, error) {
	if len(dimsData)%4 != 0 {
		return nil, fmt.Errorf("invalid dimensions payload: %d bytes is not a multiple of 4", len(dimsData))
	}

	dimCount := len(dimsData) / 4
	dimensions := make([]int, dimCount)
	for i := 0; i < dimCount; i++ {
		raw := p.Header.Order.Uint32(dimsData[i*4 : (i+1)*4])
		if raw > math.MaxInt32 {
			return nil, fmt.Errorf("invalid dimension value %d at index %d: dimensions are an int32 array", raw, i)
		}
		dimensions[i] = int(raw)
	}

	return dimensions, nil
}

// readData reads data for a given tag.
func (p *Parser) readData(tag *DataTag) ([]byte, error) {
	// For small format, data is already captured in the tag
	if tag.IsSmall {
		return tag.SmallData, nil
	}

	// Regular format: read data from stream
	data, err := p.readPayload(tag.Size)
	if err != nil {
		return nil, err
	}
	p.pos += int64(tag.Size)

	// Skip padding to 8-byte boundary
	padding := (8 - tag.Size%8) % 8
	if padding > 0 {
		_, _ = io.CopyN(io.Discard, p.r, int64(padding))
		p.pos += int64(padding)
	}

	return data, nil
}

// skipData skips over data for a given tag.
func (p *Parser) skipData(tag *DataTag) {
	// For small format, data was already read with the tag - nothing to skip
	if tag.IsSmall {
		return
	}

	// Regular format: skip data and padding
	_, _ = io.CopyN(io.Discard, p.r, int64(tag.Size))
	p.pos += int64(tag.Size)

	padding := (8 - tag.Size%8) % 8
	if padding > 0 {
		_, _ = io.CopyN(io.Discard, p.r, int64(padding))
		p.pos += int64(padding)
	}
}
