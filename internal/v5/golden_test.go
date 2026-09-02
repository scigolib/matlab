package v5

import (
	"bytes"
	"encoding/binary"
	"math"
	"testing"

	"github.com/scigolib/matlab/types"
)

// writeTestVariableFresh creates one Writer with the given description and
// endian indicator, writes v, and returns the complete byte slice
// (header + variable data). endian must be "IM" (little-endian) or "MI"
// (big-endian).
func writeTestVariableFresh(t *testing.T, description string, v *types.Variable, endian string) []byte {
	t.Helper()
	var buf bytes.Buffer
	w, err := NewWriter(&buf, description, endian)
	if err != nil {
		t.Fatalf("NewWriter(%q) unexpected error: %v", endian, err)
	}
	if err := w.WriteVariable(v); err != nil {
		t.Fatalf("WriteVariable() unexpected error: %v", err)
	}
	return buf.Bytes()
}

// TestGolden_Header verifies the 128-byte MAT-file v5 header layout for
// little-endian ("IM") mode.
//
// Header spec (MAT-File Format v5):
//   - Bytes 0-115:   description text, zero-padded
//   - Bytes 116-123: subsystem data offset (zeros for standard files)
//   - Bytes 124-125: version 0x0100 stored in LE → [0x00, 0x01]
//   - Bytes 126-127: endian indicator "IM" [0x49, 0x4D] for little-endian
func TestGolden_Header(t *testing.T) {
	var buf bytes.Buffer
	_, err := NewWriter(&buf, "Test", "IM")
	if err != nil {
		t.Fatalf("NewWriter() error: %v", err)
	}

	got := buf.Bytes()
	if len(got) != 128 {
		t.Fatalf("header length = %d, want 128", len(got))
	}

	// Bytes 0-3: "Test"
	if string(got[0:4]) != "Test" {
		t.Errorf("bytes 0-3 = %q, want %q", string(got[0:4]), "Test")
	}

	// Bytes 4-115: zero (null padding after description)
	for i := 4; i < 116; i++ {
		if got[i] != 0x00 {
			t.Errorf("description padding byte[%d] = 0x%02X, want 0x00", i, got[i])
		}
	}

	// Bytes 116-123: subsystem data offset = all zeros
	for i := 116; i < 124; i++ {
		if got[i] != 0x00 {
			t.Errorf("subsystem offset byte[%d] = 0x%02X, want 0x00", i, got[i])
		}
	}

	// Bytes 124-125: version 0x0100 stored in little-endian → [0x00, 0x01]
	if got[124] != 0x00 || got[125] != 0x01 {
		t.Errorf("version bytes [124:126] = [0x%02X, 0x%02X], want [0x00, 0x01]",
			got[124], got[125])
	}
	version := binary.LittleEndian.Uint16(got[124:126])
	if version != 0x0100 {
		t.Errorf("version = 0x%04X, want 0x0100", version)
	}

	// Bytes 126-127: "IM" (0x49, 0x4D)
	if got[126] != 0x49 || got[127] != 0x4D {
		t.Errorf("endian indicator bytes [126:128] = [0x%02X, 0x%02X], want [0x49, 0x4D] (\"IM\")",
			got[126], got[127])
	}
	if string(got[126:128]) != "IM" {
		t.Errorf("endian indicator = %q, want %q", string(got[126:128]), "IM")
	}
}

// TestGolden_ArrayFlags_Double verifies the exact binary layout of the array
// flags sub-element for a non-complex, non-sparse double variable.
//
// Per MAT-File Format v5 spec, the array flags sub-element is always written
// in regular format (8-byte data, so it does not qualify for Small Data Element):
//   - Tag:    miUINT32 (type=6), size=8  → [0x06,0x00,0x00,0x00, 0x08,0x00,0x00,0x00] (LE)
//   - Word#1: bits 0-7 = class, bits 8-11 = flags → class=mxDOUBLE_CLASS(6), flags=0
//     → uint32 = 0x00000006 → LE bytes: [0x06, 0x00, 0x00, 0x00]
//   - Word#2: nzmax = 0 → [0x00, 0x00, 0x00, 0x00]
//   - Total:  16 bytes (8 tag + 8 data, exactly 8-byte aligned, no padding)
func TestGolden_ArrayFlags_Double(t *testing.T) {
	var buf bytes.Buffer
	w, err := NewWriter(&buf, "", "IM")
	if err != nil {
		t.Fatalf("NewWriter() error: %v", err)
	}

	v := &types.Variable{
		Name:      "x",
		DataType:  types.Double,
		IsComplex: false,
		IsSparse:  false,
	}

	flagBytes := w.encodeArrayFlags(v)

	// Total size: 8-byte tag + 8-byte data = 16 bytes (regular format, data > 4 bytes)
	if len(flagBytes) != 16 {
		t.Fatalf("encodeArrayFlags() len = %d, want 16", len(flagBytes))
	}

	// Tag bytes [0:8]:
	//   type = miUINT32 = 6 → LE: [0x06, 0x00, 0x00, 0x00]
	//   size = 8           → LE: [0x08, 0x00, 0x00, 0x00]
	wantTag := []byte{0x06, 0x00, 0x00, 0x00, 0x08, 0x00, 0x00, 0x00}
	if !bytes.Equal(flagBytes[0:8], wantTag) {
		t.Errorf("tag bytes [0:8] = % 02X, want % 02X", flagBytes[0:8], wantTag)
	}

	// Data word #1 [8:12]: class=6 (mxDOUBLE_CLASS), flags=0 → combined=0x00000006
	//   LE bytes: [0x06, 0x00, 0x00, 0x00]
	wantWord1 := []byte{0x06, 0x00, 0x00, 0x00}
	if !bytes.Equal(flagBytes[8:12], wantWord1) {
		t.Errorf("array flags word#1 [8:12] = % 02X, want % 02X", flagBytes[8:12], wantWord1)
	}
	word1 := binary.LittleEndian.Uint32(flagBytes[8:12])
	// class in bits 0-7 must be mxDOUBLE_CLASS = 6
	if word1&0xFF != mxDOUBLE_CLASS {
		t.Errorf("class bits = 0x%02X, want 0x%02X (mxDOUBLE_CLASS)", word1&0xFF, mxDOUBLE_CLASS)
	}
	// flag bits 8-11 must be zero (non-complex, non-sparse)
	if word1>>8&0xF != 0 {
		t.Errorf("flag bits [8:11] = 0x%X, want 0x0 (non-complex, non-sparse)", word1>>8&0xF)
	}

	// Data word #2 [12:16]: nzmax = 0 → [0x00, 0x00, 0x00, 0x00]
	wantWord2 := []byte{0x00, 0x00, 0x00, 0x00}
	if !bytes.Equal(flagBytes[12:16], wantWord2) {
		t.Errorf("array flags word#2 (nzmax) [12:16] = % 02X, want % 02X", flagBytes[12:16], wantWord2)
	}
}

// TestGolden_ArrayFlags_ComplexDouble verifies the exact binary layout of the
// array flags sub-element for a complex double variable.
//
// Complex flag is bit 11 = 0x0800. Combined with class:
//   - Word#1: class=mxDOUBLE_CLASS(6) | flags=0x0800 → uint32 = 0x00000806
//     → LE bytes: [0x06, 0x08, 0x00, 0x00]
//   - Word#2: nzmax = 0 → [0x00, 0x00, 0x00, 0x00]
func TestGolden_ArrayFlags_ComplexDouble(t *testing.T) {
	var buf bytes.Buffer
	w, err := NewWriter(&buf, "", "IM")
	if err != nil {
		t.Fatalf("NewWriter() error: %v", err)
	}

	v := &types.Variable{
		Name:      "z",
		DataType:  types.Double,
		IsComplex: true,
		IsSparse:  false,
	}

	flagBytes := w.encodeArrayFlags(v)

	if len(flagBytes) != 16 {
		t.Fatalf("encodeArrayFlags() len = %d, want 16", len(flagBytes))
	}

	// Word#1 [8:12]: 0x00000806 in LE → [0x06, 0x08, 0x00, 0x00]
	wantWord1 := []byte{0x06, 0x08, 0x00, 0x00}
	if !bytes.Equal(flagBytes[8:12], wantWord1) {
		t.Errorf("complex array flags word#1 [8:12] = % 02X, want % 02X", flagBytes[8:12], wantWord1)
	}

	word1 := binary.LittleEndian.Uint32(flagBytes[8:12])
	// class in bits 0-7
	if word1&0xFF != mxDOUBLE_CLASS {
		t.Errorf("class bits = 0x%02X, want 0x%02X (mxDOUBLE_CLASS)", word1&0xFF, mxDOUBLE_CLASS)
	}
	// complex bit (bit 11 = 0x0800) must be set
	if word1&0x0800 == 0 {
		t.Errorf("complex flag bit (0x0800) not set in word#1 = 0x%08X", word1)
	}
	// sparse bit (bit 10 = 0x0400) must NOT be set for a non-sparse variable
	if word1&0x0400 != 0 {
		t.Errorf("sparse flag bit (0x0400) unexpectedly set for non-sparse variable")
	}

	// Word#2 [12:16]: nzmax = 0
	wantWord2 := []byte{0x00, 0x00, 0x00, 0x00}
	if !bytes.Equal(flagBytes[12:16], wantWord2) {
		t.Errorf("complex array flags word#2 (nzmax) [12:16] = % 02X, want % 02X", flagBytes[12:16], wantWord2)
	}
}

// TestGolden_Dimensions verifies the exact binary layout of the dimensions
// sub-element for a 2x3 matrix.
//
// Dimensions sub-element uses regular format (8 bytes data > 4-byte SDE limit):
//   - Tag:  miINT32 (type=5), size=8 → [0x05,0x00,0x00,0x00, 0x08,0x00,0x00,0x00] (LE)
//   - Data: [2 as int32 LE, 3 as int32 LE]
//     → [0x02,0x00,0x00,0x00, 0x03,0x00,0x00,0x00]
//   - Total: 16 bytes (8 tag + 8 data, 8-byte aligned, no extra padding)
func TestGolden_Dimensions(t *testing.T) {
	var buf bytes.Buffer
	w, err := NewWriter(&buf, "", "IM")
	if err != nil {
		t.Fatalf("NewWriter() error: %v", err)
	}

	dimBytes := w.encodeDimensions([]int{2, 3})

	// Total: 8-byte tag + 8-byte data = 16 bytes
	if len(dimBytes) != 16 {
		t.Fatalf("encodeDimensions([2,3]) len = %d, want 16", len(dimBytes))
	}

	// Tag bytes [0:8]:
	//   type = miINT32 = 5 → LE: [0x05, 0x00, 0x00, 0x00]
	//   size = 8           → LE: [0x08, 0x00, 0x00, 0x00]
	wantTag := []byte{0x05, 0x00, 0x00, 0x00, 0x08, 0x00, 0x00, 0x00}
	if !bytes.Equal(dimBytes[0:8], wantTag) {
		t.Errorf("dimensions tag [0:8] = % 02X, want % 02X", dimBytes[0:8], wantTag)
	}

	// Dimension 0 = 2: LE int32 → [0x02, 0x00, 0x00, 0x00]
	wantDim0 := []byte{0x02, 0x00, 0x00, 0x00}
	if !bytes.Equal(dimBytes[8:12], wantDim0) {
		t.Errorf("dim[0]=2 bytes [8:12] = % 02X, want % 02X", dimBytes[8:12], wantDim0)
	}
	d0 := binary.LittleEndian.Uint32(dimBytes[8:12])
	if d0 != 2 {
		t.Errorf("dim[0] value = %d, want 2", d0)
	}

	// Dimension 1 = 3: LE int32 → [0x03, 0x00, 0x00, 0x00]
	wantDim1 := []byte{0x03, 0x00, 0x00, 0x00}
	if !bytes.Equal(dimBytes[12:16], wantDim1) {
		t.Errorf("dim[1]=3 bytes [12:16] = % 02X, want % 02X", dimBytes[12:16], wantDim1)
	}
	d1 := binary.LittleEndian.Uint32(dimBytes[12:16])
	if d1 != 3 {
		t.Errorf("dim[1] value = %d, want 3", d1)
	}
}

// TestGolden_VariableName verifies the exact binary layout of the name
// sub-element for a variable named "x".
//
// A 1-byte name uses the Small Data Element (SDE) format, which packs
// both the tag and data into a single 8-byte word:
//   - SDE word [0:4]: (size << 16) | dataType stored in native byte order
//     → packed value = (1 << 16) | miINT8(1) = 0x00010001
//     → LE bytes [0:4]: [0x01, 0x00, 0x01, 0x00]
//   - Data [4:8]: 0x78 ('x') followed by 3 zero padding bytes
//   - Total: 8 bytes (vs 16 bytes for regular format)
//
// SDE detection: the reader checks that the upper 16 bits of the first uint32
// are in range [1, 4]. Here upper bits = 0x0001 = 1, so SDE is used.
func TestGolden_VariableName(t *testing.T) {
	var buf bytes.Buffer
	w, err := NewWriter(&buf, "", "IM")
	if err != nil {
		t.Fatalf("NewWriter() error: %v", err)
	}

	nameBytes := w.encodeName("x")

	// SDE format: 8 bytes total (not 16, because data is 1 byte ≤ 4)
	if len(nameBytes) != 8 {
		t.Fatalf("encodeName(\"x\") len = %d, want 8 (Small Data Element format)", len(nameBytes))
	}

	// SDE packed word [0:4]: 0x00010001 in LE → [0x01, 0x00, 0x01, 0x00]
	// upper 16 bits = 0x0001 = size (1)
	// lower 16 bits = 0x0001 = miINT8 (type 1)
	wantSDE := []byte{0x01, 0x00, 0x01, 0x00}
	if !bytes.Equal(nameBytes[0:4], wantSDE) {
		t.Errorf("SDE packed word [0:4] = % 02X, want % 02X", nameBytes[0:4], wantSDE)
	}

	// Verify the SDE structure via uint32 decode
	packed := binary.LittleEndian.Uint32(nameBytes[0:4])
	sdeType := packed & 0xFFFF
	sdeSize := packed >> 16
	if sdeType != miINT8 {
		t.Errorf("SDE type = %d, want %d (miINT8)", sdeType, miINT8)
	}
	if sdeSize != 1 {
		t.Errorf("SDE size = %d, want 1", sdeSize)
	}

	// Data byte [4]: 'x' = 0x78
	if nameBytes[4] != 0x78 {
		t.Errorf("name byte [4] = 0x%02X, want 0x78 ('x')", nameBytes[4])
	}

	// Padding bytes [5:8]: all zeros (3 bytes to fill out the 4-byte data field)
	for i := 5; i < 8; i++ {
		if nameBytes[i] != 0x00 {
			t.Errorf("SDE padding byte [%d] = 0x%02X, want 0x00", i, nameBytes[i])
		}
	}
}

// TestGolden_FullVariable verifies the COMPLETE byte-level output for a
// simple double scalar named "a" with value 3.14 written in little-endian.
//
// Full layout (192 bytes total):
//
//	Offset  Size  Content
//	0       128   Header ("Test" + zero padding, subsystem zeros, version, endian)
//	128     8     miMATRIX tag (type=14 LE, size=56 LE)
//	136     16    Array flags sub-element (regular format: miUINT32 tag + 8-byte data)
//	152     16    Dimensions sub-element (regular format: miINT32 tag + [1,1])
//	168     8     Name sub-element (SDE format: packed word + 'a' + 3-byte padding)
//	176     16    Real data sub-element (regular format: miDOUBLE tag + 8-byte double)
//
// Matrix content = 16 + 16 + 8 + 16 = 56 bytes (8-byte aligned, no outer padding).
//
// SDE format is used for the 1-byte name because 1 ≤ 4 (the SDE threshold).
// Regular format is used for all other sub-elements (8-byte payloads).
//
// IEEE 754 double 3.14 = 0x40091EB851EB851F; LE bytes: [0x1F, 0x85, 0xEB, 0x51, 0xB8, 0x1E, 0x09, 0x40].
func TestGolden_FullVariable(t *testing.T) {
	v := &types.Variable{
		Name:       "a",
		Dimensions: []int{1, 1},
		DataType:   types.Double,
		Data:       []float64{3.14},
	}

	got := writeTestVariableFresh(t, "Test", v, "IM")

	// Total size: 128 (header) + 8 (miMATRIX tag) + 56 (content) = 192
	if len(got) != 192 {
		t.Fatalf("total output length = %d, want 192", len(got))
	}

	// --- Header (bytes 0-127) ---

	// Description "Test" at bytes 0-3, then zeros through byte 115
	if string(got[0:4]) != "Test" {
		t.Errorf("header description [0:4] = %q, want %q", string(got[0:4]), "Test")
	}
	for i := 4; i < 116; i++ {
		if got[i] != 0x00 {
			t.Errorf("header description padding [%d] = 0x%02X, want 0x00", i, got[i])
		}
	}
	// Subsystem data offset zeros (bytes 116-123)
	for i := 116; i < 124; i++ {
		if got[i] != 0x00 {
			t.Errorf("subsystem offset byte [%d] = 0x%02X, want 0x00", i, got[i])
		}
	}
	// Version 0x0100 in LE → [0x00, 0x01] at bytes 124-125
	if got[124] != 0x00 || got[125] != 0x01 {
		t.Errorf("version [124:126] = [0x%02X, 0x%02X], want [0x00, 0x01]", got[124], got[125])
	}
	// Endian "IM" [0x49, 0x4D] at bytes 126-127
	if got[126] != 0x49 || got[127] != 0x4D {
		t.Errorf("endian [126:128] = [0x%02X, 0x%02X], want [0x49, 0x4D] (\"IM\")", got[126], got[127])
	}

	// --- miMATRIX tag (bytes 128-135) ---
	// type = miMATRIX = 14 LE → [0x0E, 0x00, 0x00, 0x00]
	// size = 56 LE             → [0x38, 0x00, 0x00, 0x00]
	// (56 = 16 flags + 16 dims + 8 name SDE + 16 data)
	wantMatrixTag := []byte{
		0x0E, 0x00, 0x00, 0x00, // type = 14 (miMATRIX)
		0x38, 0x00, 0x00, 0x00, // size = 56
	}
	if !bytes.Equal(got[128:136], wantMatrixTag) {
		t.Errorf("miMATRIX tag [128:136] = % 02X, want % 02X", got[128:136], wantMatrixTag)
	}
	matrixType := binary.LittleEndian.Uint32(got[128:132])
	if matrixType != miMATRIX {
		t.Errorf("miMATRIX type = %d, want %d (miMATRIX)", matrixType, miMATRIX)
	}
	matrixSize := binary.LittleEndian.Uint32(got[132:136])
	if matrixSize != 56 {
		t.Errorf("miMATRIX content size = %d, want 56", matrixSize)
	}

	// --- Array flags sub-element (bytes 136-151) ---
	// Regular format: tag(8) + data(8) = 16 bytes
	// Tag: miUINT32(6) LE, size=8 LE
	// Word#1: class=mxDOUBLE_CLASS(6), flags=0 → 0x00000006 LE → [0x06,0x00,0x00,0x00]
	// Word#2: nzmax=0 → [0x00,0x00,0x00,0x00]
	wantFlags := []byte{
		0x06, 0x00, 0x00, 0x00, // tag type = miUINT32 = 6
		0x08, 0x00, 0x00, 0x00, // tag size = 8
		0x06, 0x00, 0x00, 0x00, // word#1: class=6 (mxDOUBLE_CLASS), flags=0
		0x00, 0x00, 0x00, 0x00, // word#2: nzmax=0
	}
	if !bytes.Equal(got[136:152], wantFlags) {
		t.Errorf("array flags [136:152] = % 02X, want % 02X", got[136:152], wantFlags)
	}

	// --- Dimensions sub-element (bytes 152-167) ---
	// Regular format: tag(8) + data(8) = 16 bytes
	// Tag: miINT32(5) LE, size=8 LE
	// dim[0]=1 LE → [0x01,0x00,0x00,0x00]
	// dim[1]=1 LE → [0x01,0x00,0x00,0x00]
	wantDims := []byte{
		0x05, 0x00, 0x00, 0x00, // tag type = miINT32 = 5
		0x08, 0x00, 0x00, 0x00, // tag size = 8
		0x01, 0x00, 0x00, 0x00, // dim[0] = 1
		0x01, 0x00, 0x00, 0x00, // dim[1] = 1
	}
	if !bytes.Equal(got[152:168], wantDims) {
		t.Errorf("dimensions [152:168] = % 02X, want % 02X", got[152:168], wantDims)
	}

	// --- Name sub-element (bytes 168-175) --- SDE format, 8 bytes
	// SDE packed word [168:172]: (size=1 << 16) | miINT8(1) = 0x00010001 in LE
	//   → [0x01, 0x00, 0x01, 0x00]
	// Data [172:176]: 'a'=0x61 followed by 3 zero padding bytes
	wantName := []byte{
		0x01, 0x00, 0x01, 0x00, // SDE: type=miINT8(1) | size=1 packed LE
		0x61, 0x00, 0x00, 0x00, // 'a' + 3-byte zero padding
	}
	if !bytes.Equal(got[168:176], wantName) {
		t.Errorf("name SDE [168:176] = % 02X, want % 02X", got[168:176], wantName)
	}

	// --- Real data sub-element (bytes 176-191) ---
	// Regular format: tag(8) + data(8) = 16 bytes
	// Tag: miDOUBLE(9) LE, size=8 LE
	// IEEE 754 double 3.14 = 0x40091EB851EB851F
	// LE bytes: [0x1F, 0x85, 0xEB, 0x51, 0xB8, 0x1E, 0x09, 0x40]
	wantData := []byte{
		0x09, 0x00, 0x00, 0x00, // tag type = miDOUBLE = 9
		0x08, 0x00, 0x00, 0x00, // tag size = 8
		0x1F, 0x85, 0xEB, 0x51, // 3.14 IEEE 754 double LE, bytes 0-3
		0xB8, 0x1E, 0x09, 0x40, // 3.14 IEEE 754 double LE, bytes 4-7
	}
	if !bytes.Equal(got[176:192], wantData) {
		t.Errorf("real data [176:192] = % 02X, want % 02X", got[176:192], wantData)
	}

	// Cross-check: decode the stored double and verify it round-trips exactly
	storedBits := binary.LittleEndian.Uint64(got[184:192])
	if storedBits != math.Float64bits(3.14) {
		t.Errorf("stored bits = 0x%016X, want 0x%016X (3.14)", storedBits, math.Float64bits(3.14))
	}
}

// TestGolden_BigEndian verifies the COMPLETE byte-level output for a simple
// double scalar named "a" with value 3.14 written in big-endian ("MI").
//
// Big-endian layout (192 bytes total, same structure as LE but with byte-swapped integers):
//
//	Offset  Size  Content
//	0       128   Header ("Test" + zero padding, version BE, endian "MI")
//	128     8     miMATRIX tag (type=14 BE, size=56 BE)
//	136     16    Array flags sub-element (regular format, all words BE)
//	152     16    Dimensions sub-element (regular format, all words BE)
//	168     8     Name sub-element (SDE format, packed word BE, 'a' byte unchanged)
//	176     16    Real data sub-element (regular format, IEEE 754 BE)
//
// Big-endian byte-swap rules relative to LE:
//   - Endian indicator: "MI" [0x4D, 0x49]
//   - Version 0x0100 BE → [0x01, 0x00]
//   - miMATRIX type=14 BE → [0x00,0x00,0x00,0x0E]; size=56 BE → [0x00,0x00,0x00,0x38]
//   - Array flags tag: miUINT32(6) BE → [0x00,0x00,0x00,0x06]; size=8 BE → [0x00,0x00,0x00,0x08]
//   - Array flags word#1: 0x00000006 BE → [0x00,0x00,0x00,0x06]
//   - Dims tag: miINT32(5) BE → [0x00,0x00,0x00,0x05]; size=8 BE → [0x00,0x00,0x00,0x08]
//   - Name SDE packed: 0x00010001 BE → [0x00,0x01,0x00,0x01]
//   - Data tag: miDOUBLE(9) BE → [0x00,0x00,0x00,0x09]; size=8 BE → [0x00,0x00,0x00,0x08]
//   - IEEE 754 double 3.14 BE: [0x40,0x09,0x1E,0xB8,0x51,0xEB,0x85,0x1F]
//
// Single-byte values ('a'=0x61, zero padding) are unaffected by endianness.
func TestGolden_BigEndian(t *testing.T) {
	v := &types.Variable{
		Name:       "a",
		Dimensions: []int{1, 1},
		DataType:   types.Double,
		Data:       []float64{3.14},
	}

	got := writeTestVariableFresh(t, "Test", v, "MI")

	// Total size must equal the LE variant (192 bytes)
	if len(got) != 192 {
		t.Fatalf("total output length = %d, want 192", len(got))
	}

	// --- Header ---

	// Description is a raw byte copy, unaffected by endianness
	if string(got[0:4]) != "Test" {
		t.Errorf("header description [0:4] = %q, want %q", string(got[0:4]), "Test")
	}

	// Version 0x0100 in big-endian → [0x01, 0x00] at bytes 124-125
	if got[124] != 0x01 || got[125] != 0x00 {
		t.Errorf("version BE [124:126] = [0x%02X, 0x%02X], want [0x01, 0x00]", got[124], got[125])
	}
	version := binary.BigEndian.Uint16(got[124:126])
	if version != 0x0100 {
		t.Errorf("version = 0x%04X, want 0x0100", version)
	}

	// Endian "MI" [0x4D, 0x49] at bytes 126-127
	if got[126] != 0x4D || got[127] != 0x49 {
		t.Errorf("endian [126:128] = [0x%02X, 0x%02X], want [0x4D, 0x49] (\"MI\")", got[126], got[127])
	}
	if string(got[126:128]) != "MI" {
		t.Errorf("endian indicator = %q, want %q", string(got[126:128]), "MI")
	}

	// --- miMATRIX tag (bytes 128-135) ---
	// type=14 BE → [0x00, 0x00, 0x00, 0x0E]
	// size=56 BE → [0x00, 0x00, 0x00, 0x38]
	wantMatrixTag := []byte{
		0x00, 0x00, 0x00, 0x0E, // type = 14 (miMATRIX) in big-endian
		0x00, 0x00, 0x00, 0x38, // size = 56 in big-endian
	}
	if !bytes.Equal(got[128:136], wantMatrixTag) {
		t.Errorf("miMATRIX tag BE [128:136] = % 02X, want % 02X", got[128:136], wantMatrixTag)
	}
	matrixType := binary.BigEndian.Uint32(got[128:132])
	if matrixType != miMATRIX {
		t.Errorf("miMATRIX type BE = %d, want %d (miMATRIX)", matrixType, miMATRIX)
	}
	matrixSize := binary.BigEndian.Uint32(got[132:136])
	if matrixSize != 56 {
		t.Errorf("miMATRIX content size BE = %d, want 56", matrixSize)
	}

	// --- Array flags sub-element (bytes 136-151) ---
	// Regular format: tag(8) + data(8) = 16 bytes
	// Tag: miUINT32(6) BE, size=8 BE
	// Word#1: class=6, flags=0 → 0x00000006 BE → [0x00, 0x00, 0x00, 0x06]
	// Word#2: nzmax=0          → [0x00, 0x00, 0x00, 0x00]
	wantFlags := []byte{
		0x00, 0x00, 0x00, 0x06, // tag type = miUINT32 = 6 in big-endian
		0x00, 0x00, 0x00, 0x08, // tag size = 8 in big-endian
		0x00, 0x00, 0x00, 0x06, // word#1: class=6 (mxDOUBLE_CLASS), flags=0 in big-endian
		0x00, 0x00, 0x00, 0x00, // word#2: nzmax=0
	}
	if !bytes.Equal(got[136:152], wantFlags) {
		t.Errorf("array flags BE [136:152] = % 02X, want % 02X", got[136:152], wantFlags)
	}
	// In BE, the class occupies the least-significant byte (byte offset 3 of the word)
	flagWord1 := binary.BigEndian.Uint32(got[144:148])
	if flagWord1&0xFF != mxDOUBLE_CLASS {
		t.Errorf("BE class bits = 0x%02X, want 0x%02X (mxDOUBLE_CLASS)", flagWord1&0xFF, mxDOUBLE_CLASS)
	}

	// --- Dimensions sub-element (bytes 152-167) ---
	// Regular format: tag(8) + data(8) = 16 bytes
	// Tag: miINT32(5) BE, size=8 BE
	// dim[0]=1 BE → [0x00, 0x00, 0x00, 0x01]
	// dim[1]=1 BE → [0x00, 0x00, 0x00, 0x01]
	wantDims := []byte{
		0x00, 0x00, 0x00, 0x05, // tag type = miINT32 = 5 in big-endian
		0x00, 0x00, 0x00, 0x08, // tag size = 8 in big-endian
		0x00, 0x00, 0x00, 0x01, // dim[0] = 1 in big-endian
		0x00, 0x00, 0x00, 0x01, // dim[1] = 1 in big-endian
	}
	if !bytes.Equal(got[152:168], wantDims) {
		t.Errorf("dimensions BE [152:168] = % 02X, want % 02X", got[152:168], wantDims)
	}

	// --- Name sub-element (bytes 168-175) --- SDE format, 8 bytes
	// SDE packed word [168:172]: 0x00010001 in BE → [0x00, 0x01, 0x00, 0x01]
	// upper 16 bits (BE) = 0x0001 = size (1)
	// lower 16 bits (BE) = 0x0001 = miINT8 (type 1)
	// Data [172:176]: 'a'=0x61 followed by 3 zero padding bytes (byte values unaffected)
	wantName := []byte{
		0x00, 0x01, 0x00, 0x01, // SDE: type=miINT8(1) | size=1 packed BE
		0x61, 0x00, 0x00, 0x00, // 'a' + 3-byte zero padding
	}
	if !bytes.Equal(got[168:176], wantName) {
		t.Errorf("name SDE BE [168:176] = % 02X, want % 02X", got[168:176], wantName)
	}

	// --- Real data sub-element (bytes 176-191) ---
	// Regular format: tag(8) + data(8) = 16 bytes
	// Tag: miDOUBLE(9) BE, size=8 BE
	// IEEE 754 double 3.14 BE: [0x40, 0x09, 0x1E, 0xB8, 0x51, 0xEB, 0x85, 0x1F]
	wantData := []byte{
		0x00, 0x00, 0x00, 0x09, // tag type = miDOUBLE = 9 in big-endian
		0x00, 0x00, 0x00, 0x08, // tag size = 8 in big-endian
		0x40, 0x09, 0x1E, 0xB8, // 3.14 IEEE 754 double BE, bytes 0-3
		0x51, 0xEB, 0x85, 0x1F, // 3.14 IEEE 754 double BE, bytes 4-7
	}
	if !bytes.Equal(got[176:192], wantData) {
		t.Errorf("real data BE [176:192] = % 02X, want % 02X", got[176:192], wantData)
	}

	// Cross-check: decode the stored double in big-endian and verify round-trip
	storedBits := binary.BigEndian.Uint64(got[184:192])
	if storedBits != math.Float64bits(3.14) {
		t.Errorf("stored bits BE = 0x%016X, want 0x%016X (3.14)", storedBits, math.Float64bits(3.14))
	}
}
