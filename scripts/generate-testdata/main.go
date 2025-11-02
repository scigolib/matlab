// Package main - Generate minimal test MAT-files
//
// This script creates minimal MATLAB test files for testdata/ directory.
// Uses our own writer to generate files (dogfooding approach).
//
// Usage: go run scripts/generate-testdata.go
package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/scigolib/matlab"
	"github.com/scigolib/matlab/types"
)

func main() {
	fmt.Println("📦 Generating MATLAB test files for testdata/")
	fmt.Println(strings.Repeat("=", 60))

	testdataDir := filepath.Join("testdata", "generated")
	if err := os.MkdirAll(testdataDir, 0755); err != nil {
		log.Fatalf("Failed to create testdata directory: %v", err)
	}

	// Test files to generate
	tests := []struct {
		filename string
		variable *types.Variable
		desc     string
	}{
		{
			filename: "simple_double.mat",
			variable: &types.Variable{
				Name:       "data",
				Dimensions: []int{5},
				DataType:   types.Double,
				Data:       []float64{1.0, 2.0, 3.0, 4.0, 5.0},
			},
			desc: "Simple 1D double array",
		},
		{
			filename: "simple_int32.mat",
			variable: &types.Variable{
				Name:       "values",
				Dimensions: []int{4},
				DataType:   types.Int32,
				Data:       []int32{10, 20, 30, 40},
			},
			desc: "Simple 1D int32 array",
		},
		{
			filename: "simple_uint8.mat",
			variable: &types.Variable{
				Name:       "bytes",
				Dimensions: []int{3},
				DataType:   types.Uint8,
				Data:       []uint8{255, 128, 0},
			},
			desc: "Simple 1D uint8 array",
		},
		{
			filename: "simple_single.mat",
			variable: &types.Variable{
				Name:       "floats",
				Dimensions: []int{3},
				DataType:   types.Single,
				Data:       []float32{1.5, 2.5, 3.5},
			},
			desc: "Simple 1D single (float32) array",
		},
		{
			filename: "complex.mat",
			variable: &types.Variable{
				Name:       "z",
				Dimensions: []int{3},
				DataType:   types.Double,
				IsComplex:  true,
				Data: &types.NumericArray{
					Real: []float64{1.0, 2.0, 3.0},
					Imag: []float64{4.0, 5.0, 6.0},
				},
			},
			desc: "Complex numbers (1+4i, 2+5i, 3+6i)",
		},
		{
			filename: "matrix_2x3.mat",
			variable: &types.Variable{
				Name:       "matrix",
				Dimensions: []int{2, 3},
				DataType:   types.Double,
				Data:       []float64{1.0, 2.0, 3.0, 4.0, 5.0, 6.0},
			},
			desc: "2x3 matrix (column-major order)",
		},
		{
			filename: "matrix_3x2.mat",
			variable: &types.Variable{
				Name:       "A",
				Dimensions: []int{3, 2},
				DataType:   types.Double,
				Data:       []float64{1.0, 4.0, 2.0, 5.0, 3.0, 6.0},
			},
			desc: "3x2 matrix (column-major order)",
		},
		{
			filename: "scalar.mat",
			variable: &types.Variable{
				Name:       "x",
				Dimensions: []int{1},
				DataType:   types.Double,
				Data:       []float64{42.0},
			},
			desc: "Scalar value (single element)",
		},
	}

	// Generate v7.3 files
	fmt.Println("\n📝 Generating v7.3 (HDF5) test files:")
	for _, test := range tests {
		filename := filepath.Join(testdataDir, test.filename)
		fmt.Printf("  - %s: %s... ", test.filename, test.desc)

		writer, err := matlab.Create(filename, matlab.Version73)
		if err != nil {
			fmt.Printf("❌ FAILED\n    Error: %v\n", err)
			continue
		}

		if err := writer.WriteVariable(test.variable); err != nil {
			fmt.Printf("❌ FAILED\n    Error: %v\n", err)
			_ = writer.Close() // Best effort cleanup on error
			continue
		}

		if err := writer.Close(); err != nil {
			fmt.Printf("❌ FAILED\n    Error: %v\n", err)
			continue
		}

		fmt.Println("✅")
	}

	// Create README
	readmePath := filepath.Join(testdataDir, "README.md")
	readme := `# MATLAB Test Data

This directory contains minimal MATLAB files for testing.

## Files

| File | Format | Description | Variable | Type | Dimensions |
|------|--------|-------------|----------|------|------------|
| simple_double.mat | v7.3 | Simple 1D double array | data | double | [5] |
| simple_int32.mat | v7.3 | Simple 1D int32 array | values | int32 | [4] |
| simple_uint8.mat | v7.3 | Simple 1D uint8 array | bytes | uint8 | [3] |
| simple_single.mat | v7.3 | Simple 1D single array | floats | single | [3] |
| complex.mat | v7.3 | Complex numbers | z | double | [3] |
| matrix_2x3.mat | v7.3 | 2x3 matrix | matrix | double | [2, 3] |
| matrix_3x2.mat | v7.3 | 3x2 matrix | A | double | [3, 2] |
| scalar.mat | v7.3 | Scalar value | x | double | [1] |

## Generation

These files were generated using our own writer implementation:

` + "```bash" + `
go run scripts/generate-testdata.go
` + "```" + `

## Testing

Use these files for:
- Reader integration tests
- Round-trip verification (write → read → compare)
- MATLAB compatibility testing
- Performance benchmarking

## Notes

- All files are v7.3 format (HDF5-based)
- Files use MATLAB_class attributes for type info
- Data is stored in column-major order (MATLAB convention)
- Complex numbers use separate real/imaginary datasets
`

	if err := os.WriteFile(readmePath, []byte(readme), 0644); err != nil {
		log.Printf("Warning: Failed to create README: %v", err)
	} else {
		fmt.Println("\n📄 Created testdata/README.md")
	}

	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("✅ Test data generation complete!")
	fmt.Printf("📁 Generated %d test files in testdata/\n", len(tests))
	fmt.Println("\nNext steps:")
	fmt.Println("  1. Run tests: go test ./...")
	fmt.Println("  2. Verify files: ls -lh testdata/")
	fmt.Println("  3. Test with MATLAB/Octave (if available)")
}
