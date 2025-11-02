package v5

import (
	"encoding/binary"
	"math"
	"reflect"
	"testing"

	"github.com/scigolib/matlab/types"
)

func TestClassToDataType(t *testing.T) {
	tests := []struct {
		name  string
		class uint32
		want  types.DataType
	}{
		{"double", mxDOUBLE_CLASS, types.Double},
		{"single", mxSINGLE_CLASS, types.Single},
		{"int8", mxINT8_CLASS, types.Int8},
		{"uint8", mxUINT8_CLASS, types.Uint8},
		{"int16", mxINT16_CLASS, types.Int16},
		{"uint16", mxUINT16_CLASS, types.Uint16},
		{"int32", mxINT32_CLASS, types.Int32},
		{"uint32", mxUINT32_CLASS, types.Uint32},
		{"int64", mxINT64_CLASS, types.Int64},
		{"uint64", mxUINT64_CLASS, types.Uint64},
		{"char", mxCHAR_CLASS, types.Char},
		{"struct", mxSTRUCT_CLASS, types.Struct},
		{"cell", mxCELL_CLASS, types.CellArray},
		{"unknown", 999, types.Unknown},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := classToDataType(tt.class)
			if got != tt.want {
				t.Errorf("classToDataType(%d) = %v, want %v", tt.class, got, tt.want)
			}
		})
	}
}

func TestConvertData_Double(t *testing.T) {
	tests := []struct {
		name  string
		data  []byte
		order binary.ByteOrder
		want  []float64
	}{
		{
			name:  "single double little endian",
			data:  []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xf0, 0x3f}, // 1.0
			order: binary.LittleEndian,
			want:  []float64{1.0},
		},
		{
			name:  "single double big endian",
			data:  []byte{0x3f, 0xf0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, // 1.0
			order: binary.BigEndian,
			want:  []float64{1.0},
		},
		{
			name: "multiple doubles little endian",
			data: func() []byte {
				b := make([]byte, 16)
				binary.LittleEndian.PutUint64(b[0:8], math.Float64bits(1.0))
				binary.LittleEndian.PutUint64(b[8:16], math.Float64bits(2.0))
				return b
			}(),
			order: binary.LittleEndian,
			want:  []float64{1.0, 2.0},
		},
		{
			name:  "empty data",
			data:  []byte{},
			order: binary.LittleEndian,
			want:  []float64{},
		},
		{
			name:  "zero value",
			data:  []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
			order: binary.LittleEndian,
			want:  []float64{0.0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Parser{Header: &Header{Order: tt.order}}
			got := p.convertData(tt.data, miDOUBLE, mxDOUBLE_CLASS)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("convertData(miDOUBLE) = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConvertData_Single(t *testing.T) {
	tests := []struct {
		name  string
		data  []byte
		order binary.ByteOrder
		want  []float32
	}{
		{
			name:  "single float32 little endian",
			data:  []byte{0x00, 0x00, 0x80, 0x3f}, // 1.0
			order: binary.LittleEndian,
			want:  []float32{1.0},
		},
		{
			name:  "single float32 big endian",
			data:  []byte{0x3f, 0x80, 0x00, 0x00}, // 1.0
			order: binary.BigEndian,
			want:  []float32{1.0},
		},
		{
			name: "multiple float32",
			data: func() []byte {
				b := make([]byte, 8)
				binary.LittleEndian.PutUint32(b[0:4], math.Float32bits(1.5))
				binary.LittleEndian.PutUint32(b[4:8], math.Float32bits(2.5))
				return b
			}(),
			order: binary.LittleEndian,
			want:  []float32{1.5, 2.5},
		},
		{
			name:  "empty data",
			data:  []byte{},
			order: binary.LittleEndian,
			want:  []float32{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Parser{Header: &Header{Order: tt.order}}
			got := p.convertData(tt.data, miSINGLE, mxSINGLE_CLASS)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("convertData(miSINGLE) = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConvertData_Int8(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		want []int8
	}{
		{
			name: "positive values",
			data: []byte{1, 2, 3},
			want: []int8{1, 2, 3},
		},
		{
			name: "negative values",
			data: []byte{0xff, 0xfe, 0xfd}, // -1, -2, -3
			want: []int8{-1, -2, -3},
		},
		{
			name: "mixed values",
			data: []byte{127, 0, 128}, // 127, 0, -128
			want: []int8{127, 0, -128},
		},
		{
			name: "empty data",
			data: []byte{},
			want: []int8{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Parser{Header: &Header{Order: binary.LittleEndian}}
			got := p.convertData(tt.data, miINT8, mxINT8_CLASS)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("convertData(miINT8) = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConvertData_UInt8(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		want []byte
	}{
		{
			name: "normal values",
			data: []byte{0, 1, 255},
			want: []byte{0, 1, 255},
		},
		{
			name: "empty data",
			data: []byte{},
			want: []byte{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Parser{Header: &Header{Order: binary.LittleEndian}}
			got := p.convertData(tt.data, miUINT8, mxUINT8_CLASS)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("convertData(miUINT8) = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConvertData_Int16(t *testing.T) {
	tests := []struct {
		name  string
		data  []byte
		order binary.ByteOrder
		want  []int16
	}{
		{
			name:  "single value little endian",
			data:  []byte{0x01, 0x00}, // 1
			order: binary.LittleEndian,
			want:  []int16{1},
		},
		{
			name:  "single value big endian",
			data:  []byte{0x00, 0x01}, // 1
			order: binary.BigEndian,
			want:  []int16{1},
		},
		{
			name:  "negative value little endian",
			data:  []byte{0xff, 0xff}, // -1
			order: binary.LittleEndian,
			want:  []int16{-1},
		},
		{
			name: "multiple values",
			data: func() []byte {
				b := make([]byte, 4)
				binary.LittleEndian.PutUint16(b[0:2], 1)
				binary.LittleEndian.PutUint16(b[2:4], 2)
				return b
			}(),
			order: binary.LittleEndian,
			want:  []int16{1, 2},
		},
		{
			name:  "empty data",
			data:  []byte{},
			order: binary.LittleEndian,
			want:  []int16{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Parser{Header: &Header{Order: tt.order}}
			got := p.convertData(tt.data, miINT16, mxINT16_CLASS)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("convertData(miINT16) = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConvertData_UInt16(t *testing.T) {
	tests := []struct {
		name  string
		data  []byte
		order binary.ByteOrder
		want  []uint16
	}{
		{
			name:  "single value little endian",
			data:  []byte{0x01, 0x00}, // 1
			order: binary.LittleEndian,
			want:  []uint16{1},
		},
		{
			name:  "max value little endian",
			data:  []byte{0xff, 0xff}, // 65535
			order: binary.LittleEndian,
			want:  []uint16{65535},
		},
		{
			name:  "empty data",
			data:  []byte{},
			order: binary.LittleEndian,
			want:  []uint16{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Parser{Header: &Header{Order: tt.order}}
			got := p.convertData(tt.data, miUINT16, mxUINT16_CLASS)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("convertData(miUINT16) = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConvertData_Int32(t *testing.T) {
	tests := []struct {
		name  string
		data  []byte
		order binary.ByteOrder
		want  []int32
	}{
		{
			name:  "single value little endian",
			data:  []byte{0x01, 0x00, 0x00, 0x00}, // 1
			order: binary.LittleEndian,
			want:  []int32{1},
		},
		{
			name:  "negative value",
			data:  []byte{0xff, 0xff, 0xff, 0xff}, // -1
			order: binary.LittleEndian,
			want:  []int32{-1},
		},
		{
			name:  "empty data",
			data:  []byte{},
			order: binary.LittleEndian,
			want:  []int32{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Parser{Header: &Header{Order: tt.order}}
			got := p.convertData(tt.data, miINT32, mxINT32_CLASS)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("convertData(miINT32) = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConvertData_UInt32(t *testing.T) {
	tests := []struct {
		name  string
		data  []byte
		order binary.ByteOrder
		want  []uint32
	}{
		{
			name:  "single value",
			data:  []byte{0x01, 0x00, 0x00, 0x00}, // 1
			order: binary.LittleEndian,
			want:  []uint32{1},
		},
		{
			name:  "max value",
			data:  []byte{0xff, 0xff, 0xff, 0xff}, // 4294967295
			order: binary.LittleEndian,
			want:  []uint32{4294967295},
		},
		{
			name:  "empty data",
			data:  []byte{},
			order: binary.LittleEndian,
			want:  []uint32{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Parser{Header: &Header{Order: tt.order}}
			got := p.convertData(tt.data, miUINT32, mxUINT32_CLASS)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("convertData(miUINT32) = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConvertData_Int64(t *testing.T) {
	tests := []struct {
		name  string
		data  []byte
		order binary.ByteOrder
		want  []int64
	}{
		{
			name:  "single value little endian",
			data:  []byte{0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, // 1
			order: binary.LittleEndian,
			want:  []int64{1},
		},
		{
			name:  "negative value",
			data:  []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}, // -1
			order: binary.LittleEndian,
			want:  []int64{-1},
		},
		{
			name:  "empty data",
			data:  []byte{},
			order: binary.LittleEndian,
			want:  []int64{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Parser{Header: &Header{Order: tt.order}}
			got := p.convertData(tt.data, miINT64, mxINT64_CLASS)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("convertData(miINT64) = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConvertData_UInt64(t *testing.T) {
	tests := []struct {
		name  string
		data  []byte
		order binary.ByteOrder
		want  []uint64
	}{
		{
			name:  "single value",
			data:  []byte{0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, // 1
			order: binary.LittleEndian,
			want:  []uint64{1},
		},
		{
			name:  "large value",
			data:  []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x7f}, // max int64
			order: binary.LittleEndian,
			want:  []uint64{9223372036854775807},
		},
		{
			name:  "empty data",
			data:  []byte{},
			order: binary.LittleEndian,
			want:  []uint64{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Parser{Header: &Header{Order: tt.order}}
			got := p.convertData(tt.data, miUINT64, mxUINT64_CLASS)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("convertData(miUINT64) = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConvertData_UTF8(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		want string
	}{
		{
			name: "simple ASCII",
			data: []byte("hello"),
			want: "hello",
		},
		{
			name: "UTF-8 characters",
			data: []byte("hello 世界"),
			want: "hello 世界",
		},
		{
			name: "empty string",
			data: []byte{},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Parser{Header: &Header{Order: binary.LittleEndian}}
			got := p.convertData(tt.data, miUTF8, mxCHAR_CLASS)
			if got != tt.want {
				t.Errorf("convertData(miUTF8) = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConvertData_Unknown(t *testing.T) {
	p := &Parser{Header: &Header{Order: binary.LittleEndian}}
	data := []byte{1, 2, 3, 4}
	got := p.convertData(data, 9999, 9999) // Unknown type
	if !reflect.DeepEqual(got, data) {
		t.Errorf("convertData(unknown) = %v, want %v", got, data)
	}
}

// BenchmarkConvertData benchmarks type conversion performance.
func BenchmarkConvertData_Double(b *testing.B) {
	data := make([]byte, 8*100) // 100 doubles
	for i := 0; i < 100; i++ {
		binary.LittleEndian.PutUint64(data[i*8:(i+1)*8], math.Float64bits(float64(i)))
	}
	p := &Parser{Header: &Header{Order: binary.LittleEndian}}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = p.convertData(data, miDOUBLE, mxDOUBLE_CLASS)
	}
}

func BenchmarkConvertData_Int32(b *testing.B) {
	data := make([]byte, 4*100) // 100 int32s
	for i := 0; i < 100; i++ {
		binary.LittleEndian.PutUint32(data[i*4:(i+1)*4], uint32(i))
	}
	p := &Parser{Header: &Header{Order: binary.LittleEndian}}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = p.convertData(data, miINT32, mxINT32_CLASS)
	}
}
