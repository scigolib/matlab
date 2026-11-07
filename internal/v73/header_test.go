package v73

import (
	"bytes"
	"errors"
	"runtime"
	"strings"
	"testing"
	"time"
)

// fakeUserBlockFile records the calls writeMATHeader makes.
type fakeUserBlockFile struct {
	writeErr, closeErr error
	written            []byte
	closed             bool
}

func (f *fakeUserBlockFile) WriteUserBlock(data []byte) error {
	f.written = data
	return f.writeErr
}

func (f *fakeUserBlockFile) Close() error {
	f.closed = true
	return f.closeErr
}

// TestWriteMATHeader checks both outcomes: the header goes into the block and
// the file stays open; on a failed write the file is closed and both errors
// reach the caller.
func TestWriteMATHeader(t *testing.T) {
	now := time.Date(2026, 11, 7, 12, 0, 0, 0, time.UTC)

	ok := &fakeUserBlockFile{}
	if err := writeMATHeader(ok, now); err != nil {
		t.Fatalf("writeMATHeader() error = %v", err)
	}
	if !bytes.Equal(ok.written, matHeader73(now)) || ok.closed {
		t.Errorf("success: written %d bytes, closed=%v; want the header and an open file", len(ok.written), ok.closed)
	}

	writeErr := errors.New("block too small")
	closeErr := errors.New("close failed")
	bad := &fakeUserBlockFile{writeErr: writeErr, closeErr: closeErr}
	err := writeMATHeader(bad, now)
	if !bad.closed {
		t.Error("failure: the file was not closed")
	}
	if !errors.Is(err, writeErr) || !errors.Is(err, closeErr) {
		t.Errorf("failure: error = %v, want it to wrap both the write and the close error", err)
	}
}

var testNow = time.Date(2026, time.November, 7, 12, 30, 0, 0, time.UTC)

func TestMatHeader73(t *testing.T) {
	h := matHeader73(testNow)

	if len(h) != matHeaderSize {
		t.Fatalf("len = %d, want %d", len(h), matHeaderSize)
	}
	if !bytes.HasPrefix(h, []byte("MATLAB 7.3 MAT-file")) {
		t.Errorf("prefix = %q", h[:19])
	}
	text := string(h[:matHeaderTextSize])
	wantStart := "MATLAB 7.3 MAT-file, Platform: " + runtime.GOOS +
		", Created on: Sat Nov  7 12:30:00 2026 HDF5 schema 1.00 ."
	if !strings.HasPrefix(text, wantStart) {
		t.Errorf("text = %q, want prefix %q", text, wantStart)
	}
	if rest := text[len(wantStart):]; strings.Trim(rest, " ") != "" {
		t.Errorf("text after the schema marker is not spaces: %q", rest)
	}
	want := []byte("        \x00\x02IM")
	if !bytes.Equal(h[116:128], want) {
		t.Errorf("bytes 116-127 = %q, want %q", h[116:128], want)
	}
}

// TestMatHeaderText_NeverTruncated pins that the text fits the 116-byte field
// for every GOOS Go knows, the longest being "dragonfly".
func TestMatHeaderText_NeverTruncated(t *testing.T) {
	for _, goos := range []string{
		"windows", "linux", "darwin", "dragonfly", "freebsd", "netbsd", "openbsd",
		"solaris", "illumos", "plan9", "aix", "android", "ios", "js", "wasip1",
	} {
		t.Run(goos, func(t *testing.T) {
			text := matHeaderText(goos, testNow)
			if len(text) > matHeaderTextSize {
				t.Fatalf("len = %d, exceeds %d", len(text), matHeaderTextSize)
			}
			if !strings.HasSuffix(text, "HDF5 schema 1.00 .") {
				t.Errorf("text = %q", text)
			}
		})
	}
	if n := len(matHeaderText("dragonfly", testNow)); n != 97 {
		t.Errorf("dragonfly text is %d bytes, want 97", n)
	}
}
