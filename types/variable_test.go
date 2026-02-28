package types

import (
	"strings"
	"testing"
)

// TestVariable_GetFloat64Array tests extracting float64 data.
func TestVariable_GetFloat64Array(t *testing.T) {
	tests := []struct {
		name      string
		variable  *Variable
		want      []float64
		wantError bool
	}{
		{
			name: "float64 direct",
			variable: &Variable{
				Name:       "test",
				Dimensions: []int{3},
				DataType:   Double,
				Data:       []float64{1.0, 2.0, 3.0},
				IsComplex:  false,
			},
			want:      []float64{1.0, 2.0, 3.0},
			wantError: false,
		},
		{
			name: "float32 conversion",
			variable: &Variable{
				Name:       "test",
				Dimensions: []int{3},
				DataType:   Single,
				Data:       []float32{1.0, 2.0, 3.0},
				IsComplex:  false,
			},
			want:      []float64{1.0, 2.0, 3.0},
			wantError: false,
		},
		{
			name: "int32 conversion",
			variable: &Variable{
				Name:       "test",
				Dimensions: []int{3},
				DataType:   Int32,
				Data:       []int32{1, 2, 3},
				IsComplex:  false,
			},
			want:      []float64{1.0, 2.0, 3.0},
			wantError: false,
		},
		{
			name: "uint8 conversion",
			variable: &Variable{
				Name:       "test",
				Dimensions: []int{3},
				DataType:   Uint8,
				Data:       []uint8{1, 2, 3},
				IsComplex:  false,
			},
			want:      []float64{1.0, 2.0, 3.0},
			wantError: false,
		},
		{
			name: "complex error",
			variable: &Variable{
				Name:       "test",
				Dimensions: []int{3},
				DataType:   Double,
				Data: &NumericArray{
					Real: []float64{1.0, 2.0, 3.0},
					Imag: []float64{4.0, 5.0, 6.0},
				},
				IsComplex: true,
			},
			want:      nil,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.variable.GetFloat64Array()
			if (err != nil) != tt.wantError {
				t.Errorf("GetFloat64Array() error = %v, wantError %v", err, tt.wantError)
				return
			}
			if !tt.wantError {
				if len(got) != len(tt.want) {
					t.Errorf("GetFloat64Array() length = %d, want %d", len(got), len(tt.want))
					return
				}
				for i := range got {
					if got[i] != tt.want[i] {
						t.Errorf("GetFloat64Array()[%d] = %v, want %v", i, got[i], tt.want[i])
					}
				}
			}
		})
	}
}

// TestVariable_GetInt32Array tests extracting int32 data.
func TestVariable_GetInt32Array(t *testing.T) {
	tests := []struct {
		name      string
		variable  *Variable
		want      []int32
		wantError bool
	}{
		{
			name: "int32 direct",
			variable: &Variable{
				Name:       "test",
				Dimensions: []int{3},
				DataType:   Int32,
				Data:       []int32{1, 2, 3},
				IsComplex:  false,
			},
			want:      []int32{1, 2, 3},
			wantError: false,
		},
		{
			name: "int8 conversion",
			variable: &Variable{
				Name:       "test",
				Dimensions: []int{3},
				DataType:   Int8,
				Data:       []int8{1, 2, 3},
				IsComplex:  false,
			},
			want:      []int32{1, 2, 3},
			wantError: false,
		},
		{
			name: "uint16 conversion",
			variable: &Variable{
				Name:       "test",
				Dimensions: []int{3},
				DataType:   Uint16,
				Data:       []uint16{1, 2, 3},
				IsComplex:  false,
			},
			want:      []int32{1, 2, 3},
			wantError: false,
		},
		{
			name: "complex error",
			variable: &Variable{
				Name:       "test",
				Dimensions: []int{3},
				DataType:   Int32,
				Data:       &NumericArray{},
				IsComplex:  true,
			},
			want:      nil,
			wantError: true,
		},
		{
			name: "float64 error",
			variable: &Variable{
				Name:       "test",
				Dimensions: []int{3},
				DataType:   Double,
				Data:       []float64{1.0, 2.0, 3.0},
				IsComplex:  false,
			},
			want:      nil,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.variable.GetInt32Array()
			if (err != nil) != tt.wantError {
				t.Errorf("GetInt32Array() error = %v, wantError %v", err, tt.wantError)
				return
			}
			if !tt.wantError {
				if len(got) != len(tt.want) {
					t.Errorf("GetInt32Array() length = %d, want %d", len(got), len(tt.want))
					return
				}
				for i := range got {
					if got[i] != tt.want[i] {
						t.Errorf("GetInt32Array()[%d] = %v, want %v", i, got[i], tt.want[i])
					}
				}
			}
		})
	}
}

// TestVariable_GetComplex128Array tests extracting complex data.
func TestVariable_GetComplex128Array(t *testing.T) {
	tests := []struct {
		name      string
		variable  *Variable
		want      []complex128
		wantError bool
	}{
		{
			name: "float64 complex",
			variable: &Variable{
				Name:       "test",
				Dimensions: []int{3},
				DataType:   Double,
				Data: &NumericArray{
					Real: []float64{1.0, 2.0, 3.0},
					Imag: []float64{4.0, 5.0, 6.0},
				},
				IsComplex: true,
			},
			want:      []complex128{1 + 4i, 2 + 5i, 3 + 6i},
			wantError: false,
		},
		{
			name: "float32 complex",
			variable: &Variable{
				Name:       "test",
				Dimensions: []int{3},
				DataType:   Single,
				Data: &NumericArray{
					Real: []float32{1.0, 2.0, 3.0},
					Imag: []float32{4.0, 5.0, 6.0},
				},
				IsComplex: true,
			},
			want:      []complex128{1 + 4i, 2 + 5i, 3 + 6i},
			wantError: false,
		},
		{
			name: "not complex error",
			variable: &Variable{
				Name:       "test",
				Dimensions: []int{3},
				DataType:   Double,
				Data:       []float64{1.0, 2.0, 3.0},
				IsComplex:  false,
			},
			want:      nil,
			wantError: true,
		},
		{
			name: "length mismatch error",
			variable: &Variable{
				Name:       "test",
				Dimensions: []int{3},
				DataType:   Double,
				Data: &NumericArray{
					Real: []float64{1.0, 2.0, 3.0},
					Imag: []float64{4.0, 5.0},
				},
				IsComplex: true,
			},
			want:      nil,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.variable.GetComplex128Array()
			if (err != nil) != tt.wantError {
				t.Errorf("GetComplex128Array() error = %v, wantError %v", err, tt.wantError)
				return
			}
			if !tt.wantError {
				if len(got) != len(tt.want) {
					t.Errorf("GetComplex128Array() length = %d, want %d", len(got), len(tt.want))
					return
				}
				for i := range got {
					if got[i] != tt.want[i] {
						t.Errorf("GetComplex128Array()[%d] = %v, want %v", i, got[i], tt.want[i])
					}
				}
			}
		})
	}
}

// TestVariable_GetScalar tests extracting scalar values.
func TestVariable_GetScalar(t *testing.T) {
	tests := []struct {
		name      string
		variable  *Variable
		want      interface{}
		wantError bool
	}{
		{
			name: "float64 scalar",
			variable: &Variable{
				Name:       "test",
				Dimensions: []int{1},
				DataType:   Double,
				Data:       []float64{42.0},
			},
			want:      42.0,
			wantError: false,
		},
		{
			name: "int32 scalar",
			variable: &Variable{
				Name:       "test",
				Dimensions: []int{1},
				DataType:   Int32,
				Data:       []int32{42},
			},
			want:      int32(42),
			wantError: false,
		},
		{
			name: "multi-element error",
			variable: &Variable{
				Name:       "test",
				Dimensions: []int{3},
				DataType:   Double,
				Data:       []float64{1.0, 2.0, 3.0},
			},
			want:      nil,
			wantError: true,
		},
		{
			name: "empty array error",
			variable: &Variable{
				Name:       "test",
				Dimensions: []int{1},
				DataType:   Double,
				Data:       []float64{},
			},
			want:      nil,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.variable.GetScalar()
			if (err != nil) != tt.wantError {
				t.Errorf("GetScalar() error = %v, wantError %v", err, tt.wantError)
				return
			}
			if !tt.wantError && got != tt.want {
				t.Errorf("GetScalar() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestDataType_String tests the String method for all DataType values.
func TestDataType_String(t *testing.T) {
	tests := []struct {
		dt   DataType
		want string
	}{
		{Double, "double"},
		{Single, "single"},
		{Int8, "int8"},
		{Uint8, "uint8"},
		{Int16, "int16"},
		{Uint16, "uint16"},
		{Int32, "int32"},
		{Uint32, "uint32"},
		{Int64, "int64"},
		{Uint64, "uint64"},
		{Char, "char"},
		{Struct, "struct"},
		{CellArray, "cell"},
		{Object, "object"},
		{Unknown, "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.dt.String(); got != tt.want {
				t.Errorf("DataType(%d).String() = %q, want %q", tt.dt, got, tt.want)
			}
		})
	}
}

// TestVariable_String tests the String representation of a Variable.
func TestVariable_String(t *testing.T) {
	tests := []struct {
		name     string
		variable *Variable
		want     string
	}{
		{
			name: "double 1D",
			variable: &Variable{
				Name:       "x",
				Dimensions: []int{3},
				DataType:   Double,
			},
			want: "x: double [3]",
		},
		{
			name: "int32 2D",
			variable: &Variable{
				Name:       "matrix",
				Dimensions: []int{2, 3},
				DataType:   Int32,
			},
			want: "matrix: int32 [2 3]",
		},
		{
			name: "char",
			variable: &Variable{
				Name:       "text",
				Dimensions: []int{1, 5},
				DataType:   Char,
			},
			want: "text: char [1 5]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.variable.String(); got != tt.want {
				t.Errorf("String() = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestVariable_GetAttribute tests attribute retrieval.
func TestVariable_GetAttribute(t *testing.T) {
	t.Run("nil map returns false", func(t *testing.T) {
		v := &Variable{Name: "test", Attributes: nil}
		val, ok := v.GetAttribute("key")
		if ok {
			t.Error("expected ok=false for nil attributes map")
		}
		if val != nil {
			t.Errorf("expected nil value, got %v", val)
		}
	})

	t.Run("existing key found", func(t *testing.T) {
		v := &Variable{
			Name: "test",
			Attributes: map[string]interface{}{
				"MATLAB_class": "double",
				"description":  "test variable",
			},
		}
		val, ok := v.GetAttribute("MATLAB_class")
		if !ok {
			t.Error("expected ok=true for existing key")
		}
		if val != "double" {
			t.Errorf("expected %q, got %v", "double", val)
		}
	})

	t.Run("missing key not found", func(t *testing.T) {
		v := &Variable{
			Name: "test",
			Attributes: map[string]interface{}{
				"existing": "value",
			},
		}
		val, ok := v.GetAttribute("nonexistent")
		if ok {
			t.Error("expected ok=false for missing key")
		}
		if val != nil {
			t.Errorf("expected nil value, got %v", val)
		}
	})
}

// TestVariable_GetFloat64Array_AllTypes tests all numeric type branches.
func TestVariable_GetFloat64Array_AllTypes(t *testing.T) {
	tests := []struct {
		name      string
		data      interface{}
		want      []float64
		wantError bool
	}{
		{
			name: "int8",
			data: []int8{-1, 0, 1, 127},
			want: []float64{-1, 0, 1, 127},
		},
		{
			name: "int16",
			data: []int16{-100, 0, 100, 32767},
			want: []float64{-100, 0, 100, 32767},
		},
		{
			name: "int64",
			data: []int64{-1000, 0, 1000},
			want: []float64{-1000, 0, 1000},
		},
		{
			name: "uint16",
			data: []uint16{0, 100, 65535},
			want: []float64{0, 100, 65535},
		},
		{
			name: "uint32",
			data: []uint32{0, 1000, 4294967295},
			want: []float64{0, 1000, 4294967295},
		},
		{
			name: "uint64",
			data: []uint64{0, 1000, 18446744073709551615},
			want: []float64{0, 1000, 18446744073709551615},
		},
		{
			name:      "unsupported type string",
			data:      "not a slice",
			wantError: true,
		},
		{
			name:      "unsupported type bool slice",
			data:      []bool{true, false},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := &Variable{
				Name:      "test",
				Data:      tt.data,
				IsComplex: false,
			}
			got, err := v.GetFloat64Array()
			if (err != nil) != tt.wantError {
				t.Fatalf("GetFloat64Array() error = %v, wantError %v", err, tt.wantError)
			}
			if !tt.wantError {
				if len(got) != len(tt.want) {
					t.Fatalf("length = %d, want %d", len(got), len(tt.want))
				}
				for i := range got {
					if got[i] != tt.want[i] {
						t.Errorf("[%d] = %v, want %v", i, got[i], tt.want[i])
					}
				}
			}
		})
	}
}

// TestVariable_GetInt32Array_AllBranches tests missing int16 and uint8 branches.
func TestVariable_GetInt32Array_AllBranches(t *testing.T) {
	tests := []struct {
		name      string
		data      interface{}
		want      []int32
		wantError bool
	}{
		{
			name: "int16 conversion",
			data: []int16{-100, 0, 100, 32767},
			want: []int32{-100, 0, 100, 32767},
		},
		{
			name: "uint8 conversion",
			data: []uint8{0, 128, 255},
			want: []int32{0, 128, 255},
		},
		{
			name:      "unsupported type uint32",
			data:      []uint32{1, 2, 3},
			wantError: true,
		},
		{
			name:      "unsupported type string",
			data:      "not a slice",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := &Variable{
				Name:      "test",
				Data:      tt.data,
				IsComplex: false,
			}
			got, err := v.GetInt32Array()
			if (err != nil) != tt.wantError {
				t.Fatalf("GetInt32Array() error = %v, wantError %v", err, tt.wantError)
			}
			if !tt.wantError {
				if len(got) != len(tt.want) {
					t.Fatalf("length = %d, want %d", len(got), len(tt.want))
				}
				for i := range got {
					if got[i] != tt.want[i] {
						t.Errorf("[%d] = %v, want %v", i, got[i], tt.want[i])
					}
				}
			}
		})
	}
}

// TestVariable_GetScalar_AllTypes tests all numeric type branches for GetScalar.
func TestVariable_GetScalar_AllTypes(t *testing.T) {
	tests := []struct {
		name      string
		data      interface{}
		want      interface{}
		wantError bool
	}{
		{
			name: "float32",
			data: []float32{3.14},
			want: float32(3.14),
		},
		{
			name: "int8",
			data: []int8{-42},
			want: int8(-42),
		},
		{
			name: "int16",
			data: []int16{-1000},
			want: int16(-1000),
		},
		{
			name: "int64",
			data: []int64{9999999},
			want: int64(9999999),
		},
		{
			name: "uint8",
			data: []uint8{255},
			want: uint8(255),
		},
		{
			name: "uint16",
			data: []uint16{65535},
			want: uint16(65535),
		},
		{
			name: "uint32",
			data: []uint32{4294967295},
			want: uint32(4294967295),
		},
		{
			name: "uint64",
			data: []uint64{18446744073709551615},
			want: uint64(18446744073709551615),
		},
		{
			name:      "unsupported type",
			data:      "not a numeric slice",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := &Variable{
				Name:       "test",
				Dimensions: []int{1},
				Data:       tt.data,
			}
			got, err := v.GetScalar()
			if (err != nil) != tt.wantError {
				t.Fatalf("GetScalar() error = %v, wantError %v", err, tt.wantError)
			}
			if !tt.wantError && got != tt.want {
				t.Errorf("GetScalar() = %v (%T), want %v (%T)", got, got, tt.want, tt.want)
			}
		})
	}
}

// TestVariable_GetComplex128Array_EdgeCases tests edge cases for complex array extraction.
func TestVariable_GetComplex128Array_EdgeCases(t *testing.T) {
	t.Run("data is not NumericArray", func(t *testing.T) {
		v := &Variable{
			Name:      "test",
			IsComplex: true,
			Data:      []float64{1.0, 2.0},
		}
		_, err := v.GetComplex128Array()
		if err == nil {
			t.Fatal("expected error for non-NumericArray data")
		}
		if !strings.Contains(err.Error(), "not *NumericArray") {
			t.Errorf("unexpected error message: %v", err)
		}
	})

	t.Run("float64 real with mismatched imag type", func(t *testing.T) {
		v := &Variable{
			Name:      "test",
			IsComplex: true,
			Data: &NumericArray{
				Real: []float64{1.0, 2.0},
				Imag: []float32{3.0, 4.0},
			},
		}
		_, err := v.GetComplex128Array()
		if err == nil {
			t.Fatal("expected error for mismatched real/imag types")
		}
		if !strings.Contains(err.Error(), "different types") {
			t.Errorf("unexpected error message: %v", err)
		}
	})

	t.Run("float32 real with mismatched imag type", func(t *testing.T) {
		v := &Variable{
			Name:      "test",
			IsComplex: true,
			Data: &NumericArray{
				Real: []float32{1.0, 2.0},
				Imag: []float64{3.0, 4.0},
			},
		}
		_, err := v.GetComplex128Array()
		if err == nil {
			t.Fatal("expected error for mismatched real/imag types")
		}
		if !strings.Contains(err.Error(), "different types") {
			t.Errorf("unexpected error message: %v", err)
		}
	})

	t.Run("float32 length mismatch", func(t *testing.T) {
		v := &Variable{
			Name:      "test",
			IsComplex: true,
			Data: &NumericArray{
				Real: []float32{1.0, 2.0, 3.0},
				Imag: []float32{4.0, 5.0},
			},
		}
		_, err := v.GetComplex128Array()
		if err == nil {
			t.Fatal("expected error for length mismatch")
		}
		if !strings.Contains(err.Error(), "different lengths") {
			t.Errorf("unexpected error message: %v", err)
		}
	})

	t.Run("unsupported complex type int32", func(t *testing.T) {
		v := &Variable{
			Name:      "test",
			IsComplex: true,
			Data: &NumericArray{
				Real: []int32{1, 2, 3},
				Imag: []int32{4, 5, 6},
			},
		}
		_, err := v.GetComplex128Array()
		if err == nil {
			t.Fatal("expected error for unsupported complex type")
		}
		if !strings.Contains(err.Error(), "unsupported complex data type") {
			t.Errorf("unexpected error message: %v", err)
		}
	})
}
