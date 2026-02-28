package matlab

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/scigolib/matlab/types"
)

// TestMatFile_GetVariable tests retrieving a variable by name.
func TestMatFile_GetVariable(t *testing.T) {
	// Open a test file.
	file, err := os.Open("testdata/generated/simple_double.mat")
	if err != nil {
		t.Fatalf("Failed to open test file: %v", err)
	}
	defer file.Close()

	matFile, err := Open(file)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}

	tests := []struct {
		name     string
		varName  string
		wantNil  bool
		wantName string
	}{
		{
			name:     "existing variable",
			varName:  "data",
			wantNil:  false,
			wantName: "data",
		},
		{
			name:    "non-existent variable",
			varName: "nonexistent",
			wantNil: true,
		},
		{
			name:    "empty string",
			varName: "",
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := matFile.GetVariable(tt.varName)
			if (got == nil) != tt.wantNil {
				t.Errorf("GetVariable(%q) = %v, wantNil = %v", tt.varName, got, tt.wantNil)
			}
			if got != nil && got.Name != tt.wantName {
				t.Errorf("GetVariable(%q).Name = %q, want %q", tt.varName, got.Name, tt.wantName)
			}
		})
	}
}

// TestMatFile_GetVariableNames tests listing all variable names.
func TestMatFile_GetVariableNames(t *testing.T) {
	file, err := os.Open("testdata/generated/simple_double.mat")
	if err != nil {
		t.Fatalf("Failed to open test file: %v", err)
	}
	defer file.Close()

	matFile, err := Open(file)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}

	names := matFile.GetVariableNames()
	if len(names) != 1 {
		t.Errorf("GetVariableNames() returned %d names, want 1", len(names))
	}
	if len(names) > 0 && names[0] != "data" {
		t.Errorf("GetVariableNames()[0] = %q, want %q", names[0], "data")
	}
}

// TestMatFile_GetVariableNames_Empty tests with empty MatFile.
func TestMatFile_GetVariableNames_Empty(t *testing.T) {
	matFile := &MatFile{
		Variables: []*types.Variable{},
	}

	names := matFile.GetVariableNames()
	if len(names) != 0 {
		t.Errorf("GetVariableNames() returned %d names, want 0", len(names))
	}
}

// TestMatFile_HasVariable tests checking if a variable exists.
func TestMatFile_HasVariable(t *testing.T) {
	file, err := os.Open("testdata/generated/simple_double.mat")
	if err != nil {
		t.Fatalf("Failed to open test file: %v", err)
	}
	defer file.Close()

	matFile, err := Open(file)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}

	tests := []struct {
		name    string
		varName string
		want    bool
	}{
		{
			name:    "existing variable",
			varName: "data",
			want:    true,
		},
		{
			name:    "non-existent variable",
			varName: "nonexistent",
			want:    false,
		},
		{
			name:    "empty string",
			varName: "",
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := matFile.HasVariable(tt.varName)
			if got != tt.want {
				t.Errorf("HasVariable(%q) = %v, want %v", tt.varName, got, tt.want)
			}
		})
	}
}

// TestOpen_TooShortData tests Open with data shorter than 128 bytes.
// io.ReadFull should fail with io.ErrUnexpectedEOF.
func TestOpen_TooShortData(t *testing.T) {
	tests := []struct {
		name string
		data []byte
	}{
		{"empty", []byte{}},
		{"3 bytes", []byte{1, 2, 3}},
		{"127 bytes", make([]byte, 127)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Open(bytes.NewReader(tt.data))
			if err == nil {
				t.Error("Open() expected error for short data, got nil")
			}
		})
	}
}

// TestOpen_UnrecognizedFormat tests Open with 128+ bytes that are neither HDF5 nor v5.
func TestOpen_UnrecognizedFormat(t *testing.T) {
	// 128 bytes of zeros: no HDF5 signature, no "MI"/"IM" at bytes 126-127.
	data := make([]byte, 256)
	_, err := Open(bytes.NewReader(data))
	if err == nil {
		t.Fatal("Open() expected error for unrecognized format, got nil")
	}
	if !errors.Is(err, ErrInvalidFormat) {
		t.Errorf("Open() error = %v, want ErrInvalidFormat", err)
	}
}

// TestOpen_V5HeaderZeroBody tests Open with a valid v5 header but no data elements.
// The parser succeeds but returns zero variables.
func TestOpen_V5HeaderZeroBody(t *testing.T) {
	// Build a 128-byte header with "MI" at bytes 126-127 and empty body.
	// With an empty body, the parser reads EOF immediately and returns no variables.
	data := make([]byte, 128)
	data[126] = 'M'
	data[127] = 'I'

	matFile, err := Open(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("Open() unexpected error = %v", err)
	}
	if len(matFile.Variables) != 0 {
		t.Errorf("Variables count = %d, want 0", len(matFile.Variables))
	}
	if matFile.Version != "5.0" {
		t.Errorf("Version = %q, want %q", matFile.Version, "5.0")
	}
}

// TestOpen_InvalidHDF5Data tests Open with valid HDF5 signature but garbage body.
func TestOpen_InvalidHDF5Data(t *testing.T) {
	// HDF5 signature: 0x89 0x48 0x44 0x46 0x0d 0x0a 0x1a 0x0a
	data := make([]byte, 256)
	copy(data, []byte{0x89, 0x48, 0x44, 0x46, 0x0d, 0x0a, 0x1a, 0x0a})
	// The rest is zeros - invalid HDF5 file, parser should fail.

	_, err := Open(bytes.NewReader(data))
	if err == nil {
		t.Fatal("Open() expected error for invalid HDF5 data, got nil")
	}
	// Error should come from v73 parsing (HDF5 library).
}

// TestOpen_V5File tests opening an actual v5 .mat file written by our v5 writer.
func TestOpen_V5File(t *testing.T) {
	// The testdata/generated/ files are v7.3 (HDF5) format.
	// Write a v5 file first, then open it.
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test_v5.mat")

	writer, err := Create(tmpFile, Version5)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	v := &types.Variable{
		Name:       "data",
		Dimensions: []int{3},
		DataType:   types.Double,
		Data:       []float64{1.0, 2.0, 3.0},
	}
	if err := writer.WriteVariable(v); err != nil {
		t.Fatalf("WriteVariable() error = %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	// Now open the v5 file.
	file, err := os.Open(tmpFile)
	if err != nil {
		t.Fatalf("Failed to open: %v", err)
	}
	defer file.Close()

	matFile, err := Open(file)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}

	if matFile.Version != "5.0" {
		t.Errorf("Version = %q, want %q", matFile.Version, "5.0")
	}

	if len(matFile.Variables) != 1 {
		t.Fatalf("Variables count = %d, want 1", len(matFile.Variables))
	}

	readVar := matFile.Variables[0]
	if readVar.Name != "data" {
		t.Errorf("Variable name = %q, want %q", readVar.Name, "data")
	}
	if readVar.DataType != types.Double {
		t.Errorf("DataType = %v, want %v", readVar.DataType, types.Double)
	}
}

// TestOpen_GeneratedHDF5Files tests opening the generated testdata files (HDF5 format).
func TestOpen_GeneratedHDF5Files(t *testing.T) {
	tests := []struct {
		filename string
		wantType types.DataType
	}{
		{"simple_double.mat", types.Double},
		{"simple_int32.mat", types.Int32},
		{"simple_single.mat", types.Single},
		{"simple_uint8.mat", types.Uint8},
		{"scalar.mat", types.Double},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			file, err := os.Open(filepath.Join("testdata", "generated", tt.filename))
			if err != nil {
				t.Fatalf("Failed to open: %v", err)
			}
			defer file.Close()

			matFile, err := Open(file)
			if err != nil {
				t.Fatalf("Open() error = %v", err)
			}

			// These files are HDF5 (v7.3) format.
			if matFile.Version != "7.3" {
				t.Errorf("Version = %q, want %q", matFile.Version, "7.3")
			}

			if len(matFile.Variables) < 1 {
				t.Fatal("Expected at least 1 variable")
			}

			v := matFile.Variables[0]
			if v.DataType != tt.wantType {
				t.Errorf("DataType = %v, want %v", v.DataType, tt.wantType)
			}
		})
	}
}

// TestOpen_V73File tests writing a v73 file then reading it back via Open.
func TestOpen_V73File(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test_open_v73.mat")

	// Write a v73 file.
	writer, err := Create(tmpFile, Version73)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	original := &types.Variable{
		Name:       "x",
		Dimensions: []int{4},
		DataType:   types.Double,
		Data:       []float64{10.0, 20.0, 30.0, 40.0},
	}

	if err := writer.WriteVariable(original); err != nil {
		t.Fatalf("WriteVariable() error = %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	// Read it back via Open.
	file, err := os.Open(tmpFile)
	if err != nil {
		t.Fatalf("Failed to open: %v", err)
	}
	defer file.Close()

	matFile, err := Open(file)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}

	if matFile.Version != "7.3" {
		t.Errorf("Version = %q, want %q", matFile.Version, "7.3")
	}
	if len(matFile.Variables) != 1 {
		t.Fatalf("Variables count = %d, want 1", len(matFile.Variables))
	}

	v := matFile.Variables[0]
	if v.Name != "x" {
		t.Errorf("Variable name = %q, want %q", v.Name, "x")
	}
	if v.DataType != types.Double {
		t.Errorf("DataType = %v, want %v", v.DataType, types.Double)
	}

	data, ok := v.Data.([]float64)
	if !ok {
		t.Fatalf("Data type = %T, want []float64", v.Data)
	}
	if len(data) != 4 {
		t.Fatalf("Data length = %d, want 4", len(data))
	}
	for i, want := range []float64{10.0, 20.0, 30.0, 40.0} {
		if data[i] != want {
			t.Errorf("Data[%d] = %v, want %v", i, data[i], want)
		}
	}
}

// TestOpen_ReaderError tests Open when the underlying reader returns an error.
func TestOpen_ReaderError(t *testing.T) {
	r := &errReader{err: errors.New("disk read error")}
	_, err := Open(r)
	if err == nil {
		t.Fatal("Open() expected error from failing reader, got nil")
	}
	if !strings.Contains(err.Error(), "disk read error") {
		t.Errorf("Open() error = %v, want to contain 'disk read error'", err)
	}
}

// errReader is an io.Reader that always returns an error.
type errReader struct {
	err error
}

func (r *errReader) Read([]byte) (int, error) {
	return 0, r.err
}
