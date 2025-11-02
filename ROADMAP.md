# MATLAB File Reader/Writer - Development Roadmap

> **Strategic Approach**: Leverage existing HDF5 library and MATLAB documentation

**Last Updated**: 2025-11-03 | **Current Version**: v0.1.1-beta (RELEASED ✅) | **Target**: v1.0.0 stable (2026)

---

## 🎯 Vision

Build a **production-ready, pure Go MATLAB file library** with comprehensive **read AND write** support for both v5 and v7.3 formats, leveraging our HDF5 library for v7.3+ files.

### Key Advantages

✅ **HDF5 Library with Write Support**
- Pure Go HDF5 implementation at `github.com/scigolib/hdf5` develop (commit 36994ac)
- **HDF5 write support already implemented** (Create, WriteDataset, WriteAttribute, Group attributes)
- **Nested datasets** and **Group attributes** support (v0.1.1-beta)
- v7.3+ read/write via thin adapter layer
- Focus development on v5 format parser and writer

✅ **Clear Specification**
- MATLAB file format is well-documented by MathWorks
- Reference implementations available (MATLAB, Octave, scipy)
- Community knowledge base

✅ **Dual Mode Support**
- **Reader**: Complete read support for v5 and v7.3 formats
- **Writer**: Create MATLAB files from Go (v7.3 DONE ✅, v5 PLANNED)
- Clear format boundaries (v5, v7.3+)
- Practical feature set for scientific computing

---

## 🚀 Version Strategy

### Philosophy: MVP → Feature Complete → Community Feedback → Stable

```
v0.1.0-beta (RELEASED ✅) → Reader v5/v7.3 + Writer v7.3 (workaround complex)
         ↓ (1 day!)
v0.1.1-beta (RELEASED ✅) → Proper MATLAB complex format + race detector fix
         ↓ (3-4 weeks)
v0.2.0 → v5 Writer + bug fixes + improvements
         ↓ (2-3 weeks)
v0.3.0 → Functional Options Pattern (flexible API)
         ↓ (2-3 weeks)
v0.4.0 → Context Support (cancellable operations)
         ↓ (1-2 months)
v0.5.0+ → Advanced features (compression, structures, cells)
         ↓ (community testing, API stabilization)
v1.0.0-rc.1 → Feature freeze, API locked
         ↓ (community feedback, 2+ months testing)
v1.0.0 STABLE → Production release (only after community approval)
         ↓ (maintenance mode)
v2.0.0 → Only if breaking changes needed
```

**Important Notes**:
- **v1.0.0** requires community feedback and API freeze
- **v2.0.0** only for breaking changes
- Pre-1.0 versions may have API changes
- Beta versions document known limitations

---

## 📊 Current Status (v0.1.1-beta - RELEASED)

### ✅ What's Working Now

**Project Infrastructure** (100%):
- ✅ Repository structure with internal/ packages
- ✅ Development tools (Makefile, .golangci.yml v2.5, 34+ linters)
- ✅ CI/CD (GitHub Actions: Linux, macOS, Windows) - ALL GREEN
- ✅ Documentation (README, CONTRIBUTING, CHANGELOG, ROADMAP)
- ✅ Git-Flow workflow, Kanban task management
- ✅ Production-quality code (golangci-lint: 0 issues)

**Reader Implementation** (85%):
- ✅ Format auto-detection (v5/v7.3)
- ✅ `Open(io.Reader)` public API
- ✅ Type system (Variable, DataType, NumericArray)
- ✅ v5 parser: streaming, all numeric types
- ✅ v73 adapter: HDF5 integration
- ⚠️ Known bugs: multi-dim arrays read as 1D, multiple vars
- ❌ Compression, structures/cells (partial)

**Writer Implementation** (55%):
- ✅ v7.3 Writer COMPLETE (HDF5-based)
- ✅ `Create()`, `WriteVariable()`, `Close()` API
- ✅ All numeric types (double, single, int8-64, uint8-64)
- ✅ **Complex numbers (proper MATLAB v7.3 format)** ✨ FIXED in v0.1.1-beta
- ✅ Multi-dimensional arrays
- ✅ Round-trip verified: write → read → ✅ PASSED
- ✅ 11 test files generated (testdata/)
- ✅ **Race detector working** (Gentoo WSL2 fix) ✨ NEW in v0.1.1-beta
- ❌ v5 Writer (TASK-011) - next milestone

**Quality Metrics**:
- ✅ Test coverage: 48.8% (30 tests, 27 passing, 90%)
- ✅ Linter: 0 errors, 0 warnings
- ✅ **Race detector: WORKING** (0 races detected) ✨ NEW
- ✅ CI/CD: All checks GREEN ✅
- ✅ Documentation: Comprehensive
- ✅ API design: 90/100 (2025 Go best practices)
- ✅ Repository: PUBLIC, Google indexing started

**Known Limitations** (documented in CHANGELOG):
- ⚠️ Reader bugs: multi-dimensional arrays, multiple variables
- ❌ v5 Writer not yet implemented
- ❌ Compression not supported
- ❌ Structures/cells not supported for writing

**Fixed in v0.1.1-beta**:
- ✅ Complex numbers now use proper MATLAB v7.3 format (group with nested datasets)
- ✅ Race detector now works in Gentoo WSL2 (external linkmode fix)

---

## 📅 Development Phases

### **Phase 1: v0.1.0-beta - MVP** ✅ COMPLETE

**Goal**: First public release with v7.3 write support

**Deliverables**:
1. ✅ Project infrastructure (CI/CD, linting, documentation)
2. ✅ v5 reader (numeric types, partial structures/cells)
3. ✅ v7.3 reader (HDF5 adapter)
4. ✅ **v7.3 writer** (HDF5 adapter)
5. ✅ Public API: `Create()`, `WriteVariable()`, `Close()`
6. ✅ All numeric types + complex numbers
7. ✅ Round-trip verification tests
8. ✅ Test data generation (11 files)
9. ✅ Production-quality code (0 linter issues)

**Tasks**: TASK-001 to TASK-010
**Duration**: Completed
**Status**: ✅ RELEASED 2025-11-02

---

### **Phase 1.1: v0.1.1-beta - Complex Format Fix** ✅ COMPLETE

**Goal**: Fix complex number format and race detector

**Deliverables**:
1. ✅ Proper MATLAB v7.3 complex format (group with nested datasets)
2. ✅ HDF5 library updated to develop (nested datasets + group attributes)
3. ✅ Race detector fix for Gentoo WSL2 (external linkmode)
4. ✅ 3 new comprehensive complex number tests
5. ✅ Full MATLAB/Octave compatibility for complex numbers
6. ✅ Documentation updates

**Changes**:
- Updated HDF5 to develop branch (commit 36994ac)
- Adapted to new `CreateGroup()` API (returns `*GroupWriter`)
- Fixed "hole in findfunctab" error with `-ldflags '-linkmode=external'`
- Removed obsolete TODO comments

**Duration**: 1 day (2025-11-03)
**Status**: ✅ RELEASED 2025-11-03

---

### **Phase 2: v0.2.0 - v5 Writer** ← NEXT

**Goal**: Complete write support for both v5 and v7.3 formats

**Planned Features**:
1. ⭐ v5 binary writer implementation
2. ⭐ Tag-Length-Value encoding
3. ⭐ All numeric types
4. ⭐ Both endianness (MI/IM)
5. ⭐ Complex numbers
6. ⭐ Proper padding and alignment
7. ⭐ Round-trip tests (v5 write → read)
8. ⭐ MATLAB/Octave compatibility validation
9. ⭐ Fix reader bugs (multi-dim arrays, multiple vars)

**Tasks**: TASK-011 (v5 Writer)
**Duration**: 3-4 weeks
**Dependencies**:
- None (complex format already fixed in v0.1.1-beta)

---

### **Phase 3: v0.3.0 - Functional Options Pattern**

**Goal**: Flexible and extensible API

**Planned Features**:
1. ⭐ Functional options for `Create()` and `Open()`
2. ⭐ `WithCompression()` option
3. ⭐ `WithEndianness()` option (v5)
4. ⭐ `WithFormat()` option (force v5 or v7.3)
5. ⭐ Backward compatibility maintained
6. ⭐ Examples and documentation

**Tasks**: TASK-012
**Duration**: 2-3 weeks
**Rationale**: Modern Go API design (2025 best practices)

---

### **Phase 4: v0.4.0 - Context Support**

**Goal**: Cancellable operations for long-running tasks

**Planned Features**:
1. ⭐ `OpenWithContext(ctx, reader)` API
2. ⭐ `WriteVariableWithContext(ctx, variable)` API
3. ⭐ Proper context cancellation handling
4. ⭐ Timeout support
5. ⭐ Progress callbacks (optional)

**Tasks**: TASK-013
**Duration**: 2-3 weeks
**Rationale**: Enterprise-grade API

---

### **Phase 5: v0.5.0+ - Advanced Features**

**Goal**: Feature completeness

**Planned Features**:
1. ⭐ Compression support (v5 GZIP, v7.3 filters)
2. ⭐ Structures (read + write)
3. ⭐ Cell arrays (read + write)
4. ⭐ Character arrays / strings (complete)
5. ⭐ Sparse matrices (full support)
6. ⭐ Performance optimization
7. ⭐ Test coverage >70%

**Duration**: 1-2 months

---

### **Phase 6: v1.0.0-rc.1 - Feature Freeze**

**Goal**: API stability and polish

**Requirements**:
- ✅ All v5 features complete
- ✅ All v7.3 features complete
- ✅ Comprehensive tests (>80% coverage)
- ✅ Performance benchmarks
- ✅ Documentation complete
- ✅ Examples for all features

**After v1.0.0-rc.1**:
- API FROZEN
- Only bug fixes
- Community testing phase (2+ months)

**Duration**: 1 month

---

### **Phase 7: v1.0.0 - Production Stable**

**Goal**: Production-ready library

**Requirements**:
- Stable for 2+ months
- No critical bugs
- Community feedback positive
- Test coverage >80%
- Documentation complete

**Guarantees**:
- ✅ API stability (no breaking changes in v1.x.x)
- ✅ Long-term support
- ✅ Semantic versioning

---

## 📚 Feature Support Roadmap

### v5 Format (MATLAB v5-v7.2)

| Feature | v0.1.0-beta | v0.2.0 | v0.3.0 | v1.0.0 |
|---------|-------------|--------|--------|--------|
| **Read** numeric arrays | ✅ | ✅ | ✅ | ✅ |
| **Read** complex numbers | ✅ | ✅ | ✅ | ✅ |
| **Read** character arrays | ⚠️ Partial | ✅ | ✅ | ✅ |
| **Read** structures | ⚠️ Partial | ⚠️ | ✅ | ✅ |
| **Read** cell arrays | ⚠️ Partial | ⚠️ | ✅ | ✅ |
| **Read** sparse matrices | ❌ | ⚠️ Header | ✅ | ✅ |
| **Read** compression | ❌ | ❌ | ✅ | ✅ |
| **Write** numeric arrays | ❌ | ✅ | ✅ | ✅ |
| **Write** complex numbers | ❌ | ✅ | ✅ | ✅ |
| **Write** character arrays | ❌ | ✅ | ✅ | ✅ |
| **Write** structures | ❌ | ❌ | ✅ | ✅ |
| **Write** cell arrays | ❌ | ❌ | ✅ | ✅ |
| **Write** compression | ❌ | ❌ | ✅ | ✅ |

### v7.3 Format (MATLAB v7.3+)

| Feature | v0.1.0-beta | v0.2.0 | v0.3.0 | v1.0.0 |
|---------|-------------|--------|--------|--------|
| **Read** HDF5 detection | ✅ | ✅ | ✅ | ✅ |
| **Read** numeric datasets | ✅ | ✅ | ✅ | ✅ |
| **Read** strings | ⚠️ Limited | ✅ | ✅ | ✅ |
| **Read** structures | ❌ | ⚠️ Basic | ✅ | ✅ |
| **Read** cell arrays | ❌ | ⚠️ Basic | ✅ | ✅ |
| **Read** attributes | ✅ | ✅ | ✅ | ✅ |
| **Write** numeric datasets | ✅ | ✅ | ✅ | ✅ |
| **Write** complex numbers | ⚠️ Workaround | ✅ | ✅ | ✅ |
| **Write** strings | ❌ | ✅ | ✅ | ✅ |
| **Write** structures | ❌ | ⚠️ Basic | ✅ | ✅ |
| **Write** cell arrays | ❌ | ⚠️ Basic | ✅ | ✅ |
| **Write** attributes | ✅ | ✅ | ✅ | ✅ |
| **Write** compression | ❌ | ❌ | ✅ | ✅ |

**Legend**:
- ✅ Full support
- ⚠️ Partial support / Known limitations
- ❌ Not implemented

---

## 🎯 Current Focus (Post v0.1.0-beta)

### Immediate Priorities (Next 1-2 Weeks)

**Decision Point**: Wait for HDF5 v0.11.5-beta or start v5 Writer?
- **Option A**: Wait 1-2 weeks for HDF5 proper complex format → v0.1.1-beta
- **Option B**: Start v5 Writer now → v0.2.0 (3-4 weeks)

**Meanwhile**:
1. **Community Engagement** ⭐
   - Monitor GitHub issues
   - Respond to questions
   - Gather feature requests
   - Collect feedback on API

2. **Bug Fixes** ⭐
   - Fix reader: multi-dimensional arrays read as 1D
   - Fix reader: can't read files with multiple datasets
   - Improve error messages
   - Add more examples

3. **Documentation** ⭐
   - Add more examples to README
   - Create tutorial / getting started guide
   - API reference documentation
   - Performance tips

4. **HDF5 Collaboration** ⭐
   - Respond to HDF5 team questions
   - Provide test files for their testing
   - Test their v0.11.5-beta when ready

---

## 📖 Dependencies

**Required**:
- Go 1.25+
- github.com/scigolib/hdf5 v0.11.4-beta (for v7.3 support)
  - Future: v0.11.5-beta will add proper complex format support

**Development**:
- golangci-lint v2.5+ (code quality)
- GitHub Actions (CI/CD)

**Testing**:
- MATLAB or Octave (for generating reference files)
- h5py (Python, for HDF5 verification)

---

## 🔬 Development Approach

**Using HDF5 Library**:
- v7.3+ support is mostly done via adapter
- Focus on v5 format writer
- Leverage proven HDF5 implementation

**Testing Strategy**:
- Unit tests for all components
- Integration tests (round-trip)
- Reference MAT-files for validation
- Performance benchmarks
- Target: >70% coverage by v1.0.0

**Quality Assurance**:
- golangci-lint with 34+ linters
- Comprehensive CI/CD (Linux, macOS, Windows)
- Pre-release check script
- Code review by senior architect agent

---

## 📞 Support

**Documentation**:
- README.md - Project overview and quick start
- CLAUDE.md - Architecture details (internal)
- CONTRIBUTING.md - Development guide
- CHANGELOG.md - Release history
- ROADMAP.md - This file

**Community**:
- GitHub Issues - Bug reports and feature requests
- GitHub Discussions - Questions and help
- Repository: https://github.com/scigolib/matlab

---

## ⛔ Out of Scope

The following features are **not planned**:

- ❌ MATLAB v4 format (obsolete, pre-1999)
- ❌ Function handles (can't be serialized to Go)
- ❌ MATLAB objects/classes (language-specific, limited value)
- ❌ External links (security concerns)
- ❌ Java objects (MATLAB-specific, no Go equivalent)

---

## 🎉 Release Notes

### v0.1.1-beta (2025-11-03) - Complex Format Fix

**What's Fixed**:
- ✅ **Proper MATLAB v7.3 complex format** (group with nested datasets)
  - Before: Flat workaround (`varname_real`, `varname_imag`)
  - After: Standard MATLAB structure (`/varname` group with `/real`, `/imag`)
- ✅ **Race detector now works** in Gentoo WSL2 (external linkmode fix)
- ✅ **Full MATLAB/Octave compatibility** for complex numbers
- ✅ HDF5 updated to develop (nested datasets + group attributes)
- ✅ 3 new comprehensive tests for complex numbers

**Quality**:
- Tests: 30 total, 27 passing (90%)
- Race detector: 0 races detected ✅
- Linter: 0 issues ✅

**Impact**: Files with complex numbers now fully compatible with MATLAB/Octave!

---

### v0.1.0-beta (2025-11-02) - First Public Release

**What's New**:
- ✅ v7.3 Writer complete (HDF5-based)
- ✅ Public API: `Create()`, `WriteVariable()`, `Close()`
- ✅ All numeric types supported
- ✅ Complex numbers (with workaround)
- ✅ Multi-dimensional arrays
- ✅ Round-trip verified
- ✅ 11 test files generated
- ✅ Production-quality code (0 linter issues)
- ✅ CI/CD all green

**Known Limitations**:
- ⚠️ Complex numbers use flat structure (HDF5 library limitation)
- ⚠️ Reader bugs: multi-dim arrays, multiple variables
- ❌ v5 Writer not yet implemented
- ❌ Compression not supported
- ❌ Structures/cells not supported for writing

**Next**: v0.2.0 will add v5 Writer and fix reader bugs

---

*Version 2.1*
*Current: v0.1.1-beta (RELEASED) | Next: v0.2.0 (v5 Writer) | Target: v1.0.0 (2026)*
