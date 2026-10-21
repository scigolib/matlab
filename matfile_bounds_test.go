package matlab

import (
	"bytes"
	"encoding/binary"
	"runtime"
	"testing"
)

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
