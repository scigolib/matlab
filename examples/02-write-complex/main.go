// Command 02-write-complex writes a complex double array to a MATLAB v7.3 file.
//
// The v7.3 writer stores a complex variable as an HDF5 group named after the
// variable, with datasets "real" and "imag" and the attributes
// MATLAB_class="double" and MATLAB_complex=1. MATLAB itself uses a compound
// datatype instead; see docs/guides/DATATYPES.md.
//
// Usage, from the repository root:
//
//	go run ./examples/02-write-complex
package main

import (
	"fmt"
	"log"

	"github.com/scigolib/matlab"
	"github.com/scigolib/matlab/types"
)

func main() {
	if err := run("complex_example.mat"); err != nil {
		log.Fatal(err)
	}
}

func run(filename string) error {
	writer, err := matlab.Create(filename, matlab.Version73)
	if err != nil {
		return fmt.Errorf("create file: %w", err)
	}

	// z = [1+2i, 3+4i, 5+6i]
	z := &types.Variable{
		Name:       "z",
		Dimensions: []int{3},
		DataType:   types.Double,
		IsComplex:  true,
		Data: &types.NumericArray{
			Real: []float64{1.0, 3.0, 5.0},
			Imag: []float64{2.0, 4.0, 6.0},
		},
	}

	if err := writer.WriteVariable(z); err != nil {
		_ = writer.Close() // best effort: the write error is the one reported
		return fmt.Errorf("write variable: %w", err)
	}

	// Close finishes the file: its error matters.
	if err := writer.Close(); err != nil {
		return fmt.Errorf("close writer: %w", err)
	}

	fmt.Println("Wrote", filename)
	fmt.Println("Inspect the layout with: h5dump", filename)
	return nil
}
