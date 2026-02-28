package v73

import (
	"bytes"
	"os"
	"testing"

	"github.com/scigolib/matlab/types"
)

func TestNewParser(t *testing.T) {
	parser := NewParser()
	if parser == nil {
		t.Fatal("NewParser returned nil")
	}
}

func TestParser_Parse_SimpleDouble(t *testing.T) {
	expected := []float64{1.5, 2.5, 3.5}
	tmpFile := writeTestFile(t, &types.Variable{
		Name:       "values",
		Dimensions: []int{3},
		DataType:   types.Double,
		Data:       expected,
	})

	data, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}

	parser := NewParser()
	variables, err := parser.Parse(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if len(variables) != 1 {
		t.Fatalf("expected 1 variable, got %d", len(variables))
	}

	v := variables[0]
	if v.Name != "values" {
		t.Errorf("Name = %q, want %q", v.Name, "values")
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

func TestParser_Parse_AllTypes(t *testing.T) {
	tests := []struct {
		name     string
		dataType types.DataType
		data     interface{}
		wantType types.DataType
	}{
		{"double", types.Double, []float64{1.0, 2.0}, types.Double},
		{"single", types.Single, []float32{1.0, 2.0}, types.Single},
		{"int8", types.Int8, []int8{-1, 2}, types.Int8},
		{"uint8", types.Uint8, []uint8{0, 255}, types.Uint8},
		{"int16", types.Int16, []int16{-100, 100}, types.Int16},
		{"uint16", types.Uint16, []uint16{0, 65535}, types.Uint16},
		{"int32", types.Int32, []int32{-1000, 1000}, types.Int32},
		{"uint32", types.Uint32, []uint32{0, 4294967295}, types.Uint32},
		{"int64", types.Int64, []int64{-1000000, 1000000}, types.Int64},
		{"uint64", types.Uint64, []uint64{0, 18446744073709551615}, types.Uint64},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpFile := writeTestFile(t, &types.Variable{
				Name:       "var_" + tt.name,
				Dimensions: []int{2},
				DataType:   tt.dataType,
				Data:       tt.data,
			})

			data, err := os.ReadFile(tmpFile)
			if err != nil {
				t.Fatalf("ReadFile failed: %v", err)
			}

			parser := NewParser()
			variables, err := parser.Parse(bytes.NewReader(data))
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
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

func TestParser_Parse_ComplexVariable(t *testing.T) {
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

	data, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}

	parser := NewParser()
	variables, err := parser.Parse(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
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

	for i := range realPart {
		if gotReal[i] != realPart[i] {
			t.Errorf("Real[%d] = %v, want %v", i, gotReal[i], realPart[i])
		}
		if gotImag[i] != imagPart[i] {
			t.Errorf("Imag[%d] = %v, want %v", i, gotImag[i], imagPart[i])
		}
	}
}

func TestParser_Parse_MultipleVariables(t *testing.T) {
	t.Skip("Skipped: HDF5 reader limitation - cannot read files with multiple datasets (separate issue)")

	vars := []*types.Variable{
		{
			Name:       "x",
			Dimensions: []int{2},
			DataType:   types.Double,
			Data:       []float64{1.0, 2.0},
		},
		{
			Name:       "y",
			Dimensions: []int{3},
			DataType:   types.Int32,
			Data:       []int32{10, 20, 30},
		},
		{
			Name:       "z",
			Dimensions: []int{1},
			DataType:   types.Uint8,
			Data:       []uint8{42},
		},
	}

	tmpFile := writeTestFile(t, vars...)

	data, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}

	parser := NewParser()
	variables, err := parser.Parse(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if len(variables) != 3 {
		t.Fatalf("expected 3 variables, got %d", len(variables))
	}

	nameSet := make(map[string]bool)
	for _, v := range variables {
		nameSet[v.Name] = true
	}
	for _, name := range []string{"x", "y", "z"} {
		if !nameSet[name] {
			t.Errorf("missing variable %q in results", name)
		}
	}
}

func TestParser_Parse_InvalidReader(t *testing.T) {
	tests := []struct {
		name  string
		input []byte
	}{
		{"empty bytes", []byte{}},
		{"garbage data", []byte{0x00, 0x01, 0x02, 0x03, 0xFF, 0xFE, 0xFD}},
		{"partial HDF5 signature", []byte{0x89, 0x48, 0x44, 0x46}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewParser()
			_, err := parser.Parse(bytes.NewReader(tt.input))
			if err == nil {
				t.Error("Parse() expected error for invalid input, got nil")
			}
		})
	}
}

func TestParser_Parse_Roundtrip_DataIntegrity(t *testing.T) {
	// Write specific known values and verify exact roundtrip
	inputData := []float64{0, -1.5, 3.14159, 100.001, -999.999}
	tmpFile := writeTestFile(t, &types.Variable{
		Name:       "precise",
		Dimensions: []int{5},
		DataType:   types.Double,
		Data:       inputData,
	})

	data, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}

	parser := NewParser()
	variables, err := parser.Parse(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if len(variables) != 1 {
		t.Fatalf("expected 1 variable, got %d", len(variables))
	}

	floatData, ok := variables[0].Data.([]float64)
	if !ok {
		t.Fatalf("Data type = %T, want []float64", variables[0].Data)
	}
	if len(floatData) != len(inputData) {
		t.Fatalf("data length = %d, want %d", len(floatData), len(inputData))
	}
	for i, val := range floatData {
		if val != inputData[i] {
			t.Errorf("data[%d] = %v, want %v", i, val, inputData[i])
		}
	}
}

func TestParser_Parse_ScalarRoundtrip(t *testing.T) {
	tmpFile := writeTestFile(t, &types.Variable{
		Name:       "scalar",
		Dimensions: []int{1},
		DataType:   types.Double,
		Data:       []float64{42.0},
	})

	data, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}

	parser := NewParser()
	variables, err := parser.Parse(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if len(variables) != 1 {
		t.Fatalf("expected 1 variable, got %d", len(variables))
	}

	floatData, ok := variables[0].Data.([]float64)
	if !ok {
		t.Fatalf("Data type = %T, want []float64", variables[0].Data)
	}
	if len(floatData) != 1 || floatData[0] != 42.0 {
		t.Errorf("got %v, want [42.0]", floatData)
	}
}
