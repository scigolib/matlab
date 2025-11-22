package v5

import (
	"bytes"
	"encoding/binary"
	"strings"
	"testing"
)

// TestDataTag_OversizedTag tests that oversized tags are rejected.
func TestDataTag_OversizedTag(t *testing.T) {
	tests := []struct {
		name        string
		size        uint32
		expectError bool
	}{
		{
			name:        "valid small size",
			size:        1024,
			expectError: false,
		},
		{
			name:        "valid 1MB size",
			size:        1024 * 1024,
			expectError: false,
		},
		{
			name:        "valid 1GB size",
			size:        1024 * 1024 * 1024,
			expectError: false,
		},
		{
			name:        "valid 2GB size (boundary)",
			size:        maxReasonableSize,
			expectError: false,
		},
		{
			name:        "invalid 2GB+1 size",
			size:        maxReasonableSize + 1,
			expectError: true,
		},
		{
			name:        "malicious max uint32 size",
			size:        0xFFFFFFFF,
			expectError: true,
		},
		{
			name:        "malicious 3GB size",
			size:        3 * 1024 * 1024 * 1024,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tag, err := createAndParseTag(tt.size)

			if tt.expectError {
				assertTagError(t, tag, err, tt.size)
			} else {
				assertTagValid(t, tag, err, tt.size)
			}
		})
	}
}

// createAndParseTag creates a tag with the given size and parses it.
func createAndParseTag(size uint32) (*DataTag, error) {
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint32(buf[0:4], uint32(miDOUBLE)) // Data type
	binary.LittleEndian.PutUint32(buf[4:8], size)             // Size

	reader := bytes.NewReader(buf)
	parser := &Parser{
		r: reader,
		Header: &Header{
			Order: binary.LittleEndian,
		},
	}

	return parser.readTag()
}

// assertTagError verifies that tag parsing failed as expected.
func assertTagError(t *testing.T, tag *DataTag, err error, size uint32) {
	t.Helper()
	if err == nil {
		t.Errorf("expected error for size %d, got nil", size)
	}
	if tag != nil {
		t.Errorf("expected nil tag for size %d, got %+v", size, tag)
	}
	if err != nil && !strings.Contains(err.Error(), "tag size too large") {
		t.Errorf("wrong error message: %v", err)
	}
}

// assertTagValid verifies that tag parsing succeeded as expected.
func assertTagValid(t *testing.T, tag *DataTag, err error, size uint32) {
	t.Helper()
	if err != nil {
		t.Errorf("unexpected error for size %d: %v", size, err)
	}
	if tag == nil {
		t.Errorf("expected valid tag for size %d, got nil", size)
		return
	}
	if tag.Size != size {
		t.Errorf("size mismatch: got %d, want %d", tag.Size, size)
	}
	if tag.IsSmall {
		t.Errorf("expected regular format tag, got small format")
	}
}

// TestDataTag_SmallFormat tests that small format tags work correctly.
func TestDataTag_SmallFormat(t *testing.T) {
	tests := []struct {
		name        string
		size        uint32
		dataType    uint32
		expectSmall bool
	}{
		{
			name:        "size 1",
			size:        1,
			dataType:    uint32(miDOUBLE),
			expectSmall: true,
		},
		{
			name:        "size 4",
			size:        4,
			dataType:    uint32(miSINGLE),
			expectSmall: true,
		},
		{
			name:        "size 5 (regular format)",
			size:        5,
			dataType:    uint32(miDOUBLE),
			expectSmall: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := make([]byte, 8)

			if tt.expectSmall {
				// Small format: size in upper 16 bits, type in lower 16 bits
				firstWord := (tt.size << 16) | (tt.dataType & 0xFFFF)
				binary.LittleEndian.PutUint32(buf[0:4], firstWord)
			} else {
				// Regular format
				binary.LittleEndian.PutUint32(buf[0:4], tt.dataType)
				binary.LittleEndian.PutUint32(buf[4:8], tt.size)
			}

			reader := bytes.NewReader(buf)
			parser := &Parser{
				r: reader,
				Header: &Header{
					Order: binary.LittleEndian,
				},
			}

			tag, err := parser.readTag()

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tag.Size != tt.size {
				t.Errorf("size mismatch: got %d, want %d", tag.Size, tt.size)
			}

			if tag.IsSmall != tt.expectSmall {
				t.Errorf("IsSmall mismatch: got %v, want %v", tag.IsSmall, tt.expectSmall)
			}
		})
	}
}
