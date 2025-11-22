# Security Policy

## Supported Versions

MATLAB File Reader/Writer is currently in stable release. We provide security updates for the following versions:

| Version | Supported          |
| ------- | ------------------ |
| 0.3.x   | :white_check_mark: |
| 0.2.x   | :white_check_mark: |
| 0.1.x   | :x:                |
| < 0.1.0 | :x:                |

Future stable releases (v1.0+) will follow semantic versioning with LTS support.

## Reporting a Vulnerability

We take security seriously. If you discover a security vulnerability in MATLAB File Reader/Writer, please report it responsibly.

### How to Report

**DO NOT** open a public GitHub issue for security vulnerabilities.

Instead, please report security issues by:

1. **Private Security Advisory** (preferred):
   https://github.com/scigolib/matlab/security/advisories/new

2. **Email** to maintainers:
   Create a private GitHub issue or contact via discussions

### What to Include

Please include the following information in your report:

- **Description** of the vulnerability
- **Steps to reproduce** the issue (include malicious MAT-file test file if applicable)
- **Affected versions** (which versions are impacted)
- **Potential impact** (DoS, information disclosure, code execution, etc.)
- **Suggested fix** (if you have one)
- **Your contact information** (for follow-up questions)

### Response Timeline

- **Initial Response**: Within 48-72 hours
- **Triage & Assessment**: Within 1 week
- **Fix & Disclosure**: Coordinated with reporter

We aim to:
1. Acknowledge receipt within 72 hours
2. Provide an initial assessment within 1 week
3. Work with you on a coordinated disclosure timeline
4. Credit you in the security advisory (unless you prefer to remain anonymous)

## Security Fixes in v0.3.0 (2025-11-21)

Three critical security issues were identified through deep code analysis and fixed in v0.3.0:

### 1. Tag Size Validation

**Severity**: High
**Type**: Memory Exhaustion / Denial of Service
**Affected Versions**: v0.1.0-beta to v0.2.0
**Fixed in**: v0.3.0 (2025-11-21)

**Vulnerability Description**:
No validation was performed on v5 format tag sizes read from untrusted MAT files. An attacker could craft a malicious MAT file with extremely large tag size values (e.g., 0xFFFFFFFF = 4GB+), causing the library to attempt massive memory allocations.

**Attack Vector**:
```
Malicious v5 MAT file:
- Tag header with size = 0xFFFFFFFF (4,294,967,295 bytes)
- Library attempts to allocate 4GB+ buffer
- System runs out of memory (DoS)
- Or library crashes (availability issue)
```

**Fix Implementation**:
```go
// internal/v5/data_tag.go:8-11
const maxReasonableSize = 2 * 1024 * 1024 * 1024 // 2GB limit

// internal/v5/data_tag.go:53-55
if size > maxReasonableSize {
    return nil, fmt.Errorf("tag size too large: %d bytes (max %d)",
        size, maxReasonableSize)
}
```

**Impact**: Prevents memory exhaustion attacks. Returns error for tags > 2GB.

**Credit**: Internal security audit (2025-11-21)

---

### 2. Dimension Overflow Check

**Severity**: High
**Type**: Integer Overflow / Buffer Overflow
**Affected Versions**: v0.1.0-beta to v0.2.0
**Fixed in**: v0.3.0 (2025-11-21)

**Vulnerability Description**:
No overflow check was performed when multiplying array dimensions to calculate total element count. An attacker could specify dimensions that overflow when multiplied, leading to incorrect buffer allocation and potential buffer overflow.

**Attack Vector**:
```
Malicious dimensions: [0xFFFFFFFF, 0xFFFFFFFF]
Multiplication: 0xFFFFFFFF * 0xFFFFFFFF
Expected result: ~1.8e19 elements (overflow)
Actual result (int64): Small positive number (wraps around)
Result: Small buffer allocated, large data written → buffer overflow
```

**Fix Implementation**:
```go
// internal/v5/writer.go:124-129
total := int64(1)
for _, d := range dims {
    if d > 0 && total > math.MaxInt/int64(d) {
        return fmt.Errorf("dimensions overflow: %v", dims)
    }
    total *= int64(d)
}

// Same fix in internal/v73/writer.go:98-103
```

**Impact**: Prevents integer overflow in dimension calculations. Validates before allocation.

**Credit**: Internal security audit (2025-11-21)

---

### 3. v73 Complex Number Reading

**Severity**: Medium (Functionality, not security)
**Type**: Data Corruption / Round-trip Failure
**Affected Versions**: v0.1.0-beta to v0.2.0
**Fixed in**: v0.3.0 (2025-11-21)

**Issue Description**:
v7.3 format complex number groups were not properly detected during reading. The `MATLAB_complex` attribute was ignored, causing complex numbers to be read incorrectly and round-trip tests to fail.

**Problem**:
```
v7.3 file structure for complex number:
/variable_name (Group)
    ├── MATLAB_class attribute = "double"
    ├── MATLAB_complex attribute = "1"
    ├── /real (Dataset)
    └── /imag (Dataset)

Problem: MATLAB_complex attribute was not checked
Result: Group treated as structure, not complex number
Impact: Round-trip write → read → verify FAILED
```

**Fix Implementation**:
```go
// internal/v73/adapter.go:50-65
// Check for MATLAB_complex attribute
for _, attr := range group.Attributes() {
    if attr.Name == "MATLAB_complex" {
        variable, err := a.convertComplexGroup(group, name)
        if err == nil {
            *variables = append(*variables, variable)
            return nil
        }
    }
}

// internal/v73/adapter.go:179-247
// New function: convertComplexGroup()
func (a *HDF5Adapter) convertComplexGroup(group *hdf5.Group, name string)
    (*types.Variable, error) {
    // Opens /real and /imag datasets
    // Creates Variable with IsComplex=true
    // Returns NumericArray{Real, Imag}
}
```

**Impact**: Fixes round-trip for v7.3 complex numbers. Proper MATLAB compatibility.

**Credit**: Round-trip testing (2025-11-21)

---

## Security Considerations for MATLAB File Parsing

MATLAB `.mat` files are complex binary formats. This library parses untrusted binary data, which introduces security risks.

### 1. Malicious MATLAB Files

**Risk**: Crafted MAT-files can exploit parsing vulnerabilities.

**Attack Vectors**:
- Integer overflow in array dimensions, matrix sizes, or buffer allocations
- Buffer overflow when reading headers, tags, or data elements
- Infinite loops in tag-length-value (TLV) structure parsing
- Resource exhaustion via deeply nested structures or massive arrays
- Compression bomb attacks (v5 compressed data, v7.3 HDF5 compression)

**Mitigation in Library**:
- ✅ Bounds checking on all size fields (v5 tag sizes, dimensions)
- ✅ Validation of MAT-file signatures and magic numbers
- ✅ Sanity checks on dimension sizes, tag lengths, and offsets
- ✅ 8-byte alignment verification in v5 format
- ✅ HDF5 signature validation for v7.3 format
- 🔄 Ongoing fuzzing and security testing (planned for v1.0)

**User Recommendations**:
```go
// ❌ BAD - Don't trust untrusted MAT-files without validation
file, _ := os.Open(userUploadedFile)
matFile, _ := matlab.Open(file)

// ✅ GOOD - Validate file size and structure first
fileInfo, err := os.Stat(filename)
if err != nil || fileInfo.Size() > maxAllowedSize {
    return errors.New("file too large")
}

file, err := os.Open(filename)
if err != nil {
    return err
}
defer file.Close()

matFile, err := matlab.Open(file)
if err != nil {
    // Parsing failed - potentially malicious file
    return err
}
```

### 2. Integer Overflow Vulnerabilities

**Risk**: MATLAB files use various integer sizes (8/16/32/64-bit) for dimensions and sizes. Overflow can lead to incorrect buffer allocations.

**Example Attack (v5 format)**:
```
Dimensions: [0xFFFFFFFF, 0xFFFFFFFF] (uint32)
Total elements: overflow → small allocation
Actual data size: huge → buffer overflow
```

**Mitigation**:
- All size fields validated before use
- Safe integer conversions with overflow checks
- Maximum reasonable limits enforced

**Current Limits**:
- Max array dimensions: 2^31 per dimension
- Max variable name length: 63 bytes (v5), 1KB (v7.3)
- Max total data size per variable: Limited by available memory

### 3. v5 Format Specific Risks

**Risk**: v5 format uses tag-length-value (TLV) encoding. Malformed tags can cause parsing errors.

**Attack Vectors**:
- Invalid tag type (not in valid range 1-15)
- Incorrect small-format detection (upper 16 bits check)
- Misaligned padding (not 8-byte aligned)
- Nested matrix elements with circular references

**Mitigation**:
- Tag type validation against known types
- Small-format detection (size 1-4 bytes in upper 16 bits)
- Padding validation and alignment checks
- Recursion depth limits for nested structures

**v5 Parser Safety**:
```go
// All tag reads include validation:
// 1. Tag type must be valid (miINT8-miUTF8)
// 2. Size must be reasonable (< 2GB)
// 3. Small format detection correct
// 4. Padding aligned to 8-byte boundary
```

### 4. v7.3 Format (HDF5) Risks

**Risk**: v7.3 files are HDF5 format. All HDF5 security considerations apply.

**Inherited from HDF5**:
- Compression bomb attacks
- B-tree parsing vulnerabilities
- Resource exhaustion via nested groups
- Integer overflow in dataset dimensions

**Mitigation**:
- Relies on `github.com/scigolib/hdf5 v0.13.1` (stable)
- HDF5 library has comprehensive security mitigations
- MATLAB_class attributes validated for type safety

**See also**: HDF5 library SECURITY.md for detailed HDF5-specific risks

### 5. Compression Vulnerabilities

**Risk**: Both v5 compressed data and v7.3 HDF5 compression can be exploited.

**Attack Vectors**:
- v5: miCOMPRESSED elements with GZIP bomb
- v7.3: HDF5 compressed datasets with extreme ratios

**Current Status**:
- v5: Compressed data not yet supported (returns error)
- v7.3: Handled by HDF5 library (has compression limits)

**Planned for v0.3+**:
- v5 compression support with ratio limits
- Decompression size validation

### 6. Resource Exhaustion

**Risk**: MATLAB files can contain large arrays or deeply nested structures.

**Attack Vectors**:
- Multi-dimensional arrays with huge dimensions
- Cell arrays with millions of elements
- Deeply nested structures (recursion)
- Multiple variables with large data

**Mitigation**:
- Array size validation before allocation
- Recursion depth limits (structures, cell arrays)
- Streaming I/O for large arrays (where possible)
- Memory allocation checks

### 7. Path Traversal (v7.3 only)

**Risk**: v7.3 files use HDF5 groups with path-like names (e.g., `/variable/real`).

**Mitigation**:
- HDF5 library sanitizes paths internally
- MATLAB adapter validates MATLAB_class attributes
- No direct filesystem operations based on HDF5 paths

**User Best Practices**:
```go
// ❌ BAD - Don't use variable names directly for filesystem operations
varName := variable.Name // Could be "../../../etc/passwd"
os.Create(varName + ".txt")

// ✅ GOOD - Sanitize and validate names
safeName := filepath.Base(variable.Name)
if !isValidName(safeName) {
    return errors.New("invalid variable name")
}
```

## Security Best Practices for Users

### Input Validation

Always validate MATLAB files from untrusted sources:

```go
// Validate file size before opening
fileInfo, err := os.Stat(filename)
if err != nil || fileInfo.Size() > maxAllowedSize {
    return errors.New("invalid file")
}

// Open with error handling
file, err := os.Open(filename)
if err != nil {
    return err
}
defer file.Close()

matFile, err := matlab.Open(file)
if err != nil {
    // File failed validation - potentially malicious
    log.Printf("Failed to open MATLAB file: %v", err)
    return err
}
```

### Resource Limits

Set limits when processing untrusted files:

```go
// Check variable dimensions before reading
for _, variable := range matFile.Variables() {
    totalElements := 1
    for _, dim := range variable.Dimensions {
        totalElements *= dim
    }

    if totalElements > maxAllowedElements {
        return errors.New("array too large")
    }

    // Safe to read now
    data := variable.Data
}
```

### Error Handling

Always check errors - parsing failures may indicate malicious files:

```go
// ❌ BAD - Ignoring errors
matFile, _ := matlab.Open(file)
variables := matFile.Variables()

// ✅ GOOD - Proper error handling
matFile, err := matlab.Open(file)
if err != nil {
    return fmt.Errorf("file open failed: %w", err)
}

variables := matFile.Variables()
if len(variables) == 0 {
    return errors.New("no variables found")
}

for _, v := range variables {
    if v.Data == nil {
        log.Printf("Variable %s has no data", v.Name)
        continue
    }
    // Process variable...
}
```

### Writing Safe MATLAB Files

When writing files, validate input data:

```go
// ✅ Validate dimensions
if len(dims) == 0 || len(dims) > 10 {
    return errors.New("invalid dimensions")
}

for _, dim := range dims {
    if dim <= 0 || dim > maxDimSize {
        return errors.New("dimension out of range")
    }
}

// ✅ Validate variable name
if len(name) == 0 || len(name) > 63 {
    return errors.New("invalid variable name length")
}

// Safe to write
writer, err := matlab.Create(filename, matlab.Version5)
if err != nil {
    return err
}
defer writer.Close()

err = writer.WriteVariable(&matlab.Variable{
    Name:       name,
    Dimensions: dims,
    Data:       data,
})
```

## Known Security Considerations

### 1. Binary Parsing Vulnerabilities (v5)

**Status**: Active mitigation via bounds checking and validation.

**Risk Level**: Medium to High

**Description**: Parsing v5 binary format involves reading tags, sizes, and data from untrusted sources. Malformed files can trigger buffer overflows or integer overflows.

**Mitigation**:
- All reads bounds-checked
- Integer overflow checks before allocations
- Tag type and size validation
- 8-byte alignment verification

### 2. Compression Bomb (v5 miCOMPRESSED)

**Status**: Not yet implemented (returns error).

**Risk Level**: Medium (when implemented)

**Description**: v5 compressed elements could contain GZIP bombs.

**Planned Mitigation (v0.3+)**:
- Decompression ratio limits (1000:1)
- Streaming decompression with size checks
- Maximum decompressed size limits

### 3. HDF5 Inherited Vulnerabilities (v7.3)

**Status**: Mitigated by HDF5 library.

**Risk Level**: Low to Medium

**Description**: v7.3 files rely on HDF5 library. Any HDF5 vulnerability affects v7.3 parsing.

**Mitigation**:
- Use latest stable HDF5 library (v0.13.1)
- HDF5 library has comprehensive security testing
- Monitor HDF5 library security advisories

### 4. Dependency Security

MATLAB File Reader/Writer dependencies:

- `github.com/scigolib/hdf5 v0.13.1` (stable) - v7.3 format support
- `github.com/stretchr/testify` (dev only) - Testing

**Monitoring**:
- ✅ Using stable HDF5 version (v0.13.1)
- 🔄 Dependabot enabled (when fully public)
- ✅ No C dependencies (pure Go + HDF5 Go library)

## Security Testing

### Current Testing

- ✅ Unit tests with malformed data (v5 tags, dimensions)
- ✅ Integration tests with real MATLAB files
- ✅ Round-trip tests (write → read → verify)
- ✅ Linting with 34+ security-focused linters
- ✅ Race detector (0 data races)

### Planned for v1.0

- 🔄 Fuzzing with go-fuzz (v5 parser, tag parsing)
- 🔄 Static analysis with gosec
- 🔄 SAST/DAST scanning in CI
- 🔄 Comparison testing against MATLAB/Octave

## Security Disclosure History

### v0.3.0 (2025-11-21)

**Three security issues fixed** (identified through internal audit):

1. **Tag Size Validation** (High)
   - **CVE**: Pending
   - **Affected**: v0.1.0-beta to v0.2.0
   - **Fixed in**: v0.3.0
   - **Type**: Memory Exhaustion / DoS
   - **Credit**: Internal security audit

2. **Dimension Overflow Check** (High)
   - **CVE**: Pending
   - **Affected**: v0.1.0-beta to v0.2.0
   - **Fixed in**: v0.3.0
   - **Type**: Integer Overflow / Buffer Overflow
   - **Credit**: Internal security audit

3. **v73 Complex Reading** (Medium)
   - **CVE**: N/A (functionality bug)
   - **Affected**: v0.1.0-beta to v0.2.0
   - **Fixed in**: v0.3.0
   - **Type**: Data Corruption / Round-trip Failure
   - **Credit**: Round-trip testing

**Recommendation**: All users should upgrade to v0.3.0 immediately.

## Security Contact

- **GitHub Security Advisory**: https://github.com/scigolib/matlab/security/advisories/new
- **Public Issues** (for non-sensitive bugs): https://github.com/scigolib/matlab/issues
- **Discussions**: https://github.com/scigolib/matlab/discussions

## Bug Bounty Program

MATLAB File Reader/Writer does not currently have a bug bounty program. We rely on responsible disclosure from the security community.

If you report a valid security vulnerability:
- ✅ Public credit in security advisory (if desired)
- ✅ Acknowledgment in CHANGELOG
- ✅ Our gratitude and recognition in README
- ✅ Priority review and quick fix

---

**Thank you for helping keep MATLAB File Reader/Writer secure!** 🔒

*Security is a journey, not a destination. We continuously improve our security posture with each release.*
