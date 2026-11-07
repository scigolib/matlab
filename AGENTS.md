# AGENTS.md: scigolib/matlab

> Pure Go reader and writer for MATLAB `.mat` files, v5 (MATLAB 5 to 7.2) and v7.3 (HDF5). No cgo.

Instructions for AI agents that use this library in their own code. To work on the library itself, see
[CONTRIBUTING.md](CONTRIBUTING.md).

## When to use it

Use it to read numeric arrays (real or complex) from `.mat` files, or to write numeric arrays that other tools
read. Do not use it for structures, cell arrays, sparse matrices, function handles or objects: they are not decoded
(see Limitations).

## Install

```bash
go get github.com/scigolib/matlab
```

Go 1.25 or later. Import `github.com/scigolib/matlab` and `github.com/scigolib/matlab/types`.

## Recipes

### Open a file and list its variables

`matlab.Open` takes an `io.Reader`, detects v5 or v7.3 itself, and returns `*matlab.MatFile`.

```go
import (
	"fmt"
	"os"

	"github.com/scigolib/matlab"
)

func listVariables(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	mat, err := matlab.Open(file)
	if err != nil {
		return fmt.Errorf("open %s: %w", path, err)
	}
	for _, v := range mat.Variables {
		fmt.Printf("%s: %s %v complex=%v\n", v.Name, v.DataType, v.Dimensions, v.IsComplex)
	}
	return nil
}
```

### Read numeric arrays

```go
import (
	"fmt"

	"github.com/scigolib/matlab"
)

func readNumbers(mat *matlab.MatFile, name string) ([]float64, error) {
	v := mat.GetVariable(name) // nil when absent
	if v == nil {
		return nil, fmt.Errorf("variable %q not found", name)
	}
	// Converts from float32 and the integer types; refuses complex data.
	return v.GetFloat64Array()
}
```

`Data` is column-major. A v5 file returns the Go slice of the storage type in the file (`[]float64`, `[]int16`, ...);
a v7.3 file always returns `[]float64` for numeric data, and reports one-dimensional `Dimensions` (a 2x3 matrix comes
back as `[6]`). `GetScalar` returns the single element of a 1x1 variable.

### Read complex data

```go
import (
	"fmt"

	"github.com/scigolib/matlab"
)

func readComplex(mat *matlab.MatFile, name string) ([]complex128, error) {
	v := mat.GetVariable(name)
	if v == nil {
		return nil, fmt.Errorf("variable %q not found", name)
	}
	if !v.IsComplex {
		return nil, fmt.Errorf("%q is not complex", name)
	}
	return v.GetComplex128Array()
}
```

### Read strings

A char variable has `DataType` `types.Char`. In a v5 file `Data` is a `string` when the element is stored as UTF-8 and
`[]uint16` when it is stored as `miUINT16`; check with a type switch. v7.3 char datasets are read as numbers first and
as strings if that fails. This path has no MATLAB-written test file, so verify it on your files.

```go
import (
	"fmt"

	"github.com/scigolib/matlab/types"
)

func describeChar(v *types.Variable) string {
	switch d := v.Data.(type) {
	case string:
		return d
	case []uint16:
		r := make([]rune, len(d))
		for i, c := range d {
			r[i] = rune(c)
		}
		return string(r)
	default:
		return fmt.Sprintf("unexpected %T", d)
	}
}
```

### Structures and cell arrays

You get the class, not the contents. In v5, `DataType` is `types.Struct` or `types.CellArray` and `Data` is not
usable. In v7.3 a structure comes back as one variable per field, named by its path (for example `/s/a`). Treat both as
unsupported.

### Write a file

`WriteVariable` needs `Name`, `Dimensions`, `DataType` and `Data`, with the slice type matching `DataType`.
Close the writer explicitly and check its error.

```go
import (
	"encoding/binary"
	"fmt"

	"github.com/scigolib/matlab"
	"github.com/scigolib/matlab/types"
)

func writeV5(path string) error {
	w, err := matlab.Create(path, matlab.Version5,
		matlab.WithEndianness(binary.BigEndian),
		matlab.WithDescription("results"))
	if err != nil {
		return err
	}
	err = w.WriteVariable(&types.Variable{
		Name:       "A",
		Dimensions: []int{2, 2},
		DataType:   types.Double,
		Data:       []float64{1, 2, 3, 4},
	})
	if err != nil {
		_ = w.Close()
		return fmt.Errorf("write A: %w", err)
	}
	return w.Close()
}

func writeV73(path string) error {
	w, err := matlab.Create(path, matlab.Version73)
	if err != nil {
		return err
	}
	err = w.WriteVariable(&types.Variable{
		Name:       "z",
		Dimensions: []int{2},
		DataType:   types.Double,
		IsComplex:  true,
		Data: &types.NumericArray{
			Real: []float64{1, 2},
			Imag: []float64{3, 4},
		},
	})
	if err != nil {
		_ = w.Close()
		return fmt.Errorf("write z: %w", err)
	}
	return w.Close()
}
```

Options: `WithEndianness` and `WithDescription` apply to v5. `WithCompression` is reserved and not implemented; it has no
effect, so do not use it to shrink files.

### Handle errors

```go
import (
	"errors"
	"io"

	"github.com/scigolib/matlab"
)

func classify(err error) string {
	switch {
	case err == nil:
		return "ok"
	case errors.Is(err, matlab.ErrInvalidFormat):
		return "not a MAT-file"
	case errors.Is(err, matlab.ErrUnsupportedVersion):
		return "MAT header version is neither v5 nor v7.3"
	case errors.Is(err, io.ErrUnexpectedEOF):
		return "truncated file"
	default:
		return "other error: " + err.Error()
	}
}
```

## Class and Go type mapping

| MATLAB class | `DataType` | Write `Data` | Read `Data` (v5 / v7.3) |
|---|---|---|---|
| double | `Double` | `[]float64` | `[]float64` / `[]float64` |
| single | `Single` | `[]float32` | `[]float32` / `[]float64` |
| int8 to int64, uint8 to uint64 | `Int8` to `Uint64` | the matching slice (`[]int16`, `[]uint32`, ...) | the matching slice / `[]float64` |
| complex | `IsComplex` true | `*types.NumericArray` | `*types.NumericArray` |
| char | `Char` | not supported | `string` or `[]uint16` (v5); see above |

Full table and limits: [docs/guides/DATATYPES.md](docs/guides/DATATYPES.md).

## Limitations and common mistakes

- `defer writer.Close()` loses the error that tells you the file is incomplete. Call `Close` and check it.
- v7.3 complex data is written as an HDF5 group with `real` and `imag`, not in MATLAB's compound layout. Do not expect
  MATLAB to load it, and files MATLAB writes with the compound layout are not read as complex.
- v7.3 files carry the 512-byte MAT header MATLAB writes; loading them in MATLAB itself is not verified.
- Structures and cell arrays are not decoded; an empty v5 cell array fails to open. Sparse matrices are not decoded: a
  v5 sparse matrix comes back with `DataType` `Unknown` and meaningless data.
- Character arrays cannot be written; `WriteVariable` returns an error for `types.Char`.
- A compressed v5 element may decompress to at most 100 MB, at a ratio of at most 1000:1; larger ones are rejected.
- `Data` is column-major. A one-dimensional variable written to v5 comes back with dimensions `[n 1]`.

## More

- [docs/guides/QUICKSTART.md](docs/guides/QUICKSTART.md), [examples/README.md](examples/README.md)
- [pkg.go.dev reference](https://pkg.go.dev/github.com/scigolib/matlab)
- Working on the library: [CONTRIBUTING.md](CONTRIBUTING.md)
