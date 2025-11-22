// Package v73 provides internal v7.3 MAT-file parsing and writing via HDF5.
package v73

import (
	"fmt"
	"math"

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
// For complex numbers, a group structure is created with nested real/imag datasets
// and appropriate MATLAB_class and MATLAB_complex attributes.
//
// Parameters:
//   - v: Variable to write (must not be nil)
//
// Returns:
//   - error: If validation or writing fails
//
// Supported types:
//   - Double, Single, Int8, Uint8, Int16, Uint16, Int32, Uint32, Int64, Uint64
//   - Complex numbers (stored as HDF5 groups with /real and /imag datasets)
func (w *Writer) WriteVariable(v *types.Variable) error {
	// Check for nil first
	if v == nil {
		return fmt.Errorf("variable cannot be nil")
	}

	// Validate input
	if err := w.validateVariable(v); err != nil {
		return fmt.Errorf("invalid variable: %w", err)
	}

	// Handle complex numbers separately (group structure with nested datasets)
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

	// Validate dimensions are positive and check for overflow
	total := int64(1)
	for i, d := range v.Dimensions {
		if d <= 0 {
			return fmt.Errorf("dimension[%d] must be positive, got %d", i, d)
		}

		// Check for overflow before multiplying
		if d > 0 && total > math.MaxInt/int64(d) {
			return fmt.Errorf("dimensions overflow (total elements too large): %v", v.Dimensions)
		}

		total *= int64(d)
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

// writeComplexVariable writes complex variable in proper MATLAB v7.3 format.
//
// MATLAB v7.3 stores complex numbers as HDF5 groups with nested datasets:
// - /varname (group with MATLAB_class and MATLAB_complex attributes)
//   - /real (dataset containing real part)
//   - /imag (dataset containing imaginary part)
//
// This matches the standard MATLAB format specification for HDF5-based .mat files.
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

	// Step 1: Create group for variable
	group, err := w.file.CreateGroup("/" + v.Name)
	if err != nil {
		return fmt.Errorf("failed to create group for complex variable: %w", err)
	}

	// Step 2: Write MATLAB metadata to group
	matlabClass := w.dataTypeToMatlabClass(v.DataType)
	if err := group.WriteAttribute("MATLAB_class", matlabClass); err != nil {
		return fmt.Errorf("failed to write MATLAB_class attribute: %w", err)
	}

	// MATLAB_complex attribute indicates this is a complex number
	if err := group.WriteAttribute("MATLAB_complex", uint8(1)); err != nil {
		return fmt.Errorf("failed to write MATLAB_complex attribute: %w", err)
	}

	// Step 3: Create nested datasets for real/imag parts
	realPath := "/" + v.Name + "/real"
	imagPath := "/" + v.Name + "/imag"

	realDataset, err := w.file.CreateDataset(realPath, hdf5Type, dims)
	if err != nil {
		return fmt.Errorf("failed to create real dataset: %w", err)
	}

	imagDataset, err := w.file.CreateDataset(imagPath, hdf5Type, dims)
	if err != nil {
		return fmt.Errorf("failed to create imaginary dataset: %w", err)
	}

	// Step 4: Write data
	if err := realDataset.Write(numArray.Real); err != nil {
		return fmt.Errorf("failed to write real data: %w", err)
	}

	if err := imagDataset.Write(numArray.Imag); err != nil {
		return fmt.Errorf("failed to write imaginary data: %w", err)
	}

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
