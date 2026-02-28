package v5

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"reflect"
	"testing"

	"github.com/scigolib/matlab/types"
)

// buildV5TestData uses the Writer to generate valid v5 binary data in memory.
func buildV5TestData(t *testing.T, variables ...*types.Variable) *bytes.Reader {
	t.Helper()
	var buf bytes.Buffer
	writer, err := NewWriter(&buf, "Test", "IM")
	if err != nil {
		t.Fatalf("NewWriter failed: %v", err)
	}
	for _, v := range variables {
		if err := writer.WriteVariable(v); err != nil {
			t.Fatalf("WriteVariable(%s) failed: %v", v.Name, err)
		}
	}
	return bytes.NewReader(buf.Bytes())
}

// buildV5TestDataEndian uses the Writer with a specified endian indicator.
func buildV5TestDataEndian(t *testing.T, endian string, variables ...*types.Variable) *bytes.Reader {
	t.Helper()
	var buf bytes.Buffer
	writer, err := NewWriter(&buf, "Test", endian)
	if err != nil {
		t.Fatalf("NewWriter failed: %v", err)
	}
	for _, v := range variables {
		if err := writer.WriteVariable(v); err != nil {
			t.Fatalf("WriteVariable(%s) failed: %v", v.Name, err)
		}
	}
	return bytes.NewReader(buf.Bytes())
}

// TestNewParser tests parser creation from valid and invalid data.
func TestNewParser(t *testing.T) {
	t.Run("valid header from writer", func(t *testing.T) {
		reader := buildV5TestData(t) // no variables, just header
		parser, err := NewParser(reader)
		if err != nil {
			t.Fatalf("NewParser() unexpected error: %v", err)
		}
		if parser == nil {
			t.Fatal("NewParser() returned nil parser")
		}
		if parser.Header == nil {
			t.Fatal("NewParser() parser.Header is nil")
		}
		if parser.Header.EndianIndicator != "IM" {
			t.Errorf("EndianIndicator = %q, want %q", parser.Header.EndianIndicator, "IM")
		}
		if parser.Header.Version != 0x0100 {
			t.Errorf("Version = 0x%04x, want 0x0100", parser.Header.Version)
		}
	})

	t.Run("too short data", func(t *testing.T) {
		reader := bytes.NewReader([]byte{1, 2, 3}) // less than 128 bytes
		_, err := NewParser(reader)
		if err == nil {
			t.Error("NewParser() expected error for too-short data, got nil")
		}
	})
}

// TestParse_EmptyReader tests that an empty reader causes NewParser to fail.
func TestParse_EmptyReader(t *testing.T) {
	reader := bytes.NewReader([]byte{})
	_, err := NewParser(reader)
	if err == nil {
		t.Error("NewParser() expected error for empty reader, got nil")
	}
}

// TestParse_TruncatedData tests parsing a file with header but no variables.
func TestParse_TruncatedData(t *testing.T) {
	reader := buildV5TestData(t) // header only, no variables
	parser, err := NewParser(reader)
	if err != nil {
		t.Fatalf("NewParser() unexpected error: %v", err)
	}

	file, err := parser.Parse()
	if err != nil {
		t.Fatalf("Parse() unexpected error: %v", err)
	}
	if len(file.Variables) != 0 {
		t.Errorf("Parse() returned %d variables, want 0", len(file.Variables))
	}
}

// TestParse_SimpleDouble tests roundtrip for a double array.
func TestParse_SimpleDouble(t *testing.T) {
	want := &types.Variable{
		Name:       "mydata",
		Dimensions: []int{1, 3},
		DataType:   types.Double,
		Data:       []float64{1.1, 2.2, 3.3},
	}

	reader := buildV5TestData(t, want)
	parser, err := NewParser(reader)
	if err != nil {
		t.Fatalf("NewParser() error: %v", err)
	}

	file, err := parser.Parse()
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}
	if len(file.Variables) != 1 {
		t.Fatalf("Parse() returned %d variables, want 1", len(file.Variables))
	}

	got := file.Variables[0]
	if got.Name != want.Name {
		t.Errorf("Name = %q, want %q", got.Name, want.Name)
	}
	if !reflect.DeepEqual(got.Dimensions, want.Dimensions) {
		t.Errorf("Dimensions = %v, want %v", got.Dimensions, want.Dimensions)
	}
	if got.DataType != want.DataType {
		t.Errorf("DataType = %v, want %v", got.DataType, want.DataType)
	}

	gotData, ok := got.Data.([]float64)
	if !ok {
		t.Fatalf("Data type = %T, want []float64", got.Data)
	}
	wantData := want.Data.([]float64)
	if !reflect.DeepEqual(gotData, wantData) {
		t.Errorf("Data = %v, want %v", gotData, wantData)
	}
}

// TestParse_Int32 tests roundtrip for an int32 array.
func TestParse_Int32(t *testing.T) {
	want := &types.Variable{
		Name:       "intvals",
		Dimensions: []int{1, 4},
		DataType:   types.Int32,
		Data:       []int32{-100, 0, 42, 1000},
	}

	reader := buildV5TestData(t, want)
	parser, err := NewParser(reader)
	if err != nil {
		t.Fatalf("NewParser() error: %v", err)
	}

	file, err := parser.Parse()
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}
	if len(file.Variables) != 1 {
		t.Fatalf("Parse() returned %d variables, want 1", len(file.Variables))
	}

	got := file.Variables[0]
	if got.Name != want.Name {
		t.Errorf("Name = %q, want %q", got.Name, want.Name)
	}
	if got.DataType != types.Int32 {
		t.Errorf("DataType = %v, want %v", got.DataType, types.Int32)
	}

	gotData, ok := got.Data.([]int32)
	if !ok {
		t.Fatalf("Data type = %T, want []int32", got.Data)
	}
	wantData := want.Data.([]int32)
	if !reflect.DeepEqual(gotData, wantData) {
		t.Errorf("Data = %v, want %v", gotData, wantData)
	}
}

// TestParse_Uint8 tests roundtrip for a uint8 array.
func TestParse_Uint8(t *testing.T) {
	want := &types.Variable{
		Name:       "bytes",
		Dimensions: []int{1, 5},
		DataType:   types.Uint8,
		Data:       []byte{10, 20, 30, 40, 50},
	}

	reader := buildV5TestData(t, want)
	parser, err := NewParser(reader)
	if err != nil {
		t.Fatalf("NewParser() error: %v", err)
	}

	file, err := parser.Parse()
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}
	if len(file.Variables) != 1 {
		t.Fatalf("Parse() returned %d variables, want 1", len(file.Variables))
	}

	got := file.Variables[0]
	if got.Name != want.Name {
		t.Errorf("Name = %q, want %q", got.Name, want.Name)
	}
	if got.DataType != types.Uint8 {
		t.Errorf("DataType = %v, want %v", got.DataType, types.Uint8)
	}

	gotData, ok := got.Data.([]byte)
	if !ok {
		t.Fatalf("Data type = %T, want []byte", got.Data)
	}
	wantData := want.Data.([]byte)
	if !reflect.DeepEqual(gotData, wantData) {
		t.Errorf("Data = %v, want %v", gotData, wantData)
	}
}

// TestParse_Single tests roundtrip for a float32 array.
func TestParse_Single(t *testing.T) {
	want := &types.Variable{
		Name:       "floats",
		Dimensions: []int{1, 3},
		DataType:   types.Single,
		Data:       []float32{1.5, 2.5, 3.5},
	}

	reader := buildV5TestData(t, want)
	parser, err := NewParser(reader)
	if err != nil {
		t.Fatalf("NewParser() error: %v", err)
	}

	file, err := parser.Parse()
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}
	if len(file.Variables) != 1 {
		t.Fatalf("Parse() returned %d variables, want 1", len(file.Variables))
	}

	got := file.Variables[0]
	if got.Name != want.Name {
		t.Errorf("Name = %q, want %q", got.Name, want.Name)
	}
	if got.DataType != types.Single {
		t.Errorf("DataType = %v, want %v", got.DataType, types.Single)
	}

	gotData, ok := got.Data.([]float32)
	if !ok {
		t.Fatalf("Data type = %T, want []float32", got.Data)
	}
	wantData := want.Data.([]float32)
	if !reflect.DeepEqual(gotData, wantData) {
		t.Errorf("Data = %v, want %v", gotData, wantData)
	}
}

// TestParse_Complex tests roundtrip for a complex variable.
func TestParse_Complex(t *testing.T) {
	want := &types.Variable{
		Name:       "cmplx",
		Dimensions: []int{1, 3},
		DataType:   types.Double,
		IsComplex:  true,
		Data: &types.NumericArray{
			Real: []float64{1.0, 2.0, 3.0},
			Imag: []float64{4.0, 5.0, 6.0},
		},
	}

	reader := buildV5TestData(t, want)
	parser, err := NewParser(reader)
	if err != nil {
		t.Fatalf("NewParser() error: %v", err)
	}

	file, err := parser.Parse()
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}
	if len(file.Variables) != 1 {
		t.Fatalf("Parse() returned %d variables, want 1", len(file.Variables))
	}

	got := file.Variables[0]
	if got.Name != want.Name {
		t.Errorf("Name = %q, want %q", got.Name, want.Name)
	}
	if !got.IsComplex {
		t.Error("IsComplex = false, want true")
	}
	if got.DataType != types.Double {
		t.Errorf("DataType = %v, want %v", got.DataType, types.Double)
	}

	gotArray, ok := got.Data.(*types.NumericArray)
	if !ok {
		t.Fatalf("Data type = %T, want *types.NumericArray", got.Data)
	}

	wantArray := want.Data.(*types.NumericArray)
	gotReal, ok := gotArray.Real.([]float64)
	if !ok {
		t.Fatalf("Real type = %T, want []float64", gotArray.Real)
	}
	wantReal := wantArray.Real.([]float64)
	if !reflect.DeepEqual(gotReal, wantReal) {
		t.Errorf("Real = %v, want %v", gotReal, wantReal)
	}

	gotImag, ok := gotArray.Imag.([]float64)
	if !ok {
		t.Fatalf("Imag type = %T, want []float64", gotArray.Imag)
	}
	wantImag := wantArray.Imag.([]float64)
	if !reflect.DeepEqual(gotImag, wantImag) {
		t.Errorf("Imag = %v, want %v", gotImag, wantImag)
	}
}

// TestParse_Matrix2D tests roundtrip for a 2D matrix.
func TestParse_Matrix2D(t *testing.T) {
	want := &types.Variable{
		Name:       "matrix",
		Dimensions: []int{2, 3},
		DataType:   types.Double,
		Data:       []float64{1, 2, 3, 4, 5, 6},
	}

	reader := buildV5TestData(t, want)
	parser, err := NewParser(reader)
	if err != nil {
		t.Fatalf("NewParser() error: %v", err)
	}

	file, err := parser.Parse()
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}
	if len(file.Variables) != 1 {
		t.Fatalf("Parse() returned %d variables, want 1", len(file.Variables))
	}

	got := file.Variables[0]
	if !reflect.DeepEqual(got.Dimensions, []int{2, 3}) {
		t.Errorf("Dimensions = %v, want [2 3]", got.Dimensions)
	}

	gotData, ok := got.Data.([]float64)
	if !ok {
		t.Fatalf("Data type = %T, want []float64", got.Data)
	}
	wantData := want.Data.([]float64)
	if !reflect.DeepEqual(gotData, wantData) {
		t.Errorf("Data = %v, want %v", gotData, wantData)
	}
}

// TestParse_MultipleVariables tests parsing multiple variables from a single stream.
func TestParse_MultipleVariables(t *testing.T) {
	v1 := &types.Variable{
		Name:       "alpha",
		Dimensions: []int{1, 2},
		DataType:   types.Double,
		Data:       []float64{1.0, 2.0},
	}
	v2 := &types.Variable{
		Name:       "beta",
		Dimensions: []int{1, 3},
		DataType:   types.Int32,
		Data:       []int32{10, 20, 30},
	}
	v3 := &types.Variable{
		Name:       "gamma",
		Dimensions: []int{1, 4},
		DataType:   types.Uint8,
		Data:       []byte{0, 1, 2, 3},
	}

	reader := buildV5TestData(t, v1, v2, v3)
	parser, err := NewParser(reader)
	if err != nil {
		t.Fatalf("NewParser() error: %v", err)
	}

	file, err := parser.Parse()
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}
	if len(file.Variables) != 3 {
		t.Fatalf("Parse() returned %d variables, want 3", len(file.Variables))
	}

	// Verify names in order
	wantNames := []string{"alpha", "beta", "gamma"}
	for i, name := range wantNames {
		if file.Variables[i].Name != name {
			t.Errorf("Variable[%d].Name = %q, want %q", i, file.Variables[i].Name, name)
		}
	}

	// Verify types
	wantTypes := []types.DataType{types.Double, types.Int32, types.Uint8}
	for i, dt := range wantTypes {
		if file.Variables[i].DataType != dt {
			t.Errorf("Variable[%d].DataType = %v, want %v", i, file.Variables[i].DataType, dt)
		}
	}
}

// TestParse_BigEndian tests roundtrip with big-endian byte order.
func TestParse_BigEndian(t *testing.T) {
	want := &types.Variable{
		Name:       "bigend",
		Dimensions: []int{1, 3},
		DataType:   types.Double,
		Data:       []float64{10.0, 20.0, 30.0},
	}

	reader := buildV5TestDataEndian(t, "MI", want)
	parser, err := NewParser(reader)
	if err != nil {
		t.Fatalf("NewParser() error: %v", err)
	}

	if parser.Header.EndianIndicator != "MI" {
		t.Errorf("EndianIndicator = %q, want %q", parser.Header.EndianIndicator, "MI")
	}

	file, err := parser.Parse()
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}
	if len(file.Variables) != 1 {
		t.Fatalf("Parse() returned %d variables, want 1", len(file.Variables))
	}

	got := file.Variables[0]
	if got.Name != want.Name {
		t.Errorf("Name = %q, want %q", got.Name, want.Name)
	}

	gotData, ok := got.Data.([]float64)
	if !ok {
		t.Fatalf("Data type = %T, want []float64", got.Data)
	}
	wantData := want.Data.([]float64)
	if !reflect.DeepEqual(gotData, wantData) {
		t.Errorf("Data = %v, want %v", gotData, wantData)
	}
}

// TestParse_AllNumericTypes is a table-driven test covering all 10 numeric types.
//
//nolint:gocognit // Table-driven test with exhaustive type coverage
func TestParse_AllNumericTypes(t *testing.T) {
	tests := []struct {
		name     string
		variable *types.Variable
	}{
		{
			name: "Double",
			variable: &types.Variable{
				Name:       "dbl",
				Dimensions: []int{1, 2},
				DataType:   types.Double,
				Data:       []float64{3.14, 2.71},
			},
		},
		{
			name: "Single",
			variable: &types.Variable{
				Name:       "sgl",
				Dimensions: []int{1, 2},
				DataType:   types.Single,
				Data:       []float32{1.5, 2.5},
			},
		},
		{
			name: "Int8",
			variable: &types.Variable{
				Name:       "i8",
				Dimensions: []int{1, 3},
				DataType:   types.Int8,
				Data:       []int8{-128, 0, 127},
			},
		},
		{
			name: "Uint8",
			variable: &types.Variable{
				Name:       "u8",
				Dimensions: []int{1, 3},
				DataType:   types.Uint8,
				Data:       []byte{0, 128, 255},
			},
		},
		{
			name: "Int16",
			variable: &types.Variable{
				Name:       "i16",
				Dimensions: []int{1, 3},
				DataType:   types.Int16,
				Data:       []int16{-32768, 0, 32767},
			},
		},
		{
			name: "Uint16",
			variable: &types.Variable{
				Name:       "u16",
				Dimensions: []int{1, 3},
				DataType:   types.Uint16,
				Data:       []uint16{0, 1000, 65535},
			},
		},
		{
			name: "Int32",
			variable: &types.Variable{
				Name:       "i32",
				Dimensions: []int{1, 2},
				DataType:   types.Int32,
				Data:       []int32{-2147483648, 2147483647},
			},
		},
		{
			name: "Uint32",
			variable: &types.Variable{
				Name:       "u32",
				Dimensions: []int{1, 2},
				DataType:   types.Uint32,
				Data:       []uint32{0, 4294967295},
			},
		},
		{
			name: "Int64",
			variable: &types.Variable{
				Name:       "i64",
				Dimensions: []int{1, 2},
				DataType:   types.Int64,
				Data:       []int64{-9223372036854775808, 9223372036854775807},
			},
		},
		{
			name: "Uint64",
			variable: &types.Variable{
				Name:       "u64",
				Dimensions: []int{1, 2},
				DataType:   types.Uint64,
				Data:       []uint64{0, 18446744073709551615},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := buildV5TestData(t, tt.variable)
			parser, err := NewParser(reader)
			if err != nil {
				t.Fatalf("NewParser() error: %v", err)
			}

			file, err := parser.Parse()
			if err != nil {
				t.Fatalf("Parse() error: %v", err)
			}
			if len(file.Variables) != 1 {
				t.Fatalf("Parse() returned %d variables, want 1", len(file.Variables))
			}

			got := file.Variables[0]
			if got.Name != tt.variable.Name {
				t.Errorf("Name = %q, want %q", got.Name, tt.variable.Name)
			}
			if got.DataType != tt.variable.DataType {
				t.Errorf("DataType = %v, want %v", got.DataType, tt.variable.DataType)
			}
			if !reflect.DeepEqual(got.Dimensions, tt.variable.Dimensions) {
				t.Errorf("Dimensions = %v, want %v", got.Dimensions, tt.variable.Dimensions)
			}
			if !reflect.DeepEqual(got.Data, tt.variable.Data) {
				t.Errorf("Data = %v, want %v", got.Data, tt.variable.Data)
			}
		})
	}
}

// TestReadData_SmallFormat verifies that short variable names (<=4 bytes) use small
// format encoding for the name sub-element, and are read back correctly.
// The writer uses regular format for all sub-elements, but the parser's readData
// also handles small format. This test uses a short name to ensure the name
// sub-element is parsed correctly regardless of format.
func TestReadData_SmallFormat(t *testing.T) {
	want := &types.Variable{
		Name:       "x", // 1 byte name
		Dimensions: []int{1, 2},
		DataType:   types.Double,
		Data:       []float64{7.0, 8.0},
	}

	reader := buildV5TestData(t, want)
	parser, err := NewParser(reader)
	if err != nil {
		t.Fatalf("NewParser() error: %v", err)
	}

	file, err := parser.Parse()
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}
	if len(file.Variables) != 1 {
		t.Fatalf("Parse() returned %d variables, want 1", len(file.Variables))
	}

	got := file.Variables[0]
	if got.Name != "x" {
		t.Errorf("Name = %q, want %q", got.Name, "x")
	}
}

// TestReadData_RegularFormat verifies that arrays larger than 4 bytes use regular
// format encoding for data sub-elements, and are read back correctly.
func TestReadData_RegularFormat(t *testing.T) {
	// Large array ensures regular format is used for the data sub-element
	data := make([]float64, 100)
	for i := range data {
		data[i] = float64(i)
	}
	want := &types.Variable{
		Name:       "largearray",
		Dimensions: []int{1, 100},
		DataType:   types.Double,
		Data:       data,
	}

	reader := buildV5TestData(t, want)
	parser, err := NewParser(reader)
	if err != nil {
		t.Fatalf("NewParser() error: %v", err)
	}

	file, err := parser.Parse()
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}
	if len(file.Variables) != 1 {
		t.Fatalf("Parse() returned %d variables, want 1", len(file.Variables))
	}

	got := file.Variables[0]
	gotData, ok := got.Data.([]float64)
	if !ok {
		t.Fatalf("Data type = %T, want []float64", got.Data)
	}
	if len(gotData) != 100 {
		t.Fatalf("len(Data) = %d, want 100", len(gotData))
	}
	for i := 0; i < 100; i++ {
		if gotData[i] != float64(i) {
			t.Errorf("Data[%d] = %v, want %v", i, gotData[i], float64(i))
			break
		}
	}
}

// TestSkipData tests that parsing continues correctly when unknown tag types
// would hypothetically be encountered. Since the Writer only produces miMATRIX
// tags, we test indirectly by verifying that valid files with multiple variables
// don't get confused by padding or alignment.
func TestSkipData(t *testing.T) {
	// Write multiple variables with different data sizes to exercise
	// alignment and padding in the parser.
	v1 := &types.Variable{
		Name:       "short",
		Dimensions: []int{1, 1},
		DataType:   types.Uint8,
		Data:       []byte{42},
	}
	v2 := &types.Variable{
		Name:       "longer",
		Dimensions: []int{1, 7},
		DataType:   types.Uint8,
		Data:       []byte{1, 2, 3, 4, 5, 6, 7},
	}
	v3 := &types.Variable{
		Name:       "doubles",
		Dimensions: []int{1, 3},
		DataType:   types.Double,
		Data:       []float64{9.9, 8.8, 7.7},
	}

	reader := buildV5TestData(t, v1, v2, v3)
	parser, err := NewParser(reader)
	if err != nil {
		t.Fatalf("NewParser() error: %v", err)
	}

	file, err := parser.Parse()
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}
	if len(file.Variables) != 3 {
		t.Fatalf("Parse() returned %d variables, want 3", len(file.Variables))
	}

	// Verify all three were parsed correctly despite different padding
	if file.Variables[0].Name != "short" {
		t.Errorf("Variable[0].Name = %q, want %q", file.Variables[0].Name, "short")
	}
	if file.Variables[1].Name != "longer" {
		t.Errorf("Variable[1].Name = %q, want %q", file.Variables[1].Name, "longer")
	}
	if file.Variables[2].Name != "doubles" {
		t.Errorf("Variable[2].Name = %q, want %q", file.Variables[2].Name, "doubles")
	}
}

// TestParse_HeaderPreserved verifies that the parsed file contains the header.
func TestParse_HeaderPreserved(t *testing.T) {
	reader := buildV5TestData(t, &types.Variable{
		Name:       "val",
		Dimensions: []int{1, 1},
		DataType:   types.Double,
		Data:       []float64{42.0},
	})

	parser, err := NewParser(reader)
	if err != nil {
		t.Fatalf("NewParser() error: %v", err)
	}

	file, err := parser.Parse()
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	if file.Header == nil {
		t.Fatal("Parse() file.Header is nil")
	}
	if file.Header != parser.Header {
		t.Error("Parse() file.Header != parser.Header")
	}
}

// TestParse_ComplexSingle tests roundtrip for complex float32 data.
func TestParse_ComplexSingle(t *testing.T) {
	want := &types.Variable{
		Name:       "csgl",
		Dimensions: []int{1, 2},
		DataType:   types.Single,
		IsComplex:  true,
		Data: &types.NumericArray{
			Real: []float32{1.0, 2.0},
			Imag: []float32{3.0, 4.0},
		},
	}

	reader := buildV5TestData(t, want)
	parser, err := NewParser(reader)
	if err != nil {
		t.Fatalf("NewParser() error: %v", err)
	}

	file, err := parser.Parse()
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}
	if len(file.Variables) != 1 {
		t.Fatalf("Parse() returned %d variables, want 1", len(file.Variables))
	}

	got := file.Variables[0]
	if !got.IsComplex {
		t.Error("IsComplex = false, want true")
	}

	gotArray, ok := got.Data.(*types.NumericArray)
	if !ok {
		t.Fatalf("Data type = %T, want *types.NumericArray", got.Data)
	}

	wantArray := want.Data.(*types.NumericArray)
	if !reflect.DeepEqual(gotArray.Real, wantArray.Real) {
		t.Errorf("Real = %v, want %v", gotArray.Real, wantArray.Real)
	}
	if !reflect.DeepEqual(gotArray.Imag, wantArray.Imag) {
		t.Errorf("Imag = %v, want %v", gotArray.Imag, wantArray.Imag)
	}
}

// TestParse_ScalarVariable tests roundtrip for a scalar (1x1) variable.
func TestParse_ScalarVariable(t *testing.T) {
	want := &types.Variable{
		Name:       "scalar",
		Dimensions: []int{1, 1},
		DataType:   types.Double,
		Data:       []float64{99.99},
	}

	reader := buildV5TestData(t, want)
	parser, err := NewParser(reader)
	if err != nil {
		t.Fatalf("NewParser() error: %v", err)
	}

	file, err := parser.Parse()
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}
	if len(file.Variables) != 1 {
		t.Fatalf("Parse() returned %d variables, want 1", len(file.Variables))
	}

	got := file.Variables[0]
	if !reflect.DeepEqual(got.Dimensions, []int{1, 1}) {
		t.Errorf("Dimensions = %v, want [1 1]", got.Dimensions)
	}

	gotData, ok := got.Data.([]float64)
	if !ok {
		t.Fatalf("Data type = %T, want []float64", got.Data)
	}
	if len(gotData) != 1 || gotData[0] != 99.99 {
		t.Errorf("Data = %v, want [99.99]", gotData)
	}
}

// TestParse_BigEndian_AllTypes tests big-endian roundtrip with multiple types.
func TestParse_BigEndian_AllTypes(t *testing.T) {
	v1 := &types.Variable{
		Name:       "bedbl",
		Dimensions: []int{1, 2},
		DataType:   types.Double,
		Data:       []float64{1.0, 2.0},
	}
	v2 := &types.Variable{
		Name:       "bei32",
		Dimensions: []int{1, 2},
		DataType:   types.Int32,
		Data:       []int32{-1, 1},
	}

	reader := buildV5TestDataEndian(t, "MI", v1, v2)
	parser, err := NewParser(reader)
	if err != nil {
		t.Fatalf("NewParser() error: %v", err)
	}

	file, err := parser.Parse()
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}
	if len(file.Variables) != 2 {
		t.Fatalf("Parse() returned %d variables, want 2", len(file.Variables))
	}

	// Verify first variable
	gotDbl, ok := file.Variables[0].Data.([]float64)
	if !ok {
		t.Fatalf("Variable[0].Data type = %T, want []float64", file.Variables[0].Data)
	}
	if !reflect.DeepEqual(gotDbl, []float64{1.0, 2.0}) {
		t.Errorf("Variable[0].Data = %v, want [1.0 2.0]", gotDbl)
	}

	// Verify second variable
	gotI32, ok := file.Variables[1].Data.([]int32)
	if !ok {
		t.Fatalf("Variable[1].Data type = %T, want []int32", file.Variables[1].Data)
	}
	if !reflect.DeepEqual(gotI32, []int32{-1, 1}) {
		t.Errorf("Variable[1].Data = %v, want [-1 1]", gotI32)
	}
}

// TestParse_UnknownTagType tests that the parser's default branch (skipData) is
// exercised when an unknown tag type is encountered, and parsing continues to
// the next valid miMATRIX element.
func TestParse_UnknownTagType(t *testing.T) {
	// Strategy: Build binary data with header + unknown tag + miMATRIX variable.
	// The parser should skip the unknown tag and parse the variable.

	// Step 1: Build valid header + miMATRIX content from the Writer.
	v := &types.Variable{
		Name:       "after",
		Dimensions: []int{1, 2},
		DataType:   types.Double,
		Data:       []float64{7.0, 8.0},
	}
	writerBuf := buildV5TestData(t, v)
	allBytes := make([]byte, writerBuf.Len())
	_, _ = writerBuf.Read(allBytes)

	// allBytes: [128-byte header][miMATRIX element]
	header := allBytes[:128]
	matrixElement := allBytes[128:]

	// Step 2: Craft an unknown tag (type=99, which is not miMATRIX/miCOMPRESSED).
	// Regular format tag: 4 bytes type + 4 bytes size + data + padding.
	unknownDataSize := uint32(16) // 16 bytes of dummy data (already 8-byte aligned)
	unknownTag := make([]byte, 8+unknownDataSize)
	binary.LittleEndian.PutUint32(unknownTag[0:4], 99) // unknown type
	binary.LittleEndian.PutUint32(unknownTag[4:8], unknownDataSize)
	// Bytes 8..24 are dummy data (zeros)

	// Step 3: Assemble: header + unknown tag + miMATRIX element
	var assembled bytes.Buffer
	assembled.Write(header)
	assembled.Write(unknownTag)
	assembled.Write(matrixElement)

	// Step 4: Parse
	parser, err := NewParser(&assembled)
	if err != nil {
		t.Fatalf("NewParser() error: %v", err)
	}

	file, err := parser.Parse()
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}
	if len(file.Variables) != 1 {
		t.Fatalf("Parse() returned %d variables, want 1", len(file.Variables))
	}

	got := file.Variables[0]
	if got.Name != "after" {
		t.Errorf("Name = %q, want %q", got.Name, "after")
	}
	gotData, ok := got.Data.([]float64)
	if !ok {
		t.Fatalf("Data type = %T, want []float64", got.Data)
	}
	if !reflect.DeepEqual(gotData, []float64{7.0, 8.0}) {
		t.Errorf("Data = %v, want [7.0 8.0]", gotData)
	}
}

// TestParse_SkipData_SmallFormat tests skipData with a small-format unknown tag.
// Small format tags have data packed in the tag itself (no extra bytes to skip).
func TestParse_SkipData_SmallFormat(t *testing.T) {
	// Build a valid variable from the Writer
	v := &types.Variable{
		Name:       "val",
		Dimensions: []int{1, 1},
		DataType:   types.Double,
		Data:       []float64{42.0},
	}
	writerBuf := buildV5TestData(t, v)
	allBytes := make([]byte, writerBuf.Len())
	_, _ = writerBuf.Read(allBytes)

	header := allBytes[:128]
	matrixElement := allBytes[128:]

	// Craft a small-format unknown tag.
	// Small format: upper 16 bits = size (1-4), lower 16 bits = type.
	// Example: type=99, size=2 -> first word = (2 << 16) | 99 = 0x00020063
	// The data is in bytes 4-7 of the 8-byte tag.
	smallTag := make([]byte, 8)
	binary.LittleEndian.PutUint32(smallTag[0:4], (2<<16)|99) // size=2, type=99
	smallTag[4] = 0xAA                                        // dummy data byte 1
	smallTag[5] = 0xBB                                        // dummy data byte 2

	var assembled bytes.Buffer
	assembled.Write(header)
	assembled.Write(smallTag)
	assembled.Write(matrixElement)

	parser, err := NewParser(&assembled)
	if err != nil {
		t.Fatalf("NewParser() error: %v", err)
	}

	file, err := parser.Parse()
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}
	if len(file.Variables) != 1 {
		t.Fatalf("Parse() returned %d variables, want 1", len(file.Variables))
	}
	if file.Variables[0].Name != "val" {
		t.Errorf("Name = %q, want %q", file.Variables[0].Name, "val")
	}
}

// TestParse_CompressedElement tests the miCOMPRESSED branch of Parse().
// We craft a zlib-compressed miMATRIX element and verify the parser decompresses
// and parses it correctly.
//
//nolint:gocognit // Test builds compressed binary data with multiple steps
func TestParse_CompressedElement(t *testing.T) {
	// Step 1: Use the Writer to build a valid miMATRIX element (without header).
	v := &types.Variable{
		Name:       "compressed",
		Dimensions: []int{1, 3},
		DataType:   types.Double,
		Data:       []float64{10.0, 20.0, 30.0},
	}

	// Write to get the raw miMATRIX bytes (skip header)
	writerBuf := buildV5TestData(t, v)
	allBytes := make([]byte, writerBuf.Len())
	_, _ = writerBuf.Read(allBytes)

	header := allBytes[:128]
	matrixBytes := allBytes[128:] // This is the miMATRIX tag + content

	// Step 2: Compress the miMATRIX element with zlib.
	var compressedBuf bytes.Buffer
	zlibWriter := zlib.NewWriter(&compressedBuf)
	_, err := zlibWriter.Write(matrixBytes)
	if err != nil {
		t.Fatalf("zlib write failed: %v", err)
	}
	if err := zlibWriter.Close(); err != nil {
		t.Fatalf("zlib close failed: %v", err)
	}
	compressedData := compressedBuf.Bytes()

	// Step 3: Build a miCOMPRESSED tag.
	compressedTag := make([]byte, 8)
	binary.LittleEndian.PutUint32(compressedTag[0:4], miCOMPRESSED)
	binary.LittleEndian.PutUint32(compressedTag[4:8], uint32(len(compressedData)))

	// Step 4: Assemble: header + miCOMPRESSED tag + compressed data
	// Note: compressed elements do NOT have padding after the data.
	var assembled bytes.Buffer
	assembled.Write(header)
	assembled.Write(compressedTag)
	assembled.Write(compressedData)

	// Step 5: Parse
	parser, err := NewParser(&assembled)
	if err != nil {
		t.Fatalf("NewParser() error: %v", err)
	}

	file, err := parser.Parse()
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}
	if len(file.Variables) != 1 {
		t.Fatalf("Parse() returned %d variables, want 1", len(file.Variables))
	}

	got := file.Variables[0]
	if got.Name != "compressed" {
		t.Errorf("Name = %q, want %q", got.Name, "compressed")
	}
	if got.DataType != types.Double {
		t.Errorf("DataType = %v, want %v", got.DataType, types.Double)
	}

	gotData, ok := got.Data.([]float64)
	if !ok {
		t.Fatalf("Data type = %T, want []float64", got.Data)
	}
	wantData := []float64{10.0, 20.0, 30.0}
	if !reflect.DeepEqual(gotData, wantData) {
		t.Errorf("Data = %v, want %v", gotData, wantData)
	}
}

// TestParse_SkipData_RegularFormatWithPadding tests that skipData correctly
// handles padding for non-8-byte-aligned data sizes.
func TestParse_SkipData_RegularFormatWithPadding(t *testing.T) {
	v := &types.Variable{
		Name:       "target",
		Dimensions: []int{1, 1},
		DataType:   types.Double,
		Data:       []float64{99.0},
	}
	writerBuf := buildV5TestData(t, v)
	allBytes := make([]byte, writerBuf.Len())
	_, _ = writerBuf.Read(allBytes)

	header := allBytes[:128]
	matrixElement := allBytes[128:]

	// Craft an unknown tag with non-aligned data size (5 bytes, needs 3 bytes padding)
	unknownDataSize := uint32(5)
	padding := (8 - unknownDataSize%8) % 8
	unknownTag := make([]byte, 8+unknownDataSize+padding)
	binary.LittleEndian.PutUint32(unknownTag[0:4], 99) // unknown type
	binary.LittleEndian.PutUint32(unknownTag[4:8], unknownDataSize)
	// Fill data with recognizable pattern
	for i := uint32(0); i < unknownDataSize; i++ {
		unknownTag[8+i] = byte(0xDE)
	}

	var assembled bytes.Buffer
	assembled.Write(header)
	assembled.Write(unknownTag)
	assembled.Write(matrixElement)

	parser, err := NewParser(&assembled)
	if err != nil {
		t.Fatalf("NewParser() error: %v", err)
	}

	file, err := parser.Parse()
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}
	if len(file.Variables) != 1 {
		t.Fatalf("Parse() returned %d variables, want 1", len(file.Variables))
	}
	if file.Variables[0].Name != "target" {
		t.Errorf("Name = %q, want %q", file.Variables[0].Name, "target")
	}
}
