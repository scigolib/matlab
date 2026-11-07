# 01-read

Opens a MAT-file with `matlab.Open`, which detects v5 and v7.3, and prints the header fields and every variable.

```bash
go run ./examples/01-read testdata/scipy/testdouble_7.4_GLNX86.mat
```

Without an argument it prints a usage line. Errors from opening and parsing are returned and reported, not ignored.
