# 04-write-v73

Writes a double vector `x` and an int32 vector `counts` to a v7.3 file in a temporary directory, then reads the file
back.

```bash
go run ./examples/04-write-v73
```

The reader returns v7.3 numeric data as `[]float64` whatever the MATLAB class, and reports a one-dimensional shape.
The class is in `DataType`. The file starts with the 512-byte MAT header; loading it in MATLAB itself is not verified.
