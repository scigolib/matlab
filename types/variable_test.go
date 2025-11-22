package types

import (
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
