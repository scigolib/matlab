# Upstream references

What MAT-file behaviour this library is built and checked against. Last reviewed 2026-11-07.

## References

| Reference | Pinned to | What we follow |
|---|---|---|
| [matio](https://github.com/tbeu/matio) | v1.6.0, commit 86c1878 | `src/mat.c` `Mat_Open`: the format is told by the header's version field, 0x0100 is v5 and 0x0200 is v7.3. `src/mat73.c` `Mat_Create73`: a v7.3 file is HDF5 behind a 512-byte user block that holds the 128-byte MAT header (our v7.3 writer does the same). The same file stores a complex v7.3 variable as a compound type with `real` and `imag` members (around lines 465 and 1615) |
| [SciPy](https://github.com/scipy/scipy) | v1.18.1, commit e4e854eaa | `scipy/io/matlab/_miobase.py` `_get_matfile_version`: major version 1 is v5, 2 is v7.3, anything else is an error. The files in `testdata/scipy` come from the SciPy project |
| [HDF5 C library](https://github.com/HDFGroup/hdf5) | develop, commit bc8f6bb02aa (2025-12-18) | `src/H5FDint.c` `H5FD_locate_signature`: the HDF5 signature is looked for at offset 0 and then at every power of two from 512, which is where a MATLAB v7.3 file has it |
| MathWorks, [MAT-File Format](https://www.mathworks.com/help/pdf_doc/matlab/matfile_format.pdf) | PDF | the v5 layout: header, data element tags, small data elements, matrix sub-elements, compression, 8-byte alignment |
| MathWorks, [MAT-file versions](https://www.mathworks.com/help/matlab/import_export/mat-file-versions.html) | web page | which MATLAB versions write v5 and v7.3 |

## Test corpus

| Directory | Content | Origin |
|---|---|---|
| `testdata/scipy` | 4 files: `testdouble_7.4_GLNX86.mat`, `testcomplex_7.4_GLNX86.mat`, `testmatrix_7.4_GLNX86.mat`, `inner_outer_tbl_param.mat` | MATLAB 7.4 files from the SciPy test data and one scientific data set; the header comment of `scipy_test.go` describes each |
| `testdata/generated` | 8 v7.3 files | written by our own writer (`go run ./scripts/generate-testdata`); `h5dump` 1.14.6 opens all of them |
| `testdata/legacy` | 1 v7.3 file | written by matlab v0.1.0-beta; kept to show that early files still read |

## Deviations from the references

- v7.3 complex variables are written and read as an HDF5 group with datasets `real` and `imag` and a `MATLAB_complex`
  attribute. MATLAB and matio use one dataset of a compound type with `real` and `imag` members. We do not read or write
  that layout.
- Structures and cell arrays are not decoded; sparse matrices are not supported.
- Character arrays cannot be written.

## How files are checked

- Files our writers produce are read back by our reader (round-trip tests in the repository root, and
  `go run ./scripts/verify-roundtrip`).
- Generated v7.3 files are opened with `h5dump` on the maintainer's machine. This is not a CI job.
- `TestGeneratedFixtures_NoHDF5Warnings` requires every generated v7.3 file to open in the `hdf5` module without a
  warning.
- v5 reading is checked against the four SciPy files; v5 writing is checked by golden byte tests
  (`internal/v5/golden_test.go`) and the small-data-element tests (`internal/v5/sde_test.go`).
- Header detection is tested with MATLAB-style 512-byte user blocks in both byte orders.
- MATLAB itself is not available to the maintainers, so loading our files in MATLAB is unverified.
