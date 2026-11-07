# MATLAB File Reader/Writer for Go

> **Pure Go implementation for reading AND writing MATLAB `.mat` files** - No CGo required

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

---

A modern, pure Go library for **reading and writing** MATLAB `.mat` files without CGo dependencies. Part of the [SciGoLib](https://github.com/scigolib) scientific computing ecosystem.

## Features

✨ **Read & Write Support**
- 📖 Read MATLAB v5-v7.2 files (traditional format)
- 📖 Read MATLAB v7.3+ files (HDF5 format)
- ✍️ **Write MATLAB v7.3+ files** (HDF5 format)
- ✍️ **Write MATLAB v5-v7.2 files** (traditional format)

🎯 **Key Capabilities**
- Simple, intuitive API
- Zero external C dependencies
- Type-safe data access
- Comprehensive error handling
- Round-trip verified (write → read → verify)

📊 **Data Types**
- All numeric types (double, single, int8-64, uint8-64)
- Complex numbers
- Multi-dimensional arrays
- Character arrays
- Compressed data (v5 reading)
- Structures and cell arrays: not decoded yet (see [Known Limitations](#known-limitations))
- Sparse matrices: not supported yet

## Installation

```bash
go get github.com/scigolib/matlab
```

## Quick Start

### Reading MAT-Files

```go
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/scigolib/matlab"
)

func main() {
	// Open MAT-file
	file, err := os.Open("data.mat")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	// Parse MAT-file (auto-detects v5 or v7.3)
	mat, err := matlab.Open(file)
	if err != nil {
		log.Fatal(err)
	}

	// Access variables
	for _, v := range mat.Variables {
		fmt.Printf("%s: %s %v\n", v.Name, v.DataType, v.Dimensions)

		// Access data based on type
		if data, ok := v.Data.([]float64); ok {
			fmt.Println("Data:", data)
		}
	}
}
```

### Writing MAT-Files

#### v7.3 Format (HDF5-based)

```go
package main

import (
	"log"

	"github.com/scigolib/matlab"
	"github.com/scigolib/matlab/types"
)

func main() {
	// Create new MAT-file (v7.3 format)
	writer, err := matlab.Create("output.mat", matlab.Version73)
	if err != nil {
		log.Fatal(err)
	}

	// Write a variable
	err = writer.WriteVariable(&types.Variable{
		Name:       "mydata",
		Dimensions: []int{3, 2},
		DataType:   types.Double,
		Data:       []float64{1.0, 2.0, 3.0, 4.0, 5.0, 6.0},
	})
	if err != nil {
		log.Fatal(err)
	}

	// Write complex numbers
	err = writer.WriteVariable(&types.Variable{
		Name:       "z",
		Dimensions: []int{2},
		DataType:   types.Double,
		IsComplex:  true,
		Data: &types.NumericArray{
			Real: []float64{1.0, 2.0},
			Imag: []float64{3.0, 4.0},
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	// Close finishes the file: check its error, do not defer it away
	if err := writer.Close(); err != nil {
		log.Fatal(err)
	}
}
```

#### v5 Format (Traditional Binary)

```go
package main

import (
	"log"

	"github.com/scigolib/matlab"
	"github.com/scigolib/matlab/types"
)

func main() {
	// Create new MAT-file (v5 format - legacy compatible)
	writer, err := matlab.Create("output_v5.mat", matlab.Version5)
	if err != nil {
		log.Fatal(err)
	}

	// Write a simple array
	err = writer.WriteVariable(&types.Variable{
		Name:       "A",
		Dimensions: []int{3},
		DataType:   types.Double,
		Data:       []float64{1.0, 2.0, 3.0},
	})
	if err != nil {
		log.Fatal(err)
	}

	// Write a matrix (multi-dimensional)
	err = writer.WriteVariable(&types.Variable{
		Name:       "B",
		Dimensions: []int{2, 3},
		DataType:   types.Double,
		Data:       []float64{1, 2, 3, 4, 5, 6},
	})
	if err != nil {
		log.Fatal(err)
	}

	// Write complex numbers
	err = writer.WriteVariable(&types.Variable{
		Name:       "C",
		Dimensions: []int{2},
		DataType:   types.Double,
		IsComplex:  true,
		Data: &types.NumericArray{
			Real: []float64{1.0, 2.0},
			Imag: []float64{3.0, 4.0},
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	// Close finishes the file: check its error, do not defer it away
	if err := writer.Close(); err != nil {
		log.Fatal(err)
	}
}
```

## Supported Features

### Reader Support

| Feature              | v5 (v5-v7.2) | v7.3+ (HDF5) |
|----------------------|--------------|--------------|
| Numeric arrays       | ✅           | ✅           |
| Complex numbers      | ✅           | ⚠️ ¹         |
| Character arrays     | ✅           | ✅           |
| Multi-dimensional    | ✅           | ✅           |
| Structures           | ⚠️ class only ² | ⚠️ fields flattened ³ |
| Cell arrays          | ⚠️ class only ² | ⚠️ class only ² |
| Sparse matrices      | ❌ Not yet   | ❌ Not yet   |
| Compression          | ✅           | ✅           |
| Function handles     | ❌ Out of scope | ❌ Out of scope |
| Objects              | ❌ Out of scope | ❌ Out of scope |

### Writer Support

| Feature              | v5 (v5-v7.2) | v7.3+ (HDF5) |
|----------------------|--------------|--------------|
| Numeric arrays       | ✅           | ✅           |
| Complex numbers      | ✅           | ⚠️ ¹         |
| Character arrays     | ✅           | ✅           |
| Multi-dimensional    | ✅           | ✅           |
| Both endianness      | ✅ MI/IM     | N/A          |
| Structures           | 📅 v0.5.0+   | 📅 v0.5.0+   |
| Cell arrays          | 📅 v0.5.0+   | 📅 v0.5.0+   |
| Compression          | 📅 v0.5.0+   | 📅 v0.5.0+   |

¹ v7.3 complex arrays are written as an HDF5 group with `real` and `imag`
datasets and a `MATLAB_complex` attribute, and this layout is read back.
MATLAB itself stores a v7.3 complex array as one dataset of a compound type
with `real` and `imag` members; that layout is not read or written yet.

² The variable's `DataType` is reported (`Struct`, `CellArray`); fields and
elements are not decoded yet, and `Data` holds the undecoded first element, not
usable values.

³ A v7.3 structure is an HDF5 group: each field comes back as its own variable
named by its path (`/s/a`), not as one `Struct` variable.

## Known Limitations

### Writer Limitations
- No compression support (planned for v0.5.0+)
- No structures/cell arrays writing (planned for v0.5.0+)
- v7.3 complex arrays use the group layout described above, not MATLAB's
  compound layout
- v7.3 files carry the 512-byte MAT header MATLAB writes in front of its own
  v7.3 files; loading them in MATLAB itself is not verified yet (no MATLAB
  available)

### Reader Limitations
- Structures and cell arrays are not decoded: v5 reports the class only, a v7.3
  structure comes back as one variable per field; an empty v5 cell array fails
  to open
- Sparse matrices are not decoded: a v5 sparse matrix comes back with
  `DataType` `Unknown` and meaningless data
- A v5 compressed element may decompress to at most 100 MB (a guard against
  compression bombs); larger variables in compressed files are rejected
- v7.3 complex arrays in MATLAB's compound layout are not recognized
- Function handles not supported (MATLAB-specific, cannot be serialized)
- Objects not supported (language-specific)

### What Works Well ✅
- ✅ **v5 writer** - All numeric types, complex numbers, multi-dimensional arrays
- ✅ **v7.3 writer** - HDF5-based writing; the files open with the HDF5 C tools (`h5dump`)
- ✅ All numeric types (double, single, int8-64, uint8-64)
- ✅ Multi-dimensional arrays (read & write)
- ✅ Round-trip verified (v5 write → read, v7.3 write → read)
- ✅ v5 files validated against reference files from SciPy
- ✅ Malformed v5 input is rejected with an error: sizes are checked against the input before any allocation, and the parser is fuzz-tested
- ✅ Cross-platform (Windows, Linux, macOS; 32-bit builds checked in CI)
- ✅ Both endianness (MI/IM for v5)

See [CHANGELOG.md](CHANGELOG.md) for version history.

## Documentation

- **[API Reference](https://pkg.go.dev/github.com/scigolib/matlab)** - Full API documentation with runnable examples
- **[CONTRIBUTING.md](CONTRIBUTING.md)** - Development workflow and project structure
- **[CHANGELOG.md](CHANGELOG.md)** - Version history and changes
- **[ROADMAP.md](ROADMAP.md)** - Future plans and development timeline

## Development

### Requirements
- Go 1.25 or later
- HDF5 library (for v7.3+ support): [`github.com/scigolib/hdf5`](https://github.com/scigolib/hdf5), a pure Go module fetched by `go`
- No external C dependencies

### Building

```bash
git clone https://github.com/scigolib/matlab.git
cd matlab

# Build MATLAB library
make build

# Run tests
make test

# Run linter
make lint

# All CI checks: formatting, tests, lint, 32-bit builds
make ci

# Regenerate the v7.3 test files
go run ./scripts/generate-testdata

# Verify round-trip
go run ./scripts/verify-roundtrip
```

### Testing

```bash
# Run all tests
make test

# Run with coverage
make test-coverage

# Run specific tests
go test ./internal/v73 -v

# Run linter
make lint
```

### Test Data

The project includes test data in `testdata/`:
- `testdata/generated/` - v7.3 files created by our writer (8 files)
- `testdata/legacy/` - A v7.3 file written by an early release, kept to prove it still reads (1 file)
- `testdata/scipy/` - Reference files from the SciPy project (4 files)

---

## Contributing

Contributions are welcome! This is a stable project and we'd love your help.

**Before contributing**:
1. Read [CONTRIBUTING.md](CONTRIBUTING.md) - Git workflow, development guidelines and project structure
2. Check [open issues](https://github.com/scigolib/matlab/issues)

**Ways to contribute**:
- 🐛 Report bugs
- 💡 Suggest features
- 📝 Improve documentation
- 🔧 Submit pull requests
- ⭐ Star the project
- 🧪 Test with real MATLAB files and report compatibility

**Priority Areas**:
- Test MATLAB/Octave compatibility with real-world files
- Read and write v7.3 complex arrays in MATLAB's compound layout
- Decode structures and cell arrays; read sparse matrices
- Add compression support for v5/v7.3 writers
- Implement structures and cell arrays writing

---

## Comparison with Other Libraries

| Feature | This Library | go-hdf5/* | matlab-go |
|---------|-------------|-----------|-----------|
| Pure Go | ✅ Yes | ❌ CGo required | ✅ Yes |
| v5-v7.2 Read | ✅ Yes | ❌ Limited | ⚠️ Partial |
| v7.3+ Read | ✅ Yes | ❌ No | ❌ No |
| **Write Support** | ✅ **v5 + v7.3 Yes** | ❌ No | ❌ No |
| Complex Numbers | ✅ Yes (v7.3: own layout ¹) | ⚠️ Limited | ❌ No |
| Maintained | ✅ Active | ❌ Inactive | ❌ Inactive |
| Cross-platform | ✅ Yes | ⚠️ Platform-specific | ✅ Yes |

---

## Related Projects

- **[HDF5 Go Library](https://github.com/scigolib/hdf5)** - Pure Go HDF5 implementation (used for v7.3+ support)
- Part of [SciGoLib](https://github.com/scigolib) - Scientific computing libraries for Go

---

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

## Acknowledgments

- The MathWorks for the MATLAB file format specification
- The HDF Group for HDF5 format specification
- [scigolib/hdf5](https://github.com/scigolib/hdf5) for HDF5 support
- SciPy project for reference test data
- All contributors to this project

---

## Support

- 📖 [Documentation](https://pkg.go.dev/github.com/scigolib/matlab) - API reference and examples
- 🐛 [Issue Tracker](https://github.com/scigolib/matlab/issues)
- 💬 [Discussions](https://github.com/scigolib/matlab/discussions)
- 📧 Contact - Via GitHub Issues

---

**Status**: ✅ **STABLE** - read and write support for numeric data in both v5 and v7.3 formats (see [Known Limitations](#known-limitations))
**Last Updated**: 2026-11-07

**Ready for**: Production use, testing, feedback, and real-world usage
**Stable API**: Minor API changes may occur in 0.x versions, major stability expected

---

*Built with ❤️ by the SciGoLib community*


## Star History

<a href="https://starhistory.io">
 <picture>
   <source media="(prefers-color-scheme: dark)" srcset="https://api.starhistory.io/png?repos=scigolib/matlab&style=dark" />
   <source media="(prefers-color-scheme: light)" srcset="https://api.starhistory.io/png?repos=scigolib/matlab&style=professional" />
   <img alt="Star History Chart" src="https://api.starhistory.io/png?repos=scigolib/matlab" width="800" />
 </picture>
</a>
