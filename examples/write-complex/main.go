package main

import (
	"fmt"
	"log"

	"github.com/scigolib/matlab"
	"github.com/scigolib/matlab/types"
)

// Example program demonstrating how to write complex numbers to a MATLAB v7.3 file.
//
// This program creates a .mat file with a complex double array and demonstrates
// the proper MATLAB v7.3 HDF5 format structure:
// - /z (group with MATLAB_class="double", MATLAB_complex=1)
//   - /real (dataset)
//   - /imag (dataset)
func main() {
	// Create output file
	filename := "complex_example.mat"
	fmt.Printf("Creating MATLAB file: %s\n", filename)

	writer, err := matlab.Create(filename, matlab.Version73)
	if err != nil {
		log.Fatalf("Failed to create file: %v", err)
	}
	defer func() {
		if err := writer.Close(); err != nil {
			log.Printf("Warning: failed to close writer: %v", err)
		}
	}()

	// Define complex variable: z = [1+2i, 3+4i, 5+6i]
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

	fmt.Println("Writing complex variable 'z' = [1+2i, 3+4i, 5+6i]")
	if err := writer.WriteVariable(z); err != nil {
		log.Fatalf("Failed to write variable: %v", err)
	}

	fmt.Println("Success! File created with proper MATLAB v7.3 format:")
	fmt.Println("  /z (group)")
	fmt.Println("    - MATLAB_class = 'double'")
	fmt.Println("    - MATLAB_complex = 1")
	fmt.Println("    /real (dataset: [1.0, 3.0, 5.0])")
	fmt.Println("    /imag (dataset: [2.0, 4.0, 6.0])")
	fmt.Println("\nYou can verify the structure with: h5dump complex_example.mat")
}
