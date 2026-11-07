package matlab

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/scigolib/hdf5"
	"github.com/scigolib/matlab/types"
)

// hdf5Warnings opens path with the HDF5 library, visits every object (each
// dataset's datatype, attributes and data) and returns the warnings collected.
// The warnings grow as structures are read, so the whole file is walked.
func hdf5Warnings(t *testing.T, path string) []hdf5.Warning {
	t.Helper()

	f, err := hdf5.Open(path)
	if err != nil {
		t.Fatalf("hdf5.Open(%s): %v", path, err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			t.Errorf("Close(%s): %v", path, err)
		}
	}()

	datasets := 0
	f.Walk(func(objPath string, obj hdf5.Object) {
		ds, ok := obj.(*hdf5.Dataset)
		if !ok {
			return
		}
		datasets++
		if _, err := ds.Datatype(); err != nil {
			t.Errorf("%s %s: Datatype: %v", path, objPath, err)
		}
		if _, err := ds.Attributes(); err != nil {
			t.Errorf("%s %s: Attributes: %v", path, objPath, err)
		}
		if _, err := ds.Read(); err != nil {
			t.Errorf("%s %s: Read: %v", path, objPath, err)
		}
	})
	if datasets == 0 {
		t.Fatalf("%s: walk visited no dataset", path)
	}

	return f.Warnings()
}

// TestGeneratedFixtures_NoHDF5Warnings guards the fixtures written by our v7.3
// writer: the HDF5 library reports no warning for a file in a form the HDF5 C
// tools can open. A warning here means the fixtures are stale (written by an
// older writer) and need `go run ./scripts/generate-testdata`.
func TestGeneratedFixtures_NoHDF5Warnings(t *testing.T) {
	paths, err := filepath.Glob(filepath.Join("testdata", "generated", "*.mat"))
	if err != nil {
		t.Fatal(err)
	}
	if len(paths) == 0 {
		t.Fatal("no fixtures found in testdata/generated")
	}

	for _, path := range paths {
		t.Run(filepath.Base(path), func(t *testing.T) {
			for _, w := range hdf5Warnings(t, path) {
				t.Errorf("unexpected HDF5 warning: %s", w)
			}
		})
	}
}

// TestOpen_LegacyHDF5GoFile reads a v7.3 file written by our writer on
// hdf5-go v0.11.4-beta (matlab v0.1.0-beta). That release stored a CRC32
// superblock checksum, object headers without a valid checksum and an early
// datatype encoding, which the HDF5 C tools reject. Files like it exist in users' hands, so they must keep reading
// correctly; the HDF5 library reports them through warnings.
func TestOpen_LegacyHDF5GoFile(t *testing.T) {
	path := filepath.Join("testdata", "legacy", "simple_double_hdf5go_v0.11.mat")

	if len(hdf5Warnings(t, path)) == 0 {
		t.Fatal("expected HDF5 warnings for a file written by hdf5-go v0.11.4-beta")
	}

	file, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = file.Close() }()

	mf, err := Open(file)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}

	if len(mf.Variables) != 1 {
		t.Fatalf("got %d variables, want 1", len(mf.Variables))
	}
	v := mf.Variables[0]
	if v.Name != "data" || v.DataType != types.Double {
		t.Fatalf("got %q %v, want \"data\" double", v.Name, v.DataType)
	}
	got, ok := v.Data.([]float64)
	if !ok {
		t.Fatalf("Data is %T, want []float64", v.Data)
	}
	want := []float64{1, 2, 3, 4, 5}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got %v, want %v", got, want)
		}
	}
}
