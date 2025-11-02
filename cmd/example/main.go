// Package main provides an example of using the MATLAB file reader library.
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/scigolib/matlab"
)

func main() {
	// Open MAT-file
	file, err := os.Open("data.mat")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close() //nolint:errcheck // Example code, cleanup on exit

	// Parse MAT-file
	mat, err := matlab.Open(file)
	if err != nil {
		log.Fatal(err)
	}

	// Print file info
	fmt.Println("MAT-file version:", mat.Version)
	fmt.Println("Description:", mat.Description)

	// Print variables
	for i, v := range mat.Variables {
		fmt.Printf("%d. %s: %s %v\n", i+1, v.Name, v.DataType, v.Dimensions)
	}
}
