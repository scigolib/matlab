// Package matlab provides a library for reading and writing MATLAB .mat files.
// It supports both v5 (MATLAB v5-v7.2) and v7.3+ (HDF5-based) formats.
package matlab

import (
	"bytes"
	"errors"
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
	// Read and check the first 128 bytes to determine format
	header := make([]byte, 128)
	if _, err := io.ReadFull(r, header); err != nil {
		return nil, err
	}

	// Create a MultiReader to re-include the header
	fullReader := io.MultiReader(bytes.NewReader(header), r)

	// Check for HDF5 format (MATLAB v7.3+)
	if isHDF5Format(header) {
		return parseV73(fullReader)
	}

	// Check for v5 format (MATLAB v5-v7.2)
	if isV5Format(header) {
		return parseV5(fullReader)
	}

	return nil, ErrInvalidFormat
}

// isHDF5Format checks for HDF5 signature.
func isHDF5Format(header []byte) bool {
	// HDF5 signature: 0x89 0x48 0x44 0x46 0x0d 0x0a 0x1a 0x0a
	return bytes.HasPrefix(header, []byte{0x89, 0x48, 0x44, 0x46, 0x0d, 0x0a, 0x1a, 0x0a})
}

// isV5Format checks for v5 format signature.
func isV5Format(header []byte) bool {
	endian := string(header[124:128])
	return endian == "IM" || endian == "MI"
}

// parseV5 parses v5 format MAT-files.
func parseV5(r io.Reader) (*MatFile, error) {
	parser, err := v5.NewParser(r)
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
