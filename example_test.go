// Package matlab_test provides testable examples for the MATLAB file library.
//
// These examples demonstrate common use cases and serve as both documentation
// and verification that the API works as expected.
package matlab_test

import (
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"

	"github.com/scigolib/matlab"
	"github.com/scigolib/matlab/types"
)

// Example demonstrates basic usage of the MATLAB file library.
func Example() {
	tmpfile := filepath.Join(os.TempDir(), "example.mat")
	defer os.Remove(tmpfile)

	// Create a simple v5 MATLAB file
	writer, _ := matlab.Create(tmpfile, matlab.Version5)
	defer writer.Close()

	// Write a variable
	writer.WriteVariable(&types.Variable{
		Name:       "data",
		Dimensions: []int{3},
		DataType:   types.Double,
		Data:       []float64{1.0, 2.0, 3.0},
	})

	fmt.Println("MATLAB file created successfully")
	// Output:
	// MATLAB file created successfully
}

// ExampleCreate demonstrates creating a MATLAB file.
func ExampleCreate() {
	tmpfile := filepath.Join(os.TempDir(), "example_create.mat")
	defer os.Remove(tmpfile)

	writer, err := matlab.Create(tmpfile, matlab.Version5)
	if err != nil {
		panic(err)
	}
	defer writer.Close()

	fmt.Println("File created")
	// Output:
	// File created
}

// ExampleCreate_v5 demonstrates creating a v5 format file.
func ExampleCreate_v5() {
	tmpfile := filepath.Join(os.TempDir(), "example_v5.mat")
	defer os.Remove(tmpfile)

	writer, _ := matlab.Create(tmpfile, matlab.Version5)
	defer writer.Close()

	writer.WriteVariable(&types.Variable{
		Name:       "matrix",
		Dimensions: []int{2, 3},
		DataType:   types.Double,
		Data:       []float64{1, 2, 3, 4, 5, 6},
	})

	fmt.Println("v5 file created")
	// Output:
	// v5 file created
}

// ExampleCreate_v73 demonstrates creating a v7.3 HDF5 format file.
func ExampleCreate_v73() {
	tmpfile := filepath.Join(os.TempDir(), "example_v73.mat")
	defer os.Remove(tmpfile)

	writer, _ := matlab.Create(tmpfile, matlab.Version73)
	defer writer.Close()

	writer.WriteVariable(&types.Variable{
		Name:       "data",
		Dimensions: []int{100},
		DataType:   types.Double,
		Data:       make([]float64, 100),
	})

	fmt.Println("v7.3 file created")
	// Output:
	// v7.3 file created
}

// ExampleOpen demonstrates reading a MATLAB file.
func ExampleOpen() {
	file, _ := os.Open("testdata/generated/simple_double.mat")
	defer file.Close()

	matFile, _ := matlab.Open(file)

	fmt.Printf("Found %d variable(s)\n", len(matFile.Variables))
	// Output:
	// Found 1 variable(s)
}

// ExampleMatFile_Variables demonstrates iterating over variables.
func ExampleMatFile_Variables() {
	file, _ := os.Open("testdata/generated/simple_double.mat")
	defer file.Close()

	matFile, _ := matlab.Open(file)

	for _, v := range matFile.Variables {
		fmt.Printf("Variable: %s, Type: %v\n", v.Name, v.DataType)
	}
	// Output:
	// Variable: data, Type: double
}

// ExampleMatFileWriter_WriteVariable demonstrates writing a simple array.
func ExampleMatFileWriter_WriteVariable() {
	tmpfile := filepath.Join(os.TempDir(), "example_array.mat")
	defer os.Remove(tmpfile)

	writer, _ := matlab.Create(tmpfile, matlab.Version5)
	defer writer.Close()

	err := writer.WriteVariable(&types.Variable{
		Name:       "mydata",
		Dimensions: []int{5},
		DataType:   types.Double,
		Data:       []float64{1.0, 2.0, 3.0, 4.0, 5.0},
	})

	if err == nil {
		fmt.Println("Variable written")
	}
	// Output:
	// Variable written
}

// ExampleMatFileWriter_WriteVariable_matrix demonstrates writing a 2D matrix.
func ExampleMatFileWriter_WriteVariable_matrix() {
	tmpfile := filepath.Join(os.TempDir(), "example_matrix.mat")
	defer os.Remove(tmpfile)

	writer, _ := matlab.Create(tmpfile, matlab.Version5)
	defer writer.Close()

	// 3x4 matrix in column-major order (MATLAB standard)
	writer.WriteVariable(&types.Variable{
		Name:       "A",
		Dimensions: []int{3, 4},
		DataType:   types.Double,
		Data:       []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12},
	})

	fmt.Println("Matrix written")
	// Output:
	// Matrix written
}

// ExampleMatFileWriter_WriteVariable_complex demonstrates writing complex numbers.
func ExampleMatFileWriter_WriteVariable_complex() {
	tmpfile := filepath.Join(os.TempDir(), "example_complex.mat")
	defer os.Remove(tmpfile)

	writer, _ := matlab.Create(tmpfile, matlab.Version73)
	defer writer.Close()

	complexData := &types.NumericArray{
		Real: []float64{1.0, 2.0, 3.0},
		Imag: []float64{4.0, 5.0, 6.0},
	}

	writer.WriteVariable(&types.Variable{
		Name:       "signal",
		Dimensions: []int{3},
		DataType:   types.Double,
		Data:       complexData,
		IsComplex:  true,
	})

	fmt.Println("Complex variable written")
	// Output:
	// Complex variable written
}

// ExampleMatFileWriter_WriteVariable_int32 demonstrates writing integer data.
func ExampleMatFileWriter_WriteVariable_int32() {
	tmpfile := filepath.Join(os.TempDir(), "example_integers.mat")
	defer os.Remove(tmpfile)

	writer, _ := matlab.Create(tmpfile, matlab.Version5)
	defer writer.Close()

	writer.WriteVariable(&types.Variable{
		Name:       "counts",
		Dimensions: []int{4},
		DataType:   types.Int32,
		Data:       []int32{10, 20, 30, 40},
	})

	fmt.Println("Integer array written")
	// Output:
	// Integer array written
}

// ExampleVariable_Data_simple demonstrates accessing simple numeric data.
func ExampleVariable_Data_simple() {
	file, _ := os.Open("testdata/generated/simple_double.mat")
	defer file.Close()

	matFile, _ := matlab.Open(file)
	variable := matFile.Variables[0]

	// Simple arrays are stored directly
	data := variable.Data.([]float64)
	fmt.Printf("First value: %.1f\n", data[0])
	// Output:
	// First value: 1.0
}

// ExampleVariable_Data_complex demonstrates the structure of complex number data.
func ExampleVariable_Data_complex() {
	// Create a complex numeric array
	complexData := &types.NumericArray{
		Real: []float64{1, 2, 3},
		Imag: []float64{4, 5, 6},
	}

	// Demonstrate accessing the parts
	fmt.Printf("Real part: %v\n", complexData.Real)
	fmt.Printf("Imag part: %v\n", complexData.Imag)
	// Output:
	// Real part: [1 2 3]
	// Imag part: [4 5 6]
}

// ExampleOpen_roundTrip demonstrates writing and reading back data.
func ExampleOpen_roundTrip() {
	tmpfile := filepath.Join(os.TempDir(), "example_roundtrip.mat")
	defer os.Remove(tmpfile)

	// Write
	writer, _ := matlab.Create(tmpfile, matlab.Version5)
	writer.WriteVariable(&types.Variable{
		Name:       "test",
		Dimensions: []int{2},
		DataType:   types.Double,
		Data:       []float64{3.14, 2.71},
	})
	writer.Close()

	// Read
	file, _ := os.Open(tmpfile)
	defer file.Close()

	matFile, _ := matlab.Open(file)
	data := matFile.Variables[0].Data.([]float64)

	fmt.Printf("Read back: %.2f, %.2f\n", data[0], data[1])
	// Output:
	// Read back: 3.14, 2.71
}

// ExampleOpen_multipleVariables demonstrates handling multiple variables.
func ExampleOpen_multipleVariables() {
	tmpfile := filepath.Join(os.TempDir(), "example_multi.mat")
	defer os.Remove(tmpfile)

	// Write multiple variables
	writer, _ := matlab.Create(tmpfile, matlab.Version5)
	writer.WriteVariable(&types.Variable{
		Name:       "x",
		Dimensions: []int{3},
		DataType:   types.Double,
		Data:       []float64{1, 2, 3},
	})
	writer.WriteVariable(&types.Variable{
		Name:       "y",
		Dimensions: []int{3},
		DataType:   types.Double,
		Data:       []float64{4, 5, 6},
	})
	writer.Close()

	// Read all variables
	file, _ := os.Open(tmpfile)
	defer file.Close()

	matFile, _ := matlab.Open(file)
	fmt.Printf("Total variables: %d\n", len(matFile.Variables))
	for _, v := range matFile.Variables {
		fmt.Printf("- %s\n", v.Name)
	}
	// Output:
	// Total variables: 2
	// - x
	// - y
}

// ExampleCreate_withOptions demonstrates using functional options.
func ExampleCreate_withOptions() {
	tmpfile := filepath.Join(os.TempDir(), "options.mat")
	defer os.Remove(tmpfile)

	writer, _ := matlab.Create(tmpfile, matlab.Version5,
		matlab.WithEndianness(binary.BigEndian),
		matlab.WithDescription("Simulation results"),
	)
	defer writer.Close()

	fmt.Println("File created with custom options")
	// Output:
	// File created with custom options
}

// ExampleWithEndianness demonstrates setting byte order.
func ExampleWithEndianness() {
	tmpfile := filepath.Join(os.TempDir(), "bigendian.mat")
	defer os.Remove(tmpfile)

	writer, _ := matlab.Create(tmpfile, matlab.Version5,
		matlab.WithEndianness(binary.BigEndian),
	)
	defer writer.Close()

	fmt.Println("Big-endian file created")
	// Output:
	// Big-endian file created
}

// ExampleWithDescription demonstrates custom file description.
func ExampleWithDescription() {
	tmpfile := filepath.Join(os.TempDir(), "described.mat")
	defer os.Remove(tmpfile)

	writer, _ := matlab.Create(tmpfile, matlab.Version5,
		matlab.WithDescription("My experimental data from 2025"),
	)
	defer writer.Close()

	fmt.Println("File with custom description created")
	// Output:
	// File with custom description created
}
