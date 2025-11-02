// Package v5 implements MATLAB v5 format parser (v5-v7.2).
package v5

import (
	"errors"
	"io"
)

// ErrCompressedNotSupported indicates compressed data is not supported.
var ErrCompressedNotSupported = errors.New("compressed MAT-files not yet supported")

// decompress would handle decompression (stub for future implementation).
//
//nolint:unused // Future implementation stub
func decompress(_ io.Reader) (io.Reader, error) {
	return nil, ErrCompressedNotSupported
}
