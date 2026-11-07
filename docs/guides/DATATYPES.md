# Datatypes

How MATLAB classes map to Go types in this library, for reading and for writing. Everything here is taken from
`types/variable.go`, `internal/v5/types.go`, `internal/v73/adapter.go` and `internal/v73/writer.go`, and the behaviour
marked "observed" was checked by writing a variable and reading it back.

## The `types.DataType` values

`types.Double`, `Single`, `Int8`, `Uint8`, `Int16`, `Uint16`, `Int32`, `Uint32`, `Int64`, `Uint64`, `Char`, `Struct`,
`CellArray`, `Object`, `Unknown`. `String()` gives the lower-case MATLAB name (`double`, `uint8`, `cell`, ...).

A `types.Variable` carries `Name`, `Dimensions`, `DataType`, `Data`, `IsComplex`, `IsSparse` and `Attributes`.
Arrays are stored column-major, as in MATLAB: for a 2x3 matrix `Data` holds the first column, then the second, then the
third.

## Numeric types

| MATLAB class | `DataType` | Write: `Data` must be | v5 read: `Data` is | v7.3 read: `Data` is |
|---|---|---|---|---|
| double | `Double` | `[]float64` | `[]float64` | `[]float64` |
| single | `Single` | `[]float32` | `[]float32` | `[]float64` (observed) |
| int8 | `Int8` | `[]int8` | `[]int8` | `[]float64` (observed) |
| uint8 | `Uint8` | `[]uint8` | `[]uint8` | `[]float64` (observed) |
| int16 | `Int16` | `[]int16` | `[]int16` | `[]float64` (observed) |
| uint16 | `Uint16` | `[]uint16` | `[]uint16` | `[]float64` (observed) |
| int32 | `Int32` | `[]int32` | `[]int32` | `[]float64` (observed) |
| uint32 | `Uint32` | `[]uint32` | `[]uint32` | `[]float64` (observed) |
| int64 | `Int64` | `[]int64` | `[]int64` | `[]float64` (observed) |
| uint64 | `Uint64` | `[]uint64` | `[]uint64` | `[]float64` (observed) |

The writers reject a `Data` slice whose type does not match `DataType` (for example `[]float64` with `Single`).

Reading a v5 file returns the Go type of the storage type recorded in the file (`miDOUBLE` to `[]float64`, `miSINGLE`
to `[]float32`, and so on; `internal/v5/types.go`, `convertData`). 

To get numbers without caring about the type, use the helpers on `*types.Variable`: `GetFloat64Array` (converts from
float32 and the integer types, refuses complex data), `GetInt32Array`, `GetComplex128Array` and `GetScalar`.

## Complex numbers

Set `IsComplex: true` and give `Data` as a `*types.NumericArray`, with `Real` and `Imag` slices of the type that
matches `DataType`.

| | v5 | v7.3 |
|---|---|---|
| Layout in the file | MATLAB's: real and imaginary parts as separate sub-elements of the matrix | an HDF5 group named after the variable with datasets `real` and `imag` and attributes `MATLAB_class` and `MATLAB_complex` |
| Read back as | `*types.NumericArray`, `IsComplex` true | `*types.NumericArray`, `IsComplex` true; `NumericArray.Dimensions` is empty (observed) |
| Checked against | golden-byte tests of the writer (`internal/v5/golden_test.go`) and the SciPy file `testcomplex_7.4_GLNX86.mat`, which is read | not the layout MATLAB writes: MATLAB uses one dataset of a compound type with `real` and `imag` members. That layout is not read or written here |

## Dimensions

- v5: the dimensions array of the file is returned. A vector written with one dimension, `[]int{3}`, is stored as
  3x1 and comes back as `[3 1]` (observed).
- v7.3: the reader reports the dataset as one-dimensional: a 2x3 variable comes back as `[6]` (observed), with the
  elements in the order the file stores them. Keep the shape yourself if you need it.

## Character arrays

- Reading, v5: a char matrix has `DataType` `Char`. `Data` follows the element's storage type: `[]uint16` for
  `miUINT16`, a `string` for `miUTF8` (`convertData`, unit-tested in `internal/v5/types_test.go`). The repository has
  no MATLAB-written char file, so this path is not checked against one.
- Reading, v7.3: `DataType` is `Char` when the dataset's `MATLAB_class` attribute is `char`. The adapter reads the
  dataset as numbers first and falls back to strings if that fails.
- Writing: not supported in either format. `WriteVariable` with `DataType` `Char` returns an error
  (`unsupported data type: char` for v5, `unsupported MATLAB data type: char` for v7.3; observed).

## Structures, cell arrays, sparse, objects

| MATLAB | v5 read | v7.3 read | Write |
|---|---|---|---|
| struct | `DataType` `Struct`; fields are not decoded and `Data` is not usable | each field comes back as its own variable named by its path (for example `/s/a`) | not supported |
| cell | `DataType` `CellArray`; elements are not decoded and `Data` is not usable; an empty cell array fails to open | class reported, elements not decoded | not supported |
| sparse | `DataType` `Unknown` and meaningless data | not supported | not supported |
| function handle, object | not supported | not supported | not supported |

## Limits on reading

- A compressed v5 element (`miCOMPRESSED`) may decompress to at most 100 MB and at a ratio of at most 1000:1; larger
  or more compressible elements are rejected with an error. This guards against compression bombs.
- A truncated v5 file fails with an error that wraps `io.ErrUnexpectedEOF`.
- `matlab.Open` rejects a MAT header whose version is neither 0x0100 (v5) nor 0x0200 (v7.3) with
  `ErrUnsupportedVersion`, and input that is neither a MAT header nor HDF5 with `ErrInvalidFormat`.

## Writer limits

- Variable names: at most 63 characters in v5 (`validateVariable` in `internal/v5/writer.go`).
- Dimensions must be positive and their product must not overflow `int64`.
- `WithCompression` is reserved and has no effect: nothing is compressed on write.
