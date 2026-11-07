// Command 04-write-v73 writes a v7.3 MAT-file with a real and an int32
// variable, then reads it back.
//
// The file starts with the 512-byte MAT header MATLAB writes in front of its
// own v7.3 files, followed by HDF5 data. Loading it in MATLAB itself has not
// been verified; h5dump opens it.
//
// Usage, from the repository root:
//
//	go run ./examples/04-write-v73
package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/scigolib/matlab"
	"github.com/scigolib/matlab/types"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	dir, err := os.MkdirTemp("", "matlab-example-")
	if err != nil {
		return fmt.Errorf("create temp dir: %w", err)
	}
	defer os.RemoveAll(dir) // best effort: temp dir cleanup
	path := filepath.Join(dir, "output_v73.mat")

	writer, err := matlab.Create(path, matlab.Version73)
	if err != nil {
		return fmt.Errorf("create file: %w", err)
	}

	vars := []*types.Variable{
		{Name: "x", Dimensions: []int{4}, DataType: types.Double, Data: []float64{0.5, 1.5, 2.5, 3.5}},
		{Name: "counts", Dimensions: []int{3}, DataType: types.Int32, Data: []int32{10, 20, 30}},
	}
	for _, v := range vars {
		if err := writer.WriteVariable(v); err != nil {
			_ = writer.Close() // best effort: the write error is the one reported
			return fmt.Errorf("write %s: %w", v.Name, err)
		}
	}
	if err := writer.Close(); err != nil {
		return fmt.Errorf("close writer: %w", err)
	}

	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("reopen file: %w", err)
	}
	defer file.Close() // read-only file: a close error loses nothing

	mat, err := matlab.Open(file)
	if err != nil {
		return fmt.Errorf("read back: %w", err)
	}
	for _, v := range mat.Variables {
		// v7.3 numeric data comes back as []float64 whatever the MATLAB class.
		fmt.Printf("%s: %s %v %T %v\n", v.Name, v.DataType, v.Dimensions, v.Data, v.Data)
	}
	return nil
}
