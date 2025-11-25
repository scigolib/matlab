// Package v5 implements MATLAB v5 format parser (v5-v7.2).
package v5

import (
	"bytes"
	"compress/zlib"
	"fmt"
	"io"
)

// maxDecompressedSize is the maximum allowed size after decompression (100MB).
// This prevents compression bomb attacks (zip bombs).
const maxDecompressedSize = 100 * 1024 * 1024 // 100MB

// maxCompressionRatio is the maximum allowed compression ratio.
// Typical zlib compression achieves 2:1 to 10:1 ratios.
// A ratio above 1000:1 suggests a potential zip bomb.
const maxCompressionRatio = 1000

// decompress decompresses zlib-compressed data from a MAT-file.
// It reads compressedSize bytes from r and returns the decompressed content.
//
// Security: Implements protection against compression bombs:
// - Maximum decompressed size limit (100MB).
// - Maximum compression ratio check (1000:1).
func decompress(r io.Reader, compressedSize uint32) ([]byte, error) {
	// Read compressed data
	compressed := make([]byte, compressedSize)
	if _, err := io.ReadFull(r, compressed); err != nil {
		return nil, fmt.Errorf("failed to read compressed data: %w", err)
	}

	// Create zlib reader
	zlibReader, err := zlib.NewReader(bytes.NewReader(compressed))
	if err != nil {
		return nil, fmt.Errorf("failed to create zlib reader: %w", err)
	}
	defer zlibReader.Close() //nolint:errcheck // Best effort cleanup

	// Read decompressed data with size limit
	var decompressed bytes.Buffer
	limited := io.LimitReader(zlibReader, maxDecompressedSize+1)
	n, err := io.Copy(&decompressed, limited)
	if err != nil {
		return nil, fmt.Errorf("failed to decompress data: %w", err)
	}

	// Check for size limit exceeded
	if n > maxDecompressedSize {
		return nil, fmt.Errorf("decompressed size exceeds limit: %d > %d bytes", n, maxDecompressedSize)
	}

	// Check compression ratio
	if compressedSize > 0 {
		ratio := float64(n) / float64(compressedSize)
		if ratio > maxCompressionRatio {
			return nil, fmt.Errorf("compression ratio too high: %.1f:1 (max %d:1)", ratio, maxCompressionRatio)
		}
	}

	return decompressed.Bytes(), nil
}
