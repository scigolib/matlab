package matlab

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/scigolib/matlab/internal/v5"
	"github.com/scigolib/matlab/types"
)

// TestRoundTrip_V5_SimpleDouble tests writing and reading back a simple double array.
func TestRoundTrip_V5_SimpleDouble(t *testing.T) {
	// Write
	var buf bytes.Buffer
	writer, err := v5.NewWriter(&buf, "Test roundtrip", "MI")
	if err != nil {
		t.Fatalf("NewWriter() error = %v", err)
	}

	original := &types.Variable{
		Name:       "A",
		Dimensions: []int{3},
		DataType:   types.Double,
		Data:       []float64{1.0, 2.0, 3.0},
	}

	err = writer.WriteVariable(original)
	if err != nil {
		t.Fatalf("WriteVariable() error = %v", err)
	}

	// Read back
	reader := bytes.NewReader(buf.Bytes())
	parser, err := v5.NewParser(reader)
	if err != nil {
		t.Fatalf("NewParser() error = %v", err)
	}

	file, err := parser.Parse()
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	// Verify
	if len(file.Variables) != 1 {
		t.Fatalf("Variables count = %d, want 1", len(file.Variables))
	}

	v := file.Variables[0]
	if v.Name != "A" {
		t.Errorf("Name = %q, want %q", v.Name, "A")
	}

	data, ok := v.Data.([]float64)
	if !ok {
		t.Fatalf("Data is not []float64, got %T", v.Data)
	}

	expected := []float64{1.0, 2.0, 3.0}
	if len(data) != len(expected) {
		t.Fatalf("Data length = %d, want %d", len(data), len(expected))
	}

	for i, val := range expected {
		if data[i] != val {
			t.Errorf("Data[%d] = %v, want %v", i, data[i], val)
		}
	}
}

// TestRoundTrip_V5_Int32 tests int32 arrays.
func TestRoundTrip_V5_Int32(t *testing.T) {
	var buf bytes.Buffer
	writer, err := v5.NewWriter(&buf, "Test", "MI")
	if err != nil {
		t.Fatalf("NewWriter() error = %v", err)
	}

	original := &types.Variable{
		Name:       "B",
		Dimensions: []int{4},
		DataType:   types.Int32,
		Data:       []int32{-100, 0, 100, 200},
	}

	err = writer.WriteVariable(original)
	if err != nil {
		t.Fatalf("WriteVariable() error = %v", err)
	}

	// Read back
	reader := bytes.NewReader(buf.Bytes())
	parser, err := v5.NewParser(reader)
	if err != nil {
		t.Fatalf("NewParser() error = %v", err)
	}

	file, err := parser.Parse()
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	v := file.Variables[0]
	data, ok := v.Data.([]int32)
	if !ok {
		t.Fatalf("Data is not []int32, got %T", v.Data)
	}

	expected := []int32{-100, 0, 100, 200}
	for i, val := range expected {
		if data[i] != val {
			t.Errorf("Data[%d] = %v, want %v", i, data[i], val)
		}
	}
}

// TestRoundTrip_V5_Complex tests complex numbers.
func TestRoundTrip_V5_Complex(t *testing.T) {
	var buf bytes.Buffer
	writer, err := v5.NewWriter(&buf, "Test", "MI")
	if err != nil {
		t.Fatalf("NewWriter() error = %v", err)
	}

	original := &types.Variable{
		Name:       "C",
		Dimensions: []int{2},
		DataType:   types.Double,
		IsComplex:  true,
		Data: &types.NumericArray{
			Real: []float64{1.0, 3.0},
			Imag: []float64{2.0, 4.0},
		},
	}

	err = writer.WriteVariable(original)
	if err != nil {
		t.Fatalf("WriteVariable() error = %v", err)
	}

	// Read back
	reader := bytes.NewReader(buf.Bytes())
	parser, err := v5.NewParser(reader)
	if err != nil {
		t.Fatalf("NewParser() error = %v", err)
	}

	file, err := parser.Parse()
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	v := file.Variables[0]
	if !v.IsComplex {
		t.Errorf("IsComplex = false, want true")
	}

	numArray, ok := v.Data.(*types.NumericArray)
	if !ok {
		t.Fatalf("Data is not *types.NumericArray, got %T", v.Data)
	}

	realData, ok := numArray.Real.([]float64)
	if !ok {
		t.Fatalf("Real is not []float64, got %T", numArray.Real)
	}

	imagData, ok := numArray.Imag.([]float64)
	if !ok {
		t.Fatalf("Imag is not []float64, got %T", numArray.Imag)
	}

	expectedReal := []float64{1.0, 3.0}
	expectedImag := []float64{2.0, 4.0}

	for i := range expectedReal {
		if realData[i] != expectedReal[i] {
			t.Errorf("Real[%d] = %v, want %v", i, realData[i], expectedReal[i])
		}
		if imagData[i] != expectedImag[i] {
			t.Errorf("Imag[%d] = %v, want %v", i, imagData[i], expectedImag[i])
		}
	}
}

// TestRoundTrip_V5_Matrix2x3 tests 2D matrix.
func TestRoundTrip_V5_Matrix2x3(t *testing.T) {
	var buf bytes.Buffer
	writer, err := v5.NewWriter(&buf, "Test", "MI")
	if err != nil {
		t.Fatalf("NewWriter() error = %v", err)
	}

	original := &types.Variable{
		Name:       "M",
		Dimensions: []int{2, 3},
		DataType:   types.Double,
		Data:       []float64{1, 2, 3, 4, 5, 6},
	}

	err = writer.WriteVariable(original)
	if err != nil {
		t.Fatalf("WriteVariable() error = %v", err)
	}

	// Read back
	reader := bytes.NewReader(buf.Bytes())
	parser, err := v5.NewParser(reader)
	if err != nil {
		t.Fatalf("NewParser() error = %v", err)
	}

	file, err := parser.Parse()
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	v := file.Variables[0]
	if len(v.Dimensions) != 2 || v.Dimensions[0] != 2 || v.Dimensions[1] != 3 {
		t.Errorf("Dimensions = %v, want [2 3]", v.Dimensions)
	}

	data, ok := v.Data.([]float64)
	if !ok {
		t.Fatalf("Data is not []float64, got %T", v.Data)
	}

	expected := []float64{1, 2, 3, 4, 5, 6}
	for i, val := range expected {
		if data[i] != val {
			t.Errorf("Data[%d] = %v, want %v", i, data[i], val)
		}
	}
}

// TestRoundTrip_V5_BigEndian tests big-endian format.
func TestRoundTrip_V5_BigEndian(t *testing.T) {
	var buf bytes.Buffer
	writer, err := v5.NewWriter(&buf, "Test", "IM") // Big-endian
	if err != nil {
		t.Fatalf("NewWriter() error = %v", err)
	}

	original := &types.Variable{
		Name:       "BE",
		Dimensions: []int{3},
		DataType:   types.Double,
		Data:       []float64{10.0, 20.0, 30.0},
	}

	err = writer.WriteVariable(original)
	if err != nil {
		t.Fatalf("WriteVariable() error = %v", err)
	}

	// Read back
	reader := bytes.NewReader(buf.Bytes())
	parser, err := v5.NewParser(reader)
	if err != nil {
		t.Fatalf("NewParser() error = %v", err)
	}

	file, err := parser.Parse()
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	// Verify endian indicator
	if file.Header.EndianIndicator != "IM" {
		t.Errorf("EndianIndicator = %q, want %q", file.Header.EndianIndicator, "IM")
	}

	v := file.Variables[0]
	data, ok := v.Data.([]float64)
	if !ok {
		t.Fatalf("Data is not []float64, got %T", v.Data)
	}

	expected := []float64{10.0, 20.0, 30.0}
	for i, val := range expected {
		if data[i] != val {
			t.Errorf("Data[%d] = %v, want %v", i, data[i], val)
		}
	}
}

// TestRoundTrip_V5_PublicAPI tests the public API with file I/O.
func TestRoundTrip_V5_PublicAPI(t *testing.T) {
	// Create temp file
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "test_v5.mat")

	// Write using public API
	writer, err := Create(filename, Version5)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	err = writer.WriteVariable(&types.Variable{
		Name:       "data",
		Dimensions: []int{5},
		DataType:   types.Double,
		Data:       []float64{1.1, 2.2, 3.3, 4.4, 5.5},
	})
	if err != nil {
		t.Fatalf("WriteVariable() error = %v", err)
	}

	err = writer.Close()
	if err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	// Read back using public API
	f, err := os.Open(filename)
	if err != nil {
		t.Fatalf("os.Open() error = %v", err)
	}
	defer f.Close()

	matfile, err := Open(f)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}

	if len(matfile.Variables) != 1 {
		t.Fatalf("Variables count = %d, want 1", len(matfile.Variables))
	}

	v := matfile.Variables[0]
	if v.Name != "data" {
		t.Errorf("Name = %q, want %q", v.Name, "data")
	}

	data, ok := v.Data.([]float64)
	if !ok {
		t.Fatalf("Data is not []float64, got %T", v.Data)
	}

	expected := []float64{1.1, 2.2, 3.3, 4.4, 5.5}
	for i, val := range expected {
		if data[i] != val {
			t.Errorf("Data[%d] = %v, want %v", i, data[i], val)
		}
	}

	// Verify file exists and has content
	info, err := os.Stat(filename)
	if err != nil {
		t.Fatalf("Stat() error = %v", err)
	}
	if info.Size() < 128 {
		t.Errorf("File size = %d, expected > 128 (header + data)", info.Size())
	}
}

// TestRoundTrip_V5_MultipleVariables tests writing multiple variables.
func TestRoundTrip_V5_MultipleVariables(t *testing.T) {
	var buf bytes.Buffer
	writer, err := v5.NewWriter(&buf, "Test", "MI")
	if err != nil {
		t.Fatalf("NewWriter() error = %v", err)
	}

	// Write multiple variables
	vars := []*types.Variable{
		{
			Name:       "var1",
			Dimensions: []int{2},
			DataType:   types.Double,
			Data:       []float64{1.0, 2.0},
		},
		{
			Name:       "var2",
			Dimensions: []int{3},
			DataType:   types.Int32,
			Data:       []int32{10, 20, 30},
		},
		{
			Name:       "var3",
			Dimensions: []int{1},
			DataType:   types.Uint8,
			Data:       []byte{255},
		},
	}

	for _, v := range vars {
		err = writer.WriteVariable(v)
		if err != nil {
			t.Fatalf("WriteVariable(%s) error = %v", v.Name, err)
		}
	}

	// Read back
	reader := bytes.NewReader(buf.Bytes())
	parser, err := v5.NewParser(reader)
	if err != nil {
		t.Fatalf("NewParser() error = %v", err)
	}

	file, err := parser.Parse()
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	// Verify count
	if len(file.Variables) != 3 {
		t.Fatalf("Variables count = %d, want 3", len(file.Variables))
	}

	// Verify var1
	v1 := file.Variables[0]
	if v1.Name != "var1" {
		t.Errorf("var1.Name = %q, want %q", v1.Name, "var1")
	}

	// Verify var2
	v2 := file.Variables[1]
	if v2.Name != "var2" {
		t.Errorf("var2.Name = %q, want %q", v2.Name, "var2")
	}

	// Verify var3
	v3 := file.Variables[2]
	if v3.Name != "var3" {
		t.Errorf("var3.Name = %q, want %q", v3.Name, "var3")
	}
}
