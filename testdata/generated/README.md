# MATLAB Test Data

This directory contains minimal MATLAB files for testing.

## Files

| File | Format | Description | Variable | Type | Dimensions |
|------|--------|-------------|----------|------|------------|
| simple_double.mat | v7.3 | Simple 1D double array | data | double | [5] |
| simple_int32.mat | v7.3 | Simple 1D int32 array | values | int32 | [4] |
| simple_uint8.mat | v7.3 | Simple 1D uint8 array | bytes | uint8 | [3] |
| simple_single.mat | v7.3 | Simple 1D single array | floats | single | [3] |
| complex.mat | v7.3 | Complex numbers | z | double | [3] |
| matrix_2x3.mat | v7.3 | 2x3 matrix | matrix | double | [2, 3] |
| matrix_3x2.mat | v7.3 | 3x2 matrix | A | double | [3, 2] |
| scalar.mat | v7.3 | Scalar value | x | double | [1] |

## Generation

These files were generated using our own writer implementation:

```bash
go run scripts/generate-testdata.go
```

## Testing

Use these files for:
- Reader integration tests
- Round-trip verification (write → read → compare)
- MATLAB compatibility testing
- Performance benchmarking

## Notes

- All files are v7.3 format (HDF5-based)
- Files use MATLAB_class attributes for type info
- Data is stored in column-major order (MATLAB convention)
- Complex numbers use separate real/imaginary datasets
