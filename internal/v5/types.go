package v5

import (
	"math"

	"github.com/scigolib/matlab/types"
)

// MATLAB data type constants.
const (
	miINT8       = 1
	miUINT8      = 2
	miINT16      = 3
	miUINT16     = 4
	miINT32      = 5
	miUINT32     = 6
	miSINGLE     = 7
	miDOUBLE     = 9
	miINT64      = 12
	miUINT64     = 13
	miMATRIX     = 14
	miCOMPRESSED = 15
	miUTF8       = 16
)

// MATLAB array class constants.
//
//nolint:revive // MATLAB official naming convention from specification
const (
	mxCELL_CLASS   = 1
	mxSTRUCT_CLASS = 2
	mxOBJECT_CLASS = 3
	mxCHAR_CLASS   = 4
	mxDOUBLE_CLASS = 6
	mxSINGLE_CLASS = 7
	mxINT8_CLASS   = 8
	mxUINT8_CLASS  = 9
	mxINT16_CLASS  = 10
	mxUINT16_CLASS = 11
	mxINT32_CLASS  = 12
	mxUINT32_CLASS = 13
	mxINT64_CLASS  = 14
	mxUINT64_CLASS = 15
)

// classToDataType converts MATLAB class to DataType.
func classToDataType(class uint32) types.DataType {
	switch class {
	case mxDOUBLE_CLASS:
		return types.Double
	case mxSINGLE_CLASS:
		return types.Single
	case mxINT8_CLASS:
		return types.Int8
	case mxUINT8_CLASS:
		return types.Uint8
	case mxINT16_CLASS:
		return types.Int16
	case mxUINT16_CLASS:
		return types.Uint16
	case mxINT32_CLASS:
		return types.Int32
	case mxUINT32_CLASS:
		return types.Uint32
	case mxINT64_CLASS:
		return types.Int64
	case mxUINT64_CLASS:
		return types.Uint64
	case mxCHAR_CLASS:
		return types.Char
	case mxSTRUCT_CLASS:
		return types.Struct
	case mxCELL_CLASS:
		return types.CellArray
	default:
		return types.Unknown
	}
}

// convertData converts raw bytes to appropriate Go type.
//
//nolint:gocognit,gocyclo,cyclop,funlen // Type conversion requires exhaustive type matching (MATLAB spec)
func (p *Parser) convertData(data []byte, dataType, _ uint32) interface{} {
	switch dataType {
	case miDOUBLE:
		count := len(data) / 8
		if count == 0 || len(data) < count*8 {
			return []float64{}
		}
		result := make([]float64, count)
		for i := 0; i < count; i++ {
			result[i] = math.Float64frombits(
				p.Header.Order.Uint64(data[i*8 : (i+1)*8]))
		}
		return result

	case miSINGLE:
		count := len(data) / 4
		if count == 0 || len(data) < count*4 {
			return []float32{}
		}
		result := make([]float32, count)
		for i := 0; i < count; i++ {
			result[i] = math.Float32frombits(
				p.Header.Order.Uint32(data[i*4 : (i+1)*4]))
		}
		return result

	case miINT8:
		result := make([]int8, len(data))
		for i := 0; i < len(data); i++ {
			result[i] = int8(data[i])
		}
		return result

	case miUINT8:
		return data

	case miINT16:
		count := len(data) / 2
		if count == 0 || len(data) < count*2 {
			return []int16{}
		}
		result := make([]int16, count)
		for i := 0; i < count; i++ {
			result[i] = int16(p.Header.Order.Uint16(data[i*2 : (i+1)*2])) //nolint:gosec // MATLAB format conversion
		}
		return result

	case miUINT16:
		count := len(data) / 2
		if count == 0 || len(data) < count*2 {
			return []uint16{}
		}
		result := make([]uint16, count)
		for i := 0; i < count; i++ {
			result[i] = p.Header.Order.Uint16(data[i*2 : (i+1)*2])
		}
		return result

	case miINT32:
		count := len(data) / 4
		if count == 0 || len(data) < count*4 {
			return []int32{}
		}
		result := make([]int32, count)
		for i := 0; i < count; i++ {
			result[i] = int32(p.Header.Order.Uint32(data[i*4 : (i+1)*4])) //nolint:gosec // MATLAB format conversion
		}
		return result

	case miUINT32:
		count := len(data) / 4
		if count == 0 || len(data) < count*4 {
			return []uint32{}
		}
		result := make([]uint32, count)
		for i := 0; i < count; i++ {
			result[i] = p.Header.Order.Uint32(data[i*4 : (i+1)*4])
		}
		return result

	case miINT64:
		count := len(data) / 8
		if count == 0 || len(data) < count*8 {
			return []int64{}
		}
		result := make([]int64, count)
		for i := 0; i < count; i++ {
			result[i] = int64(p.Header.Order.Uint64(data[i*8 : (i+1)*8])) //nolint:gosec // MATLAB format conversion
		}
		return result

	case miUINT64:
		count := len(data) / 8
		if count == 0 || len(data) < count*8 {
			return []uint64{}
		}
		result := make([]uint64, count)
		for i := 0; i < count; i++ {
			result[i] = p.Header.Order.Uint64(data[i*8 : (i+1)*8])
		}
		return result

	case miUTF8:
		return string(data)

	default:
		// For unsupported types, return raw bytes
		return data
	}
}
