package v5

import (
	"encoding/binary"
	"testing"
)

func TestParseHeader(t *testing.T) {
	// Note: Endian indicator interpretation:
	// - "IM" = file created on little-endian system → use LittleEndian
	// - "MI" = file created on big-endian system → use BigEndian
	// This is because the 16-bit value 0x4D49 ("MI") is stored as [0x49, 0x4D]
	// on little-endian systems, which reads as "IM".
	tests := []struct {
		name        string
		header      []byte
		wantDesc    string
		wantVersion uint16
		wantEndian  string
		wantOrder   binary.ByteOrder
		wantErr     bool
	}{
		{
			name:        "valid little endian v5",
			header:      makeHeader("MATLAB 5.0 MAT-file", 0x0100, "IM"),
			wantDesc:    "MATLAB 5.0 MAT-file",
			wantVersion: 0x0100,
			wantEndian:  "IM",
			wantOrder:   binary.LittleEndian,
			wantErr:     false,
		},
		{
			name:        "valid big endian v5",
			header:      makeHeader("MATLAB 5.0 MAT-file", 0x0100, "MI"),
			wantDesc:    "MATLAB 5.0 MAT-file",
			wantVersion: 0x0100,
			wantEndian:  "MI",
			wantOrder:   binary.BigEndian,
			wantErr:     false,
		},
		{
			name:        "description with trailing nulls",
			header:      makeHeader("Test file\x00\x00\x00", 0x0100, "IM"),
			wantDesc:    "Test file",
			wantVersion: 0x0100,
			wantEndian:  "IM",
			wantOrder:   binary.LittleEndian,
			wantErr:     false,
		},
		{
			name:        "empty description",
			header:      makeHeader("", 0x0100, "IM"),
			wantDesc:    "",
			wantVersion: 0x0100,
			wantEndian:  "IM",
			wantOrder:   binary.LittleEndian,
			wantErr:     false,
		},
		{
			name:        "v7.2 format",
			header:      makeHeader("MATLAB 7.0 MAT-file", 0x0100, "IM"),
			wantDesc:    "MATLAB 7.0 MAT-file",
			wantVersion: 0x0100,
			wantEndian:  "IM",
			wantOrder:   binary.LittleEndian,
			wantErr:     false,
		},
		{
			name:    "invalid endian indicator",
			header:  makeHeader("Test", 0x0100, "XX"),
			wantErr: true,
		},
		{
			name:    "invalid endian indicator - empty",
			header:  makeHeader("Test", 0x0100, "\x00\x00"),
			wantErr: true,
		},
		{
			name:    "invalid endian indicator - partial",
			header:  makeHeader("Test", 0x0100, "M\x00"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseHeader(tt.header)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseHeader() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}

			if got.Description != tt.wantDesc {
				t.Errorf("Description = %q, want %q", got.Description, tt.wantDesc)
			}
			if got.Version != tt.wantVersion {
				t.Errorf("Version = 0x%04x, want 0x%04x", got.Version, tt.wantVersion)
			}
			if got.EndianIndicator != tt.wantEndian {
				t.Errorf("EndianIndicator = %q, want %q", got.EndianIndicator, tt.wantEndian)
			}
			if got.Order != tt.wantOrder {
				t.Errorf("Order = %v, want %v", got.Order, tt.wantOrder)
			}
		})
	}
}

// TestParseHeaderByteOrderVerification verifies that byte order is correctly detected
// and used for version number parsing.
func TestParseHeaderByteOrderVerification(t *testing.T) {
	// Note: "IM" = little-endian, "MI" = big-endian
	tests := []struct {
		name        string
		endian      string
		version     uint16
		wantVersion uint16
	}{
		{
			name:        "little endian version parsing",
			endian:      "IM",
			version:     0x0100,
			wantVersion: 0x0100,
		},
		{
			name:        "big endian version parsing",
			endian:      "MI",
			version:     0x0100,
			wantVersion: 0x0100,
		},
		{
			name:        "little endian different version",
			endian:      "IM",
			version:     0x0200,
			wantVersion: 0x0200,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			header := makeHeader("Test", tt.version, tt.endian)
			got, err := parseHeader(header)
			if err != nil {
				t.Fatalf("parseHeader() unexpected error: %v", err)
			}

			if got.Version != tt.wantVersion {
				t.Errorf("Version = 0x%04x, want 0x%04x", got.Version, tt.wantVersion)
			}

			// Verify byte order matches endian indicator
			// "IM" = little-endian, "MI" = big-endian
			if tt.endian == "IM" && got.Order != binary.LittleEndian {
				t.Error("Expected LittleEndian for 'IM' indicator")
			}
			if tt.endian == "MI" && got.Order != binary.BigEndian {
				t.Error("Expected BigEndian for 'MI' indicator")
			}
		})
	}
}

// TestParseHeaderLongDescription tests handling of maximum-length descriptions.
func TestParseHeaderLongDescription(t *testing.T) {
	// Description field is 116 bytes (0-115)
	longDesc := string(make([]byte, 116))
	for i := range longDesc {
		longDesc = longDesc[:i] + "A" + longDesc[i+1:]
	}

	header := makeHeader(longDesc, 0x0100, "IM") // Use "IM" for little-endian
	got, err := parseHeader(header)
	if err != nil {
		t.Fatalf("parseHeader() unexpected error: %v", err)
	}

	if len(got.Description) > 116 {
		t.Errorf("Description length = %d, want <= 116", len(got.Description))
	}
}

// makeHeader creates a test MAT-file header (128 bytes).
// Note: "IM" = little-endian, "MI" = big-endian.
func makeHeader(desc string, version uint16, endian string) []byte {
	header := make([]byte, 128)

	// Description (bytes 0-115)
	copy(header, desc)

	// Determine byte order from endian indicator
	// "IM" = little-endian, "MI" = big-endian
	var order binary.ByteOrder
	switch endian {
	case "IM":
		order = binary.LittleEndian
	case "MI":
		order = binary.BigEndian
	default:
		// For invalid endian, use little endian but write invalid indicator
		order = binary.LittleEndian
	}

	// Version (bytes 124-125)
	order.PutUint16(header[124:126], version)

	// Endian indicator (bytes 126-127)
	copy(header[126:128], endian)

	return header
}

// BenchmarkParseHeader benchmarks header parsing performance.
func BenchmarkParseHeader(b *testing.B) {
	header := makeHeader("MATLAB 5.0 MAT-file", 0x0100, "IM") // Use "IM" for little-endian

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = parseHeader(header)
	}
}
