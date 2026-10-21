package v5

import (
	"bytes"
	"encoding/binary"
	"math"
	"strings"
	"testing"
)

// matrixWithDims builds a miMATRIX element carrying the array flags and the
// given dimensions payload. The element ends right after the dimensions, so a
// parser that accepts the payload fails later on the missing name tag — which
// is how these tests tell "rejected as malformed dimensions" from "accepted".
func matrixWithDims(dims []byte) []byte {
	var payload bytes.Buffer

	flags := make([]byte, 8)
	binary.LittleEndian.PutUint32(flags[0:4], mxDOUBLE_CLASS)
	payload.Write(regularTag(miUINT32, 8))
	payload.Write(flags)

	payload.Write(regularTag(miINT32, uint32(len(dims))))
	payload.Write(dims)
	payload.Write(make([]byte, (8-len(dims)%8)%8))

	return streamBytes(testHeader(), regularTag(miMATRIX, uint32(payload.Len())), payload.Bytes())
}

// parseMatrixDims parses a file whose only element carries the given
// dimensions payload.
func parseMatrixDims(t *testing.T, dims []byte) error {
	t.Helper()

	parser, err := NewParser(bytes.NewReader(matrixWithDims(dims)))
	if err != nil {
		t.Fatalf("NewParser() error: %v", err)
	}

	_, err = parser.Parse()
	return err
}

// TestParseMatrix_RejectsDimensionAboveInt32 verifies the dimensions payload is
// validated as an int32 array, as the MAT v5 specification defines it. Without
// the check a value above MaxInt32 is a large positive int on 64-bit and a
// negative int on 32-bit — the same file would describe two different arrays.
func TestParseMatrix_RejectsDimensionAboveInt32(t *testing.T) {
	dims := make([]byte, 8)
	binary.LittleEndian.PutUint32(dims[0:4], 1)
	binary.LittleEndian.PutUint32(dims[4:8], 0xFFFFFFFF) // -1 as int32

	err := parseMatrixDims(t, dims)

	if err == nil || !strings.Contains(err.Error(), "dimension") {
		t.Fatalf("Parse() error = %v, want a dimension validation error (dimensions are an int32 array)", err)
	}
}

// TestParseMatrix_RejectsUnalignedDimensions verifies a dimensions payload whose
// length is not a multiple of the 4-byte element size is rejected instead of
// silently dropping the trailing bytes.
func TestParseMatrix_RejectsUnalignedDimensions(t *testing.T) {
	dims := make([]byte, 5)
	binary.LittleEndian.PutUint32(dims[0:4], 1)

	err := parseMatrixDims(t, dims)

	if err == nil || !strings.Contains(err.Error(), "dimension") {
		t.Fatalf("Parse() error = %v, want a dimensions payload alignment error", err)
	}
}

// TestParseMatrix_AcceptsMaxInt32Dimension pins the upper boundary: the largest
// valid int32 dimension must not be rejected by the new validation.
func TestParseMatrix_AcceptsMaxInt32Dimension(t *testing.T) {
	dims := make([]byte, 8)
	binary.LittleEndian.PutUint32(dims[0:4], 1)
	binary.LittleEndian.PutUint32(dims[4:8], math.MaxInt32)

	err := parseMatrixDims(t, dims)

	if err != nil && strings.Contains(err.Error(), "dimension") {
		t.Fatalf("Parse() rejected the largest valid int32 dimension: %v", err)
	}
}
