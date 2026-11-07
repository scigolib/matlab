// Command 01-read opens a MAT-file (v5 or v7.3) and lists its variables.
//
// Usage, from the repository root:
//
//	go run ./examples/01-read testdata/scipy/testdouble_7.4_GLNX86.mat
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/scigolib/matlab"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: 01-read <file.mat>")
		return
	}

	if err := run(os.Args[1]); err != nil {
		log.Fatal(err)
	}
}

func run(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("open file: %w", err)
	}
	defer file.Close() // read-only file: a close error loses nothing

	mat, err := matlab.Open(file)
	if err != nil {
		return fmt.Errorf("read MAT-file: %w", err)
	}

	fmt.Println("MAT-file version:", mat.Version)
	if mat.Endian != "" {
		fmt.Println("Endian:", mat.Endian)
	}
	if mat.Description != "" {
		fmt.Println("Description:", mat.Description)
	}
	fmt.Printf("Found %d variables:\n", len(mat.Variables))

	for i, v := range mat.Variables {
		fmt.Printf("\n%d. %s\n", i+1, v)
		fmt.Printf("  Dimensions: %v\n", v.Dimensions)
		fmt.Printf("  Type: %s\n", v.DataType)
		fmt.Printf("  Complex: %v, Sparse: %v\n", v.IsComplex, v.IsSparse)

		// Print scalar values
		if len(v.Dimensions) == 2 && v.Dimensions[0] == 1 && v.Dimensions[1] == 1 {
			fmt.Printf("  Value: %v\n", v.Data)
		}
	}
	return nil
}
