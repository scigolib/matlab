package v73

import (
	"bytes"
	"runtime"
	"time"
)

// MATLAB writes a v7.3 MAT-file as an HDF5 file behind a 512-byte user block.
// The first 128 bytes of that block are the MAT header (the same layout as the
// v5 header), the rest is zero. matio does the same in Mat_Create73
// (src/mat73.c: H5Pset_userblock(512), then the header text, 8 spaces for the
// subsystem offset, version 0x0200 and the endian indicator).
const (
	// userBlockSize is the HDF5 user block size MATLAB uses.
	userBlockSize = 512
	// matHeaderSize is the MAT header length: text, subsystem offset, version, endian.
	matHeaderSize = 128
	// matHeaderTextSize is the length of the descriptive text field.
	matHeaderTextSize = 116
	// matVersion73 is the version field value that marks a v7.3 file.
	matVersion73 = 0x0200
)

// matHeaderText is the descriptive text in MATLAB's own wording. Even the
// longest GOOS name ("dragonfly") leaves it under matHeaderTextSize, so the
// text is never truncated; the tests pin that.
func matHeaderText(goos string, now time.Time) string {
	return "MATLAB 7.3 MAT-file, Platform: " + goos + ", Created on: " +
		now.UTC().Format(time.ANSIC) + " HDF5 schema 1.00 ."
}

// matHeader73 builds the 128-byte MAT header of a v7.3 file. The text is
// space padded to 116 bytes, the subsystem offset (bytes 116-123) is spaces,
// and bytes 124-127 hold the version and the "IM" endian indicator. The
// version is written little-endian, so the bytes are 00 02 'I' 'M', which is
// what matio writes on a little-endian host.
func matHeader73(now time.Time) []byte {
	h := bytes.Repeat([]byte{' '}, matHeaderSize)
	copy(h[:matHeaderTextSize], matHeaderText(runtime.GOOS, now))
	h[124] = matVersion73 & 0xff
	h[125] = matVersion73 >> 8
	h[126] = 'I'
	h[127] = 'M'
	return h
}
