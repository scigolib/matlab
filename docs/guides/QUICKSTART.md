# Quick Start

Install, read a MAT-file, write one. Both formats go through the same two entry points: `matlab.Open` to read
(it detects v5 and v7.3) and `matlab.Create` to write.

## Install

```bash
go get github.com/scigolib/matlab
```

Go 1.25 or later. There is no C dependency; v7.3 support comes from the pure-Go
[`github.com/scigolib/hdf5`](https://github.com/scigolib/hdf5) module, which `go` fetches by itself.

## Read a file

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

	mat, err := matlab.Open(file)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("version:", mat.Version)
	for _, v := range mat.Variables {
		fmt.Printf("%s: %s %v\n", v.Name, v.DataType, v.Dimensions)
	}

	// Look a variable up by name and get its numbers as []float64.
	if v := mat.GetVariable("x"); v != nil {
		values, err := v.GetFloat64Array()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(values)
	}
}
```

`GetFloat64Array` converts from float32 and the integer types and refuses complex data; for complex variables use
`GetComplex128Array`. What `Data` holds for each class is in [DATATYPES.md](DATATYPES.md).

## Write a v5 file

```go
package main

import (
	"log"

	"github.com/scigolib/matlab"
	"github.com/scigolib/matlab/types"
)

func main() {
	writer, err := matlab.Create("output_v5.mat", matlab.Version5)
	if err != nil {
		log.Fatal(err)
	}

	// Data is column-major: this is the 2x3 matrix [1 3 5; 2 4 6].
	err = writer.WriteVariable(&types.Variable{
		Name:       "B",
		Dimensions: []int{2, 3},
		DataType:   types.Double,
		Data:       []float64{1, 2, 3, 4, 5, 6},
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

`matlab.Create` takes options: `WithEndianness(binary.BigEndian)` and `WithDescription("...")` for v5 files.
`WithCompression` is reserved and has no effect; nothing is compressed on write.

## Write a v7.3 file

```go
package main

import (
	"log"

	"github.com/scigolib/matlab"
	"github.com/scigolib/matlab/types"
)

func main() {
	writer, err := matlab.Create("output_v73.mat", matlab.Version73)
	if err != nil {
		log.Fatal(err)
	}

	// A complex variable: Real and Imag are slices of the type DataType names.
	err = writer.WriteVariable(&types.Variable{
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
		_ = writer.Close()
		log.Fatal(err)
	}

	if err := writer.Close(); err != nil {
		log.Fatal(err)
	}
}
```

The file starts with the 512-byte MAT header MATLAB writes in front of its own v7.3 files (v0.3.19). `h5dump`
opens it; loading it in MATLAB itself is not verified. The complex variable is stored as a group, not in MATLAB's
compound layout (see [DATATYPES.md](DATATYPES.md)).

## Handle errors

```go
package main

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/scigolib/matlab"
)

func main() {
	file, err := os.Open("data.mat")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer file.Close()

	_, err = matlab.Open(file)
	switch {
	case err == nil:
		fmt.Println("ok")
	case errors.Is(err, matlab.ErrInvalidFormat):
		fmt.Println("not a MAT-file")
	case errors.Is(err, matlab.ErrUnsupportedVersion):
		fmt.Println("a MAT header with a version other than v5 or v7.3")
	case errors.Is(err, io.ErrUnexpectedEOF):
		fmt.Println("the file is truncated")
	default:
		fmt.Println("read failed:", err)
	}
}
```

## Next

- [Datatypes](DATATYPES.md): classes and Go types, v5 and v7.3 differences, limitations
- [Examples](../../examples/README.md): runnable programs
- [API reference](https://pkg.go.dev/github.com/scigolib/matlab)
