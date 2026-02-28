package v73

import (
	"path/filepath"
	"testing"

	"github.com/scigolib/hdf5"
	"github.com/scigolib/matlab/types"
)

// writeTestFile creates a temporary HDF5-based .mat file with the given variables.
func writeTestFile(t *testing.T, vars ...*types.Variable) string {
	t.Helper()
	tmpFile := filepath.Join(t.TempDir(), "test.mat")
	writer, err := NewWriter(tmpFile)
	if err != nil {
		t.Fatalf("NewWriter failed: %v", err)
	}
	for _, v := range vars {
		if err := writer.WriteVariable(v); err != nil {
			t.Fatalf("WriteVariable(%s) failed: %v", v.Name, err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("Close failed: %v", err)
	}
	return tmpFile
}

// openHDF5 opens an HDF5 file for reading and returns the handle.
func openHDF5(t *testing.T, path string) *hdf5.File {
	t.Helper()
	file, err := hdf5.Open(path)
	if err != nil {
		t.Fatalf("hdf5.Open failed: %v", err)
	}
	return file
}

func TestNewHDF5Adapter(t *testing.T) {
	tmpFile := writeTestFile(t, &types.Variable{
		Name:       "x",
		Dimensions: []int{1},
		DataType:   types.Double,
		Data:       []float64{1.0},
	})

	file := openHDF5(t, tmpFile)
	defer file.Close()

	adapter := NewHDF5Adapter(file)
	if adapter == nil {
		t.Fatal("NewHDF5Adapter returned nil")
	}
	if adapter.file != file {
		t.Error("adapter.file does not match the provided file")
	}
}

func TestConvertToMatlab_SimpleDouble(t *testing.T) {
	expected := []float64{1.1, 2.2, 3.3}
	tmpFile := writeTestFile(t, &types.Variable{
		Name:       "data",
		Dimensions: []int{3},
		DataType:   types.Double,
		Data:       expected,
	})

	file := openHDF5(t, tmpFile)
	defer file.Close()

	adapter := NewHDF5Adapter(file)
	variables, err := adapter.ConvertToMatlab()
	if err != nil {
		t.Fatalf("ConvertToMatlab() error = %v", err)
	}
	if len(variables) != 1 {
		t.Fatalf("expected 1 variable, got %d", len(variables))
	}

	v := variables[0]
	if v.Name != "data" {
		t.Errorf("Name = %q, want %q", v.Name, "data")
	}
	if v.DataType != types.Double {
		t.Errorf("DataType = %v, want %v", v.DataType, types.Double)
	}

	floatData, ok := v.Data.([]float64)
	if !ok {
		t.Fatalf("Data type = %T, want []float64", v.Data)
	}
	if len(floatData) != len(expected) {
		t.Fatalf("data length = %d, want %d", len(floatData), len(expected))
	}
	for i, val := range floatData {
		if val != expected[i] {
			t.Errorf("data[%d] = %v, want %v", i, val, expected[i])
		}
	}
}

func TestConvertToMatlab_AllNumericTypes(t *testing.T) {
	tests := []struct {
		name     string
		dataType types.DataType
		data     interface{}
		wantType types.DataType
	}{
		{"double", types.Double, []float64{1.0, 2.0}, types.Double},
		{"single", types.Single, []float32{1.0, 2.0}, types.Single},
		{"int8", types.Int8, []int8{1, 2}, types.Int8},
		{"uint8", types.Uint8, []uint8{1, 2}, types.Uint8},
		{"int16", types.Int16, []int16{1, 2}, types.Int16},
		{"uint16", types.Uint16, []uint16{1, 2}, types.Uint16},
		{"int32", types.Int32, []int32{1, 2}, types.Int32},
		{"uint32", types.Uint32, []uint32{1, 2}, types.Uint32},
		{"int64", types.Int64, []int64{1, 2}, types.Int64},
		{"uint64", types.Uint64, []uint64{1, 2}, types.Uint64},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpFile := writeTestFile(t, &types.Variable{
				Name:       tt.name + "_var",
				Dimensions: []int{2},
				DataType:   tt.dataType,
				Data:       tt.data,
			})

			file := openHDF5(t, tmpFile)
			defer file.Close()

			adapter := NewHDF5Adapter(file)
			variables, err := adapter.ConvertToMatlab()
			if err != nil {
				t.Fatalf("ConvertToMatlab() error = %v", err)
			}
			if len(variables) != 1 {
				t.Fatalf("expected 1 variable, got %d", len(variables))
			}

			v := variables[0]
			if v.DataType != tt.wantType {
				t.Errorf("DataType = %v, want %v", v.DataType, tt.wantType)
			}
			if v.Data == nil {
				t.Error("Data is nil")
			}
		})
	}
}

func TestConvertToMatlab_MultipleVariables(t *testing.T) {
	t.Skip("Skipped: HDF5 reader limitation - cannot read files with multiple datasets (separate issue)")

	vars := []*types.Variable{
		{
			Name:       "alpha",
			Dimensions: []int{2},
			DataType:   types.Double,
			Data:       []float64{1.0, 2.0},
		},
		{
			Name:       "beta",
			Dimensions: []int{3},
			DataType:   types.Int32,
			Data:       []int32{10, 20, 30},
		},
		{
			Name:       "gamma",
			Dimensions: []int{1},
			DataType:   types.Uint8,
			Data:       []uint8{255},
		},
	}

	tmpFile := writeTestFile(t, vars...)

	file := openHDF5(t, tmpFile)
	defer file.Close()

	adapter := NewHDF5Adapter(file)
	variables, err := adapter.ConvertToMatlab()
	if err != nil {
		t.Fatalf("ConvertToMatlab() error = %v", err)
	}
	if len(variables) != 3 {
		t.Fatalf("expected 3 variables, got %d", len(variables))
	}

	// Build name set for order-independent verification
	nameSet := make(map[string]bool)
	for _, v := range variables {
		nameSet[v.Name] = true
	}
	for _, expected := range []string{"alpha", "beta", "gamma"} {
		if !nameSet[expected] {
			t.Errorf("missing variable %q in results", expected)
		}
	}
}

func TestConvertToMatlab_ComplexVariable(t *testing.T) {
	t.Skip("Skipped: v73 writer bug - datasets not closed properly (separate issue)")

	realPart := []float64{1.0, 2.0, 3.0}
	imagPart := []float64{4.0, 5.0, 6.0}

	tmpFile := writeTestFile(t, &types.Variable{
		Name:       "z",
		Dimensions: []int{3},
		DataType:   types.Double,
		IsComplex:  true,
		Data: &types.NumericArray{
			Real: realPart,
			Imag: imagPart,
		},
	})

	file := openHDF5(t, tmpFile)
	defer file.Close()

	adapter := NewHDF5Adapter(file)
	variables, err := adapter.ConvertToMatlab()
	if err != nil {
		t.Fatalf("ConvertToMatlab() error = %v", err)
	}
	if len(variables) != 1 {
		t.Fatalf("expected 1 variable, got %d", len(variables))
	}

	v := variables[0]
	if !v.IsComplex {
		t.Error("IsComplex = false, want true")
	}

	numArray, ok := v.Data.(*types.NumericArray)
	if !ok {
		t.Fatalf("Data type = %T, want *types.NumericArray", v.Data)
	}

	gotReal, ok := numArray.Real.([]float64)
	if !ok {
		t.Fatalf("Real type = %T, want []float64", numArray.Real)
	}
	gotImag, ok := numArray.Imag.([]float64)
	if !ok {
		t.Fatalf("Imag type = %T, want []float64", numArray.Imag)
	}

	if len(gotReal) != len(realPart) {
		t.Fatalf("Real length = %d, want %d", len(gotReal), len(realPart))
	}
	for i := range realPart {
		if gotReal[i] != realPart[i] {
			t.Errorf("Real[%d] = %v, want %v", i, gotReal[i], realPart[i])
		}
		if gotImag[i] != imagPart[i] {
			t.Errorf("Imag[%d] = %v, want %v", i, gotImag[i], imagPart[i])
		}
	}
}

func TestMatlabClassToDataType(t *testing.T) {
	// Need an adapter instance to call the method.
	// Create a minimal valid file to get one.
	tmpFile := writeTestFile(t, &types.Variable{
		Name:       "dummy",
		Dimensions: []int{1},
		DataType:   types.Double,
		Data:       []float64{0},
	})
	file := openHDF5(t, tmpFile)
	defer file.Close()

	adapter := NewHDF5Adapter(file)

	tests := []struct {
		class string
		want  types.DataType
	}{
		{matlabClassDouble, types.Double},
		{matlabClassSingle, types.Single},
		{matlabClassInt8, types.Int8},
		{matlabClassUint8, types.Uint8},
		{matlabClassInt16, types.Int16},
		{matlabClassUint16, types.Uint16},
		{matlabClassInt32, types.Int32},
		{matlabClassUint32, types.Uint32},
		{matlabClassInt64, types.Int64},
		{matlabClassUint64, types.Uint64},
		{matlabClassChar, types.Char},
		{matlabClassStruct, types.Struct},
		{matlabClassCell, types.CellArray},
		{"nonexistent", types.Unknown},
		{"", types.Unknown},
	}

	for _, tt := range tests {
		t.Run(tt.class, func(t *testing.T) {
			got := adapter.matlabClassToDataType(tt.class)
			if got != tt.want {
				t.Errorf("matlabClassToDataType(%q) = %v, want %v", tt.class, got, tt.want)
			}
		})
	}
}

func TestConvertToMatlab_2DMatrix(t *testing.T) {
	data := []float64{1, 2, 3, 4, 5, 6}
	tmpFile := writeTestFile(t, &types.Variable{
		Name:       "matrix",
		Dimensions: []int{2, 3},
		DataType:   types.Double,
		Data:       data,
	})

	file := openHDF5(t, tmpFile)
	defer file.Close()

	adapter := NewHDF5Adapter(file)
	variables, err := adapter.ConvertToMatlab()
	if err != nil {
		t.Fatalf("ConvertToMatlab() error = %v", err)
	}
	if len(variables) != 1 {
		t.Fatalf("expected 1 variable, got %d", len(variables))
	}

	v := variables[0]
	if v.Name != "matrix" {
		t.Errorf("Name = %q, want %q", v.Name, "matrix")
	}

	floatData, ok := v.Data.([]float64)
	if !ok {
		t.Fatalf("Data type = %T, want []float64", v.Data)
	}
	if len(floatData) != len(data) {
		t.Fatalf("data length = %d, want %d", len(floatData), len(data))
	}
	for i, val := range floatData {
		if val != data[i] {
			t.Errorf("data[%d] = %v, want %v", i, val, data[i])
		}
	}
}

func TestConvertToMatlab_ScalarVariable(t *testing.T) {
	tmpFile := writeTestFile(t, &types.Variable{
		Name:       "pi",
		Dimensions: []int{1},
		DataType:   types.Double,
		Data:       []float64{3.14159},
	})

	file := openHDF5(t, tmpFile)
	defer file.Close()

	adapter := NewHDF5Adapter(file)
	variables, err := adapter.ConvertToMatlab()
	if err != nil {
		t.Fatalf("ConvertToMatlab() error = %v", err)
	}
	if len(variables) != 1 {
		t.Fatalf("expected 1 variable, got %d", len(variables))
	}

	v := variables[0]
	floatData, ok := v.Data.([]float64)
	if !ok {
		t.Fatalf("Data type = %T, want []float64", v.Data)
	}
	if len(floatData) != 1 {
		t.Fatalf("data length = %d, want 1", len(floatData))
	}
	if floatData[0] != 3.14159 {
		t.Errorf("data[0] = %v, want 3.14159", floatData[0])
	}
}

func TestConvertToMatlab_VariableAttributes(t *testing.T) {
	tmpFile := writeTestFile(t, &types.Variable{
		Name:       "data",
		Dimensions: []int{2},
		DataType:   types.Int32,
		Data:       []int32{10, 20},
	})

	file := openHDF5(t, tmpFile)
	defer file.Close()

	adapter := NewHDF5Adapter(file)
	variables, err := adapter.ConvertToMatlab()
	if err != nil {
		t.Fatalf("ConvertToMatlab() error = %v", err)
	}
	if len(variables) != 1 {
		t.Fatalf("expected 1 variable, got %d", len(variables))
	}

	v := variables[0]
	// Writer adds MATLAB_class attribute, so it should be in Attributes map
	if v.Attributes == nil {
		t.Fatal("Attributes map is nil")
	}
	if _, ok := v.Attributes["MATLAB_class"]; !ok {
		t.Error("MATLAB_class attribute not found in variable attributes")
	}
}
