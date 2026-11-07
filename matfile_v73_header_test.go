package matlab

import (
	"bytes"
	"encoding/binary"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

// matlabUserBlock builds the 512-byte block MATLAB puts in front of the HDF5
// superblock of a v7.3 MAT-file: a 128-byte MAT header (116 bytes of text,
// 8 bytes subsystem offset, 2-byte version, 2-byte endian indicator), zero
// padded to the HDF5 user-block size. version is written in the byte order
// named by endian ("IM" little-endian, "MI" big-endian).
func matlabUserBlock(t *testing.T, version uint16, endian string) []byte {
	t.Helper()

	block := make([]byte, 512)
	text := "MATLAB 7.3 MAT-file, Platform: GLNXA64, Created on: Mon Jan  1 00:00:00 2024 HDF5 schema 1.00 ."
	copy(block[:116], bytes.Repeat([]byte{' '}, 116))
	copy(block[:116], text)
	copy(block[116:124], bytes.Repeat([]byte{' '}, 8))

	var order binary.ByteOrder = binary.LittleEndian
	if endian == "MI" {
		order = binary.BigEndian
	}
	order.PutUint16(block[124:126], version)
	copy(block[126:128], endian)

	return block
}

func openBytes(t *testing.T, data []byte) *MatFile {
	t.Helper()

	mf, err := Open(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	return mf
}

// TestOpen_MATLABv73UserBlock opens every file our v7.3 writer produced with a
// MATLAB-style 512-byte user block prepended (the layout MATLAB itself writes,
// HDF5 signature at offset 512) and expects the same content as the plain file.
func TestOpen_MATLABv73UserBlock(t *testing.T) {
	files, err := filepath.Glob(filepath.Join("testdata", "generated", "*.mat"))
	if err != nil || len(files) == 0 {
		t.Fatalf("no generated test files: %v", err)
	}

	cases := []struct {
		name    string
		version uint16
		endian  string
	}{
		{"little-endian IM", 0x0200, "IM"},
		{"big-endian MI", 0x0200, "MI"},
	}

	for _, file := range files {
		orig, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		want := openBytes(t, orig)

		for _, tc := range cases {
			t.Run(filepath.Base(file)+"/"+tc.name, func(t *testing.T) {
				data := append(matlabUserBlock(t, tc.version, tc.endian), orig...)
				got := openBytes(t, data)

				if got.Version != "7.3" {
					t.Errorf("Version = %q, want 7.3", got.Version)
				}
				compareVariables(t, got, want)
			})
		}
	}
}

func compareVariables(t *testing.T, got, want *MatFile) {
	t.Helper()

	if len(got.Variables) != len(want.Variables) {
		t.Fatalf("got %d variables, want %d", len(got.Variables), len(want.Variables))
	}
	for i, wv := range want.Variables {
		gv := got.Variables[i]
		if gv.Name != wv.Name {
			t.Errorf("var %d: name %q, want %q", i, gv.Name, wv.Name)
		}
		if gv.DataType != wv.DataType {
			t.Errorf("%s: DataType %v, want %v", wv.Name, gv.DataType, wv.DataType)
		}
		if !reflect.DeepEqual(gv.Dimensions, wv.Dimensions) {
			t.Errorf("%s: Dimensions %v, want %v", wv.Name, gv.Dimensions, wv.Dimensions)
		}
		if !reflect.DeepEqual(gv.Data, wv.Data) {
			t.Errorf("%s: Data %v, want %v", wv.Name, gv.Data, wv.Data)
		}
	}
}

// TestOpen_UnsupportedVersion checks that a valid endian indicator with a
// version other than 0x0100 or 0x0200 is rejected, as scipy and matio do.
func TestOpen_UnsupportedVersion(t *testing.T) {
	for _, endian := range []string{"IM", "MI"} {
		t.Run(endian, func(t *testing.T) {
			header := matlabUserBlock(t, 0x0300, endian)[:128]
			_, err := Open(bytes.NewReader(header))
			if !errors.Is(err, ErrUnsupportedVersion) {
				t.Fatalf("err = %v, want ErrUnsupportedVersion", err)
			}
		})
	}
}
