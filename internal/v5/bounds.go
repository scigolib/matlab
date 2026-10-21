package v5

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"reflect"
)

// Payload bounds.
//
// MAT v5 stores an element's size as a uint32, so the file itself can claim up
// to 4 GiB for one element. The header is untrusted input: an allocation sized
// from it is a memory-exhaustion vector (the class fixed in hdf5 v0.14.2,
// GHSA-8p88-h28q-w6m9). Two independent bounds apply before a size is used:
//
//   - an absolute cap (maxTagSize) that holds even when the stream length is
//     unknown, and
//   - a region check against the bytes actually available (Parser.limit), which
//     is the real bound wherever the input length can be determined.

// maxTagSize is the hard cap on a single element payload: 2 GiB. It is typed
// because an untyped 2^31 does not fit in int on 32-bit targets, where using it
// as an int — as a fmt argument, for example — fails to compile.
const maxTagSize uint32 = 2 << 30

// maxInt is the largest value representable by int on the target platform.
// make([]byte, n) panics when n exceeds it, and a panic is a crash rather than
// an error, so a 32-bit build must reject such a size before allocating.
const maxInt = int(^uint(0) >> 1)

// ReaderSize reports the total size of r when the reader can tell, and
// (0, false) when it cannot.
//
// This mirrors H5FD_get_eof and hdf5's utils.ReaderSize: the format-level end
// of the stream when the driver knows it (Stat for *os.File, Size for
// *bytes.Reader and io.SectionReader), otherwise nothing. The value is an upper
// bound, not an invariant: a reader positioned past the start of its backing
// store reports the whole store, and a file may be truncated after the call.
// An over-stated bound only weakens the check; it never lets an allocation
// exceed the real input.
func ReaderSize(r io.Reader) (int64, bool) {
	if isNilReader(r) {
		return 0, false
	}

	switch s := r.(type) {
	case interface{ Stat() (fs.FileInfo, error) }:
		info, err := s.Stat()
		if err != nil || info == nil {
			return 0, false
		}
		if info.Size() < 0 {
			return 0, false
		}
		return info.Size(), true
	case interface{ Size() int64 }:
		if s.Size() < 0 {
			return 0, false
		}
		return s.Size(), true
	case interface{ Len() int }:
		return int64(s.Len()), true
	default:
		return 0, false
	}
}

// isNilReader reports whether r is nil, including a typed nil pointer such as
// (*bytes.Reader)(nil). The latter is a non-nil interface whose method call
// would panic, so the plain r == nil test is not enough.
func isNilReader(r io.Reader) bool {
	if r == nil {
		return true
	}

	v := reflect.ValueOf(r)
	switch v.Kind() { //nolint:exhaustive // Only nil-able kinds need the IsNil call.
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return v.IsNil()
	default:
		return false
	}
}

// validatePayload checks a size taken from a tag before it is used to allocate,
// skip or read. readTag is the only producer of DataTag values, so validating
// there gives every consumer the same guarantee.
//
// p.pos already includes the 8 tag bytes when this runs; the reported offset is
// the one the tag starts at.
func (p *Parser) validatePayload(size uint32) error {
	if size > maxTagSize {
		return fmt.Errorf("tag size too large: %d bytes (max %d)", size, maxTagSize)
	}

	if uint64(size) > uint64(maxInt) {
		return fmt.Errorf("tag size %d exceeds the platform limit (max %d)", size, maxInt)
	}

	if p.limit > 0 {
		remaining := p.limit - p.pos
		if remaining < 0 || int64(size) > remaining {
			return fmt.Errorf("tag size beyond end of file: %d bytes at offset %d, %d available",
				size, p.pos-8, remaining)
		}
	}

	return nil
}

// readPayload reads a validated payload of size bytes from the stream.
//
// The size must have passed validatePayload — readTag guarantees that for every
// DataTag. The guarantee is what makes the exact allocation below bounded by
// the real input. When the stream length is unknown no region check was
// possible, so the data is accumulated as it arrives instead of being
// preallocated from the claim: a truncated stream then costs only the bytes it
// actually delivers.
func (p *Parser) readPayload(size uint32) ([]byte, error) {
	if size == 0 {
		return nil, nil
	}

	if p.limit > 0 {
		data := make([]byte, size)
		if _, err := io.ReadFull(p.r, data); err != nil {
			return nil, err
		}
		return data, nil
	}

	var buf bytes.Buffer
	if _, err := io.CopyN(&buf, p.r, int64(size)); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
