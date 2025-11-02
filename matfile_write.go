package matlab

import (
	"errors"
	"fmt"

	"github.com/scigolib/matlab/internal/v73"
	"github.com/scigolib/matlab/types"
)

// Version represents MAT-file format version for writing.
type Version int

const (
	// Version5 represents v5-v7.2 format (binary format).
	// Note: v5 writer is not yet implemented (planned for v0.3.0).
	Version5 Version = 5

	// Version73 represents v7.3+ format (HDF5-based).
	// This is the recommended format for new files.
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

	// v5 specific (future)
	// v5writer *v5.Writer
}

// Create creates a new MATLAB file for writing.
//
// Parameters:
//   - filename: Path to the file to create (will overwrite if exists)
//   - version: MAT-file format version (Version5 or Version73)
//
// Returns:
//   - *MatFileWriter: Handle to the created file
//   - error: If file creation fails or version is unsupported
//
// Example:
//
//	writer, err := matlab.Create("output.mat", matlab.Version73)
//	if err != nil {
//	    return err
//	}
//	defer writer.Close()
//
//	err = writer.WriteVariable(&types.Variable{
//	    Name:       "mydata",
//	    Dimensions: []int{3},
//	    DataType:   types.Double,
//	    Data:       []float64{1.0, 2.0, 3.0},
//	})
func Create(filename string, version Version) (*MatFileWriter, error) {
	if filename == "" {
		return nil, errors.New("filename cannot be empty")
	}

	switch version {
	case Version73:
		return createV73(filename)
	case Version5:
		return nil, errors.New("v5 writer not yet implemented (planned for v0.3.0)")
	default:
		return nil, fmt.Errorf("unsupported MAT-file version: %d", version)
	}
}

// createV73 creates a v7.3 format writer.
func createV73(filename string) (*MatFileWriter, error) {
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
		return errors.New("v5 writer not yet implemented")
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
		// Future: close v5 writer
		return nil
	default:
		return nil
	}
}
