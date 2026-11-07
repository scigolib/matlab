// Package matlab provides a library for reading and writing MATLAB .mat files.
// It supports both v5 (MATLAB v5-v7.2) and v7.3+ (HDF5-based) formats.
package matlab

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"

	"github.com/scigolib/matlab/internal/v5"
	"github.com/scigolib/matlab/internal/v73"
	"github.com/scigolib/matlab/types"
)

// ErrUnsupportedVersion indicates an unsupported MAT-file version.
var ErrUnsupportedVersion = errors.New("unsupported MAT-file version")

// ErrInvalidFormat indicates an invalid MAT-file format.
var ErrInvalidFormat = errors.New("invalid MAT-file format")

// MatFile represents a parsed MAT-file.
type MatFile struct {
	Version     string            // MAT-file version (e.g., "5.0", "7.3")
	Endian      string            // Byte order indicator ("MI" or "IM")
	Description string            // File description from header
	Variables   []*types.Variable // List of variables in the file
}

// Open reads and parses a MAT-file from an io.Reader.
func Open(r io.Reader) (*MatFile, error) {
	// The stream size must be taken before anything is read: a reader that
	// only reports Len (bytes.Buffer) tells the unread remainder, so a size
	// taken after the header would understate the file by 128 bytes. It must
	// also be taken before the header is prepended, since io.MultiReader hides
	// it. The v5 parser needs the real bound to validate element sizes against
	// the bytes actually present (ADR-003); an unknown size leaves only the
	// absolute payload cap as a bound.
	streamSize, _ := v5.ReaderSize(r)

	// Read and check the first 128 bytes to determine format
	header := make([]byte, 128)
	if _, err := io.ReadFull(r, header); err != nil {
		return nil, err
	}

	// Create a MultiReader to re-include the header
	fullReader := io.MultiReader(bytes.NewReader(header), r)

	// A bare HDF5 file (signature at offset 0), as our own v7.3 writer produces.
	if isHDF5Format(header) {
		return parseV73(fullReader)
	}

	// Otherwise the first 128 bytes are a MAT header, and its version field
	// tells v5 from v7.3. MATLAB writes v7.3 files as HDF5 behind a 512-byte
	// user block that holds this header, so the HDF5 signature is at offset 512.
	if isV5Format(header) {
		version := matHeaderVersion(header)
		switch version {
		case matVersion5:
			return parseV5(fullReader, streamSize)
		case matVersion73:
			// hdf5 locates the superblock behind the user block itself.
			return parseV73(fullReader)
		default:
			return nil, fmt.Errorf("%w: 0x%04x", ErrUnsupportedVersion, version)
		}
	}

	return nil, ErrInvalidFormat
}

// MAT header version field (bytes 124-125, read in the byte order named by the
// endian indicator). Same values as matio v1.6.0 Mat_Open (src/mat.c, version
// 0x0100 = v5, 0x0200 = v7.3 via H5Fopen) and scipy v1.18.1
// _get_matfile_version (scipy/io/matlab/_miobase.py: major 1 = v5, 2 = v7.3,
// anything else is an error).
const (
	matVersion5  uint16 = 0x0100
	matVersion73 uint16 = 0x0200
)

// matHeaderVersion reads the version field of a MAT header whose endian
// indicator has already been validated by isV5Format.
func matHeaderVersion(header []byte) uint16 {
	if string(header[126:128]) == "MI" {
		return binary.BigEndian.Uint16(header[124:126])
	}
	return binary.LittleEndian.Uint16(header[124:126])
}

// isHDF5Format checks for HDF5 signature.
func isHDF5Format(header []byte) bool {
	// HDF5 signature: 0x89 0x48 0x44 0x46 0x0d 0x0a 0x1a 0x0a
	return bytes.HasPrefix(header, []byte{0x89, 0x48, 0x44, 0x46, 0x0d, 0x0a, 0x1a, 0x0a})
}

// isV5Format checks for v5 format signature.
func isV5Format(header []byte) bool {
	// Check endian indicator at bytes 126-127
	endian := string(header[126:128])
	return endian == "IM" || endian == "MI"
}

// parseV5 parses v5 format MAT-files. streamSize is the length of the input the
// fullReader wraps, or 0 when it could not be determined.
func parseV5(r io.Reader, streamSize int64) (*MatFile, error) {
	parser, err := v5.NewParserSized(r, streamSize)
	if err != nil {
		return nil, err
	}

	v5File, err := parser.Parse()
	if err != nil {
		return nil, err
	}

	return &MatFile{
		Version:     "5.0",
		Endian:      v5File.Header.EndianIndicator,
		Description: v5File.Header.Description,
		Variables:   v5File.Variables,
	}, nil
}

// parseV73 parses v7.3 format MAT-files (HDF5-based).
func parseV73(r io.Reader) (*MatFile, error) {
	parser := v73.NewParser()
	variables, err := parser.Parse(r)
	if err != nil {
		return nil, err
	}

	return &MatFile{
		Version:   "7.3",
		Variables: variables,
	}, nil
}

// GetVariable retrieves a variable by name.
// Returns nil if the variable is not found.
//
// Example:
//
//	matFile, err := matlab.Open(file)
//	if err != nil {
//	    return err
//	}
//	data := matFile.GetVariable("mydata")
//	if data != nil {
//	    fmt.Println("Found:", data.Name)
//	}
func (m *MatFile) GetVariable(name string) *types.Variable {
	for _, v := range m.Variables {
		if v.Name == name {
			return v
		}
	}
	return nil
}

// GetVariableNames returns the names of all variables in the file.
// Useful for listing available data without accessing values.
//
// Example:
//
//	matFile, err := matlab.Open(file)
//	if err != nil {
//	    return err
//	}
//	names := matFile.GetVariableNames()
//	fmt.Println("Variables:", names) // ["x", "y", "z"]
func (m *MatFile) GetVariableNames() []string {
	names := make([]string, len(m.Variables))
	for i, v := range m.Variables {
		names[i] = v.Name
	}
	return names
}

// HasVariable checks if a variable with the given name exists.
//
// Example:
//
//	if matFile.HasVariable("results") {
//	    data := matFile.GetVariable("results")
//	    // process data
//	}
func (m *MatFile) HasVariable(name string) bool {
	return m.GetVariable(name) != nil
}
