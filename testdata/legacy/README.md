# Legacy test files

Files written by earlier releases, kept to prove they still read correctly.
Do not regenerate them.

| File | Written by | Variable | Content |
|------|------------|----------|---------|
| simple_double_hdf5go_v0.11.mat | matlab v0.1.0-beta on hdf5-go v0.11.4-beta | data | double [1, 2, 3, 4, 5] |

hdf5-go v0.13.9 and earlier wrote v7.3 files that the HDF5 C tools (h5dump,
h5py, MATLAB) cannot open: a CRC32 superblock checksum, object headers without
a valid checksum and an early datatype encoding. github.com/scigolib/hdf5
v0.14.3 reads such files and reports each structure through `File.Warnings()`.
To make a file readable by the C tools, read it and write it again with the
current release.
