package v5

import (
	"bytes"
	"testing"
)

// FuzzParse exercises the v5 parser on arbitrary input. Every size, count and
// offset it acts on comes from the file, so a panic here is a crash on a
// corrupt file — the class of bug that bounds validation exists to prevent.
func FuzzParse(f *testing.F) {
	flags := make([]byte, 8)
	dims := make([]byte, 8)

	f.Add(testHeader())
	f.Add(streamBytes(testHeader(), regularTag(miMATRIX, 0)))
	f.Add(streamBytes(testHeader(), regularTag(miMATRIX, 8), make([]byte, 8)))
	f.Add(streamBytes(testHeader(), regularTag(miMATRIX, 16), flags, dims))
	f.Add(streamBytes(testHeader(), regularTag(miCOMPRESSED, 4), make([]byte, 4)))
	f.Add(streamBytes(testHeader(), regularTag(miUINT32, 0xFFFFFFFF)))

	f.Fuzz(func(_ *testing.T, data []byte) {
		parser, err := NewParser(bytes.NewReader(data))
		if err != nil {
			return
		}

		// The parser must return an error or a value; a panic is the failure.
		_, _ = parser.Parse()
	})
}
