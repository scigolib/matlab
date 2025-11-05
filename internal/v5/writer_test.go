package v5

import (
	"bytes"
	"encoding/binary"
	"math"
	"testing"

	"github.com/scigolib/matlab/types"
)

// TestNewWriter tests writer creation with both endianness.
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
			endian:      "MI",
			wantErr:     false,
		},
		{
			name:        "big endian",
			description: "Test file big endian",
			endian:      "IM",
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
			endian:      "MI",
			wantErr:     false,
		},
		{
			name:        "long description",
			description: string(make([]byte, 200)), // 200 bytes, should truncate to 116
			endian:      "MI",
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
			endian:      "MI",
			wantVersion: 0x0100,
		},
		{
			name:        "big endian header",
			description: "Test MAT-file",
			endian:      "IM",
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
			var order binary.ByteOrder
			if tt.endian == "MI" {
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
	w, err := NewWriter(&buf, "Test", "MI")
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
	w, err := NewWriter(&buf, "Test", "MI")
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
	w, err := NewWriter(&buf, "Test", "MI")
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
	w, err := NewWriter(&buf, "Test", "MI")
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
			w, err := NewWriter(&buf, "Test", "MI")
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
