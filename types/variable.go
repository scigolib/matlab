package types

import "fmt"

// DataType represents MATLAB data types.
type DataType int

// MATLAB data type constants.
const (
	Double DataType = iota
	Single
	Int8
	Uint8
	Int16
	Uint16
	Int32
	Uint32
	Int64
	Uint64
	Char
	Struct
	CellArray
	Object
	Unknown
)

func (d DataType) String() string {
	return [...]string{
		"double", "single", "int8", "uint8", "int16", "uint16",
		"int32", "uint32", "int64", "uint64", "char", "struct", "cell", "object", "unknown",
	}[d]
}

// Variable represents a MATLAB variable.
type Variable struct {
	Name       string                 // Variable name
	Dimensions []int                  // Array dimensions
	DataType   DataType               // Data type identifier
	Data       interface{}            // Actual data
	IsComplex  bool                   // True for complex numbers
	IsSparse   bool                   // True for sparse matrices
	Attributes map[string]interface{} // Additional metadata
}

// String returns a string representation of the variable.
func (v *Variable) String() string {
	return fmt.Sprintf("%s: %s %v", v.Name, v.DataType, v.Dimensions)
}

// GetAttribute retrieves an attribute by name.
func (v *Variable) GetAttribute(name string) (interface{}, bool) {
	if v.Attributes == nil {
		return nil, false
	}
	val, ok := v.Attributes[name]
	return val, ok
}

// GetFloat64Array extracts variable data as []float64.
// Supports automatic conversion from float32, int types.
// Returns error if data is complex or incompatible type.
//
// Example:
//
//	variable := matFile.GetVariable("data")
//	data, err := variable.GetFloat64Array()
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println(data) // [1.0, 2.0, 3.0]
//
//nolint:gocyclo,cyclop // Type conversion requires checking all numeric types
func (v *Variable) GetFloat64Array() ([]float64, error) {
	if v.IsComplex {
		return nil, fmt.Errorf("cannot convert complex data to []float64, use GetComplex128Array()")
	}

	switch data := v.Data.(type) {
	case []float64:
		return data, nil

	case []float32:
		result := make([]float64, len(data))
		for i, val := range data {
			result[i] = float64(val)
		}
		return result, nil

	case []int8:
		result := make([]float64, len(data))
		for i, val := range data {
			result[i] = float64(val)
		}
		return result, nil

	case []int16:
		result := make([]float64, len(data))
		for i, val := range data {
			result[i] = float64(val)
		}
		return result, nil

	case []int32:
		result := make([]float64, len(data))
		for i, val := range data {
			result[i] = float64(val)
		}
		return result, nil

	case []int64:
		result := make([]float64, len(data))
		for i, val := range data {
			result[i] = float64(val)
		}
		return result, nil

	case []uint8:
		result := make([]float64, len(data))
		for i, val := range data {
			result[i] = float64(val)
		}
		return result, nil

	case []uint16:
		result := make([]float64, len(data))
		for i, val := range data {
			result[i] = float64(val)
		}
		return result, nil

	case []uint32:
		result := make([]float64, len(data))
		for i, val := range data {
			result[i] = float64(val)
		}
		return result, nil

	case []uint64:
		result := make([]float64, len(data))
		for i, val := range data {
			result[i] = float64(val)
		}
		return result, nil

	default:
		return nil, fmt.Errorf("cannot convert %T to []float64", v.Data)
	}
}

// GetInt32Array extracts variable data as []int32.
// Supports conversion from smaller integer types.
// Returns error if data contains non-integer values or is out of range.
func (v *Variable) GetInt32Array() ([]int32, error) {
	if v.IsComplex {
		return nil, fmt.Errorf("cannot convert complex data to []int32")
	}

	switch data := v.Data.(type) {
	case []int32:
		return data, nil

	case []int8:
		result := make([]int32, len(data))
		for i, val := range data {
			result[i] = int32(val)
		}
		return result, nil

	case []int16:
		result := make([]int32, len(data))
		for i, val := range data {
			result[i] = int32(val)
		}
		return result, nil

	case []uint8:
		result := make([]int32, len(data))
		for i, val := range data {
			result[i] = int32(val)
		}
		return result, nil

	case []uint16:
		result := make([]int32, len(data))
		for i, val := range data {
			result[i] = int32(val)
		}
		return result, nil

	default:
		return nil, fmt.Errorf("cannot convert %T to []int32", v.Data)
	}
}

// GetComplex128Array extracts complex variable data as []complex128.
// Returns error if data is not complex.
//
// Example:
//
//	variable := matFile.GetVariable("signal")
//	data, err := variable.GetComplex128Array()
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println(data) // [(1+4i), (2+5i), (3+6i)]
func (v *Variable) GetComplex128Array() ([]complex128, error) {
	if !v.IsComplex {
		return nil, fmt.Errorf("variable is not complex")
	}

	numArray, ok := v.Data.(*NumericArray)
	if !ok {
		return nil, fmt.Errorf("complex data is not *NumericArray")
	}

	// Handle different numeric types for real part.
	switch realPart := numArray.Real.(type) {
	case []float64:
		imagPart, ok := numArray.Imag.([]float64)
		if !ok {
			return nil, fmt.Errorf("real and imag have different types")
		}
		if len(realPart) != len(imagPart) {
			return nil, fmt.Errorf("real and imag arrays have different lengths")
		}
		result := make([]complex128, len(realPart))
		for i := range realPart {
			result[i] = complex(realPart[i], imagPart[i])
		}
		return result, nil

	case []float32:
		imagPart, ok := numArray.Imag.([]float32)
		if !ok {
			return nil, fmt.Errorf("real and imag have different types")
		}
		if len(realPart) != len(imagPart) {
			return nil, fmt.Errorf("real and imag arrays have different lengths")
		}
		result := make([]complex128, len(realPart))
		for i := range realPart {
			result[i] = complex(float64(realPart[i]), float64(imagPart[i]))
		}
		return result, nil

	default:
		return nil, fmt.Errorf("unsupported complex data type: %T", numArray.Real)
	}
}

// GetScalar extracts a scalar value (single element).
// Returns error if variable has more than one element.
//
// Example:
//
//	variable := matFile.GetVariable("temperature")
//	value, err := variable.GetScalar()
//	if err != nil {
//	    log.Fatal(err)
//	}
//	temp := value.(float64)
//
//nolint:gocognit,gocyclo,cyclop // Type extraction requires checking all numeric types
func (v *Variable) GetScalar() (interface{}, error) {
	// Calculate total elements.
	totalElements := 1
	for _, dim := range v.Dimensions {
		totalElements *= dim
	}

	if totalElements != 1 {
		return nil, fmt.Errorf("variable has %d elements, not a scalar", totalElements)
	}

	// Extract first element based on type.
	switch data := v.Data.(type) {
	case []float64:
		if len(data) > 0 {
			return data[0], nil
		}
	case []float32:
		if len(data) > 0 {
			return data[0], nil
		}
	case []int8:
		if len(data) > 0 {
			return data[0], nil
		}
	case []int16:
		if len(data) > 0 {
			return data[0], nil
		}
	case []int32:
		if len(data) > 0 {
			return data[0], nil
		}
	case []int64:
		if len(data) > 0 {
			return data[0], nil
		}
	case []uint8:
		if len(data) > 0 {
			return data[0], nil
		}
	case []uint16:
		if len(data) > 0 {
			return data[0], nil
		}
	case []uint32:
		if len(data) > 0 {
			return data[0], nil
		}
	case []uint64:
		if len(data) > 0 {
			return data[0], nil
		}
	default:
		return nil, fmt.Errorf("cannot extract scalar from %T", v.Data)
	}

	return nil, fmt.Errorf("data slice is empty")
}
