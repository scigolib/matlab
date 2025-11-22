package v5

import (
	"fmt"
	"io"
)

// maxReasonableSize defines the maximum allowed tag size (2GB).
// This prevents memory exhaustion attacks from malicious MAT-files
// with extremely large size values.
const maxReasonableSize = 2 * 1024 * 1024 * 1024 // 2GB

// DataTag represents a data element tag.
type DataTag struct {
	DataType uint32 // Data type identifier
	Size     uint32 // Data size in bytes
	IsSmall  bool   // True for small data elements
}

// readTag reads a data tag from the stream.
//
// MAT-file v5 uses two tag formats:
//   - Small format (8 bytes total): Upper 16 bits of first word = size (1-4),
//     lower 16 bits = type, bytes 4-7 = packed data.
//   - Regular format (8 bytes tag + N bytes data): bytes 0-3 = type, bytes 4-7 = size.
func (p *Parser) readTag() (*DataTag, error) {
	buf := make([]byte, 8)
	if _, err := io.ReadFull(p.r, buf); err != nil {
		return nil, err
	}
	p.pos += 8

	firstWord := p.Header.Order.Uint32(buf[0:4])

	// Check for small format: upper 16 bits contain size (1-4)
	// Lower 16 bits contain data type
	size := firstWord >> 16
	if size > 0 && size <= 4 {
		// Small format
		dataType := firstWord & 0xFFFF
		return &DataTag{
			DataType: dataType,
			Size:     size,
			IsSmall:  true,
		}, nil
	}

	// Regular format: entire first word is type, second word is size
	dataType := firstWord
	size = p.Header.Order.Uint32(buf[4:8])

	// Validate size to prevent memory exhaustion attacks
	if size > maxReasonableSize {
		return nil, fmt.Errorf("tag size too large: %d bytes (max %d)", size, maxReasonableSize)
	}

	return &DataTag{
		DataType: dataType,
		Size:     size,
		IsSmall:  false,
	}, nil
}
