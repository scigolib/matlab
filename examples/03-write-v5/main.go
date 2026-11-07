// Command 03-write-v5 writes a v5 MAT-file with a vector and a matrix, then
// reads it back.
//
// Usage, from the repository root:
//
//	go run ./examples/03-write-v5
package main

import (
	"encoding/binary"
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
	path := filepath.Join(dir, "output_v5.mat")

	writer, err := matlab.Create(path, matlab.Version5,
		matlab.WithEndianness(binary.BigEndian),
		matlab.WithDescription("Example data"))
	if err != nil {
		return fmt.Errorf("create file: %w", err)
	}

	vars := []*types.Variable{
		{Name: "A", Dimensions: []int{3}, DataType: types.Double, Data: []float64{1, 2, 3}},
		// Data is column-major: [1 3 5; 2 4 6] for a 2x3 matrix.
		{Name: "B", Dimensions: []int{2, 3}, DataType: types.Double, Data: []float64{1, 2, 3, 4, 5, 6}},
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
	fmt.Printf("version %s, byte order %s\n", mat.Version, mat.Endian)
	for _, v := range mat.Variables {
		fmt.Printf("%s: %s %v %v\n", v.Name, v.DataType, v.Dimensions, v.Data)
	}
	return nil
}
