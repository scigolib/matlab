// Package matlab_test provides testable examples for the MATLAB file library.
//
// These examples demonstrate common use cases and serve as both documentation
// and verification that the API works as expected. Every error is checked:
// examples are the code users copy. They panic on an error instead of calling
// log.Fatal, so the deferred cleanup (temporary directories, open files) still
// runs and the failing example reports the error.
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
	dir, err := os.MkdirTemp("", "matlab-example-")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(dir)

	// Create a simple v5 MATLAB file
	writer, err := matlab.Create(filepath.Join(dir, "example.mat"), matlab.Version5)
	if err != nil {
		panic(err)
	}

	// Write a variable
	err = writer.WriteVariable(&types.Variable{
		Name:       "data",
		Dimensions: []int{3},
		DataType:   types.Double,
		Data:       []float64{1.0, 2.0, 3.0},
	})
	if err != nil {
		panic(err)
	}

	// Close writes the remaining data: its error must be checked too.
	if err := writer.Close(); err != nil {
		panic(err)
	}

	fmt.Println("MATLAB file created successfully")
	// Output:
	// MATLAB file created successfully
}

// ExampleCreate demonstrates creating a MATLAB file.
func ExampleCreate() {
	dir, err := os.MkdirTemp("", "matlab-example-")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(dir)

	writer, err := matlab.Create(filepath.Join(dir, "example_create.mat"), matlab.Version5)
	if err != nil {
		panic(err)
	}
	if err := writer.Close(); err != nil {
		panic(err)
	}

	fmt.Println("File created")
	// Output:
	// File created
}

// ExampleCreate_v5 demonstrates creating a v5 format file.
func ExampleCreate_v5() {
	dir, err := os.MkdirTemp("", "matlab-example-")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(dir)

	writer, err := matlab.Create(filepath.Join(dir, "example_v5.mat"), matlab.Version5)
	if err != nil {
		panic(err)
	}

	err = writer.WriteVariable(&types.Variable{
		Name:       "matrix",
		Dimensions: []int{2, 3},
		DataType:   types.Double,
		Data:       []float64{1, 2, 3, 4, 5, 6},
	})
	if err != nil {
		panic(err)
	}
	if err := writer.Close(); err != nil {
		panic(err)
	}

	fmt.Println("v5 file created")
	// Output:
	// v5 file created
}

// ExampleCreate_v73 demonstrates creating a v7.3 HDF5 format file.
func ExampleCreate_v73() {
	dir, err := os.MkdirTemp("", "matlab-example-")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(dir)

	writer, err := matlab.Create(filepath.Join(dir, "example_v73.mat"), matlab.Version73)
	if err != nil {
		panic(err)
	}

	err = writer.WriteVariable(&types.Variable{
		Name:       "data",
		Dimensions: []int{100},
		DataType:   types.Double,
		Data:       make([]float64, 100),
	})
	if err != nil {
		panic(err)
	}
	if err := writer.Close(); err != nil {
		panic(err)
	}

	fmt.Println("v7.3 file created")
	// Output:
	// v7.3 file created
}

// ExampleOpen demonstrates reading a MATLAB file.
func ExampleOpen() {
	file, err := os.Open("testdata/generated/simple_double.mat")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	matFile, err := matlab.Open(file)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Found %d variable(s)\n", len(matFile.Variables))
	// Output:
	// Found 1 variable(s)
}

// ExampleMatFile_Variables demonstrates iterating over variables.
func ExampleMatFile_Variables() {
	file, err := os.Open("testdata/generated/simple_double.mat")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	matFile, err := matlab.Open(file)
	if err != nil {
		panic(err)
	}

	for _, v := range matFile.Variables {
		fmt.Printf("Variable: %s, Type: %v\n", v.Name, v.DataType)
	}
	// Output:
	// Variable: data, Type: double
}

// ExampleMatFileWriter_WriteVariable demonstrates writing a simple array.
func ExampleMatFileWriter_WriteVariable() {
	dir, err := os.MkdirTemp("", "matlab-example-")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(dir)

	writer, err := matlab.Create(filepath.Join(dir, "example_array.mat"), matlab.Version5)
	if err != nil {
		panic(err)
	}

	err = writer.WriteVariable(&types.Variable{
		Name:       "mydata",
		Dimensions: []int{5},
		DataType:   types.Double,
		Data:       []float64{1.0, 2.0, 3.0, 4.0, 5.0},
	})
	if err != nil {
		panic(err)
	}
	if err := writer.Close(); err != nil {
		panic(err)
	}

	fmt.Println("Variable written")
	// Output:
	// Variable written
}

// ExampleMatFileWriter_WriteVariable_matrix demonstrates writing a 2D matrix.
func ExampleMatFileWriter_WriteVariable_matrix() {
	dir, err := os.MkdirTemp("", "matlab-example-")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(dir)

	writer, err := matlab.Create(filepath.Join(dir, "example_matrix.mat"), matlab.Version5)
	if err != nil {
		panic(err)
	}

	// 3x4 matrix in column-major order (MATLAB standard)
	err = writer.WriteVariable(&types.Variable{
		Name:       "A",
		Dimensions: []int{3, 4},
		DataType:   types.Double,
		Data:       []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12},
	})
	if err != nil {
		panic(err)
	}
	if err := writer.Close(); err != nil {
		panic(err)
	}

	fmt.Println("Matrix written")
	// Output:
	// Matrix written
}

// ExampleMatFileWriter_WriteVariable_complex demonstrates writing complex numbers.
func ExampleMatFileWriter_WriteVariable_complex() {
	dir, err := os.MkdirTemp("", "matlab-example-")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(dir)

	writer, err := matlab.Create(filepath.Join(dir, "example_complex.mat"), matlab.Version73)
	if err != nil {
		panic(err)
	}

	complexData := &types.NumericArray{
		Real: []float64{1.0, 2.0, 3.0},
		Imag: []float64{4.0, 5.0, 6.0},
	}

	err = writer.WriteVariable(&types.Variable{
		Name:       "signal",
		Dimensions: []int{3},
		DataType:   types.Double,
		Data:       complexData,
		IsComplex:  true,
	})
	if err != nil {
		panic(err)
	}
	if err := writer.Close(); err != nil {
		panic(err)
	}

	fmt.Println("Complex variable written")
	// Output:
	// Complex variable written
}

// ExampleMatFileWriter_WriteVariable_int32 demonstrates writing integer data.
func ExampleMatFileWriter_WriteVariable_int32() {
	dir, err := os.MkdirTemp("", "matlab-example-")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(dir)

	writer, err := matlab.Create(filepath.Join(dir, "example_integers.mat"), matlab.Version5)
	if err != nil {
		panic(err)
	}

	err = writer.WriteVariable(&types.Variable{
		Name:       "counts",
		Dimensions: []int{4},
		DataType:   types.Int32,
		Data:       []int32{10, 20, 30, 40},
	})
	if err != nil {
		panic(err)
	}
	if err := writer.Close(); err != nil {
		panic(err)
	}

	fmt.Println("Integer array written")
	// Output:
	// Integer array written
}

// ExampleVariable_Data_simple demonstrates accessing simple numeric data.
func ExampleVariable_Data_simple() {
	file, err := os.Open("testdata/generated/simple_double.mat")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	matFile, err := matlab.Open(file)
	if err != nil {
		panic(err)
	}

	variable := matFile.GetVariable("data")
	if variable == nil {
		panic("variable \"data\" not found")
	}

	// Simple arrays are stored directly
	data, ok := variable.Data.([]float64)
	if !ok {
		panic(fmt.Sprintf("data is %T, want []float64", variable.Data))
	}
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
	dir, err := os.MkdirTemp("", "matlab-example-")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(dir)
	path := filepath.Join(dir, "example_roundtrip.mat")

	// Write
	writer, err := matlab.Create(path, matlab.Version5)
	if err != nil {
		panic(err)
	}
	err = writer.WriteVariable(&types.Variable{
		Name:       "test",
		Dimensions: []int{2},
		DataType:   types.Double,
		Data:       []float64{3.14, 2.71},
	})
	if err != nil {
		panic(err)
	}
	if err := writer.Close(); err != nil {
		panic(err)
	}

	// Read
	file, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	matFile, err := matlab.Open(file)
	if err != nil {
		panic(err)
	}
	variable := matFile.GetVariable("test")
	if variable == nil {
		panic("variable \"test\" not found")
	}
	data, ok := variable.Data.([]float64)
	if !ok {
		panic(fmt.Sprintf("test is %T, want []float64", variable.Data))
	}

	fmt.Printf("Read back: %.2f, %.2f\n", data[0], data[1])
	// Output:
	// Read back: 3.14, 2.71
}

// ExampleOpen_multipleVariables demonstrates handling multiple variables.
func ExampleOpen_multipleVariables() {
	dir, err := os.MkdirTemp("", "matlab-example-")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(dir)
	path := filepath.Join(dir, "example_multi.mat")

	// Write multiple variables
	writer, err := matlab.Create(path, matlab.Version5)
	if err != nil {
		panic(err)
	}
	for _, v := range []*types.Variable{
		{Name: "x", Dimensions: []int{3}, DataType: types.Double, Data: []float64{1, 2, 3}},
		{Name: "y", Dimensions: []int{3}, DataType: types.Double, Data: []float64{4, 5, 6}},
	} {
		if err := writer.WriteVariable(v); err != nil {
			panic(err)
		}
	}
	if err := writer.Close(); err != nil {
		panic(err)
	}

	// Read all variables
	file, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	matFile, err := matlab.Open(file)
	if err != nil {
		panic(err)
	}
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
	dir, err := os.MkdirTemp("", "matlab-example-")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(dir)

	writer, err := matlab.Create(filepath.Join(dir, "options.mat"), matlab.Version5,
		matlab.WithEndianness(binary.BigEndian),
		matlab.WithDescription("Simulation results"),
	)
	if err != nil {
		panic(err)
	}
	if err := writer.Close(); err != nil {
		panic(err)
	}

	fmt.Println("File created with custom options")
	// Output:
	// File created with custom options
}

// ExampleWithEndianness demonstrates setting byte order.
func ExampleWithEndianness() {
	dir, err := os.MkdirTemp("", "matlab-example-")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(dir)

	writer, err := matlab.Create(filepath.Join(dir, "bigendian.mat"), matlab.Version5,
		matlab.WithEndianness(binary.BigEndian),
	)
	if err != nil {
		panic(err)
	}
	if err := writer.Close(); err != nil {
		panic(err)
	}

	fmt.Println("Big-endian file created")
	// Output:
	// Big-endian file created
}

// ExampleWithDescription demonstrates custom file description.
func ExampleWithDescription() {
	dir, err := os.MkdirTemp("", "matlab-example-")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(dir)

	writer, err := matlab.Create(filepath.Join(dir, "described.mat"), matlab.Version5,
		matlab.WithDescription("My experimental data from 2025"),
	)
	if err != nil {
		panic(err)
	}
	if err := writer.Close(); err != nil {
		panic(err)
	}

	fmt.Println("File with custom description created")
	// Output:
	// File with custom description created
}
