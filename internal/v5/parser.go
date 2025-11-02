package v5

import (
	"bytes"
	"errors"
	"io"

	"github.com/scigolib/matlab/types"
)

// Parser handles parsing of v5 MAT-files.
type Parser struct {
	r      io.Reader
	Header *Header
	pos    int64
}

// Mat5File represents a parsed v5 MAT-file.
type Mat5File struct {
	Header    *Header
	Variables []*types.Variable
}

// NewParser creates a new v5 parser.
func NewParser(r io.Reader) (*Parser, error) {
	p := &Parser{r: r}
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
			return nil, errors.New("compressed data not yet supported")
		default:
			p.skipData(tag)
		}
	}

	return file, nil
}

// parseMatrix parses a matrix element.
func (p *Parser) parseMatrix(tag *DataTag) (*types.Variable, error) {
	data := make([]byte, tag.Size)
	if _, err := io.ReadFull(p.r, data); err != nil {
		return nil, err
	}
	p.pos += int64(tag.Size)

	sub := &Parser{
		r:      bytes.NewReader(data),
		Header: p.Header,
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

	flags := p.Header.Order.Uint32(flagsData[:4])
	class := p.Header.Order.Uint32(flagsData[4:8])
	isComplex := (flags & 0x0800) != 0

	// Read dimensions
	dimsTag, err := p.readTag()
	if err != nil {
		return nil, err
	}
	dimsData, err := p.readData(dimsTag)
	if err != nil {
		return nil, err
	}

	dimCount := len(dimsData) / 4
	dimensions := make([]int, dimCount)
	for i := 0; i < dimCount; i++ {
		dimensions[i] = int(p.Header.Order.Uint32(dimsData[i*4 : (i+1)*4]))
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

// readData reads data for a given tag.
func (p *Parser) readData(tag *DataTag) ([]byte, error) {
	data := make([]byte, tag.Size)
	if _, err := io.ReadFull(p.r, data); err != nil {
		return nil, err
	}
	p.pos += int64(tag.Size)

	// Skip padding for large elements
	if !tag.IsSmall {
		padding := (8 - tag.Size%8) % 8
		if padding > 0 {
			_, _ = io.CopyN(io.Discard, p.r, int64(padding))
			p.pos += int64(padding)
		}
	}

	return data, nil
}

// skipData skips over data for a given tag.
func (p *Parser) skipData(tag *DataTag) {
	_, _ = io.CopyN(io.Discard, p.r, int64(tag.Size))
	p.pos += int64(tag.Size)

	if !tag.IsSmall {
		padding := (8 - tag.Size%8) % 8
		if padding > 0 {
			_, _ = io.CopyN(io.Discard, p.r, int64(padding))
			p.pos += int64(padding)
		}
	}
}
