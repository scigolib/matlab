# MATLAB File Reader/Writer for Go

> **Pure Go implementation for reading AND writing MATLAB `.mat` files** - No CGo required

[![GitHub Release](https://img.shields.io/github/v/release/scigolib/matlab?include_prereleases&style=flat-square&logo=github&color=blue)](https://github.com/scigolib/matlab/releases/latest)
[![Go Version](https://img.shields.io/badge/Go-1.25%2B-00ADD8?style=flat-square&logo=go)](https://go.dev/dl/)
[![Go Reference](https://pkg.go.dev/badge/github.com/scigolib/matlab.svg)](https://pkg.go.dev/github.com/scigolib/matlab)
[![GitHub Actions](https://img.shields.io/github/actions/workflow/status/scigolib/matlab/test.yml?branch=main&style=flat-square&logo=github-actions&label=CI)](https://github.com/scigolib/matlab/actions)
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
- ✍️ **Write MATLAB v7.3+ files** (HDF5 format) - NEW!
- ✍️ Write v5 format (coming in v0.2.0)

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
- Structures (partial support)
- Cell arrays (partial support)

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
	defer writer.Close()

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
}
```

## Supported Features

### Reader Support

| Feature              | v5 (v5-v7.2) | v7.3+ (HDF5) |
|----------------------|--------------|--------------|
| Numeric arrays       | ✅           | ✅           |
| Complex numbers      | ✅           | ✅           |
| Character arrays     | ✅           | ✅           |
| Multi-dimensional    | ⚠️ Partial   | ✅           |
| Structures           | ⚠️ Partial   | ⚠️ Partial   |
| Cell arrays          | ⚠️ Partial   | ⚠️ Partial   |
| Sparse matrices      | ❌           | ⚠️ Limited   |
| Compression          | ❌           | ❌           |
| Function handles     | ❌           | ❌           |
| Objects              | ❌           | ❌           |

### Writer Support (v0.1.1-beta)

| Feature              | v5 (v5-v7.2) | v7.3+ (HDF5) |
|----------------------|--------------|--------------|
| Numeric arrays       | 📋 Planned   | ✅           |
| Complex numbers      | 📋 Planned   | ✅*          |
| Character arrays     | 📋 Planned   | ✅           |
| Multi-dimensional    | 📋 Planned   | ✅           |
| Structures           | ❌ Future    | ❌ Future    |
| Cell arrays          | ❌ Future    | ❌ Future    |
| Compression          | ❌ Future    | ❌ Future    |

## Known Limitations (v0.1.1-beta)

### Writer Limitations
- **v5 format writing not yet implemented** (coming in v0.2.0)
- No compression support yet
- No structures/cell arrays writing yet

### Reader Issues
- Some bugs with multi-dimensional arrays (being fixed)
- Limited support for structures and cell arrays
- Multiple variables in one file may have issues

### What Works Well ✅
- All numeric types (double, single, int8-64, uint8-64)
- Multi-dimensional arrays (write)
- **Complex numbers** (proper MATLAB v7.3 format) ✨ FIXED in v0.1.1-beta
- Round-trip write → read verified
- Cross-platform (Windows, Linux, macOS)

See [CHANGELOG.md](CHANGELOG.md) for detailed limitations and planned fixes.

## Documentation

- **[Getting Started](docs/)** - Basic usage examples
- **[API Reference](https://pkg.go.dev/github.com/scigolib/matlab)** - Full API documentation
- **[Architecture](.claude/CLAUDE.md)** - Internal architecture and design decisions
- **[CHANGELOG.md](CHANGELOG.md)** - Version history and changes
- **[ROADMAP.md](ROADMAP.md)** - Future plans and development timeline

## Development

### Requirements
- Go 1.25 or later
- HDF5 library (for v7.3+ support): `github.com/scigolib/hdf5` develop branch (commit 36994ac)
- No external C dependencies

### Building

```bash
# Clone repositories (side-by-side)
cd D:\projects\scigolibs
git clone https://github.com/scigolib/hdf5.git
git clone https://github.com/scigolib/matlab.git

# Build MATLAB library
cd matlab
make build

# Run tests
make test

# Run linter
make lint

# Generate test data
go run scripts/generate-testdata/main.go

# Verify round-trip
go run scripts/verify-roundtrip/main.go
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
- `testdata/generated/` - Files created by our writer (8 files)
- `testdata/scipy/` - Reference files from SciPy project (3 files)

---

## Contributing

Contributions are welcome! This is an early-stage beta project and we'd love your help.

**Before contributing**:
1. Read [CONTRIBUTING.md](CONTRIBUTING.md) - Git workflow and development guidelines
2. Check [open issues](https://github.com/scigolib/matlab/issues)
3. Review the architecture in `.claude/CLAUDE.md`

**Ways to contribute**:
- 🐛 Report bugs
- 💡 Suggest features
- 📝 Improve documentation
- 🔧 Submit pull requests
- ⭐ Star the project
- 🧪 Test with real MATLAB files and report compatibility

**Priority Areas**:
- Implement v5 writer (TASK-011)
- Fix reader bugs (multi-dimensional arrays, multiple variables)
- Test MATLAB/Octave compatibility
- Improve test coverage (target: 80%+)

---

## Comparison with Other Libraries

| Feature | This Library | go-hdf5/* | matlab-go |
|---------|-------------|-----------|-----------|
| Pure Go | ✅ Yes | ❌ CGo required | ✅ Yes |
| v5-v7.2 Read | ✅ Yes | ❌ Limited | ⚠️ Partial |
| v7.3+ Read | ✅ Yes | ❌ No | ❌ No |
| **Write Support** | ✅ **v7.3 Yes** | ❌ No | ❌ No |
| Complex Numbers | ✅ Yes | ⚠️ Limited | ❌ No |
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

- 📖 Documentation - See `.claude/CLAUDE.md` for architecture details
- 🐛 [Issue Tracker](https://github.com/scigolib/matlab/issues)
- 💬 Discussions - GitHub Discussions (coming soon)
- 📧 Contact - Via GitHub Issues

---

**Status**: Beta - Read and Write support for v7.3 format (proper complex numbers!)
**Version**: v0.1.1-beta
**Last Updated**: 2025-11-03

**Ready for**: Testing, feedback, and real-world usage
**Not ready for**: Production use (API may change)

---

*Built with ❤️ by the SciGoLib community*
