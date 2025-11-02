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
