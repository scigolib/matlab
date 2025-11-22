package matlab

import (
	"os"
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
