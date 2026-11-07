# Architecture overview

The library has one public entry point per direction and two format backends. Detection happens at the entry point,
so callers never name a format when reading.

```
matfile.go          Open(io.Reader): detects the format, returns *MatFile
matfile_write.go    Create(filename, version, opts...): returns *MatFileWriter
options.go          WithEndianness, WithDescription, WithCompression (reserved)
types/              Variable, DataType, NumericArray: the public data model
internal/v5/        v5 (MATLAB 5 to 7.2) reader and writer
internal/v73/       v7.3 reader and writer, on top of github.com/scigolib/hdf5
```

## Format detection

`Open` reads the first 128 bytes. If they start with the HDF5 signature it parses the file as v7.3 (this is what
files written before v0.3.19 look like). Otherwise it expects a MAT header and reads its version field: 0x0100 is v5,
0x0200 is v7.3. Any other version fails with `ErrUnsupportedVersion`, and input that is neither fails with
`ErrInvalidFormat`. The version rule follows matio and SciPy; see [UPSTREAM.md](../../UPSTREAM.md). MATLAB writes v7.3
files as HDF5 behind a 512-byte user block that holds the MAT header, and the `hdf5` module finds the signature behind
it.

## internal/v5

| File | Role |
|---|---|
| `parser.go`, `matrix.go`, `data_tag.go` | streaming reader: tag-length-value elements, nested matrix elements |
| `header.go` | the 128-byte header: description, version, byte order (`MI` or `IM`) |
| `compressed.go` | `miCOMPRESSED` elements (zlib), with a size and ratio cap |
| `bounds.go` | sizes read from the file are checked against the bytes present before any allocation |
| `types.go` | MATLAB class and storage type constants, conversion to Go slices |
| `writer.go` | the v5 writer |

## internal/v73

| File | Role |
|---|---|
| `parser.go`, `adapter.go` | walk the HDF5 tree with the `hdf5` module and turn datasets and groups into variables; the class comes from the `MATLAB_class` attribute |
| `header.go` | the MAT header written in the user block |
| `writer.go` | the v7.3 writer: one dataset per variable, a group with `real` and `imag` for complex variables |

## Tests and tools

- `*_test.go` in the root: public API, round trips, SciPy reference files, header detection
- `internal/*/golden_test.go`, `fuzz_test.go`: byte-exact v5 output and fuzzing of the v5 parser
- `testdata/`: reference files and files written by our writer; see [UPSTREAM.md](../../UPSTREAM.md)
- `scripts/generate-testdata`, `scripts/verify-roundtrip`: regenerate the v7.3 test files and check a round trip
