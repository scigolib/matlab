// Package types provides common data structures for MATLAB variables and arrays.
package types

// Array represents a MATLAB array.
type Array interface {
	Dims() []int           // Array dimensions
	Size() int             // Total number of elements
	ElementType() DataType // Data type of elements
}

// NumericArray represents a numeric array.
type NumericArray struct {
	Real       interface{} // Real part data (slice of numbers)
	Imag       interface{} // Imaginary part data (optional)
	Dimensions []int       // Array dimensions
	Type       DataType    // Data type
}

// Dims returns the array dimensions.
func (n NumericArray) Dims() []int { return n.Dimensions }

// Size returns the total number of elements.
func (n NumericArray) Size() int { return numElements(n.Dimensions) }

// ElementType returns the data type of elements.
func (n NumericArray) ElementType() DataType { return n.Type }

// CharArray represents a character array.
type CharArray struct {
	Data       []rune
	Dimensions []int
}

// Dims returns the array dimensions.
func (c CharArray) Dims() []int { return c.Dimensions }

// Size returns the total number of elements.
func (c CharArray) Size() int { return numElements(c.Dimensions) }

// ElementType returns the data type of elements.
func (c CharArray) ElementType() DataType { return Char }

// numElements calculates total elements from dimensions.
func numElements(dims []int) int {
	if len(dims) == 0 {
		return 0
	}
	elements := 1
	for _, d := range dims {
		elements *= d
	}
	return elements
}
