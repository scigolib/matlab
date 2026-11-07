package matlab

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/scigolib/hdf5"
	"github.com/scigolib/matlab/types"
)

// TestCreate_V73WritesMATHeader checks the layout MATLAB writes: a 512-byte
// user block whose first 128 bytes are the MAT header, HDF5 data behind it.
func TestCreate_V73WritesMATHeader(t *testing.T) {
	path := filepath.Join(t.TempDir(), "header.mat")

	w, err := Create(path, Version73)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	want := &types.Variable{
		Name:       "x",
		Dimensions: []int{3},
		DataType:   types.Double,
		Data:       []float64{1, 2, 3},
	}
	if err := w.WriteVariable(want); err != nil {
		t.Fatalf("WriteVariable: %v", err)
	}
	if err := w.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(raw[:19], []byte("MATLAB 7.3 MAT-file")) {
		t.Errorf("header text = %q", raw[:19])
	}
	if !bytes.Equal(raw[124:128], []byte{0x00, 0x02, 'I', 'M'}) {
		t.Errorf("version/endian = % x", raw[124:128])
	}
	if !bytes.Equal(raw[512:520], []byte("\x89HDF\r\n\x1a\n")) {
		t.Errorf("bytes 512-519 = % x, want HDF5 signature", raw[512:520])
	}

	mf := openBytes(t, raw)
	if mf.Version != "7.3" {
		t.Errorf("Version = %q", mf.Version)
	}
	if len(mf.Variables) != 1 {
		t.Fatalf("got %d variables", len(mf.Variables))
	}
	got := mf.Variables[0]
	if got.Name != "x" || got.DataType != types.Double {
		t.Errorf("got %s %v", got.Name, got.DataType)
	}

	f, err := hdf5.Open(path)
	if err != nil {
		t.Fatalf("hdf5.Open: %v", err)
	}
	defer func() { _ = f.Close() }()
	if ws := f.Warnings(); len(ws) != 0 {
		t.Errorf("hdf5 warnings: %v", ws)
	}
}
