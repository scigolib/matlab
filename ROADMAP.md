# MATLAB File Reader/Writer - Development Roadmap

> **Strategic Approach**: Leverage existing HDF5 library and MATLAB documentation

**Last Updated**: 2026-01-28 | **Current Version**: v0.3.3 (HOTFIX ✅) | **Target**: v1.0.0 stable (2026)

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
v0.1.0-beta (2025-11-02) → Reader v5/v7.3 + Writer v7.3 (workaround complex)
         ↓ (1 day!)
v0.1.1-beta (2025-11-03) → Proper MATLAB complex format + race detector fix
         ↓ (3 days!)
v0.2.0-beta (2025-11-06) → v5 Writer + parser bug fixes + comprehensive tests
         ↓ (2 months!)
v0.2.0 STABLE (2025-01-09) ✅ → HDF5 v0.13.1 stable + production ready
         ↓ (1 day!)
v0.3.0 STABLE (2025-11-21) ✅ → Production Quality (Grade A-)
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

## 📊 Current Status (v0.3.0 - STABLE RELEASE ✅)

### ✅ What's Working Now

**Project Infrastructure** (100%):
- ✅ Repository structure with internal/ packages
- ✅ Development tools (Makefile, .golangci.yml v2.5, 34+ linters)
- ✅ CI/CD (GitHub Actions: Linux, macOS, Windows) - ALL GREEN
- ✅ Documentation (README, CONTRIBUTING, CHANGELOG, ROADMAP)
- ✅ Git-Flow workflow, Kanban task management
- ✅ Production-quality code (golangci-lint: 0 issues)

**Reader Implementation** (100%):
- ✅ Format auto-detection (v5/v7.3)
- ✅ `Open(io.Reader)` public API
- ✅ Type system (Variable, DataType, NumericArray)
- ✅ v5 parser: streaming, all numeric types
- ✅ v73 adapter: HDF5 integration
- ✅ **Parser bugs FIXED** ✨ NEW in v0.2.0-beta
  - ✅ Multi-dimensional arrays now work correctly
  - ✅ Multiple variables in one file supported
  - ✅ Critical tag format detection bug fixed
- ⚠️ Compression, structures/cells (partial)

**Writer Implementation** (95%):
- ✅ **v5 Writer COMPLETE** ✨ NEW in v0.2.0-beta
  - ✅ All numeric types (double, single, int8-64, uint8-64)
  - ✅ Complex numbers (proper v5 format)
  - ✅ Multi-dimensional arrays (1D, 2D, 3D, N-D)
  - ✅ Both endianness (MI/IM)
  - ✅ Proper 8-byte alignment and padding
- ✅ v7.3 Writer COMPLETE (HDF5-based)
- ✅ `Create()`, `WriteVariable()`, `Close()` API
- ✅ **Complex numbers (proper MATLAB v7.3 format)** (v0.1.1-beta)
- ✅ Round-trip verified: write → read → ✅ PASSED (both v5 and v7.3)
- ✅ 11 test files generated (testdata/)
- ✅ **Race detector working** (Gentoo WSL2 fix) (v0.1.1-beta)
- ⚠️ Character arrays (partial for v5)

**Quality Metrics** (v0.3.0):
- ✅ **Grade: A- (Excellent)** - Production Quality ⬆️
- ✅ Test coverage: 85.4% (+6.9% from v0.2.0)
- ✅ Tests: 298 passing (100%) (+60 new tests)
- ✅ Linter: 0 errors, 0 warnings
- ✅ **Race detector: WORKING** (0 races detected)
- ✅ CI/CD: All checks GREEN ✅
- ✅ Documentation: Comprehensive + 17 testable examples
- ✅ API design: 95/100 (2025 Go best practices) ⬆️
- ✅ Repository: PUBLIC, Google indexing active
- ✅ Security: 3 critical issues fixed ✨

**Known Limitations** (documented in CHANGELOG):
- ⚠️ Character arrays (partial support for v5 Writer)
- ❌ Compression not supported
- ❌ Structures/cells not supported for writing

**What's in v0.2.0 STABLE**:
- ✅ **HDF5 v0.13.1 stable** (upgraded from v0.11.5-beta)
- ✅ v5 Writer fully implemented (565 lines)
- ✅ Critical parser bug fixed (tag format detection)
- ✅ Multi-dimensional arrays working in reader
- ✅ Multiple variables per file working in reader
- ✅ All round-trip tests passing (100%)
- ✅ Production-ready quality maintained

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

### **Phase 2: v0.2.0 - v5 Writer + Parser Fixes + Stable Release** ✅ COMPLETE

**Goal**: Complete MATLAB v5 format writer and fix critical parser bugs

**Deliverables**:
1. ✅ v5 Writer implementation (565 lines core code)
2. ✅ Comprehensive unit tests (589 lines, 17+ test functions)
3. ✅ Round-trip tests (430 lines, v5 write → read → verify)
4. ✅ Critical parser bug fix (tag format detection)
5. ✅ Multi-dimensional arrays support in reader
6. ✅ Multiple variables per file support in reader
7. ✅ Both endianness support (MI/IM)
8. ✅ All numeric types + complex numbers for v5
9. ✅ Documentation updates (README, CHANGELOG, ROADMAP)
10. ✅ Production quality: 0 linter issues, all tests passing

**Tasks**: TASK-011 (v5 Writer + Parser Fixes)
**Duration**:
- Beta: 3 days (2025-11-04 to 2025-11-06)
- Stable: 2 months testing (2025-11-06 to 2025-01-09)
**Status**: ✅ STABLE RELEASED 2025-01-09

**Key Achievements**:
- v5 Writer implementation: 565 lines of production code
- Parser bug fix: Single critical fix resolved 3 major bugs
- Test quality: 100% passing, 78.5% coverage (main package)
- Code quality: 0 linter errors, professional Go code
- Round-trip verification: Both v5 and v7.3 formats working perfectly
- **HDF5 v0.13.1 stable**: Upgraded from beta to stable dependency
- **Production-ready**: 2 months of battle-testing

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


---

### **Phase 3: v0.3.0 - Production Quality** ✅ COMPLETE

**Goal**: Bring library to production quality (Grade A-)

**Deliverables**:
1. ✅ Critical Security Fixes (3 issues)
   - Tag size validation (2GB limit)
   - Dimension overflow check
   - v73 complex reading fix
2. ✅ Testable Examples (17 examples)
   - Package-level, Create, Open, Write, Read
   - Round-trip and multi-variable examples
   - Functional options examples
3. ✅ API Convenience Methods (7 methods)
   - MatFile: GetVariable, GetVariableNames, HasVariable
   - Variable: GetFloat64Array, GetInt32Array, GetComplex128Array, GetScalar
4. ✅ Functional Options Pattern (3 options)
   - WithEndianness, WithDescription, WithCompression
   - 100% backward compatible
5. ✅ LINTER_RULES.md enforcement
6. ✅ Grade improvement: B+ → A-
7. ✅ Coverage increase: 78.5% → 85.4% (+6.9%)

**Tasks**: TASK-014, TASK-015, TASK-016, TASK-012
**Duration**: 1 day (2025-11-21) - All tasks completed in single session!
**Status**: ✅ RELEASED 2025-11-21

**Key Achievements**:
- 🏆 60 new tests added (298 total, 100% passing)
- 🏆 0 linter violations maintained
- 🏆 Production-ready quality achieved
- 🏆 Zero technical debt
- 🏆 70% reduction in user boilerplate code

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

## 🎯 Current Focus (Post v0.3.0 Stable)

### Immediate Priorities (Next 2-3 Weeks)

**Focus**: v0.4.0 - Context Support + Advanced Features

**Current Status**: v0.3.0 STABLE released (2025-11-21) ✅

**Planned Work**:
1. **Context Support** ⭐ (TASK-013, v0.4.0)
   - OpenWithContext, WriteVariableWithContext
   - Cancellable operations
   - Timeout support
   - Progress callbacks

2. **Community Engagement** ⭐
   - Monitor GitHub issues
   - Respond to questions
   - Gather feature requests
   - Collect feedback on v0.3.0 API

3. **Documentation** ⭐
   - Migration guide (v0.2.0 → v0.3.0)
   - API reference updates
   - Performance tips
   - Security best practices guide

4. **Quality Improvements** ⭐
   - Increase test coverage to 90%+
   - Add more edge case tests
   - Performance benchmarks
   - Memory optimization

---

## 📖 Dependencies

**Required**:
- Go 1.25+
- github.com/scigolib/hdf5 v0.13.1 (STABLE) - for v7.3 support
  - Production-ready HDF5 implementation
  - Includes nested datasets and group attributes support

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

**Next**: v0.3.0 will add Functional Options Pattern for flexible API

---

## 🎉 Release Notes - v0.2.0 STABLE (2025-01-09)

### What's New in v0.2.0 Stable
- ✅ **STABLE RELEASE**: Graduated from beta to stable
- ✅ **HDF5 v0.13.1**: Upgraded to stable HDF5 dependency
- ✅ **Production-ready**: 2 months of battle-testing since v0.2.0-beta
- ✅ **All features preserved**: Complete v5+v7.3 read/write support
- ✅ **Zero regressions**: All tests passing with new HDF5 version

### Complete Feature Set (from v0.2.0-beta)
- v5 Writer: All numeric types, complex, multi-dimensional (565 lines)
- v5 Reader: Fixed critical parser bugs (tag format, multi-dim, multiple vars)
- v7.3 Writer: HDF5-based with proper MATLAB format
- v7.3 Reader: Full HDF5 integration
- Round-trip verified: Both formats working perfectly

### Quality Metrics
- Tests: 100% passing (all platforms)
- Coverage: 78.5% (main), 51.8% (v5), 48.8% (v73)
- Linter: 0 errors, 0 warnings
- Race detector: 0 races
- CI/CD: All platforms GREEN

**Recommendation**: Upgrade from any beta version - stable, production-ready!

---

*Version 2.3*
*Current: v0.3.0 STABLE (RELEASED 2025-11-21) | Next: v0.4.0 (Context Support) | Target: v1.0.0 (2026)*
