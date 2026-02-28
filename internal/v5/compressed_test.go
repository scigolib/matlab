package v5

import (
	"bytes"
	"compress/zlib"
	"testing"
)

// TestDecompress_ValidData compresses data with zlib, then decompresses and verifies.
func TestDecompress_ValidData(t *testing.T) {
	// Prepare test data
	original := []byte("Hello, this is test data for zlib compression in MAT-file v5 format!")

	// Compress with zlib
	var compressedBuf bytes.Buffer
	zlibWriter := zlib.NewWriter(&compressedBuf)
	if _, err := zlibWriter.Write(original); err != nil {
		t.Fatalf("zlib.Write() error: %v", err)
	}
	if err := zlibWriter.Close(); err != nil {
		t.Fatalf("zlib.Close() error: %v", err)
	}
	compressed := compressedBuf.Bytes()

	// Decompress using our function
	reader := bytes.NewReader(compressed)
	result, err := decompress(reader, uint32(len(compressed)))
	if err != nil {
		t.Fatalf("decompress() unexpected error: %v", err)
	}

	if !bytes.Equal(result, original) {
		t.Errorf("decompress() = %q, want %q", result, original)
	}
}

// TestDecompress_InvalidZlib tests that garbage bytes produce an error.
func TestDecompress_InvalidZlib(t *testing.T) {
	garbage := []byte{0xDE, 0xAD, 0xBE, 0xEF, 0x01, 0x02, 0x03, 0x04}
	reader := bytes.NewReader(garbage)
	_, err := decompress(reader, uint32(len(garbage)))
	if err == nil {
		t.Error("decompress() expected error for invalid zlib data, got nil")
	}
}

// TestDecompress_EmptyInput tests that an empty reader produces an error.
func TestDecompress_EmptyInput(t *testing.T) {
	// compressedSize > 0 but empty reader => io.ReadFull should fail
	reader := bytes.NewReader([]byte{})
	_, err := decompress(reader, 10)
	if err == nil {
		t.Error("decompress() expected error for empty input, got nil")
	}
}

// TestDecompress_CorruptedData tests that a valid zlib header with corrupted body produces an error.
func TestDecompress_CorruptedData(t *testing.T) {
	// First, create valid compressed data
	original := []byte("some data to compress")
	var compressedBuf bytes.Buffer
	zlibWriter := zlib.NewWriter(&compressedBuf)
	if _, err := zlibWriter.Write(original); err != nil {
		t.Fatalf("zlib.Write() error: %v", err)
	}
	if err := zlibWriter.Close(); err != nil {
		t.Fatalf("zlib.Close() error: %v", err)
	}
	compressed := compressedBuf.Bytes()

	// Corrupt the body (keep first 2 bytes which are the zlib header)
	if len(compressed) > 4 {
		for i := 2; i < len(compressed); i++ {
			compressed[i] = 0xFF
		}
	}

	reader := bytes.NewReader(compressed)
	_, err := decompress(reader, uint32(len(compressed)))
	if err == nil {
		t.Error("decompress() expected error for corrupted zlib data, got nil")
	}
}

// TestDecompress_TruncatedReader tests that a reader shorter than compressedSize produces an error.
func TestDecompress_TruncatedReader(t *testing.T) {
	// Provide only 3 bytes but claim size is 100
	reader := bytes.NewReader([]byte{0x78, 0x9C, 0x00})
	_, err := decompress(reader, 100)
	if err == nil {
		t.Error("decompress() expected error for truncated reader, got nil")
	}
}
