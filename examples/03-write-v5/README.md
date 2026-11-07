# 03-write-v5

Writes a vector `A` and a 2x3 matrix `B` to a v5 file in a temporary directory, with `WithEndianness` and
`WithDescription`, then reads the file back and prints it.

```bash
go run ./examples/03-write-v5
```

`Data` is column-major, as in MATLAB. A one-dimensional variable comes back from a v5 file with dimensions `[3 1]`.
