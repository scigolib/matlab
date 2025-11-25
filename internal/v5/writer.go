package v5

import (
	"encoding/binary"
	"fmt"
	"io"
	"math"

	"github.com/scigolib/matlab/types"
)

// Writer handles writing v5 MAT-files.
//
// The writer creates MAT-files in the v5 binary format (compatible with
// MATLAB versions v5.0 through v7.2). Files are written with a 128-byte
// header followed by data elements in Tag-Length-Value (TLV) format.
//
// All data elements are aligned to 8-byte boundaries as per the MAT-File
// Format v5 specification. The writer supports both little-endian ("IM")
// and big-endian ("MI") byte ordering.
type Writer struct {
	w      io.Writer
	header *Header
	pos    int64
}

// NewWriter creates a new v5 writer.
//
// The header is written immediately upon creation. All subsequent variable
// writes will be appended to the file in the order they are written.
//
// Parameters:
//   - w: io.Writer to write to (typically a file opened with os.Create)
//   - description: Optional description text (max 116 bytes, truncated if longer)
//   - endian: Byte order - "IM" for little-endian or "MI" for big-endian
//
// Returns:
//   - *Writer: Configured writer with header already written
//   - error: If endian indicator is invalid or header write fails
//
// Example:
//
//	f, _ := os.Create("output.mat")
//	defer f.Close()
//	writer, err := NewWriter(f, "Created by scigolib/matlab", "IM")
func NewWriter(w io.Writer, description, endian string) (*Writer, error) {
	// Validate endian indicator
	// "IM" = little-endian, "MI" = big-endian
	// (The 16-bit value 0x4D49 "MI" is stored as [0x49, 0x4D] "IM" on little-endian systems)
	var order binary.ByteOrder
	switch endian {
	case "IM":
		order = binary.LittleEndian
	case "MI":
		order = binary.BigEndian
	default:
		return nil, fmt.Errorf("invalid endian indicator: %q (must be IM or MI)", endian)
	}

	writer := &Writer{
		w: w,
		header: &Header{
			Description:     description,
			Version:         0x0100, // MAT-file version 5.0
			EndianIndicator: endian,
			Order:           order,
		},
		pos: 0,
	}

	// Write header immediately
	if err := writer.writeHeader(); err != nil {
		return nil, fmt.Errorf("failed to write header: %w", err)
	}

	return writer, nil
}

// WriteVariable writes a MATLAB variable to the file.
//
// The variable is written as a miMATRIX data element containing nested
// sub-elements for array flags, dimensions, name, and data. Complex numbers
// are written with separate real and imaginary data sub-elements.
//
// Parameters:
//   - v: Variable to write (must not be nil)
//
// Returns:
//   - error: If validation fails or writing fails
//
// Supported types:
//   - Double, Single, Int8, Uint8, Int16, Uint16, Int32, Uint32, Int64, Uint64
//   - Complex numbers (use types.NumericArray with Real/Imag)
//   - Multi-dimensional arrays
func (w *Writer) WriteVariable(v *types.Variable) error {
	// Validate variable
	if err := w.validateVariable(v); err != nil {
		return fmt.Errorf("invalid variable: %w", err)
	}

	// Write as miMATRIX data element
	return w.writeMatrix(v)
}

// validateVariable checks if variable is valid for v5 format.
func (w *Writer) validateVariable(v *types.Variable) error {
	if v.Name == "" {
		return fmt.Errorf("variable name is required")
	}
	if len(v.Name) > 63 {
		return fmt.Errorf("variable name too long (max 63 characters): %d", len(v.Name))
	}
	if len(v.Dimensions) == 0 {
		return fmt.Errorf("dimensions are required")
	}
	if v.Data == nil {
		return fmt.Errorf("data is required")
	}

	// Validate dimensions are positive and check for overflow
	total := int64(1)
	for i, d := range v.Dimensions {
		if d <= 0 {
			return fmt.Errorf("dimension[%d] must be positive, got %d", i, d)
		}

		// Check for overflow before multiplying
		if d > 0 && total > math.MaxInt/int64(d) {
			return fmt.Errorf("dimensions overflow (total elements too large): %v", v.Dimensions)
		}

		total *= int64(d)
	}

	return nil
}

// writeHeader writes the 128-byte MAT-file header.
//
// Header structure:
// - Bytes 0-115: Description text (null-terminated/padded)
// - Bytes 116-123: Subsystem data offset (zeros for standard files)
// - Bytes 124-125: Version (0x0100)
// - Bytes 126-127: Endian indicator ("MI" or "IM").
func (w *Writer) writeHeader() error {
	header := make([]byte, 128)

	// Description (bytes 0-115) - truncate if too long
	desc := w.header.Description
	if len(desc) > 116 {
		desc = desc[:116]
	}
	copy(header, desc)
	// Remaining bytes are already zero (null padding)

	// Subsystem data offset (bytes 116-123) - zeros for standard files
	// Already zero from make()

	// Version (bytes 124-125)
	w.header.Order.PutUint16(header[124:126], w.header.Version)

	// Endian indicator (bytes 126-127)
	copy(header[126:128], w.header.EndianIndicator)

	// Write to stream
	n, err := w.w.Write(header)
	if err != nil {
		return err
	}
	if n != 128 {
		return fmt.Errorf("wrote %d bytes, expected 128", n)
	}

	w.pos += 128
	return nil
}

// writeMatrix writes a complete matrix data element.
//
// The matrix is written as a miMATRIX data element containing:
// 1. Array flags (8 bytes)
// 2. Dimensions array (int32 array)
// 3. Array name (int8/UTF-8 string)
// 4. Real part data
// 5. Imaginary part data (if complex)
//
// All data is padded to 8-byte boundaries.
func (w *Writer) writeMatrix(v *types.Variable) error {
	// Step 1: Encode matrix content to buffer (to calculate total size)
	content, err := w.encodeMatrixContent(v)
	if err != nil {
		return fmt.Errorf("failed to encode matrix content: %w", err)
	}

	// Step 2: Write miMATRIX tag (8 bytes)
	//nolint:gosec // G115: Content length is from encoded buffer, safe conversion
	if err := w.writeTag(miMATRIX, uint32(len(content))); err != nil {
		return fmt.Errorf("failed to write matrix tag: %w", err)
	}

	// Step 3: Write content
	n, err := w.w.Write(content)
	if err != nil {
		return err
	}
	w.pos += int64(n)

	// Step 4: Write padding to 8-byte boundary
	padding := (8 - len(content)%8) % 8
	if padding > 0 {
		zeros := make([]byte, padding)
		np, err := w.w.Write(zeros)
		if err != nil {
			return err
		}
		w.pos += int64(np)
	}

	return nil
}

// encodeMatrixContent encodes all matrix sub-elements to a byte buffer.
//
// Returns the complete matrix content as a single byte slice.
// This is used to calculate the total size for the miMATRIX tag.
func (w *Writer) encodeMatrixContent(v *types.Variable) ([]byte, error) {
	var buf []byte

	// Sub-element 1: Array Flags (8 bytes)
	flags := w.encodeArrayFlags(v)
	buf = append(buf, flags...)

	// Sub-element 2: Dimensions Array
	dims := w.encodeDimensions(v.Dimensions)
	buf = append(buf, dims...)

	// Sub-element 3: Array Name
	name := w.encodeName(v.Name)
	buf = append(buf, name...)

	// Sub-element 4: Real Data
	realData, err := w.encodeData(v, false)
	if err != nil {
		return nil, err
	}
	buf = append(buf, realData...)

	// Sub-element 5: Imaginary Data (if complex)
	if v.IsComplex {
		imagData, err := w.encodeData(v, true)
		if err != nil {
			return nil, err
		}
		buf = append(buf, imagData...)
	}

	return buf, nil
}

// encodeArrayFlags encodes array flags sub-element.
//
// The array flags contain:
// - Bytes 0-3: Flags (complex bit, sparse bit, etc.)
// - Bytes 4-7: MATLAB class (mxDOUBLE_CLASS, etc.)
func (w *Writer) encodeArrayFlags(v *types.Variable) []byte {
	// Build flags
	var flags uint32
	if v.IsComplex {
		flags |= 0x0800 // Complex bit (bit 11)
	}
	if v.IsSparse {
		flags |= 0x0400 // Sparse bit (bit 10)
	}

	class := w.dataTypeToClass(v.DataType)

	// Create 8-byte data: flags + class
	data := make([]byte, 8)
	w.header.Order.PutUint32(data[0:4], flags)
	w.header.Order.PutUint32(data[4:8], class)

	// Wrap in miUINT32 tag
	return w.wrapInTag(miUINT32, data)
}

// encodeDimensions encodes dimensions array sub-element.
//
// Dimensions are written as an int32 array wrapped in a data element tag.
func (w *Writer) encodeDimensions(dims []int) []byte {
	// Convert to int32 array
	data := make([]byte, len(dims)*4)
	for i, d := range dims {
		//nolint:gosec // G115: Dimensions already validated as positive, safe conversion
		w.header.Order.PutUint32(data[i*4:(i+1)*4], uint32(d))
	}

	// Wrap in data element tag
	return w.wrapInTag(miINT32, data)
}

// encodeName encodes array name sub-element.
//
// The name is written as an int8/UTF-8 string wrapped in a data element tag.
func (w *Writer) encodeName(name string) []byte {
	data := []byte(name)
	return w.wrapInTag(miINT8, data)
}

// encodeData encodes real or imaginary data.
//
// For complex numbers, this is called twice:
// - Once with imaginary=false for real part
// - Once with imaginary=true for imaginary part.
//
//nolint:gocognit,gocyclo,cyclop,funlen,nestif // Type dispatch requires exhaustive type matching (MATLAB spec)
func (w *Writer) encodeData(v *types.Variable, imaginary bool) ([]byte, error) {
	var rawData []byte
	var dataType uint32

	// Get data to encode
	var data interface{}
	if v.IsComplex {
		numArray, ok := v.Data.(*types.NumericArray)
		if !ok {
			return nil, fmt.Errorf("complex variable must have *types.NumericArray, got %T", v.Data)
		}
		if imaginary {
			if numArray.Imag == nil {
				return nil, fmt.Errorf("complex variable missing imaginary part")
			}
			data = numArray.Imag
		} else {
			if numArray.Real == nil {
				return nil, fmt.Errorf("complex variable missing real part")
			}
			data = numArray.Real
		}
	} else {
		data = v.Data
	}

	// Encode based on type
	switch v.DataType {
	case types.Double:
		dataType = miDOUBLE
		arr, ok := data.([]float64)
		if !ok {
			return nil, fmt.Errorf("expected []float64 for Double, got %T", data)
		}
		rawData = w.encodeFloat64Array(arr)

	case types.Single:
		dataType = miSINGLE
		arr, ok := data.([]float32)
		if !ok {
			return nil, fmt.Errorf("expected []float32 for Single, got %T", data)
		}
		rawData = w.encodeFloat32Array(arr)

	case types.Int8:
		dataType = miINT8
		arr, ok := data.([]int8)
		if !ok {
			return nil, fmt.Errorf("expected []int8 for Int8, got %T", data)
		}
		rawData = w.encodeInt8Array(arr)

	case types.Uint8:
		dataType = miUINT8
		arr, ok := data.([]byte)
		if !ok {
			return nil, fmt.Errorf("expected []byte for Uint8, got %T", data)
		}
		rawData = arr

	case types.Int16:
		dataType = miINT16
		arr, ok := data.([]int16)
		if !ok {
			return nil, fmt.Errorf("expected []int16 for Int16, got %T", data)
		}
		rawData = w.encodeInt16Array(arr)

	case types.Uint16:
		dataType = miUINT16
		arr, ok := data.([]uint16)
		if !ok {
			return nil, fmt.Errorf("expected []uint16 for Uint16, got %T", data)
		}
		rawData = w.encodeUint16Array(arr)

	case types.Int32:
		dataType = miINT32
		arr, ok := data.([]int32)
		if !ok {
			return nil, fmt.Errorf("expected []int32 for Int32, got %T", data)
		}
		rawData = w.encodeInt32Array(arr)

	case types.Uint32:
		dataType = miUINT32
		arr, ok := data.([]uint32)
		if !ok {
			return nil, fmt.Errorf("expected []uint32 for Uint32, got %T", data)
		}
		rawData = w.encodeUint32Array(arr)

	case types.Int64:
		dataType = miINT64
		arr, ok := data.([]int64)
		if !ok {
			return nil, fmt.Errorf("expected []int64 for Int64, got %T", data)
		}
		rawData = w.encodeInt64Array(arr)

	case types.Uint64:
		dataType = miUINT64
		arr, ok := data.([]uint64)
		if !ok {
			return nil, fmt.Errorf("expected []uint64 for Uint64, got %T", data)
		}
		rawData = w.encodeUint64Array(arr)

	default:
		return nil, fmt.Errorf("unsupported data type: %v", v.DataType)
	}

	// Wrap in tag
	return w.wrapInTag(dataType, rawData), nil
}

// wrapInTag wraps data in a data element tag.
//
// Always uses regular format (8-byte tag + N-byte data + padding).
// Small format is not used for matrix sub-elements to maintain compatibility
// with the parser's readData implementation.
func (w *Writer) wrapInTag(dataType uint32, data []byte) []byte {
	//nolint:gosec // G115: Data length is bounded by actual data size, safe conversion
	size := uint32(len(data))

	// Regular format: tag (8 bytes) + data + padding to 8-byte boundary
	padding := (8 - size%8) % 8
	buf := make([]byte, 8+size+padding)

	// Tag
	w.header.Order.PutUint32(buf[0:4], dataType)
	w.header.Order.PutUint32(buf[4:8], size)

	// Data
	copy(buf[8:8+size], data)

	// Padding is already zero from make()

	return buf
}

// writeTag writes a data element tag (8 bytes) for miMATRIX.
//
// This is only used for the top-level miMATRIX tag.
// Nested tags are written via wrapInTag.
func (w *Writer) writeTag(dataType, size uint32) error {
	buf := make([]byte, 8)
	w.header.Order.PutUint32(buf[0:4], dataType)
	w.header.Order.PutUint32(buf[4:8], size)

	n, err := w.w.Write(buf)
	if err != nil {
		return err
	}
	w.pos += int64(n)
	return nil
}

// Type encoding helpers - convert Go slices to byte arrays.

func (w *Writer) encodeFloat64Array(data []float64) []byte {
	buf := make([]byte, len(data)*8)
	for i, val := range data {
		w.header.Order.PutUint64(buf[i*8:(i+1)*8], math.Float64bits(val))
	}
	return buf
}

func (w *Writer) encodeFloat32Array(data []float32) []byte {
	buf := make([]byte, len(data)*4)
	for i, val := range data {
		w.header.Order.PutUint32(buf[i*4:(i+1)*4], math.Float32bits(val))
	}
	return buf
}

func (w *Writer) encodeInt8Array(data []int8) []byte {
	buf := make([]byte, len(data))
	for i, val := range data {
		buf[i] = byte(val)
	}
	return buf
}

func (w *Writer) encodeInt16Array(data []int16) []byte {
	buf := make([]byte, len(data)*2)
	for i, val := range data {
		//nolint:gosec // G115: int16 to uint16 is safe, preserves bit pattern
		w.header.Order.PutUint16(buf[i*2:(i+1)*2], uint16(val))
	}
	return buf
}

func (w *Writer) encodeUint16Array(data []uint16) []byte {
	buf := make([]byte, len(data)*2)
	for i, val := range data {
		w.header.Order.PutUint16(buf[i*2:(i+1)*2], val)
	}
	return buf
}

func (w *Writer) encodeInt32Array(data []int32) []byte {
	buf := make([]byte, len(data)*4)
	for i, val := range data {
		//nolint:gosec // G115: int32 to uint32 is safe, preserves bit pattern
		w.header.Order.PutUint32(buf[i*4:(i+1)*4], uint32(val))
	}
	return buf
}

func (w *Writer) encodeUint32Array(data []uint32) []byte {
	buf := make([]byte, len(data)*4)
	for i, val := range data {
		w.header.Order.PutUint32(buf[i*4:(i+1)*4], val)
	}
	return buf
}

func (w *Writer) encodeInt64Array(data []int64) []byte {
	buf := make([]byte, len(data)*8)
	for i, val := range data {
		//nolint:gosec // G115: int64 to uint64 is safe, preserves bit pattern
		w.header.Order.PutUint64(buf[i*8:(i+1)*8], uint64(val))
	}
	return buf
}

func (w *Writer) encodeUint64Array(data []uint64) []byte {
	buf := make([]byte, len(data)*8)
	for i, val := range data {
		w.header.Order.PutUint64(buf[i*8:(i+1)*8], val)
	}
	return buf
}

// dataTypeToClass converts types.DataType to MATLAB class constant.
func (w *Writer) dataTypeToClass(dt types.DataType) uint32 {
	switch dt {
	case types.Double:
		return mxDOUBLE_CLASS
	case types.Single:
		return mxSINGLE_CLASS
	case types.Int8:
		return mxINT8_CLASS
	case types.Uint8:
		return mxUINT8_CLASS
	case types.Int16:
		return mxINT16_CLASS
	case types.Uint16:
		return mxUINT16_CLASS
	case types.Int32:
		return mxINT32_CLASS
	case types.Uint32:
		return mxUINT32_CLASS
	case types.Int64:
		return mxINT64_CLASS
	case types.Uint64:
		return mxUINT64_CLASS
	default:
		return mxDOUBLE_CLASS // Fallback
	}
}
