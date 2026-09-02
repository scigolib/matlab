package v5

import (
	"bytes"
	"encoding/binary"
	"testing"

	"github.com/scigolib/matlab/types"
)

// TestSDE_WrapInTag_SmallData verifies that wrapInTag produces exactly 8 bytes
// for data of length 1-4, using the Small Data Element (SDE) format described
// in the MAT-file v5 specification.
//
// SDE layout (8 bytes total):
//
//	bytes 0-3: packed uint32 = (size << 16) | dataType
//	bytes 4-7: data bytes, zero-padded to fill 4 bytes
//
//nolint:gocognit // Table-driven test with per-byte SDE field validation
func TestSDE_WrapInTag_SmallData(t *testing.T) {
	var buf bytes.Buffer
	w, err := NewWriter(&buf, "Test", "IM") // "IM" = little-endian
	if err != nil {
		t.Fatalf("NewWriter() error = %v", err)
	}

	tests := []struct {
		name     string
		dataType uint32
		data     []byte
	}{
		{"1 byte miINT8", miINT8, []byte{0x7F}},
		{"2 bytes miINT16", miINT16, []byte{0x01, 0x02}},
		{"3 bytes miINT8", miINT8, []byte{0x0A, 0x0B, 0x0C}},
		{"4 bytes miINT32", miINT32, []byte{0x01, 0x02, 0x03, 0x04}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := w.wrapInTag(tt.dataType, tt.data)

			// SDE is always exactly 8 bytes.
			if len(result) != 8 {
				t.Fatalf("wrapInTag SDE total size = %d, want 8", len(result))
			}

			// The reader detects SDE by checking firstWord >> 16 in the file's byte
			// order. Decode the packed first word the same way the reader will.
			firstWord := binary.LittleEndian.Uint32(result[0:4])

			// Upper 16 bits = data size.
			gotSize := firstWord >> 16
			if gotSize != uint32(len(tt.data)) {
				t.Errorf("upper 16 bits (size) = %d, want %d", gotSize, len(tt.data))
			}

			// Lower 16 bits = data type.
			gotType := firstWord & 0xFFFF
			if gotType != tt.dataType {
				t.Errorf("lower 16 bits (dataType) = %d, want %d", gotType, tt.dataType)
			}

			// Data bytes occupy result[4 : 4+size]; remaining bytes must be zero.
			if !bytes.Equal(result[4:4+len(tt.data)], tt.data) {
				t.Errorf("data bytes = %v, want %v", result[4:4+len(tt.data)], tt.data)
			}
			for i := 4 + len(tt.data); i < 8; i++ {
				if result[i] != 0 {
					t.Errorf("padding byte[%d] = 0x%02x, want 0x00", i, result[i])
				}
			}
		})
	}
}

// TestSDE_WrapInTag_RegularData verifies that wrapInTag uses regular format
// (8-byte tag header + data + 8-byte-aligned padding) for data longer than 4 bytes.
//
//nolint:gocognit // Table-driven test with per-byte tag field validation
func TestSDE_WrapInTag_RegularData(t *testing.T) {
	var buf bytes.Buffer
	w, err := NewWriter(&buf, "Test", "IM") // "IM" = little-endian
	if err != nil {
		t.Fatalf("NewWriter() error = %v", err)
	}

	tests := []struct {
		name     string
		dataType uint32
		data     []byte
		wantSize int // 8 (tag) + len(data) + padding-to-8
	}{
		{"5 bytes", miUINT8, []byte{1, 2, 3, 4, 5}, 16}, // 8+5 → pad to 16
		{"8 bytes", miDOUBLE, make([]byte, 8), 16},      // 8+8 = 16, no pad
		{"16 bytes", miDOUBLE, make([]byte, 16), 24},    // 8+16 = 24, no pad
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := w.wrapInTag(tt.dataType, tt.data)

			if len(result) != tt.wantSize {
				t.Fatalf("wrapInTag regular total size = %d, want %d", len(result), tt.wantSize)
			}

			// Regular format: first 4 bytes = dataType (full uint32), NOT packed.
			// The upper 16 bits of the first uint32 must be zero (distinguishes from SDE).
			firstWord := binary.LittleEndian.Uint32(result[0:4])
			if firstWord>>16 != 0 {
				t.Errorf("upper 16 bits of first word = %d, want 0 (regular format requires no SDE marker)", firstWord>>16)
			}
			if firstWord != tt.dataType {
				t.Errorf("first word (dataType) = %d, want %d", firstWord, tt.dataType)
			}

			// Second uint32 = data size.
			sizeWord := binary.LittleEndian.Uint32(result[4:8])
			if sizeWord != uint32(len(tt.data)) {
				t.Errorf("size word = %d, want %d", sizeWord, len(tt.data))
			}

			// Data starts at byte 8.
			if !bytes.Equal(result[8:8+len(tt.data)], tt.data) {
				t.Errorf("data bytes mismatch")
			}

			// Padding bytes must be zero.
			for i := 8 + len(tt.data); i < len(result); i++ {
				if result[i] != 0 {
					t.Errorf("padding byte[%d] = 0x%02x, want 0x00", i, result[i])
				}
			}
		})
	}
}

// TestSDE_WrapInTag_EmptyData verifies that zero-length data always uses
// regular format. An SDE with size=0 would be ambiguous (same upper-16-bit
// pattern as a regular tag with type < 0x10000).
func TestSDE_WrapInTag_EmptyData(t *testing.T) {
	var buf bytes.Buffer
	w, err := NewWriter(&buf, "Test", "IM")
	if err != nil {
		t.Fatalf("NewWriter() error = %v", err)
	}

	result := w.wrapInTag(miINT8, []byte{})

	// Empty data: regular format, 8-byte tag with size=0, no data, no padding.
	if len(result) != 8 {
		t.Fatalf("wrapInTag empty total size = %d, want 8", len(result))
	}

	// Upper 16 bits of first word must be 0 — regular format, not SDE.
	firstWord := binary.LittleEndian.Uint32(result[0:4])
	if firstWord>>16 != 0 {
		t.Errorf("upper 16 bits = %d, want 0 (regular format for empty data)", firstWord>>16)
	}
	if firstWord != miINT8 {
		t.Errorf("dataType word = %d, want %d", firstWord, miINT8)
	}

	sizeWord := binary.LittleEndian.Uint32(result[4:8])
	if sizeWord != 0 {
		t.Errorf("size word = %d, want 0", sizeWord)
	}
}

// TestSDE_Roundtrip writes a variable with a short name (triggering SDE for the
// name sub-element) and reads it back via the parser. This exercises the
// writer-SDE → reader-SDE path end-to-end.
//
//nolint:gocognit // Table-driven test with full variable round-trip validation
func TestSDE_Roundtrip(t *testing.T) {
	tests := []struct {
		name    string
		varName string
		data    []float64
		dims    []int
	}{
		{
			name:    "short name 1 char triggers SDE",
			varName: "x",
			data:    []float64{1.0, 2.0, 3.0},
			dims:    []int{1, 3},
		},
		{
			name:    "short name 3 chars triggers SDE",
			varName: "abc",
			data:    []float64{42.0},
			dims:    []int{1, 1},
		},
		{
			name:    "short name 4 chars triggers SDE boundary",
			varName: "vars",
			data:    []float64{-1.0, 0.0, 1.0},
			dims:    []int{1, 3},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			w, err := NewWriter(&buf, "SDE roundtrip test", "IM")
			if err != nil {
				t.Fatalf("NewWriter() error: %v", err)
			}

			v := &types.Variable{
				Name:       tt.varName,
				Dimensions: tt.dims,
				DataType:   types.Double,
				Data:       tt.data,
			}

			if err := w.WriteVariable(v); err != nil {
				t.Fatalf("WriteVariable() error: %v", err)
			}

			// Parse back using the v5 parser (which already handles SDE via readTag).
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

			// Name must survive the round-trip.
			if got.Name != tt.varName {
				t.Errorf("Name = %q, want %q", got.Name, tt.varName)
			}

			// Dimensions must match.
			if len(got.Dimensions) != len(tt.dims) {
				t.Fatalf("Dimensions len = %d, want %d", len(got.Dimensions), len(tt.dims))
			}
			for i, d := range tt.dims {
				if got.Dimensions[i] != d {
					t.Errorf("Dimensions[%d] = %d, want %d", i, got.Dimensions[i], d)
				}
			}

			// Data values must match exactly.
			gotData, ok := got.Data.([]float64)
			if !ok {
				t.Fatalf("Data type = %T, want []float64", got.Data)
			}
			if len(gotData) != len(tt.data) {
				t.Fatalf("Data len = %d, want %d", len(gotData), len(tt.data))
			}
			for i, want := range tt.data {
				if gotData[i] != want {
					t.Errorf("Data[%d] = %v, want %v", i, gotData[i], want)
				}
			}
		})
	}
}

// TestSDE_Roundtrip_BigEndian verifies SDE round-trip correctness with big-endian
// byte order. The packed uint32 is written and read using the same order, so the
// reader's firstWord >> 16 extraction must still yield the correct size.
func TestSDE_Roundtrip_BigEndian(t *testing.T) {
	var buf bytes.Buffer
	w, err := NewWriter(&buf, "SDE big-endian test", "MI") // "MI" = big-endian
	if err != nil {
		t.Fatalf("NewWriter(MI) error: %v", err)
	}

	v := &types.Variable{
		Name:       "y",
		Dimensions: []int{1, 2},
		DataType:   types.Double,
		Data:       []float64{3.14, 2.71},
	}

	if err := w.WriteVariable(v); err != nil {
		t.Fatalf("WriteVariable() error: %v", err)
	}

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
	if got.Name != "y" {
		t.Errorf("Name = %q, want %q", got.Name, "y")
	}

	gotData, ok := got.Data.([]float64)
	if !ok {
		t.Fatalf("Data type = %T, want []float64", got.Data)
	}
	want := []float64{3.14, 2.71}
	for i, wv := range want {
		if gotData[i] != wv {
			t.Errorf("Data[%d] = %v, want %v", i, gotData[i], wv)
		}
	}
}

// TestSDE_FileSizeReduction verifies that writing with SDE produces a smaller
// file than what pure regular format would produce for the same variable.
//
// We compare against a hand-computed baseline: for a variable named "x" (1 byte),
// the name sub-element is 1 byte of miINT8 data.
//   - SDE:     8 bytes  (packed tag + 1 data byte + 3 zero bytes)
//   - Regular: 16 bytes (8 tag + 1 data + 7 padding)
//
// Saving of 8 bytes per <= 4-byte sub-element. For a typical scalar variable
// with name "x" and dims [1,1], we save bytes on:
//   - Name sub-element (1 byte → SDE saves 8 bytes vs regular)
//   - Dims [1,1]: 8 bytes → regular format, no SDE for this case
//
// We verify the actual file is smaller than the equivalent regular-format size.
func TestSDE_FileSizeReduction(t *testing.T) {
	// Write the same variable twice: once with SDE-enabled writer (current),
	// and compute what a regular-only writer would produce.

	var sdeBuf bytes.Buffer
	w, err := NewWriter(&sdeBuf, "Size test", "IM")
	if err != nil {
		t.Fatalf("NewWriter() error: %v", err)
	}

	v := &types.Variable{
		Name:       "x", // 1-byte name → SDE applies
		Dimensions: []int{1, 1},
		DataType:   types.Double,
		Data:       []float64{42.0},
	}

	if err := w.WriteVariable(v); err != nil {
		t.Fatalf("WriteVariable() error: %v", err)
	}

	sdeFileSize := sdeBuf.Len()

	// Compute the expected regular-format size for the same variable:
	// Header: 128 bytes
	// miMATRIX outer tag: 8 bytes
	// Content sub-elements (regular format):
	//   Array flags:   8 (tag) + 8 (data) = 16 bytes
	//   Dimensions:    8 (tag) + 8 (data, 2×int32) = 16 bytes
	//   Name "x":      8 (tag) + 1 (data) + 7 (padding) = 16 bytes
	//   Real data:     8 (tag) + 8 (float64) = 16 bytes
	// Total content: 16+16+16+16 = 64 bytes → outer tag size = 64
	// Total file: 128 + 8 + 64 = 200 bytes (regular-format baseline)
	regularOnlySize := 128 + 8 + 16 + 16 + 16 + 16

	// With SDE, the 1-byte name sub-element shrinks from 16 to 8 bytes.
	// Dimensions [1,1] is 2×int32 = 8 bytes → regular format (> 4 bytes).
	// So SDE saves exactly 8 bytes vs regular-only.
	expectedSDESize := regularOnlySize - 8 // 192 bytes

	if sdeFileSize != expectedSDESize {
		t.Errorf("SDE file size = %d bytes, want %d bytes (saving 8 bytes vs regular-only %d)", sdeFileSize, expectedSDESize, regularOnlySize)
	}

	// Verify SDE file is strictly smaller than the regular-only baseline.
	if sdeFileSize >= regularOnlySize {
		t.Errorf("SDE file (%d bytes) is not smaller than regular-only (%d bytes)", sdeFileSize, regularOnlySize)
	}
}
