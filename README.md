# MATLAB File Reader/Writer for Go

A pure Go implementation of the MATLAB `.mat` file formats: read and write MAT-files without MATLAB, the MATLAB
runtime or cgo. Part of the [SciGoLib](https://github.com/scigolib) scientific computing libraries.

[![GitHub Release](https://img.shields.io/github/v/release/scigolib/matlab?include_prereleases&style=flat-square&logo=github&color=blue)](https://github.com/scigolib/matlab/releases/latest)
[![Go Version](https://img.shields.io/badge/Go-1.25%2B-00ADD8?style=flat-square&logo=go)](https://go.dev/dl/)
[![Go Reference](https://pkg.go.dev/badge/github.com/scigolib/matlab.svg)](https://pkg.go.dev/github.com/scigolib/matlab)
[![GitHub Actions](https://img.shields.io/github/actions/workflow/status/scigolib/matlab/test.yml?branch=main&style=flat-square&logo=github-actions&label=CI)](https://github.com/scigolib/matlab/actions)
[![Codecov](https://img.shields.io/codecov/c/github/scigolib/matlab?style=flat-square&logo=codecov)](https://codecov.io/gh/scigolib/matlab)
[![Go Report Card](https://goreportcard.com/badge/github.com/scigolib/matlab?style=flat-square)](https://goreportcard.com/report/github.com/scigolib/matlab)
[![License](https://img.shields.io/badge/license-MIT-blue.svg?style=flat-square)](LICENSE)
[![GitHub Stars](https://img.shields.io/github/stars/scigolib/matlab?style=flat-square&logo=github)](https://github.com/scigolib/matlab/stargazers)
[![GitHub Issues](https://img.shields.io/github/issues/scigolib/matlab?style=flat-square&logo=github)](https://github.com/scigolib/matlab/issues)
[![Discussions](https://img.shields.io/github/discussions/scigolib/matlab?style=flat-square&logo=github&color=purple)](https://github.com/scigolib/matlab/discussions)

A MAT-file is the file format MATLAB uses to save workspace variables. Engineers and scientists use it to exchange
measurements, simulation results and trained parameters, and many tools besides MATLAB read and write it: GNU Octave,
SciPy, and libraries for most languages. There are two families. Version 5 (MATLAB 5 to 7.2) is a binary
tag-length-value format, optionally zlib-compressed. Version 7.3 is an HDF5 file with a 512-byte MAT header in front.
This library reads both and writes both, so a Go program can exchange data with MATLAB users without a C toolchain.

---

## Features

- **Pure Go**: no cgo and no C dependencies; cross-compiles like any Go program. 32-bit builds
  (`linux/386`, `linux/arm`, `windows/386`) are checked in CI.
- **One entry point for reading**: `matlab.Open` detects v5 and v7.3 by the MAT header.
- **Two writers**: `matlab.Create` with `matlab.Version5` or `matlab.Version73`.
- **Numeric data**: double, single, int8 to int64 and uint8 to uint64, multi-dimensional, real and complex.
- **Compressed v5 files** are read (zlib), with a cap on the decompressed size.
- **Both byte orders** in v5 files, read and write.
- **Hardened input handling**: sizes read from a file are checked against the bytes present before memory is
  allocated, and the v5 parser is fuzz-tested.
- **Checked against references**: SciPy test files, golden-byte tests of the v5 writer, `h5dump` on generated v7.3
  files; see [UPSTREAM.md](UPSTREAM.md).

---

## Quick Start

```bash
go get github.com/scigolib/matlab
```

### Read a file

```go
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/scigolib/matlab"
)

func main() {
	file, err := os.Open("data.mat")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	// Detects v5 and v7.3 by itself.
	mat, err := matlab.Open(file)
	if err != nil {
		log.Fatal(err)
	}

	for _, v := range mat.Variables {
		fmt.Printf("%s: %s %v\n", v.Name, v.DataType, v.Dimensions)

		if data, err := v.GetFloat64Array(); err == nil {
			fmt.Println("data:", data)
		}
	}
}
```

### Write a file

```go
package main

import (
	"log"

	"github.com/scigolib/matlab"
	"github.com/scigolib/matlab/types"
)

func main() {
	// matlab.Version5 writes the traditional format instead.
	writer, err := matlab.Create("output.mat", matlab.Version73)
	if err != nil {
		log.Fatal(err)
	}

	err = writer.WriteVariable(&types.Variable{
		Name:       "mydata",
		Dimensions: []int{3, 2},
		DataType:   types.Double,
		Data:       []float64{1, 2, 3, 4, 5, 6}, // column-major
	})
	if err != nil {
		_ = writer.Close()
		log.Fatal(err)
	}

	// Close finishes the file: check its error, do not defer it away.
	if err := writer.Close(); err != nil {
		log.Fatal(err)
	}
}
```

Complex data, byte order and the file description are covered in the
[Quick Start guide](docs/guides/QUICKSTART.md); runnable programs are in [examples](examples/README.md).

---

## Documentation

### Getting started
- [Quick Start](docs/guides/QUICKSTART.md): read, write, options and error handling
- [Examples](examples/README.md): four runnable programs

### Reference
- [Datatypes](docs/guides/DATATYPES.md): how MATLAB classes map to Go types, v5 and v7.3 differences, limitations
- [API reference](https://pkg.go.dev/github.com/scigolib/matlab) on pkg.go.dev
- [UPSTREAM.md](UPSTREAM.md): the references this library is built and checked against
- [CHANGELOG.md](CHANGELOG.md) and [ROADMAP.md](ROADMAP.md)

### Advanced
- [Architecture overview](docs/architecture/OVERVIEW.md): packages and how format detection works
- [Engineering quality](docs/QUALITY.md): the standard the library is held to and how it is enforced
- [AGENTS.md](AGENTS.md) and [llms.txt](llms.txt): the same guidance for AI coding agents

---

## Current Status

v0.3.19. Test coverage from `go test -cover`: 97.1% (root package), 97.4% (`internal/v5`), 86.8% (`internal/v73`),
100% (`types`).

### Reading

| Feature | v5 (v5 to v7.2) | v7.3 (HDF5) |
|---|---|---|
| Numeric arrays | yes | yes, returned as `[]float64` |
| Complex numbers | yes | group layout only (see below) |
| Character arrays | `DataType` `Char`; `Data` is a `string` or `[]uint16` by storage type | `DataType` `Char` when `MATLAB_class` is `char` |
| Multi-dimensional | yes | reported as one-dimensional |
| Compression | yes (zlib) | filters are decoded by the hdf5 module |
| Structures | class only | one variable per field |
| Cell arrays | class only | class only |
| Sparse matrices | not decoded | not supported |

- A v7.3 file may start with MATLAB's 512-byte MAT header; `Open` reads it (v0.3.18). A bare HDF5 file also opens.
- A MAT header whose version is neither 0x0100 nor 0x0200 fails with `ErrUnsupportedVersion`; input that is neither a MAT
  header nor HDF5 fails with `ErrInvalidFormat`; a truncated v5 file fails with an error that wraps
  `io.ErrUnexpectedEOF`.

### Writing

| Feature | v5 | v7.3 |
|---|---|---|
| Numeric arrays, all ten types | yes | yes |
| Complex numbers | yes | group layout, not MATLAB's (see below) |
| Multi-dimensional | yes | yes |
| Byte order | `WithEndianness`, little-endian by default | n/a |
| Character arrays | not supported | not supported |
| Structures, cell arrays | not supported | not supported |
| Compression | not supported | not supported |

- v7.3 files start with the 512-byte MAT header MATLAB writes in front of its own v7.3 files (v0.3.19).
- `WithCompression` is reserved and has no effect.

### Validation

- 283 top-level tests pass (`go test -v ./...`, counting `--- PASS` lines of top-level tests, run on 2026-11-07).
- v5 reading is checked against four reference files from SciPy's test data; the v5 writer against golden byte tests.
- Generated v7.3 files are opened with `h5dump` 1.14.6 locally (not in CI), and every generated file must open in the
  [hdf5](https://github.com/scigolib/hdf5) module without a warning.
- MATLAB itself is not available to the maintainers: loading our files in MATLAB is not verified.

### Limitations

- v7.3 complex arrays are written and read as an HDF5 group with `real` and `imag` datasets and a `MATLAB_complex`
  attribute. MATLAB stores them as one dataset of a compound type with `real` and `imag` members; that layout is not
  read or written, so MATLAB-written complex v7.3 arrays are not recognized as complex.
- Structures and cell arrays are not decoded. In v5, `DataType` is `Struct` or `CellArray` and `Data` is not usable;
  an empty v5 cell array fails to open. In v7.3 a structure comes back as one variable per field, named by its path
  (`/s/a`).
- Sparse matrices are not decoded: a v5 sparse matrix comes back with `DataType` `Unknown` and meaningless data.
- v7.3 reading returns numeric data as `[]float64` whatever the class, and reports one-dimensional dimensions.
- Character arrays cannot be written; `WriteVariable` returns an error for `types.Char`.
- A v5 compressed element may decompress to at most 100 MB, at a ratio of at most 1000:1; larger ones are rejected.
- Function handles and objects are not supported.

See [ROADMAP.md](ROADMAP.md) for what comes next and [docs/guides/DATATYPES.md](docs/guides/DATATYPES.md) for the details.

---

## Development

### Requirements
- Go 1.25 or later
- [`github.com/scigolib/hdf5`](https://github.com/scigolib/hdf5) for v7.3, a pure Go module that `go` fetches
- No C dependencies. The race detector needs cgo, so `make test-race` needs a C compiler.

### Building and testing

```bash
git clone https://github.com/scigolib/matlab.git
cd matlab

make build        # go build ./...
make test         # tests with coverage
make lint         # golangci-lint
make cross-build  # 32-bit targets
make ci           # formatting check, tests, lint, 32-bit builds

go run ./scripts/generate-testdata   # regenerate the v7.3 test files
go run ./scripts/verify-roundtrip    # write-and-read check
```

### Test data

- `testdata/scipy/`: 4 reference files from the SciPy project
- `testdata/generated/`: 8 v7.3 files written by our writer
- `testdata/legacy/`: 1 v7.3 file written by an early release, kept to prove it still reads

---

## Contributing

Contributions are welcome: bug reports with a sample file, fixes, documentation, and MAT-files from your own field
that the library should read. Before a pull request:

1. Read [CONTRIBUTING.md](CONTRIBUTING.md) for the workflow and the checks
2. Check the [open issues](https://github.com/scigolib/matlab/issues)
3. Look at the [architecture overview](docs/architecture/OVERVIEW.md)

Open areas: reading and writing v7.3 complex arrays in MATLAB's compound layout, decoding structures and cell arrays,
reading sparse matrices, writing character arrays, and compression in the writers.

---

## Other Go libraries

| | This library | [ready-steady/mat](https://github.com/ready-steady/mat) |
|---|---|---|
| Description | reader and writer of MAT-files, v5 and v7.3 | "Reader and writer of MATLAB MAT-files" (repository description) |
| Repository | active | last change March 2015 |

The comparison covers only what the repository pages state; the other library was not tested. For HDF5 itself see
[scigolib/hdf5](https://github.com/scigolib/hdf5), which this library uses for v7.3.

---

## MATLAB Resources

- [MAT-File Format](https://www.mathworks.com/help/pdf_doc/matlab/matfile_format.pdf) (MathWorks, the v5 format)
- [MAT-file versions](https://www.mathworks.com/help/matlab/import_export/mat-file-versions.html) (MathWorks)
- [matio](https://github.com/tbeu/matio), a C library for MAT-files that we compare against; see [UPSTREAM.md](UPSTREAM.md)

---

## License

MIT License; see [LICENSE](LICENSE).

---

## Acknowledgments

- The MathWorks, for the MAT-file format documentation
- The HDF Group, for the HDF5 format and the C library the v7.3 files are checked against
- The SciPy project, for the reference test files
- matio, whose handling of the MAT header we follow
- Everyone who reported issues and contributed code

---

## Support

- [Documentation](docs/guides/QUICKSTART.md): guides and architecture
- [Issue tracker](https://github.com/scigolib/matlab/issues): bugs and feature requests
- [Discussions](https://github.com/scigolib/matlab/discussions): questions
- [API reference](https://pkg.go.dev/github.com/scigolib/matlab)

---

## Star History

<a href="https://starhistory.io">
 <picture>
   <source media="(prefers-color-scheme: dark)" srcset="https://api.starhistory.io/png?repos=scigolib/matlab&style=dark" />
   <source media="(prefers-color-scheme: light)" srcset="https://api.starhistory.io/png?repos=scigolib/matlab&style=professional" />
   <img alt="Star History Chart" src="https://api.starhistory.io/png?repos=scigolib/matlab" width="800" />
 </picture>
</a>
