package matlab

import (
	"encoding/binary"
	"errors"
	"fmt"
	"os"

	"github.com/scigolib/matlab/internal/v5"
	"github.com/scigolib/matlab/internal/v73"
	"github.com/scigolib/matlab/types"
)

// Version represents MAT-file format version for writing.
type Version int

const (
	// Version5 represents v5-v7.2 format (binary format).
	// Recommended for smaller files and maximum compatibility.
	Version5 Version = 5

	// Version73 represents v7.3+ format (HDF5-based).
	// Recommended for large files and modern MATLAB versions.
	Version73 Version = 73
)

// MatFileWriter represents a MATLAB file opened for writing.
//
// The writer automatically selects the appropriate backend based on
// the requested version (v5 or v7.3). After creating a writer, use
// WriteVariable to add variables to the file, then Close to finalize.
type MatFileWriter struct {
	filename string
	version  Version

	// v7.3 specific
	v73writer *v73.Writer

	// v5 specific
	v5writer *v5.Writer
	v5file   *os.File
}

// Create creates a new MATLAB file for writing with optional configuration.
//
// The version parameter specifies the format: Version5 or Version73.
// Optional parameters can be provided using functional options.
//
// Supported options:
//   - WithEndianness(binary.ByteOrder) - v5 byte order (default: LittleEndian)
//   - WithDescription(string) - v5 file description (max 116 bytes)
//   - WithCompression(int) - compression level 0-9 (not yet implemented)
//
// Example (basic):
//
//	writer, err := matlab.Create("output.mat", matlab.Version5)
//
// Example (with options):
//
//	writer, err := matlab.Create("output.mat", matlab.Version5,
//	    matlab.WithEndianness(binary.BigEndian),
//	    matlab.WithDescription("Simulation results"))
//
// Example (v7.3):
//
//	writer, err := matlab.Create("output.mat", matlab.Version73)
func Create(filename string, version Version, opts ...Option) (*MatFileWriter, error) {
	if filename == "" {
		return nil, errors.New("filename cannot be empty")
	}

	// Apply default config
	cfg := defaultConfig()
	applyOptions(cfg, opts)

	// Create based on version
	switch version {
	case Version73:
		return createV73(filename, cfg)
	case Version5:
		return createV5(filename, cfg)
	default:
		return nil, fmt.Errorf("unsupported MAT-file version: %d", version)
	}
}

// createV73 creates a v7.3 format writer with configuration.
func createV73(filename string, cfg *config) (*MatFileWriter, error) {
	// Note: v73 doesn't use endianness or description (HDF5 handles that)
	_ = cfg // Avoid unused parameter warning

	writer, err := v73.NewWriter(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to create v7.3 writer: %w", err)
	}

	return &MatFileWriter{
		filename:  filename,
		version:   Version73,
		v73writer: writer,
	}, nil
}

// createV5 creates a v5 format writer with configuration.
func createV5(filename string, cfg *config) (*MatFileWriter, error) {
	// Create file
	//nolint:gosec // G304: filename is provided by user for MAT-file creation, expected behavior
	f, err := os.Create(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to create file: %w", err)
	}

	// Determine endianness string
	var endian string
	if cfg.endianness == binary.LittleEndian {
		endian = "MI"
	} else {
		endian = "IM"
	}

	// Create v5 writer (writes header immediately) with config
	writer, err := v5.NewWriter(f, cfg.description, endian)
	if err != nil {
		//nolint:errcheck,gosec // G104: File cleanup after error, error logged elsewhere
		f.Close()
		return nil, fmt.Errorf("failed to create v5 writer: %w", err)
	}

	return &MatFileWriter{
		filename: filename,
		version:  Version5,
		v5writer: writer,
		v5file:   f,
	}, nil
}

// WriteVariable writes a variable to the MATLAB file.
//
// The variable must have valid Name, Dimensions, DataType, and Data fields.
// The data is written immediately to the underlying storage.
//
// Parameters:
//   - v: Variable to write (must not be nil)
//
// Returns:
//   - error: If writing fails or variable is invalid
//
// Supported types:
//   - Double, Single (float64, float32)
//   - Int8, Uint8, Int16, Uint16, Int32, Uint32, Int64, Uint64
//   - Complex numbers (use types.NumericArray with Real/Imag)
//
// Example:
//
//	// Simple array
//	writer.WriteVariable(&types.Variable{
//	    Name:       "A",
//	    Dimensions: []int{2, 3},
//	    DataType:   types.Double,
//	    Data:       []float64{1, 2, 3, 4, 5, 6},
//	})
//
//	// Complex numbers
//	writer.WriteVariable(&types.Variable{
//	    Name:       "C",
//	    Dimensions: []int{2},
//	    DataType:   types.Double,
//	    IsComplex:  true,
//	    Data: &types.NumericArray{
//	        Real: []float64{1.0, 3.0},
//	        Imag: []float64{2.0, 4.0},
//	    },
//	})
func (w *MatFileWriter) WriteVariable(v *types.Variable) error {
	if v == nil {
		return errors.New("variable cannot be nil")
	}

	switch w.version {
	case Version73:
		if w.v73writer == nil {
			return errors.New("v7.3 writer is not initialized")
		}
		return w.v73writer.WriteVariable(v)
	case Version5:
		if w.v5writer == nil {
			return errors.New("v5 writer is not initialized")
		}
		return w.v5writer.WriteVariable(v)
	default:
		return fmt.Errorf("unsupported version: %d", w.version)
	}
}

// Close closes the MATLAB file and flushes all data to disk.
//
// After calling Close, the writer cannot be used anymore. Any subsequent
// calls to WriteVariable or Close will fail.
//
// It is safe to call Close multiple times - subsequent calls will be no-ops.
//
// Returns:
//   - error: If flushing or closing fails
func (w *MatFileWriter) Close() error {
	switch w.version {
	case Version73:
		if w.v73writer != nil {
			err := w.v73writer.Close()
			w.v73writer = nil // Mark as closed
			return err
		}
		return nil
	case Version5:
		if w.v5file != nil {
			err := w.v5file.Close()
			w.v5writer = nil // Mark as closed
			w.v5file = nil
			return err
		}
		return nil
	default:
		return nil
	}
}
