package matlab

import (
	"encoding/binary"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWithEndianness(t *testing.T) {
	tests := []struct {
		name     string
		order    binary.ByteOrder
		expected binary.ByteOrder
	}{
		{
			name:     "little endian",
			order:    binary.LittleEndian,
			expected: binary.LittleEndian,
		},
		{
			name:     "big endian",
			order:    binary.BigEndian,
			expected: binary.BigEndian,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := defaultConfig()
			opt := WithEndianness(tt.order)
			opt(cfg)

			assert.Equal(t, tt.expected, cfg.endianness)
		})
	}
}

func TestWithDescription(t *testing.T) {
	tests := []struct {
		name     string
		desc     string
		expected string
	}{
		{
			name:     "short description",
			desc:     "Test file",
			expected: "Test file",
		},
		{
			name:     "long description (truncated)",
			desc:     string(make([]byte, 200)), // 200 bytes
			expected: string(make([]byte, 116)), // Truncated to 116
		},
		{
			name:     "empty description",
			desc:     "",
			expected: "",
		},
		{
			name:     "exactly 116 bytes",
			desc:     string(make([]byte, 116)),
			expected: string(make([]byte, 116)),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := defaultConfig()
			opt := WithDescription(tt.desc)
			opt(cfg)

			assert.Equal(t, tt.expected, cfg.description)
		})
	}
}

func TestWithCompression(t *testing.T) {
	tests := []struct {
		name     string
		level    int
		expected int
	}{
		{"no compression", 0, 0},
		{"medium compression", 5, 5},
		{"max compression", 9, 9},
		{"negative (clamped to 0)", -5, 0},
		{"too high (clamped to 9)", 15, 9},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := defaultConfig()
			opt := WithCompression(tt.level)
			opt(cfg)

			assert.Equal(t, tt.expected, cfg.compression)
		})
	}
}

func TestCreate_WithOptions(t *testing.T) {
	tmpfile := filepath.Join(t.TempDir(), "options.mat")

	// Create with custom options
	writer, err := Create(tmpfile, Version5,
		WithEndianness(binary.BigEndian),
		WithDescription("Custom description"),
	)
	require.NoError(t, err)
	require.NoError(t, writer.Close())

	// Read back and verify header
	file, err := os.Open(tmpfile)
	require.NoError(t, err)
	defer file.Close()

	header := make([]byte, 128)
	_, err = file.Read(header)
	require.NoError(t, err)

	// Check description
	desc := string(header[0:116])
	assert.Contains(t, desc, "Custom description")

	// Check endianness (bytes 126-127)
	assert.Equal(t, byte('I'), header[126])
	assert.Equal(t, byte('M'), header[127])
}

func TestCreate_BackwardCompatibility(t *testing.T) {
	tmpfile := filepath.Join(t.TempDir(), "compat.mat")

	// Old API (no options) should still work
	writer, err := Create(tmpfile, Version5)
	require.NoError(t, err)
	require.NoError(t, writer.Close())

	// Verify file was created
	_, err = os.Stat(tmpfile)
	assert.NoError(t, err)
}

func TestCreate_V5_DefaultEndianness(t *testing.T) {
	tmpfile := filepath.Join(t.TempDir(), "default_endian.mat")

	// Create without specifying endianness
	writer, err := Create(tmpfile, Version5)
	require.NoError(t, err)
	require.NoError(t, writer.Close())

	// Read back and verify default is little-endian
	file, err := os.Open(tmpfile)
	require.NoError(t, err)
	defer file.Close()

	header := make([]byte, 128)
	_, err = file.Read(header)
	require.NoError(t, err)

	// Check default endianness (bytes 126-127) should be "MI" (little-endian)
	assert.Equal(t, byte('M'), header[126])
	assert.Equal(t, byte('I'), header[127])
}

func TestCreate_V5_DefaultDescription(t *testing.T) {
	tmpfile := filepath.Join(t.TempDir(), "default_desc.mat")

	// Create without specifying description
	writer, err := Create(tmpfile, Version5)
	require.NoError(t, err)
	require.NoError(t, writer.Close())

	// Read back and verify default description
	file, err := os.Open(tmpfile)
	require.NoError(t, err)
	defer file.Close()

	header := make([]byte, 128)
	_, err = file.Read(header)
	require.NoError(t, err)

	// Check description
	desc := string(header[0:116])
	assert.Contains(t, desc, "MATLAB MAT-file, created by scigolib/matlab")
}

func TestCreate_MultipleOptions(t *testing.T) {
	tmpfile := filepath.Join(t.TempDir(), "multiple_opts.mat")

	// Create with multiple options
	writer, err := Create(tmpfile, Version5,
		WithEndianness(binary.BigEndian),
		WithDescription("Test with multiple options"),
		WithCompression(6), // Future feature, should not affect v5 for now
	)
	require.NoError(t, err)
	require.NoError(t, writer.Close())

	// Verify file was created successfully
	_, err = os.Stat(tmpfile)
	assert.NoError(t, err)
}

func TestDefaultConfig(t *testing.T) {
	cfg := defaultConfig()

	// Check default values
	assert.Equal(t, "MATLAB MAT-file, created by scigolib/matlab", cfg.description)
	assert.Equal(t, binary.LittleEndian, cfg.endianness)
	assert.Equal(t, 0, cfg.compression)
}

func TestApplyOptions(t *testing.T) {
	cfg := defaultConfig()

	// Apply multiple options
	opts := []Option{
		WithEndianness(binary.BigEndian),
		WithDescription("Modified"),
		WithCompression(7),
	}
	applyOptions(cfg, opts)

	// Verify all options were applied
	assert.Equal(t, binary.BigEndian, cfg.endianness)
	assert.Equal(t, "Modified", cfg.description)
	assert.Equal(t, 7, cfg.compression)
}

func TestApplyOptions_Empty(t *testing.T) {
	cfg := defaultConfig()
	originalDesc := cfg.description
	originalEndian := cfg.endianness
	originalCompression := cfg.compression

	// Apply no options
	applyOptions(cfg, nil)

	// Verify defaults remain unchanged
	assert.Equal(t, originalDesc, cfg.description)
	assert.Equal(t, originalEndian, cfg.endianness)
	assert.Equal(t, originalCompression, cfg.compression)
}
