package v5

import (
	"encoding/binary"
	"errors"
	"strings"
)

// Header represents a MAT-file header.
type Header struct {
	Description     string           // File description
	Version         uint16           // MAT-file version
	EndianIndicator string           // Endian indicator ("MI" or "IM")
	Order           binary.ByteOrder // Byte order
}

// parseHeader parses the MAT-file header.
func parseHeader(data []byte) (*Header, error) {
	hdr := &Header{
		Description:     strings.TrimRight(string(data[:116]), "\x00"),
		EndianIndicator: string(data[126:128]),
	}

	// Determine byte order
	switch hdr.EndianIndicator {
	case "MI":
		hdr.Order = binary.LittleEndian
	case "IM":
		hdr.Order = binary.BigEndian
	default:
		return nil, errors.New("invalid endian indicator")
	}

	// Parse version
	hdr.Version = hdr.Order.Uint16(data[124:126])
	return hdr, nil
}
