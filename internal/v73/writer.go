// Package v73 provides internal v7.3 MAT-file parsing and writing via HDF5.
package v73

import (
	"fmt"

	"github.com/scigolib/hdf5"
	"github.com/scigolib/matlab/types"
)

// Writer handles writing v7.3 MAT-files (HDF5 format).
//
// The writer creates HDF5 files with MATLAB-compatible attributes.
// Each MATLAB variable is stored as an HDF5 dataset with a MATLAB_class
// attribute indicating the original MATLAB type.
type Writer struct {
	file *hdf5.FileWriter
}

// NewWriter creates a new v7.3 writer.
//
// The file will be created in HDF5 format compatible with MATLAB v7.3+.
// If the file exists, it will be truncated and overwritten.
//
// Parameters:
//   - filename: Path to the file to create
//
// Returns:
//   - *Writer: The created writer
//   - error: If file creation fails
func NewWriter(filename string) (*Writer, error) {
	// Create HDF5 file with truncate mode (overwrite if exists)
	file, err := hdf5.CreateForWrite(filename, hdf5.CreateTruncate)
	if err != nil {
		return nil, fmt.Errorf("failed to create HDF5 file: %w", err)
	}

	return &Writer{file: file}, nil
}

// WriteVariable writes a MATLAB variable as HDF5 dataset with proper attributes.
//
// The variable is validated before writing. The data is converted to the
// appropriate HDF5 datatype and written as a dataset. A MATLAB_class
// attribute is added to indicate the MATLAB type.
//
// Parameters:
//   - v: Variable to write (must not be nil)
//
// Returns:
//   - error: If validation or writing fails
//
// Supported types:
//   - Double, Single, Int8, Uint8, Int16, Uint16, Int32, Uint32, Int64, Uint64
//   - Complex numbers (future implementation)
func (w *Writer) WriteVariable(v *types.Variable) error {
	// Check for nil first
	if v == nil {
		return fmt.Errorf("variable cannot be nil")
	}

	// Validate input
	if err := w.validateVariable(v); err != nil {
		return fmt.Errorf("invalid variable: %w", err)
	}

	// Handle complex numbers separately (future implementation)
	if v.IsComplex {
		return w.writeComplexVariable(v)
	}

	// Write as regular dataset
	return w.writeSimpleVariable(v)
}

// validateVariable checks if variable has all required fields.
func (w *Writer) validateVariable(v *types.Variable) error {
	if v.Name == "" {
		return fmt.Errorf("variable name is required")
	}
	if len(v.Dimensions) == 0 {
		return fmt.Errorf("variable dimensions are required")
	}
	if v.Data == nil {
		return fmt.Errorf("variable data is required")
	}
	return nil
}

// writeSimpleVariable writes non-complex variable as HDF5 dataset.
func (w *Writer) writeSimpleVariable(v *types.Variable) error {
	// Step 1: Convert dimensions to uint64 (HDF5 API requirement)
	dims := make([]uint64, len(v.Dimensions))
	for i, d := range v.Dimensions {
		if d <= 0 {
			return fmt.Errorf("invalid dimension at index %d: %d (must be positive)", i, d)
		}
		dims[i] = uint64(d)
	}

	// Step 2: Map MATLAB type to HDF5 datatype
	hdf5Type, err := w.dataTypeToHDF5(v.DataType)
	if err != nil {
		return fmt.Errorf("unsupported data type: %w", err)
	}

	// Step 3: Create dataset using HDF5 API
	dataset, err := w.file.CreateDataset("/"+v.Name, hdf5Type, dims)
	if err != nil {
		return fmt.Errorf("failed to create dataset: %w", err)
	}

	// Step 4: Write data
	if err := dataset.Write(v.Data); err != nil {
		return fmt.Errorf("failed to write data: %w", err)
	}

	// Step 5: Add MATLAB_class attribute
	matlabClass := w.dataTypeToMatlabClass(v.DataType)
	if err := dataset.WriteAttribute("MATLAB_class", matlabClass); err != nil {
		return fmt.Errorf("failed to write MATLAB_class attribute: %w", err)
	}

	return nil
}

// writeComplexVariable writes complex variable (real + imaginary).
//
// WORKAROUND: Due to HDF5 library limitations, complex numbers are stored as
// two separate datasets at root level:
// - varname_real: real part
// - varname_imag: imaginary part
//
// Standard MATLAB format uses groups (/varname/real, /varname/imag), but the
// HDF5 library doesn't yet support nested datasets or group attributes.
// See: docs/dev/notes/BUG_REPORT_HDF5_GROUP_ATTRIBUTES.md.
func (w *Writer) writeComplexVariable(v *types.Variable) error {
	// Extract real and imaginary parts
	numArray, ok := v.Data.(*types.NumericArray)
	if !ok {
		return fmt.Errorf("complex variable must have *types.NumericArray data, got %T", v.Data)
	}

	if numArray.Real == nil || numArray.Imag == nil {
		return fmt.Errorf("complex variable must have both Real and Imag parts")
	}

	// Convert dimensions to uint64
	dims := make([]uint64, len(v.Dimensions))
	for i, d := range v.Dimensions {
		if d <= 0 {
			return fmt.Errorf("invalid dimension at index %d: %d (must be positive)", i, d)
		}
		dims[i] = uint64(d)
	}

	// Map MATLAB type to HDF5 datatype
	hdf5Type, err := w.dataTypeToHDF5(v.DataType)
	if err != nil {
		return fmt.Errorf("unsupported data type: %w", err)
	}

	// Create real dataset (flat structure workaround)
	realName := v.Name + "_real"
	realDataset, err := w.file.CreateDataset("/"+realName, hdf5Type, dims)
	if err != nil {
		return fmt.Errorf("failed to create real dataset: %w", err)
	}

	// Write real data
	if err := realDataset.Write(numArray.Real); err != nil {
		return fmt.Errorf("failed to write real data: %w", err)
	}

	// Add MATLAB_class attribute to real dataset
	matlabClass := w.dataTypeToMatlabClass(v.DataType)
	if err := realDataset.WriteAttribute("MATLAB_class", matlabClass); err != nil {
		return fmt.Errorf("failed to write MATLAB_class attribute: %w", err)
	}

	// Create imaginary dataset (flat structure workaround)
	imagName := v.Name + "_imag"
	imagDataset, err := w.file.CreateDataset("/"+imagName, hdf5Type, dims)
	if err != nil {
		return fmt.Errorf("failed to create imaginary dataset: %w", err)
	}

	// Write imaginary data
	if err := imagDataset.Write(numArray.Imag); err != nil {
		return fmt.Errorf("failed to write imaginary data: %w", err)
	}

	// Add MATLAB_class attribute to imaginary dataset
	if err := imagDataset.WriteAttribute("MATLAB_class", matlabClass); err != nil {
		return fmt.Errorf("failed to write MATLAB_class attribute: %w", err)
	}

	// Note: MATLAB_complex attribute would normally indicate this is a complex variable,
	// but HDF5 library has issues writing multiple attributes.
	// The _real/_imag suffix is sufficient for identification.

	return nil
}

// dataTypeToHDF5 converts MATLAB DataType to HDF5 Datatype.
func (w *Writer) dataTypeToHDF5(dt types.DataType) (hdf5.Datatype, error) {
	switch dt {
	case types.Double:
		return hdf5.Float64, nil
	case types.Single:
		return hdf5.Float32, nil
	case types.Int8:
		return hdf5.Int8, nil
	case types.Uint8:
		return hdf5.Uint8, nil
	case types.Int16:
		return hdf5.Int16, nil
	case types.Uint16:
		return hdf5.Uint16, nil
	case types.Int32:
		return hdf5.Int32, nil
	case types.Uint32:
		return hdf5.Uint32, nil
	case types.Int64:
		return hdf5.Int64, nil
	case types.Uint64:
		return hdf5.Uint64, nil
	default:
		return 0, fmt.Errorf("unsupported MATLAB data type: %v", dt)
	}
}

// dataTypeToMatlabClass converts types.DataType to MATLAB class string.
//
// This string is stored in the MATLAB_class attribute and tells MATLAB
// what type the variable should be when loading the file.
func (w *Writer) dataTypeToMatlabClass(dt types.DataType) string {
	switch dt {
	case types.Double:
		return matlabClassDouble
	case types.Single:
		return "single"
	case types.Int8:
		return "int8"
	case types.Uint8:
		return "uint8"
	case types.Int16:
		return "int16"
	case types.Uint16:
		return "uint16"
	case types.Int32:
		return "int32"
	case types.Uint32:
		return "uint32"
	case types.Int64:
		return "int64"
	case types.Uint64:
		return "uint64"
	case types.Char:
		return "char"
	default:
		return matlabClassDouble // Default fallback
	}
}

// Close closes the underlying HDF5 file.
//
// After calling Close, the writer cannot be used anymore.
// It is safe to call Close multiple times.
//
// Returns:
//   - error: If closing the HDF5 file fails
func (w *Writer) Close() error {
	if w.file != nil {
		err := w.file.Close()
		w.file = nil // Mark as closed
		return err
	}
	return nil
}
