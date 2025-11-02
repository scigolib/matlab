package v73

import (
	"fmt"
	"io"
	"os"

	"github.com/scigolib/hdf5"
	"github.com/scigolib/matlab/types"
)

// Parser handles parsing of v7.3 MAT-files (HDF5 format).
type Parser struct{}

// NewParser creates a new v7.3 parser.
func NewParser() *Parser {
	return &Parser{}
}

// Parse reads the HDF5-based MAT-file.
// Since the HDF5 library requires a file path, we create a temporary file
// from the io.Reader, parse it, and then clean up.
func (p *Parser) Parse(r io.Reader) ([]*types.Variable, error) {
	// Create temporary file
	tmpFile, err := os.CreateTemp("", "matfile-*.tmp")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()

	// Ensure cleanup
	defer func() {
		_ = tmpFile.Close()
		_ = os.Remove(tmpPath)
	}()

	// Copy reader to temp file
	if _, err := io.Copy(tmpFile, r); err != nil {
		return nil, fmt.Errorf("failed to copy data to temp file: %w", err)
	}

	// Close to flush before opening with HDF5
	if err := tmpFile.Close(); err != nil {
		return nil, fmt.Errorf("failed to close temp file: %w", err)
	}

	// Open with HDF5
	file, err := hdf5.Open(tmpPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open HDF5 file: %w", err)
	}
	defer file.Close() //nolint:errcheck // Best effort cleanup

	// Create adapter and convert to MATLAB variables
	adapter := NewHDF5Adapter(file)
	return adapter.ConvertToMatlab()
}
