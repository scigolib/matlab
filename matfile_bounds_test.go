package matlab

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"testing"
)

// readerKinds wraps the same bytes in each kind of reader Open must accept. The
// size probe takes a different path for each: Stat for a file, Size for
// bytes.Reader, Len (the unread remainder) for bytes.Buffer, nothing at all for
// a reader that hides its type.
func readerKinds(t *testing.T, data []byte) map[string]io.Reader {
	t.Helper()

	path := filepath.Join(t.TempDir(), "in.mat")
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = f.Close() })

	return map[string]io.Reader{
		"os.File":      f,
		"bytes.Reader": bytes.NewReader(data),
		"bytes.Buffer": bytes.NewBuffer(data),
		"unknown size": struct{ io.Reader }{bytes.NewReader(data)},
	}
}

// TestOpen_EveryReaderKind opens the v5 reference files and the v7.3 test
// files through every kind of reader. The stream size must be measured before
// Open consumes the 128-byte header: a bytes.Buffer reports only its unread remainder, so a size taken
// after the header understated the file by 128 bytes and valid files were
// rejected as running past their end.
func TestOpen_EveryReaderKind(t *testing.T) {
	v5Files, err := filepath.Glob(filepath.Join("testdata", "scipy", "*.mat"))
	if err != nil {
		t.Fatal(err)
	}
	v73Files, err := filepath.Glob(filepath.Join("testdata", "generated", "*.mat"))
	if err != nil {
		t.Fatal(err)
	}
	if len(v5Files) == 0 || len(v73Files) == 0 {
		t.Fatalf("test files missing: %d v5, %d v7.3", len(v5Files), len(v73Files))
	}
	paths := slices.Concat(v5Files, v73Files)

	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		for kind, r := range readerKinds(t, data) {
			t.Run(filepath.Base(path)+"/"+kind, func(t *testing.T) {
				mf, err := Open(r)
				if err != nil {
					t.Fatalf("Open() error = %v", err)
				}
				if len(mf.Variables) == 0 {
					t.Error("Open() returned no variables")
				}
			})
		}
	}
}

// TestOpen_TruncatedFileIsUnexpectedEOF cuts a real file short and opens it
// through every kind of reader: whether the cut is caught by the size check
// (known length) or by the read itself (unknown length), the error is a
// truncated file, io.ErrUnexpectedEOF.
func TestOpen_TruncatedFileIsUnexpectedEOF(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("testdata", "scipy", "testdouble_7.4_GLNX86.mat"))
	if err != nil {
		t.Fatal(err)
	}
	cut := data[:len(data)-3]

	for kind, r := range readerKinds(t, cut) {
		t.Run(kind, func(t *testing.T) {
			_, err := Open(r)
			if err == nil {
				t.Fatal("Open() accepted a truncated file")
			}
			if !errors.Is(err, io.ErrUnexpectedEOF) {
				t.Errorf("Open() error = %v, want io.ErrUnexpectedEOF", err)
			}
		})
	}
}

// TestOpen_TruncatedTagDoesNotAllocate is the end-to-end regression test for
// allocations sized from file metadata (the class fixed in hdf5 v0.14.2,
// ADR-003): a 136-byte input whose element tag claims a large payload must fail
// without allocating the claim.
func TestOpen_TruncatedTagDoesNotAllocate(t *testing.T) {
	const claim = 64 << 20 // 64 MiB, large enough to measure and safe for the runner

	in := make([]byte, 128+8)
	in[126], in[127] = 'I', 'M'                    // little-endian v5 indicator
	binary.LittleEndian.PutUint32(in[128:132], 14) // miMATRIX
	binary.LittleEndian.PutUint32(in[132:136], claim)

	var before, after runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&before)

	_, err := Open(bytes.NewReader(in))

	runtime.ReadMemStats(&after)

	if err == nil {
		t.Fatal("Open() accepted a file whose element claims bytes the file does not contain")
	}

	if delta := after.TotalAlloc - before.TotalAlloc; delta > 8<<20 {
		t.Errorf("%d-byte input claiming %d bytes allocated %d: element sizes must be validated against the input before allocation",
			len(in), claim, delta)
	}
}
