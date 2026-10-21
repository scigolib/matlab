package v5

import (
	"bytes"
	"encoding/binary"
	"io"
	"runtime"
	"strings"
	"testing"
)

// testClaim is the payload size a crafted tag claims while the stream carries
// nothing: large enough to measure, small enough that a pre-fix run cannot OOM
// the test runner.
const testClaim = 4 << 20 // 4 MiB

// allocTolerance is the headroom over the input size allowed for the runtime's
// own allocations while parsing. A pre-fix run allocates testClaim bytes, which
// is three orders of magnitude above it.
const allocTolerance = 1 << 20 // 1 MiB

// testHeader builds a 128-byte little-endian MAT-file header.
func testHeader() []byte {
	header := make([]byte, 128)
	header[126], header[127] = 'I', 'M'
	return header
}

// regularTag encodes an 8-byte regular-format data element tag.
func regularTag(dataType, size uint32) []byte {
	tag := make([]byte, 8)
	binary.LittleEndian.PutUint32(tag[0:4], dataType)
	binary.LittleEndian.PutUint32(tag[4:8], size)
	return tag
}

// streamBytes joins parts into one MAT-file byte stream.
func streamBytes(parts ...[]byte) []byte {
	return bytes.Join(parts, nil)
}

// allocatedBytes reports how many bytes the process allocated while f ran.
// TotalAlloc is cumulative and never decreases, so the delta over f is the cost
// of f alone. The test binary runs its cases sequentially.
func allocatedBytes(f func()) uint64 {
	var before, after runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&before)
	f()
	runtime.ReadMemStats(&after)
	return after.TotalAlloc - before.TotalAlloc
}

// parseWithNewParser runs the whole parse inside an allocation measurement and
// reports the bytes allocated along with the parse error.
func parseWithNewParser(in []byte) (uint64, error) {
	var err error
	alloc := allocatedBytes(func() {
		var parser *Parser
		parser, err = NewParser(bytes.NewReader(in))
		if err == nil {
			_, err = parser.Parse()
		}
	})

	return alloc, err
}

// TestParser_TruncatedPayloadDoesNotAllocate is the regression test for the
// allocation-amplification class (hdf5 issue #68, ADR-003): a tag's size comes
// from the file and must be checked against the bytes actually present before
// it is used to size an allocation.
func TestParser_TruncatedPayloadDoesNotAllocate(t *testing.T) {
	in := streamBytes(testHeader(), regularTag(miMATRIX, testClaim))

	alloc, err := parseWithNewParser(in)

	if err == nil {
		t.Fatal("Parse() accepted a tag whose payload is not in the stream")
	}

	if alloc > allocTolerance {
		t.Errorf("parsing %d bytes that claim %d allocated %d bytes: the tag size must be validated against the stream before use (ADR-003)",
			len(in), testClaim, alloc)
	}
}

// TestParser_UnknownSizeTruncatedPayload covers a reader that cannot report its
// length (io.MultiReader and readers embedding another io.Reader do that): no
// region check is possible, so the payload must be accumulated as it arrives and
// a truncated stream must not cost the claimed size.
func TestParser_UnknownSizeTruncatedPayload(t *testing.T) {
	in := streamBytes(testHeader(), regularTag(miMATRIX, testClaim))

	// Embedding io.Reader hides the concrete type's Size method.
	var hidden io.Reader = struct{ io.Reader }{bytes.NewReader(in)}

	var err error
	alloc := allocatedBytes(func() {
		var parser *Parser
		parser, err = NewParser(hidden)
		if err == nil {
			_, err = parser.Parse()
		}
	})

	if err == nil {
		t.Fatal("Parse() accepted a tag whose payload is not in the stream")
	}
	if alloc > allocTolerance {
		t.Errorf("parsing %d bytes of unknown length that claim %d allocated %d bytes", len(in), testClaim, alloc)
	}
}

// TestParser_MatrixElementBeyondPayload verifies that an inner element claiming
// more than the matrix payload it lives in is rejected, not allocated.
func TestParser_MatrixElementBeyondPayload(t *testing.T) {
	var payload bytes.Buffer

	// Array flags: class double, not complex.
	flags := make([]byte, 8)
	binary.LittleEndian.PutUint32(flags[0:4], mxDOUBLE_CLASS)
	payload.Write(regularTag(miUINT32, 8))
	payload.Write(flags)

	// Dimensions: 1x1.
	dims := make([]byte, 8)
	binary.LittleEndian.PutUint32(dims[0:4], 1)
	binary.LittleEndian.PutUint32(dims[4:8], 1)
	payload.Write(regularTag(miINT32, 8))
	payload.Write(dims)

	// Name: one byte plus padding to the 8-byte boundary.
	name := make([]byte, 8)
	name[0] = 'x'
	payload.Write(regularTag(miINT8, 1))
	payload.Write(name)

	// Real data: claims a payload the matrix does not contain.
	payload.Write(regularTag(miDOUBLE, testClaim))

	in := streamBytes(testHeader(), regularTag(miMATRIX, uint32(payload.Len())), payload.Bytes())

	alloc, err := parseWithNewParser(in)

	if err == nil {
		t.Fatal("Parse() accepted a matrix element whose data tag points past the matrix payload")
	}

	if alloc > allocTolerance {
		t.Errorf("parsing a matrix whose inner tag claims %d allocated %d bytes: the inner tag must be bounded by the matrix payload (ADR-003)",
			testClaim, alloc)
	}
}

// TestParser_CompressedPayloadBeyondInput verifies the compressed element path
// is bounded by the same rule as every other payload.
func TestParser_CompressedPayloadBeyondInput(t *testing.T) {
	in := streamBytes(testHeader(), regularTag(miCOMPRESSED, testClaim))

	alloc, err := parseWithNewParser(in)

	if err == nil {
		t.Fatal("Parse() accepted a compressed element whose data is not in the stream")
	}

	if alloc > allocTolerance {
		t.Errorf("parsing %d bytes that claim %d compressed bytes allocated %d bytes: the compressed size must be bounded by the stream (ADR-003)",
			len(in), testClaim, alloc)
	}
}

// TestParser_PayloadBeyondLimit pins the region arithmetic: a payload that
// exactly fits the stream is accepted, one byte more is rejected with the
// end-of-file error, and the parser never reads from the stream in the second
// case.
func TestParser_PayloadBeyondLimit(t *testing.T) {
	tests := []struct {
		name    string
		size    uint32
		wantErr string
	}{
		{name: "exact fit", size: 4},
		{name: "one byte beyond", size: 5, wantErr: "beyond end of file"},
		{name: "far beyond", size: 1 << 20, wantErr: "beyond end of file"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checkPayloadFits(t, tt.size, tt.wantErr)
		})
	}
}

// TestParserSized_ExplicitLimit covers the path matlab.Open takes: the reader
// hides the stream length (io.MultiReader does), so the bound is supplied
// explicitly to the parser instead of being detected from the reader.
func TestParserSized_ExplicitLimit(t *testing.T) {
	in := streamBytes(testHeader(), regularTag(miMATRIX, testClaim))

	// Embedding io.Reader keeps the concrete type's Size method out of the
	// method set, reproducing what io.MultiReader does to the caller's reader.
	var hidden io.Reader = struct{ io.Reader }{bytes.NewReader(in)}

	parser, err := NewParserSized(hidden, int64(len(in)))
	if err != nil {
		t.Fatalf("NewParserSized() error: %v", err)
	}

	_, err = parser.Parse()

	if err == nil {
		t.Fatal("Parse() accepted a tag whose payload is not in the stream")
	}
	if !strings.Contains(err.Error(), "beyond end of file") {
		t.Errorf("Parse() error = %q, want it to report the payload beyond the explicit limit", err)
	}
}

// TestReaderSize pins which readers report their size and which do not.
func TestReaderSize(t *testing.T) {
	tests := []struct {
		name   string
		reader io.Reader
		size   int64
		known  bool
	}{
		{name: "bytes.Reader", reader: bytes.NewReader(make([]byte, 42)), size: 42, known: true},
		{name: "strings.Reader", reader: strings.NewReader("0123456789"), size: 10, known: true},
		{name: "MultiReader hides the size", reader: io.MultiReader(bytes.NewReader(make([]byte, 7))), known: false},
		{name: "embedded reader hides the size", reader: struct{ io.Reader }{bytes.NewReader(make([]byte, 7))}, known: false},
		{name: "nil", reader: nil, known: false},
		{name: "typed nil", reader: (*bytes.Reader)(nil), known: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			size, known := ReaderSize(tt.reader)
			if known != tt.known {
				t.Fatalf("ReaderSize() known = %v, want %v", known, tt.known)
			}
			if known && size != tt.size {
				t.Errorf("ReaderSize() = %d, want %d", size, tt.size)
			}
		})
	}
}

// checkPayloadFits parses a tag whose payload is the only thing after the
// header: four bytes of real data. wantErr is the substring the read error must
// contain; empty means the tag is expected to be accepted.
func checkPayloadFits(t *testing.T, size uint32, wantErr string) {
	t.Helper()

	const available = 4
	in := streamBytes(testHeader(), regularTag(miMATRIX, size), make([]byte, available))

	parser, err := NewParser(bytes.NewReader(in))
	if err != nil {
		t.Fatalf("NewParser() error: %v", err)
	}

	tag, err := parser.readTag()

	if wantErr == "" {
		if err != nil {
			t.Fatalf("readTag() error for a tag that fits the stream: %v", err)
		}
		if tag.Size != size {
			t.Errorf("tag.Size = %d, want %d", tag.Size, size)
		}
		return
	}

	if err == nil {
		t.Fatalf("readTag() accepted a tag of %d bytes with %d available", size, available)
	}
	if !strings.Contains(err.Error(), wantErr) {
		t.Errorf("readTag() error = %q, want it to contain %q", err, wantErr)
	}
}
