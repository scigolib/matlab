package v5

import (
	"io"
)

// DataTag represents a data element tag.
type DataTag struct {
	DataType uint32 // Data type identifier
	Size     uint32 // Data size in bytes
	IsSmall  bool   // True for small data elements
}

// readTag reads a data tag from the stream.
func (p *Parser) readTag() (*DataTag, error) {
	buf := make([]byte, 8)
	if _, err := io.ReadFull(p.r, buf); err != nil {
		return nil, err
	}
	p.pos += 8

	firstWord := p.Header.Order.Uint32(buf[:4])
	if firstWord == 0xffffffff {
		return p.readLargeTag(buf[4:])
	}
	return p.readSmallTag(firstWord)
}

// readSmallTag reads a small data element tag.
func (p *Parser) readSmallTag(tagPrefix uint32) (*DataTag, error) {
	return &DataTag{
		DataType: tagPrefix & 0x0000ffff,
		Size:     tagPrefix >> 16,
		IsSmall:  true,
	}, nil
}

// readLargeTag reads a large data element tag.
func (p *Parser) readLargeTag(suffix []byte) (*DataTag, error) {
	buf := make([]byte, 8)
	if _, err := io.ReadFull(p.r, buf); err != nil {
		return nil, err
	}
	p.pos += 8

	return &DataTag{
		DataType: p.Header.Order.Uint32(buf[:4]),
		Size:     p.Header.Order.Uint32(suffix),
		IsSmall:  false,
	}, nil
}
