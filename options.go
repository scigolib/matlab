package matlab

import (
	"encoding/binary"
)

// config holds optional configuration for Create.
type config struct {
	// v5-specific options
	description string           // File description (max 116 bytes)
	endianness  binary.ByteOrder // Byte order (LittleEndian or BigEndian)

	// Compression options (both formats)
	compression int // 0-9, 0=none, 9=max (future feature)
}

// Option configures optional parameters for Create.
type Option func(*config)

// WithEndianness sets the byte order for v5 files.
// Valid values: binary.LittleEndian, binary.BigEndian
//
// Default: binary.LittleEndian
//
// Example:
//
//	writer, _ := matlab.Create("file.mat", matlab.Version5,
//	    matlab.WithEndianness(binary.BigEndian))
func WithEndianness(order binary.ByteOrder) Option {
	return func(c *config) {
		c.endianness = order
	}
}

// WithDescription sets the file description (v5 only, max 116 bytes).
// If longer than 116 bytes, it will be truncated.
//
// Default: "MATLAB MAT-file, created by scigolib/matlab vX.X.X"
//
// Example:
//
//	writer, _ := matlab.Create("file.mat", matlab.Version5,
//	    matlab.WithDescription("Simulation results"))
func WithDescription(desc string) Option {
	return func(c *config) {
		if len(desc) > 116 {
			desc = desc[:116] // Truncate to fit v5 header
		}
		c.description = desc
	}
}

// WithCompression enables compression with specified level (0-9).
// 0 = no compression, 9 = maximum compression
//
// Note: Compression is not yet implemented. This option is reserved for future use.
//
// Default: 0 (no compression)
//
// Example:
//
//	writer, _ := matlab.Create("file.mat", matlab.Version73,
//	    matlab.WithCompression(6))
func WithCompression(level int) Option {
	return func(c *config) {
		if level < 0 {
			level = 0
		} else if level > 9 {
			level = 9
		}
		c.compression = level
	}
}

// defaultConfig returns configuration with default values.
func defaultConfig() *config {
	return &config{
		description: "MATLAB MAT-file, created by scigolib/matlab",
		endianness:  binary.LittleEndian,
		compression: 0,
	}
}

// applyOptions applies Option functions to config.
func applyOptions(cfg *config, opts []Option) {
	for _, opt := range opts {
		opt(cfg)
	}
}
