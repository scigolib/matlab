# 02-write-complex

Writes `z = [1+2i, 3+4i, 5+6i]` to `complex_example.mat` in the current directory, in v7.3 format.

```bash
go run ./examples/02-write-complex
h5dump complex_example.mat     # optional: needs the HDF5 command-line tools
```

The v7.3 writer stores a complex variable as a group `/z` with the datasets `real` and `imag` and the attributes
`MATLAB_class` and `MATLAB_complex`. MATLAB itself uses one dataset of a compound type; this library does not read or
write that layout. See [DATATYPES.md](../../docs/guides/DATATYPES.md).

The program closes the writer explicitly and returns the error from `Close`.
