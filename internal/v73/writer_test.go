package v73

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/scigolib/matlab/types"
)

func TestNewWriter(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test_new_writer.mat")

	writer, err := NewWriter(tmpFile)
	if err != nil {
		t.Fatalf("NewWriter() error = %v", err)
	}
	defer writer.Close()

	// Verify file was created
	if _, err := os.Stat(tmpFile); os.IsNotExist(err) {
		t.Error("File was not created")
	}
}

func TestNewWriter_InvalidPath(t *testing.T) {
	// Try to create file in non-existent directory
	invalidPath := filepath.Join("nonexistent", "dir", "test.mat")

	_, err := NewWriter(invalidPath)
	if err == nil {
		t.Error("NewWriter() expected error for invalid path, got nil")
	}
}

func TestValidateVariable(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.mat")

	writer, err := NewWriter(tmpFile)
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}
	defer writer.Close()

	tests := []struct {
		name    string
		input   *types.Variable
		wantErr bool
	}{
		{
			name: "valid variable",
			input: &types.Variable{
				Name:       "data",
				Dimensions: []int{3},
				DataType:   types.Double,
				Data:       []float64{1.0, 2.0, 3.0},
			},
			wantErr: false,
		},
		{
			name: "empty name",
			input: &types.Variable{
				Name:       "",
				Dimensions: []int{3},
				Data:       []float64{1.0, 2.0, 3.0},
			},
			wantErr: true,
		},
		{
			name: "empty dimensions",
			input: &types.Variable{
				Name:       "data",
				Dimensions: []int{},
				Data:       []float64{1.0, 2.0, 3.0},
			},
			wantErr: true,
		},
		{
			name: "nil data",
			input: &types.Variable{
				Name:       "data",
				Dimensions: []int{3},
				Data:       nil,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := writer.validateVariable(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateVariable() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDataTypeToMatlabClass(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.mat")

	writer, err := NewWriter(tmpFile)
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}
	defer writer.Close()

	tests := []struct {
		dataType types.DataType
		want     string
	}{
		{types.Double, "double"},
		{types.Single, "single"},
		{types.Int8, "int8"},
		{types.Uint8, "uint8"},
		{types.Int16, "int16"},
		{types.Uint16, "uint16"},
		{types.Int32, "int32"},
		{types.Uint32, "uint32"},
		{types.Int64, "int64"},
		{types.Uint64, "uint64"},
		{types.Char, "char"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := writer.dataTypeToMatlabClass(tt.dataType)
			if got != tt.want {
				t.Errorf("dataTypeToMatlabClass() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestDataTypeToHDF5(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.mat")

	writer, err := NewWriter(tmpFile)
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}
	defer writer.Close()

	tests := []struct {
		name     string
		dataType types.DataType
		wantErr  bool
	}{
		{"double", types.Double, false},
		{"single", types.Single, false},
		{"int8", types.Int8, false},
		{"uint8", types.Uint8, false},
		{"int16", types.Int16, false},
		{"uint16", types.Uint16, false},
		{"int32", types.Int32, false},
		{"uint32", types.Uint32, false},
		{"int64", types.Int64, false},
		{"uint64", types.Uint64, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := writer.dataTypeToHDF5(tt.dataType)
			if (err != nil) != tt.wantErr {
				t.Errorf("dataTypeToHDF5() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestWriteVariable_InvalidDimensions(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.mat")

	writer, err := NewWriter(tmpFile)
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}
	defer writer.Close()

	// Test negative dimension
	v := &types.Variable{
		Name:       "data",
		Dimensions: []int{-1},
		DataType:   types.Double,
		Data:       []float64{1.0},
	}

	err = writer.WriteVariable(v)
	if err == nil {
		t.Error("WriteVariable() expected error for negative dimension, got nil")
	}

	// Test zero dimension
	v.Dimensions = []int{0}
	err = writer.WriteVariable(v)
	if err == nil {
		t.Error("WriteVariable() expected error for zero dimension, got nil")
	}
}

func TestWriteVariable_ComplexSupported(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.mat")

	writer, err := NewWriter(tmpFile)
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}
	defer writer.Close()

	v := &types.Variable{
		Name:       "data",
		Dimensions: []int{2},
		DataType:   types.Double,
		IsComplex:  true,
		Data: &types.NumericArray{
			Real: []float64{1.0, 2.0},
			Imag: []float64{3.0, 4.0},
		},
	}

	err = writer.WriteVariable(v)
	if err != nil {
		t.Errorf("WriteVariable() complex numbers should be supported (with workaround), got error: %v", err)
	}
}

func TestWriteVariable_NilVariable(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.mat")

	writer, err := NewWriter(tmpFile)
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}
	defer writer.Close()

	err = writer.WriteVariable(nil)
	if err == nil {
		t.Error("WriteVariable() expected error for nil variable, got nil")
	}
}

func TestClose_MultipleCallsSafe(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.mat")

	writer, err := NewWriter(tmpFile)
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	// Close once
	if err := writer.Close(); err != nil {
		t.Errorf("First Close() error = %v", err)
	}

	// Close again - should be safe
	if err := writer.Close(); err != nil {
		t.Errorf("Second Close() error = %v", err)
	}
}

func TestWriteVariable_SimpleDouble(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test_simple_double.mat")

	writer, err := NewWriter(tmpFile)
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}
	defer writer.Close()

	v := &types.Variable{
		Name:       "A",
		Dimensions: []int{3},
		DataType:   types.Double,
		Data:       []float64{1.0, 2.0, 3.0},
	}

	err = writer.WriteVariable(v)
	if err != nil {
		t.Errorf("WriteVariable() error = %v", err)
	}
}

func TestWriteVariable_MultipleTypes(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test_multiple_types.mat")

	writer, err := NewWriter(tmpFile)
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}
	defer writer.Close()

	tests := []struct {
		name     string
		variable *types.Variable
	}{
		{
			"double",
			&types.Variable{
				Name:       "double_data",
				Dimensions: []int{3},
				DataType:   types.Double,
				Data:       []float64{1.1, 2.2, 3.3},
			},
		},
		{
			"single",
			&types.Variable{
				Name:       "single_data",
				Dimensions: []int{2},
				DataType:   types.Single,
				Data:       []float32{1.5, 2.5},
			},
		},
		{
			"int32",
			&types.Variable{
				Name:       "int32_data",
				Dimensions: []int{4},
				DataType:   types.Int32,
				Data:       []int32{10, 20, 30, 40},
			},
		},
		{
			"uint8",
			&types.Variable{
				Name:       "uint8_data",
				Dimensions: []int{5},
				DataType:   types.Uint8,
				Data:       []uint8{100, 101, 102, 103, 104},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := writer.WriteVariable(tt.variable)
			if err != nil {
				t.Errorf("WriteVariable() error = %v", err)
			}
		})
	}
}

func TestWriteVariable_2DMatrix(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test_2d_matrix.mat")

	writer, err := NewWriter(tmpFile)
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}
	defer writer.Close()

	v := &types.Variable{
		Name:       "M",
		Dimensions: []int{2, 3}, // 2x3 matrix
		DataType:   types.Double,
		Data:       []float64{1, 2, 3, 4, 5, 6},
	}

	err = writer.WriteVariable(v)
	if err != nil {
		t.Errorf("WriteVariable() error = %v", err)
	}
}

func TestWriteVariable_3DArray(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test_3d_array.mat")

	writer, err := NewWriter(tmpFile)
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}
	defer writer.Close()

	v := &types.Variable{
		Name:       "A3D",
		Dimensions: []int{2, 3, 4}, // 2x3x4 array
		DataType:   types.Int32,
		Data:       make([]int32, 24), // 2*3*4 = 24 elements
	}

	// Fill with sequential values
	for i := range v.Data.([]int32) {
		v.Data.([]int32)[i] = int32(i + 1)
	}

	err = writer.WriteVariable(v)
	if err != nil {
		t.Errorf("WriteVariable() error = %v", err)
	}
}
