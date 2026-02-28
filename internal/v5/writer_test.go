package v5

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
	"testing"

	"github.com/scigolib/matlab/types"
)

// TestNewWriter tests writer creation with both endianness.
// Note: "IM" = little-endian, "MI" = big-endian.
func TestNewWriter(t *testing.T) {
	tests := []struct {
		name        string
		description string
		endian      string
		wantErr     bool
	}{
		{
			name:        "little endian",
			description: "Test file little endian",
			endian:      "IM",
			wantErr:     false,
		},
		{
			name:        "big endian",
			description: "Test file big endian",
			endian:      "MI",
			wantErr:     false,
		},
		{
			name:        "invalid endian",
			description: "Test file",
			endian:      "XX",
			wantErr:     true,
		},
		{
			name:        "empty description",
			description: "",
			endian:      "IM",
			wantErr:     false,
		},
		{
			name:        "long description",
			description: string(make([]byte, 200)), // 200 bytes, should truncate to 116
			endian:      "IM",
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			writer, err := NewWriter(&buf, tt.description, tt.endian)

			if tt.wantErr {
				if err == nil {
					t.Errorf("NewWriter() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("NewWriter() unexpected error = %v", err)
			}

			if writer == nil {
				t.Fatal("NewWriter() returned nil writer")
			}

			// Verify header was written (128 bytes)
			if buf.Len() != 128 {
				t.Errorf("Header size = %d, want 128", buf.Len())
			}
		})
	}
}

// TestWriteHeader tests header writing in detail.
// Note: "IM" = little-endian, "MI" = big-endian.
//
//nolint:gocognit // Table-driven test with comprehensive header validation
func TestWriteHeader(t *testing.T) {
	tests := []struct {
		name        string
		description string
		endian      string
		wantVersion uint16
	}{
		{
			name:        "little endian header",
			description: "Test MAT-file",
			endian:      "IM",
			wantVersion: 0x0100,
		},
		{
			name:        "big endian header",
			description: "Test MAT-file",
			endian:      "MI",
			wantVersion: 0x0100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := NewWriter(&buf, tt.description, tt.endian)
			if err != nil {
				t.Fatalf("NewWriter() error = %v", err)
			}

			header := buf.Bytes()

			// Verify header size
			if len(header) != 128 {
				t.Errorf("Header size = %d, want 128", len(header))
			}

			// Verify description (bytes 0-115)
			desc := string(bytes.TrimRight(header[0:116], "\x00"))
			if desc != tt.description {
				t.Errorf("Description = %q, want %q", desc, tt.description)
			}

			// Verify subsystem data offset is zeros (bytes 116-123)
			subsys := header[116:124]
			if !bytes.Equal(subsys, make([]byte, 8)) {
				t.Errorf("Subsystem data offset not zero: %v", subsys)
			}

			// Verify endian indicator (bytes 126-127)
			endian := string(header[126:128])
			if endian != tt.endian {
				t.Errorf("Endian = %q, want %q", endian, tt.endian)
			}

			// Verify version (bytes 124-125)
			// "IM" = little-endian, "MI" = big-endian
			var order binary.ByteOrder
			if tt.endian == "IM" {
				order = binary.LittleEndian
			} else {
				order = binary.BigEndian
			}
			version := order.Uint16(header[124:126])
			if version != tt.wantVersion {
				t.Errorf("Version = 0x%04x, want 0x%04x", version, tt.wantVersion)
			}
		})
	}
}

// TestEncodeFloat64Array tests float64 array encoding.
func TestEncodeFloat64Array(t *testing.T) {
	var buf bytes.Buffer
	w, err := NewWriter(&buf, "Test", "IM") // "IM" = little-endian
	if err != nil {
		t.Fatalf("NewWriter() error = %v", err)
	}

	tests := []struct {
		name string
		data []float64
	}{
		{
			name: "simple array",
			data: []float64{1.0, 2.0, 3.0},
		},
		{
			name: "single element",
			data: []float64{42.0},
		},
		{
			name: "empty array",
			data: []float64{},
		},
		{
			name: "negative values",
			data: []float64{-1.5, -2.7, -3.9},
		},
		{
			name: "special values",
			data: []float64{math.NaN(), math.Inf(1), math.Inf(-1), 0.0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoded := w.encodeFloat64Array(tt.data)

			// Verify size
			expectedSize := len(tt.data) * 8
			if len(encoded) != expectedSize {
				t.Errorf("Encoded size = %d, want %d", len(encoded), expectedSize)
			}

			// Decode and verify values
			for i := 0; i < len(tt.data); i++ {
				bits := binary.LittleEndian.Uint64(encoded[i*8 : (i+1)*8])
				val := math.Float64frombits(bits)

				expected := tt.data[i]
				// Special handling for NaN
				if math.IsNaN(expected) {
					if !math.IsNaN(val) {
						t.Errorf("Value[%d] = %v, want NaN", i, val)
					}
				} else if val != expected {
					t.Errorf("Value[%d] = %v, want %v", i, val, expected)
				}
			}
		})
	}
}

// TestEncodeFloat32Array tests float32 array encoding.
func TestEncodeFloat32Array(t *testing.T) {
	var buf bytes.Buffer
	w, err := NewWriter(&buf, "Test", "IM") // "IM" = little-endian
	if err != nil {
		t.Fatalf("NewWriter() error = %v", err)
	}

	data := []float32{1.5, 2.5, 3.5}
	encoded := w.encodeFloat32Array(data)

	// Verify size
	if len(encoded) != 12 {
		t.Errorf("Encoded size = %d, want 12", len(encoded))
	}

	// Decode and verify
	for i := 0; i < 3; i++ {
		bits := binary.LittleEndian.Uint32(encoded[i*4 : (i+1)*4])
		val := math.Float32frombits(bits)
		if val != data[i] {
			t.Errorf("Value[%d] = %v, want %v", i, val, data[i])
		}
	}
}

// TestEncodeInt32Array tests int32 array encoding.
func TestEncodeInt32Array(t *testing.T) {
	var buf bytes.Buffer
	w, err := NewWriter(&buf, "Test", "IM") // "IM" = little-endian
	if err != nil {
		t.Fatalf("NewWriter() error = %v", err)
	}

	tests := []struct {
		name string
		data []int32
	}{
		{
			name: "positive values",
			data: []int32{1, 2, 3},
		},
		{
			name: "negative values",
			data: []int32{-1, -2, -3},
		},
		{
			name: "mixed values",
			data: []int32{-100, 0, 100},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoded := w.encodeInt32Array(tt.data)

			// Verify size
			expectedSize := len(tt.data) * 4
			if len(encoded) != expectedSize {
				t.Errorf("Encoded size = %d, want %d", len(encoded), expectedSize)
			}

			// Decode and verify
			for i := 0; i < len(tt.data); i++ {
				val := int32(binary.LittleEndian.Uint32(encoded[i*4 : (i+1)*4]))
				if val != tt.data[i] {
					t.Errorf("Value[%d] = %v, want %v", i, val, tt.data[i])
				}
			}
		})
	}
}

// TestWrapInTag tests data element tag wrapping.
//
//nolint:gocognit,nestif // Table-driven test with format-specific validation
func TestWrapInTag(t *testing.T) {
	var buf bytes.Buffer
	w, err := NewWriter(&buf, "Test", "IM") // "IM" = little-endian
	if err != nil {
		t.Fatalf("NewWriter() error = %v", err)
	}

	tests := []struct {
		name        string
		dataType    uint32
		data        []byte
		wantSize    int // Total size including tag and padding
		wantSmall   bool
		wantPadding int
	}{
		{
			name:        "1 byte - uses regular format",
			dataType:    miUINT8,
			data:        []byte{42},
			wantSize:    16, // 8 (tag) + 1 (data) + 7 (padding)
			wantSmall:   false,
			wantPadding: 7,
		},
		{
			name:        "4 bytes - uses regular format",
			dataType:    miINT32,
			data:        []byte{1, 2, 3, 4},
			wantSize:    16, // 8 (tag) + 4 (data) + 4 (padding)
			wantSmall:   false,
			wantPadding: 4,
		},
		{
			name:        "regular format - 5 bytes",
			dataType:    miUINT8,
			data:        []byte{1, 2, 3, 4, 5},
			wantSize:    16, // 8 (tag) + 5 (data) + 3 (padding)
			wantSmall:   false,
			wantPadding: 3,
		},
		{
			name:        "regular format - 8 bytes aligned",
			dataType:    miDOUBLE,
			data:        make([]byte, 8),
			wantSize:    16, // 8 (tag) + 8 (data) + 0 (padding)
			wantSmall:   false,
			wantPadding: 0,
		},
		{
			name:        "regular format - 10 bytes",
			dataType:    miINT8,
			data:        make([]byte, 10),
			wantSize:    24, // 8 (tag) + 10 (data) + 6 (padding)
			wantSmall:   false,
			wantPadding: 6,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wrapped := w.wrapInTag(tt.dataType, tt.data)

			// Verify total size
			if len(wrapped) != tt.wantSize {
				t.Errorf("Total size = %d, want %d", len(wrapped), tt.wantSize)
			}

			if tt.wantSmall {
				// Small format: verify tag structure
				dtype := binary.LittleEndian.Uint16(wrapped[0:2])
				size := binary.LittleEndian.Uint16(wrapped[2:4])
				if uint32(dtype) != tt.dataType {
					t.Errorf("DataType = %d, want %d", dtype, tt.dataType)
				}
				if int(size) != len(tt.data) {
					t.Errorf("Size = %d, want %d", size, len(tt.data))
				}
				// Verify data is in bytes 4-7
				if !bytes.Equal(wrapped[4:4+len(tt.data)], tt.data) {
					t.Errorf("Data mismatch in small format")
				}
			} else {
				// Regular format: verify tag structure
				dtype := binary.LittleEndian.Uint32(wrapped[0:4])
				size := binary.LittleEndian.Uint32(wrapped[4:8])
				if dtype != tt.dataType {
					t.Errorf("DataType = %d, want %d", dtype, tt.dataType)
				}
				if int(size) != len(tt.data) {
					t.Errorf("Size = %d, want %d", size, len(tt.data))
				}
				// Verify data
				if !bytes.Equal(wrapped[8:8+len(tt.data)], tt.data) {
					t.Errorf("Data mismatch in regular format")
				}
				// Verify padding is zeros
				if tt.wantPadding > 0 {
					padding := wrapped[8+len(tt.data):]
					expectedPadding := make([]byte, tt.wantPadding)
					if !bytes.Equal(padding, expectedPadding) {
						t.Errorf("Padding not zero: %v", padding)
					}
				}
			}
		})
	}
}

// TestWriteVariable tests writing complete variables.
//
//nolint:gocognit // Table-driven test with comprehensive variable validation
func TestWriteVariable(t *testing.T) {
	tests := []struct {
		name     string
		variable *types.Variable
		wantErr  bool
		errMsg   string
	}{
		{
			name: "simple double array",
			variable: &types.Variable{
				Name:       "A",
				Dimensions: []int{3},
				DataType:   types.Double,
				Data:       []float64{1.0, 2.0, 3.0},
			},
			wantErr: false,
		},
		{
			name: "2D matrix",
			variable: &types.Variable{
				Name:       "B",
				Dimensions: []int{2, 3},
				DataType:   types.Double,
				Data:       []float64{1, 2, 3, 4, 5, 6},
			},
			wantErr: false,
		},
		{
			name: "int32 array",
			variable: &types.Variable{
				Name:       "C",
				Dimensions: []int{4},
				DataType:   types.Int32,
				Data:       []int32{-1, 0, 1, 2},
			},
			wantErr: false,
		},
		{
			name: "uint8 array",
			variable: &types.Variable{
				Name:       "D",
				Dimensions: []int{5},
				DataType:   types.Uint8,
				Data:       []byte{10, 20, 30, 40, 50},
			},
			wantErr: false,
		},
		{
			name: "complex double array",
			variable: &types.Variable{
				Name:       "E",
				Dimensions: []int{2},
				DataType:   types.Double,
				IsComplex:  true,
				Data: &types.NumericArray{
					Real: []float64{1.0, 3.0},
					Imag: []float64{2.0, 4.0},
				},
			},
			wantErr: false,
		},
		{
			name: "empty name",
			variable: &types.Variable{
				Name:       "",
				Dimensions: []int{1},
				DataType:   types.Double,
				Data:       []float64{1.0},
			},
			wantErr: true,
			errMsg:  "variable name is required",
		},
		{
			name: "name too long",
			variable: &types.Variable{
				Name:       string(make([]byte, 64)), // 64 chars, max is 63
				Dimensions: []int{1},
				DataType:   types.Double,
				Data:       []float64{1.0},
			},
			wantErr: true,
			errMsg:  "variable name too long",
		},
		{
			name: "no dimensions",
			variable: &types.Variable{
				Name:       "F",
				Dimensions: []int{},
				DataType:   types.Double,
				Data:       []float64{1.0},
			},
			wantErr: true,
			errMsg:  "dimensions are required",
		},
		{
			name: "invalid dimension",
			variable: &types.Variable{
				Name:       "G",
				Dimensions: []int{3, 0, 2}, // Zero dimension
				DataType:   types.Double,
				Data:       []float64{1.0},
			},
			wantErr: true,
			errMsg:  "dimension[1] must be positive",
		},
		{
			name: "nil data",
			variable: &types.Variable{
				Name:       "H",
				Dimensions: []int{1},
				DataType:   types.Double,
				Data:       nil,
			},
			wantErr: true,
			errMsg:  "data is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			w, err := NewWriter(&buf, "Test", "IM") // "IM" = little-endian
			if err != nil {
				t.Fatalf("NewWriter() error = %v", err)
			}

			err = w.WriteVariable(tt.variable)

			if tt.wantErr {
				if err == nil {
					t.Errorf("WriteVariable() expected error containing %q, got nil", tt.errMsg)
					return
				}
				if tt.errMsg != "" && !contains(err.Error(), tt.errMsg) {
					t.Errorf("WriteVariable() error = %q, want error containing %q", err.Error(), tt.errMsg)
				}
				return
			}

			if err != nil {
				t.Errorf("WriteVariable() unexpected error = %v", err)
			}

			// Verify something was written (header + data)
			if buf.Len() <= 128 {
				t.Errorf("Buffer size = %d, expected > 128 (header + variable data)", buf.Len())
			}
		})
	}
}

// TestBothEndianness tests that both endianness produce valid output.
func TestBothEndianness(t *testing.T) {
	endians := []string{"MI", "IM"}

	for _, endian := range endians {
		t.Run(endian, func(t *testing.T) {
			var buf bytes.Buffer
			w, err := NewWriter(&buf, "Test", endian)
			if err != nil {
				t.Fatalf("NewWriter() error = %v", err)
			}

			v := &types.Variable{
				Name:       "test",
				Dimensions: []int{3},
				DataType:   types.Double,
				Data:       []float64{1.0, 2.0, 3.0},
			}

			err = w.WriteVariable(v)
			if err != nil {
				t.Fatalf("WriteVariable() error = %v", err)
			}

			// Verify header endian indicator
			header := buf.Bytes()[:128]
			endianIndicator := string(header[126:128])
			if endianIndicator != endian {
				t.Errorf("Endian indicator = %q, want %q", endianIndicator, endian)
			}
		})
	}
}

// contains checks if s contains substr.
func contains(s, substr string) bool {
	return bytes.Contains([]byte(s), []byte(substr))
}

// TestEncodeInt8Array tests int8 array encoding.
func TestEncodeInt8Array(t *testing.T) {
	var buf bytes.Buffer
	w, err := NewWriter(&buf, "Test", "IM")
	if err != nil {
		t.Fatalf("NewWriter() error = %v", err)
	}

	data := []int8{-128, 0, 127}
	encoded := w.encodeInt8Array(data)

	// Verify size: 1 byte per element
	if len(encoded) != 3 {
		t.Errorf("Encoded size = %d, want 3", len(encoded))
	}

	// Decode and verify values
	for i, want := range data {
		got := int8(encoded[i])
		if got != want {
			t.Errorf("Value[%d] = %d, want %d", i, got, want)
		}
	}
}

// TestEncodeInt16Array tests int16 array encoding.
func TestEncodeInt16Array(t *testing.T) {
	var buf bytes.Buffer
	w, err := NewWriter(&buf, "Test", "IM")
	if err != nil {
		t.Fatalf("NewWriter() error = %v", err)
	}

	data := []int16{-32768, 0, 32767}
	encoded := w.encodeInt16Array(data)

	// Verify size: 2 bytes per element
	if len(encoded) != 6 {
		t.Errorf("Encoded size = %d, want 6", len(encoded))
	}

	// Decode and verify values
	for i, want := range data {
		got := int16(binary.LittleEndian.Uint16(encoded[i*2 : (i+1)*2]))
		if got != want {
			t.Errorf("Value[%d] = %d, want %d", i, got, want)
		}
	}
}

// TestEncodeUint16Array tests uint16 array encoding.
func TestEncodeUint16Array(t *testing.T) {
	var buf bytes.Buffer
	w, err := NewWriter(&buf, "Test", "IM")
	if err != nil {
		t.Fatalf("NewWriter() error = %v", err)
	}

	data := []uint16{0, 1000, 65535}
	encoded := w.encodeUint16Array(data)

	// Verify size: 2 bytes per element
	if len(encoded) != 6 {
		t.Errorf("Encoded size = %d, want 6", len(encoded))
	}

	// Decode and verify values
	for i, want := range data {
		got := binary.LittleEndian.Uint16(encoded[i*2 : (i+1)*2])
		if got != want {
			t.Errorf("Value[%d] = %d, want %d", i, got, want)
		}
	}
}

// TestEncodeUint32Array tests uint32 array encoding.
func TestEncodeUint32Array(t *testing.T) {
	var buf bytes.Buffer
	w, err := NewWriter(&buf, "Test", "IM")
	if err != nil {
		t.Fatalf("NewWriter() error = %v", err)
	}

	data := []uint32{0, 100000, math.MaxUint32}
	encoded := w.encodeUint32Array(data)

	// Verify size: 4 bytes per element
	if len(encoded) != 12 {
		t.Errorf("Encoded size = %d, want 12", len(encoded))
	}

	// Decode and verify values
	for i, want := range data {
		got := binary.LittleEndian.Uint32(encoded[i*4 : (i+1)*4])
		if got != want {
			t.Errorf("Value[%d] = %d, want %d", i, got, want)
		}
	}
}

// TestEncodeInt64Array tests int64 array encoding.
func TestEncodeInt64Array(t *testing.T) {
	var buf bytes.Buffer
	w, err := NewWriter(&buf, "Test", "IM")
	if err != nil {
		t.Fatalf("NewWriter() error = %v", err)
	}

	data := []int64{math.MinInt64, 0, math.MaxInt64}
	encoded := w.encodeInt64Array(data)

	// Verify size: 8 bytes per element
	if len(encoded) != 24 {
		t.Errorf("Encoded size = %d, want 24", len(encoded))
	}

	// Decode and verify values
	for i, want := range data {
		got := int64(binary.LittleEndian.Uint64(encoded[i*8 : (i+1)*8]))
		if got != want {
			t.Errorf("Value[%d] = %d, want %d", i, got, want)
		}
	}
}

// TestEncodeUint64Array tests uint64 array encoding.
func TestEncodeUint64Array(t *testing.T) {
	var buf bytes.Buffer
	w, err := NewWriter(&buf, "Test", "IM")
	if err != nil {
		t.Fatalf("NewWriter() error = %v", err)
	}

	data := []uint64{0, 1000000, math.MaxUint64}
	encoded := w.encodeUint64Array(data)

	// Verify size: 8 bytes per element
	if len(encoded) != 24 {
		t.Errorf("Encoded size = %d, want 24", len(encoded))
	}

	// Decode and verify values
	for i, want := range data {
		got := binary.LittleEndian.Uint64(encoded[i*8 : (i+1)*8])
		if got != want {
			t.Errorf("Value[%d] = %d, want %d", i, got, want)
		}
	}
}

// TestDataTypeToClass_AllTypes tests all dataTypeToClass mappings including unknown fallback.
func TestDataTypeToClass_AllTypes(t *testing.T) {
	var buf bytes.Buffer
	w, err := NewWriter(&buf, "Test", "IM")
	if err != nil {
		t.Fatalf("NewWriter() error = %v", err)
	}

	tests := []struct {
		name      string
		dataType  types.DataType
		wantClass uint32
	}{
		{"Double", types.Double, mxDOUBLE_CLASS},
		{"Single", types.Single, mxSINGLE_CLASS},
		{"Int8", types.Int8, mxINT8_CLASS},
		{"Uint8", types.Uint8, mxUINT8_CLASS},
		{"Int16", types.Int16, mxINT16_CLASS},
		{"Uint16", types.Uint16, mxUINT16_CLASS},
		{"Int32", types.Int32, mxINT32_CLASS},
		{"Uint32", types.Uint32, mxUINT32_CLASS},
		{"Int64", types.Int64, mxINT64_CLASS},
		{"Uint64", types.Uint64, mxUINT64_CLASS},
		{"Unknown falls back to Double", types.Unknown, mxDOUBLE_CLASS},
		{"Char falls back to Double", types.Char, mxDOUBLE_CLASS},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := w.dataTypeToClass(tt.dataType)
			if got != tt.wantClass {
				t.Errorf("dataTypeToClass(%v) = %d, want %d", tt.dataType, got, tt.wantClass)
			}
		})
	}
}

// TestEncodeData_AllTypes tests encodeData for all 10 numeric types.
func TestEncodeData_AllTypes(t *testing.T) {
	var buf bytes.Buffer
	w, err := NewWriter(&buf, "Test", "IM")
	if err != nil {
		t.Fatalf("NewWriter() error = %v", err)
	}

	tests := []struct {
		name     string
		variable *types.Variable
	}{
		{
			name: "Double",
			variable: &types.Variable{
				Name: "a", Dimensions: []int{2}, DataType: types.Double,
				Data: []float64{1.0, 2.0},
			},
		},
		{
			name: "Single",
			variable: &types.Variable{
				Name: "b", Dimensions: []int{2}, DataType: types.Single,
				Data: []float32{1.0, 2.0},
			},
		},
		{
			name: "Int8",
			variable: &types.Variable{
				Name: "c", Dimensions: []int{2}, DataType: types.Int8,
				Data: []int8{-1, 1},
			},
		},
		{
			name: "Uint8",
			variable: &types.Variable{
				Name: "d", Dimensions: []int{2}, DataType: types.Uint8,
				Data: []byte{0, 255},
			},
		},
		{
			name: "Int16",
			variable: &types.Variable{
				Name: "e", Dimensions: []int{2}, DataType: types.Int16,
				Data: []int16{-100, 100},
			},
		},
		{
			name: "Uint16",
			variable: &types.Variable{
				Name: "f", Dimensions: []int{2}, DataType: types.Uint16,
				Data: []uint16{0, 65535},
			},
		},
		{
			name: "Int32",
			variable: &types.Variable{
				Name: "g", Dimensions: []int{2}, DataType: types.Int32,
				Data: []int32{-100, 100},
			},
		},
		{
			name: "Uint32",
			variable: &types.Variable{
				Name: "h", Dimensions: []int{2}, DataType: types.Uint32,
				Data: []uint32{0, math.MaxUint32},
			},
		},
		{
			name: "Int64",
			variable: &types.Variable{
				Name: "i", Dimensions: []int{2}, DataType: types.Int64,
				Data: []int64{math.MinInt64, math.MaxInt64},
			},
		},
		{
			name: "Uint64",
			variable: &types.Variable{
				Name: "j", Dimensions: []int{2}, DataType: types.Uint64,
				Data: []uint64{0, math.MaxUint64},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoded, err := w.encodeData(tt.variable, false)
			if err != nil {
				t.Errorf("encodeData(%s) unexpected error: %v", tt.name, err)
			}
			if len(encoded) == 0 {
				t.Errorf("encodeData(%s) returned empty result", tt.name)
			}
		})
	}
}

// TestEncodeData_ComplexMissingParts tests error when complex data has nil Real or Imag.
func TestEncodeData_ComplexMissingParts(t *testing.T) {
	var buf bytes.Buffer
	w, err := NewWriter(&buf, "Test", "IM")
	if err != nil {
		t.Fatalf("NewWriter() error = %v", err)
	}

	tests := []struct {
		name      string
		variable  *types.Variable
		imaginary bool
		errMsg    string
	}{
		{
			name: "missing real part",
			variable: &types.Variable{
				Name: "x", Dimensions: []int{2}, DataType: types.Double,
				IsComplex: true,
				Data: &types.NumericArray{
					Real: nil,
					Imag: []float64{1.0, 2.0},
				},
			},
			imaginary: false,
			errMsg:    "missing real part",
		},
		{
			name: "missing imag part",
			variable: &types.Variable{
				Name: "y", Dimensions: []int{2}, DataType: types.Double,
				IsComplex: true,
				Data: &types.NumericArray{
					Real: []float64{1.0, 2.0},
					Imag: nil,
				},
			},
			imaginary: true,
			errMsg:    "missing imaginary part",
		},
		{
			name: "complex with wrong Data type",
			variable: &types.Variable{
				Name: "z", Dimensions: []int{2}, DataType: types.Double,
				IsComplex: true,
				Data:      []float64{1.0, 2.0}, // Not *NumericArray
			},
			imaginary: false,
			errMsg:    "must have *types.NumericArray",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := w.encodeData(tt.variable, tt.imaginary)
			if err == nil {
				t.Errorf("encodeData() expected error containing %q, got nil", tt.errMsg)
				return
			}
			if !contains(err.Error(), tt.errMsg) {
				t.Errorf("encodeData() error = %q, want containing %q", err.Error(), tt.errMsg)
			}
		})
	}
}

// TestEncodeData_WrongSliceType tests error when data slice type does not match DataType.
func TestEncodeData_WrongSliceType(t *testing.T) {
	var buf bytes.Buffer
	w, err := NewWriter(&buf, "Test", "IM")
	if err != nil {
		t.Fatalf("NewWriter() error = %v", err)
	}

	tests := []struct {
		name     string
		dataType types.DataType
		data     interface{}
		errMsg   string
	}{
		{"Double with int32 data", types.Double, []int32{1, 2}, "expected []float64"},
		{"Single with float64 data", types.Single, []float64{1.0}, "expected []float32"},
		{"Int8 with byte data", types.Int8, []byte{1, 2}, "expected []int8"},
		{"Uint8 with int8 data", types.Uint8, []int8{1, 2}, "expected []byte"},
		{"Int16 with int32 data", types.Int16, []int32{1, 2}, "expected []int16"},
		{"Uint16 with uint32 data", types.Uint16, []uint32{1, 2}, "expected []uint16"},
		{"Int32 with int64 data", types.Int32, []int64{1, 2}, "expected []int32"},
		{"Uint32 with uint64 data", types.Uint32, []uint64{1, 2}, "expected []uint32"},
		{"Int64 with int32 data", types.Int64, []int32{1, 2}, "expected []int64"},
		{"Uint64 with uint32 data", types.Uint64, []uint32{1, 2}, "expected []uint64"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := &types.Variable{
				Name: "x", Dimensions: []int{2}, DataType: tt.dataType,
				Data: tt.data,
			}
			_, err := w.encodeData(v, false)
			if err == nil {
				t.Errorf("encodeData() expected error containing %q, got nil", tt.errMsg)
				return
			}
			if !contains(err.Error(), tt.errMsg) {
				t.Errorf("encodeData() error = %q, want containing %q", err.Error(), tt.errMsg)
			}
		})
	}
}

// TestEncodeData_UnsupportedType tests error for unsupported DataType.
func TestEncodeData_UnsupportedType(t *testing.T) {
	var buf bytes.Buffer
	w, err := NewWriter(&buf, "Test", "IM")
	if err != nil {
		t.Fatalf("NewWriter() error = %v", err)
	}

	v := &types.Variable{
		Name:       "x",
		Dimensions: []int{1},
		DataType:   types.Unknown,
		Data:       []float64{1.0},
	}

	_, err = w.encodeData(v, false)
	if err == nil {
		t.Error("encodeData() expected error for unsupported type, got nil")
		return
	}
	if !contains(err.Error(), "unsupported data type") {
		t.Errorf("encodeData() error = %q, want containing %q", err.Error(), "unsupported data type")
	}
}

// TestWriteVariable_AllNumericTypes tests writing variables of all 10 numeric types.
func TestWriteVariable_AllNumericTypes(t *testing.T) {
	tests := []struct {
		name     string
		variable *types.Variable
	}{
		{
			name: "Double",
			variable: &types.Variable{
				Name: "a", Dimensions: []int{3}, DataType: types.Double,
				Data: []float64{1.0, 2.0, 3.0},
			},
		},
		{
			name: "Single",
			variable: &types.Variable{
				Name: "b", Dimensions: []int{3}, DataType: types.Single,
				Data: []float32{1.0, 2.0, 3.0},
			},
		},
		{
			name: "Int8",
			variable: &types.Variable{
				Name: "c", Dimensions: []int{3}, DataType: types.Int8,
				Data: []int8{-128, 0, 127},
			},
		},
		{
			name: "Uint8",
			variable: &types.Variable{
				Name: "d", Dimensions: []int{3}, DataType: types.Uint8,
				Data: []byte{0, 128, 255},
			},
		},
		{
			name: "Int16",
			variable: &types.Variable{
				Name: "e", Dimensions: []int{3}, DataType: types.Int16,
				Data: []int16{-32768, 0, 32767},
			},
		},
		{
			name: "Uint16",
			variable: &types.Variable{
				Name: "f", Dimensions: []int{3}, DataType: types.Uint16,
				Data: []uint16{0, 1000, 65535},
			},
		},
		{
			name: "Int32",
			variable: &types.Variable{
				Name: "g", Dimensions: []int{3}, DataType: types.Int32,
				Data: []int32{-2147483648, 0, 2147483647},
			},
		},
		{
			name: "Uint32",
			variable: &types.Variable{
				Name: "h", Dimensions: []int{3}, DataType: types.Uint32,
				Data: []uint32{0, 100000, math.MaxUint32},
			},
		},
		{
			name: "Int64",
			variable: &types.Variable{
				Name: "i", Dimensions: []int{3}, DataType: types.Int64,
				Data: []int64{math.MinInt64, 0, math.MaxInt64},
			},
		},
		{
			name: "Uint64",
			variable: &types.Variable{
				Name: "j", Dimensions: []int{3}, DataType: types.Uint64,
				Data: []uint64{0, 1000000, math.MaxUint64},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			w, err := NewWriter(&buf, "Test", "IM")
			if err != nil {
				t.Fatalf("NewWriter() error = %v", err)
			}

			err = w.WriteVariable(tt.variable)
			if err != nil {
				t.Fatalf("WriteVariable(%s) unexpected error: %v", tt.name, err)
			}

			// Buffer should be larger than 128 bytes (header) since variable data was written
			if buf.Len() <= 128 {
				t.Errorf("Buffer size = %d, expected > 128 (header + variable data)", buf.Len())
			}
		})
	}
}

// TestWriteMatrix_BigEndian writes a variable with big-endian and verifies the
// output is non-empty and different from little-endian output.
func TestWriteMatrix_BigEndian(t *testing.T) {
	v := &types.Variable{
		Name:       "bevar",
		Dimensions: []int{1, 3},
		DataType:   types.Double,
		Data:       []float64{1.0, 2.0, 3.0},
	}

	// Write with big-endian
	var beBuf bytes.Buffer
	beWriter, err := NewWriter(&beBuf, "Test", "MI")
	if err != nil {
		t.Fatalf("NewWriter(MI) error: %v", err)
	}
	if err := beWriter.WriteVariable(v); err != nil {
		t.Fatalf("WriteVariable(MI) error: %v", err)
	}

	// Write with little-endian
	var leBuf bytes.Buffer
	leWriter, err := NewWriter(&leBuf, "Test", "IM")
	if err != nil {
		t.Fatalf("NewWriter(IM) error: %v", err)
	}
	if err := leWriter.WriteVariable(v); err != nil {
		t.Fatalf("WriteVariable(IM) error: %v", err)
	}

	// Both should produce non-empty output beyond the header
	if beBuf.Len() <= 128 {
		t.Errorf("Big-endian buffer size = %d, expected > 128", beBuf.Len())
	}
	if leBuf.Len() <= 128 {
		t.Errorf("Little-endian buffer size = %d, expected > 128", leBuf.Len())
	}

	// The outputs should differ (different byte ordering)
	if bytes.Equal(beBuf.Bytes(), leBuf.Bytes()) {
		t.Error("Big-endian and little-endian outputs should differ")
	}
}

// TestWriteHeader_BigEndian_Bytes verifies that big-endian header has "MI" at bytes 126-127
// and that version is encoded correctly in big-endian.
func TestWriteHeader_BigEndian_Bytes(t *testing.T) {
	var buf bytes.Buffer
	_, err := NewWriter(&buf, "Big Endian Test", "MI")
	if err != nil {
		t.Fatalf("NewWriter(MI) error: %v", err)
	}

	header := buf.Bytes()
	if len(header) != 128 {
		t.Fatalf("Header size = %d, want 128", len(header))
	}

	// Verify endian indicator is "MI"
	endian := string(header[126:128])
	if endian != "MI" {
		t.Errorf("Endian indicator = %q, want %q", endian, "MI")
	}

	// Verify version in big-endian: 0x0100 stored as [0x01, 0x00]
	version := binary.BigEndian.Uint16(header[124:126])
	if version != 0x0100 {
		t.Errorf("Version = 0x%04x, want 0x0100", version)
	}
}

// TestEncodeData_ComplexSingle tests encoding complex single (float32) data.
func TestEncodeData_ComplexSingle(t *testing.T) {
	var buf bytes.Buffer
	w, err := NewWriter(&buf, "Test", "IM")
	if err != nil {
		t.Fatalf("NewWriter() error: %v", err)
	}

	v := &types.Variable{
		Name:       "csgl",
		Dimensions: []int{1, 2},
		DataType:   types.Single,
		IsComplex:  true,
		Data: &types.NumericArray{
			Real: []float32{1.0, 2.0},
			Imag: []float32{3.0, 4.0},
		},
	}

	// Encode real part
	realEncoded, err := w.encodeData(v, false)
	if err != nil {
		t.Fatalf("encodeData(real) error: %v", err)
	}
	if len(realEncoded) == 0 {
		t.Error("encodeData(real) returned empty result")
	}

	// Encode imaginary part
	imagEncoded, err := w.encodeData(v, true)
	if err != nil {
		t.Fatalf("encodeData(imag) error: %v", err)
	}
	if len(imagEncoded) == 0 {
		t.Error("encodeData(imag) returned empty result")
	}
}

// TestEncodeMatrixContent_ComplexSingle tests full matrix content encoding for complex single.
func TestEncodeMatrixContent_ComplexSingle(t *testing.T) {
	var buf bytes.Buffer
	w, err := NewWriter(&buf, "Test", "IM")
	if err != nil {
		t.Fatalf("NewWriter() error: %v", err)
	}

	v := &types.Variable{
		Name:       "csgl",
		Dimensions: []int{1, 2},
		DataType:   types.Single,
		IsComplex:  true,
		Data: &types.NumericArray{
			Real: []float32{1.0, 2.0},
			Imag: []float32{3.0, 4.0},
		},
	}

	content, err := w.encodeMatrixContent(v)
	if err != nil {
		t.Fatalf("encodeMatrixContent() error: %v", err)
	}
	if len(content) == 0 {
		t.Error("encodeMatrixContent() returned empty result")
	}
}

// TestWriteVariable_NilVariable tests that WriteVariable with nil returns an error.
func TestWriteVariable_NilVariable(t *testing.T) {
	var buf bytes.Buffer
	w, err := NewWriter(&buf, "Test", "IM")
	if err != nil {
		t.Fatalf("NewWriter() error: %v", err)
	}

	// Passing nil should panic or error. Since validateVariable dereferences v,
	// a nil variable will cause a panic. We test that it doesn't silently succeed.
	defer func() {
		if r := recover(); r == nil {
			// If no panic, check that error was returned
		}
	}()

	// This will likely panic since v.Name dereferences nil
	// If it doesn't panic, it should return an error
	err = w.WriteVariable(nil)
	if err == nil {
		t.Error("WriteVariable(nil) expected error or panic, got nil")
	}
}

// TestEncodeArrayFlags_Sparse tests that the sparse flag bit is set correctly.
func TestEncodeArrayFlags_Sparse(t *testing.T) {
	var buf bytes.Buffer
	w, err := NewWriter(&buf, "Test", "IM")
	if err != nil {
		t.Fatalf("NewWriter() error: %v", err)
	}

	v := &types.Variable{
		Name:       "sp",
		Dimensions: []int{1, 3},
		DataType:   types.Double,
		Data:       []float64{1.0, 2.0, 3.0},
		IsSparse:   true,
	}

	flags := w.encodeArrayFlags(v)

	// flags should be: 8-byte tag + 8-byte data = 16 bytes
	if len(flags) != 16 {
		t.Fatalf("encodeArrayFlags() len = %d, want 16", len(flags))
	}

	// Read the flags data (after the 8-byte tag)
	flagsWord := binary.LittleEndian.Uint32(flags[8:12])
	if flagsWord&0x0400 == 0 {
		t.Error("Sparse flag bit (0x0400) not set")
	}
}

// TestEncodeArrayFlags_ComplexAndSparse tests that both complex and sparse bits are set.
func TestEncodeArrayFlags_ComplexAndSparse(t *testing.T) {
	var buf bytes.Buffer
	w, err := NewWriter(&buf, "Test", "IM")
	if err != nil {
		t.Fatalf("NewWriter() error: %v", err)
	}

	v := &types.Variable{
		Name:       "csp",
		Dimensions: []int{1, 2},
		DataType:   types.Double,
		IsComplex:  true,
		IsSparse:   true,
		Data: &types.NumericArray{
			Real: []float64{1.0, 2.0},
			Imag: []float64{3.0, 4.0},
		},
	}

	flags := w.encodeArrayFlags(v)

	// Read the flags data (after the 8-byte tag)
	flagsWord := binary.LittleEndian.Uint32(flags[8:12])
	if flagsWord&0x0800 == 0 {
		t.Error("Complex flag bit (0x0800) not set")
	}
	if flagsWord&0x0400 == 0 {
		t.Error("Sparse flag bit (0x0400) not set")
	}
}

// TestWriteMatrix_BigEndian_Roundtrip writes with big-endian and verifies roundtrip parsing.
func TestWriteMatrix_BigEndian_Roundtrip(t *testing.T) {
	v := &types.Variable{
		Name:       "beround",
		Dimensions: []int{1, 4},
		DataType:   types.Int32,
		Data:       []int32{100, 200, 300, 400},
	}

	var buf bytes.Buffer
	w, err := NewWriter(&buf, "Test", "MI")
	if err != nil {
		t.Fatalf("NewWriter(MI) error: %v", err)
	}
	if err := w.WriteVariable(v); err != nil {
		t.Fatalf("WriteVariable() error: %v", err)
	}

	// Parse back
	parser, err := NewParser(bytes.NewReader(buf.Bytes()))
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
	if got.Name != "beround" {
		t.Errorf("Name = %q, want %q", got.Name, "beround")
	}

	gotData, ok := got.Data.([]int32)
	if !ok {
		t.Fatalf("Data type = %T, want []int32", got.Data)
	}
	wantData := []int32{100, 200, 300, 400}
	for i, want := range wantData {
		if gotData[i] != want {
			t.Errorf("Data[%d] = %d, want %d", i, gotData[i], want)
		}
	}
}

// TestWriteTag_ErrorOnWrite tests writeTag error propagation when the writer fails.
func TestWriteTag_ErrorOnWrite(t *testing.T) {
	fw := &failWriter{failAfter: 0} // fail immediately
	w := &Writer{
		w: fw,
		header: &Header{
			Order:           binary.LittleEndian,
			EndianIndicator: "IM",
			Version:         0x0100,
		},
	}

	err := w.writeTag(miMATRIX, 100)
	if err == nil {
		t.Error("writeTag() expected error from failing writer, got nil")
	}
}

// TestWriteMatrix_ErrorOnContentWrite tests error propagation when writing matrix content fails.
func TestWriteMatrix_ErrorOnContentWrite(t *testing.T) {
	// Writer that fails after writing the tag (8 bytes)
	fw := &failWriter{failAfter: 8}
	w := &Writer{
		w: fw,
		header: &Header{
			Order:           binary.LittleEndian,
			EndianIndicator: "IM",
			Version:         0x0100,
		},
	}

	v := &types.Variable{
		Name:       "x",
		Dimensions: []int{1, 2},
		DataType:   types.Double,
		Data:       []float64{1.0, 2.0},
	}

	err := w.writeMatrix(v)
	if err == nil {
		t.Error("writeMatrix() expected error from failing writer, got nil")
	}
}

// TestWriteHeader_ErrorOnWrite tests writeHeader error propagation.
func TestWriteHeader_ErrorOnWrite(t *testing.T) {
	fw := &failWriter{failAfter: 0}
	_, err := NewWriter(fw, "Test", "IM")
	if err == nil {
		t.Error("NewWriter() expected error from failing writer, got nil")
	}
}

// failWriter is a writer that fails after writing a certain number of bytes.
type failWriter struct {
	written   int
	failAfter int
}

func (f *failWriter) Write(p []byte) (int, error) {
	if f.written >= f.failAfter {
		return 0, errWriteFailed
	}
	n := len(p)
	if f.written+n > f.failAfter {
		n = f.failAfter - f.written
	}
	f.written += n
	if n < len(p) {
		return n, errWriteFailed
	}
	return n, nil
}

var errWriteFailed = fmt.Errorf("simulated write failure")

// TestWriteMatrix_EncodeContentError tests error propagation when encodeMatrixContent
// fails inside writeMatrix (covers the "failed to encode matrix content" return path).
func TestWriteMatrix_EncodeContentError(t *testing.T) {
	var buf bytes.Buffer
	w, err := NewWriter(&buf, "Test", "IM")
	if err != nil {
		t.Fatalf("NewWriter() error: %v", err)
	}

	// Variable that will fail during encodeMatrixContent because imaginary data
	// has wrong type, causing encodeData to fail.
	v := &types.Variable{
		Name:       "bad",
		Dimensions: []int{1, 2},
		DataType:   types.Double,
		IsComplex:  true,
		Data: &types.NumericArray{
			Real: []float64{1.0, 2.0},
			Imag: []int32{3, 4}, // wrong type
		},
	}

	err = w.writeMatrix(v)
	if err == nil {
		t.Error("writeMatrix() expected error for bad imaginary data type, got nil")
	}
	if err != nil && !contains(err.Error(), "failed to encode matrix content") {
		t.Errorf("writeMatrix() error = %q, want containing %q", err.Error(), "failed to encode matrix content")
	}
}

// TestWriteMatrix_WriteTagError tests error propagation when writeTag fails
// inside writeMatrix (covers the "failed to write matrix tag" return path).
func TestWriteMatrix_WriteTagError(t *testing.T) {
	fw := &failWriter{failAfter: 0} // fail immediately on first write
	w := &Writer{
		w: fw,
		header: &Header{
			Order:           binary.LittleEndian,
			EndianIndicator: "IM",
			Version:         0x0100,
		},
	}

	v := &types.Variable{
		Name:       "x",
		Dimensions: []int{1, 2},
		DataType:   types.Double,
		Data:       []float64{1.0, 2.0},
	}

	err := w.writeMatrix(v)
	if err == nil {
		t.Error("writeMatrix() expected error from failing writeTag, got nil")
	}
	if err != nil && !contains(err.Error(), "failed to write matrix tag") {
		t.Errorf("writeMatrix() error = %q, want containing %q", err.Error(), "failed to write matrix tag")
	}
}

// TestWriteMatrix_HappyPathWithPadding tests the normal path where content has non-zero
// padding (content length is not 8-byte aligned).
func TestWriteMatrix_HappyPathWithPadding(t *testing.T) {
	// Uint8 with 3 elements produces content that requires padding
	v := &types.Variable{
		Name:       "padtest",
		Dimensions: []int{1, 3},
		DataType:   types.Uint8,
		Data:       []byte{1, 2, 3},
	}

	var buf bytes.Buffer
	w, err := NewWriter(&buf, "Test", "IM")
	if err != nil {
		t.Fatalf("NewWriter() error: %v", err)
	}

	err = w.WriteVariable(v)
	if err != nil {
		t.Fatalf("WriteVariable() error: %v", err)
	}

	// Verify that the output length is 8-byte aligned after header
	dataLen := buf.Len() - 128
	if dataLen%8 != 0 {
		t.Errorf("Data length %d is not 8-byte aligned", dataLen)
	}
}

// TestWriteMatrix_PaddingErrorPath tests error when padding write fails.
func TestWriteMatrix_PaddingErrorPath(t *testing.T) {
	v := &types.Variable{
		Name:       "p",
		Dimensions: []int{1, 1},
		DataType:   types.Uint8,
		Data:       []byte{42},
	}

	// First, figure out how many bytes a successful write produces
	var measureBuf bytes.Buffer
	mw, _ := NewWriter(&measureBuf, "Test", "IM")
	_ = mw.WriteVariable(v)
	totalWritten := measureBuf.Len() - 128 // subtract header

	// failWriter that fails right before the padding write would complete.
	// The miMATRIX tag is 8 bytes, then the content, then padding.
	// We need to fail during the padding write. The content is written in
	// one shot, so we fail after tag (8) + content but before padding completes.
	// Let's fail after content write but before padding.
	// Total = tag(8) + content + padding
	// We want to fail at: tag(8) + content (no padding room)
	failAt := totalWritten - 1 // fail during the last byte(s) of padding

	fw := &failWriter{failAfter: failAt}
	w := &Writer{
		w: fw,
		header: &Header{
			Order:           binary.LittleEndian,
			EndianIndicator: "IM",
			Version:         0x0100,
		},
	}

	err := w.writeMatrix(v)
	if err == nil {
		t.Error("writeMatrix() expected error during padding write, got nil")
	}
}

// TestWriteHeader_ShortWrite tests the "wrote N bytes, expected 128" error path.
func TestWriteHeader_ShortWrite(t *testing.T) {
	// shortWriter writes only partial data without error
	sw := &shortWriter{maxBytes: 64}
	w := &Writer{
		w: sw,
		header: &Header{
			Description:     "Test",
			Version:         0x0100,
			EndianIndicator: "IM",
			Order:           binary.LittleEndian,
		},
	}

	err := w.writeHeader()
	if err == nil {
		t.Error("writeHeader() expected error for short write, got nil")
	}
	if err != nil && !contains(err.Error(), "expected 128") {
		t.Errorf("writeHeader() error = %q, want containing %q", err.Error(), "expected 128")
	}
}

// shortWriter is a writer that writes only up to maxBytes and reports success.
type shortWriter struct {
	maxBytes int
	written  int
}

func (s *shortWriter) Write(p []byte) (int, error) {
	remaining := s.maxBytes - s.written
	if remaining <= 0 {
		return 0, nil // no error, but no bytes written
	}
	n := len(p)
	if n > remaining {
		n = remaining
	}
	s.written += n
	return n, nil
}

// TestEncodeMatrixContent_RealDataError tests error propagation when real data encoding fails.
func TestEncodeMatrixContent_RealDataError(t *testing.T) {
	var buf bytes.Buffer
	w, err := NewWriter(&buf, "Test", "IM")
	if err != nil {
		t.Fatalf("NewWriter() error: %v", err)
	}

	// Variable where the data slice type does not match DataType
	v := &types.Variable{
		Name:       "bad",
		Dimensions: []int{1, 2},
		DataType:   types.Double,
		Data:       []int32{1, 2}, // wrong type, should be []float64
	}

	_, err = w.encodeMatrixContent(v)
	if err == nil {
		t.Error("encodeMatrixContent() expected error for wrong real data type, got nil")
	}
}

// TestEncodeMatrixContent_ImagError tests error propagation when imaginary data encoding fails.
func TestEncodeMatrixContent_ImagError(t *testing.T) {
	var buf bytes.Buffer
	w, err := NewWriter(&buf, "Test", "IM")
	if err != nil {
		t.Fatalf("NewWriter() error: %v", err)
	}

	// Complex variable where real part is valid but imaginary part has wrong type
	v := &types.Variable{
		Name:       "bad",
		Dimensions: []int{1, 2},
		DataType:   types.Double,
		IsComplex:  true,
		Data: &types.NumericArray{
			Real: []float64{1.0, 2.0},
			Imag: []int32{3, 4}, // wrong type for imaginary part
		},
	}

	_, err = w.encodeMatrixContent(v)
	if err == nil {
		t.Error("encodeMatrixContent() expected error for wrong imag data type, got nil")
	}
}

// TestValidateDimensions_Overflow tests dimension overflow detection.
func TestValidateDimensions_Overflow(t *testing.T) {
	tests := []struct {
		name        string
		dims        []int
		expectError bool
	}{
		{
			name:        "valid small dimensions",
			dims:        []int{10, 20, 30},
			expectError: false,
		},
		{
			name:        "valid 1D large array",
			dims:        []int{1000000},
			expectError: false,
		},
		{
			name:        "valid 2D matrix",
			dims:        []int{1000, 1000},
			expectError: false,
		},
		{
			name:        "overflow with huge dimensions",
			dims:        []int{math.MaxInt / 2, 3}, // (MaxInt/2) * 3 overflows
			expectError: true,
		},
		{
			name:        "overflow with many dimensions",
			dims:        []int{2000, 2000, 2000, 2000, 2000, 2000}, // 2000^6 overflows int64
			expectError: true,
		},
		{
			name:        "negative dimension",
			dims:        []int{10, -5, 20},
			expectError: true,
		},
		{
			name:        "zero dimension",
			dims:        []int{10, 0, 20},
			expectError: true,
		},
		{
			name:        "empty dimensions",
			dims:        []int{},
			expectError: true,
		},
		{
			name:        "single large valid dimension",
			dims:        []int{math.MaxInt / 2}, // Valid, won't overflow
			expectError: false,
		},
		{
			name:        "2D that overflows",
			dims:        []int{100000000, 100000000}, // 10^16 elements, won't overflow on 64-bit
			expectError: false,
		},
		{
			name:        "3D that overflows",
			dims:        []int{100000, 100000, 1000}, // 10^13 elements, won't overflow
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a minimal variable with test dimensions
			variable := &types.Variable{
				Name:       "test",
				Dimensions: tt.dims,
				Data:       []float64{1.0}, // Dummy data
				DataType:   types.Double,
			}

			// Create writer
			buf := &bytes.Buffer{}
			writer, err := NewWriter(buf, "Test", "IM") // "IM" = little-endian
			if err != nil {
				t.Fatalf("failed to create writer: %v", err)
			}

			// Validate
			err = writer.validateVariable(variable)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error for dimensions %v, got nil", tt.dims)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error for dimensions %v: %v", tt.dims, err)
				}
			}
		})
	}
}
