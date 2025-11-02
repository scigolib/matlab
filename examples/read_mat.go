package main

import (
	"fmt"
	"log"
	"os"

	"github.com/scigolib/matlab"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: read_mat <file.mat>")
		return
	}

	file, err := os.Open(os.Args[1])
	if err != nil {
		log.Fatal("Error opening file:", err)
	}
	defer file.Close()

	mat, err := matlab.Open(file)
	if err != nil {
		log.Fatal("Error reading MAT-file:", err)
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

		// Print simple values
		if len(v.Dimensions) == 2 && v.Dimensions[0] == 1 && v.Dimensions[1] == 1 {
			fmt.Printf("  Value: %v\n", v.Data)
		}
	}
}
