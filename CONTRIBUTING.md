# Contributing to MATLAB File Reader/Writer

Thank you for considering contributing to the MATLAB File Reader/Writer! This document outlines the development workflow and guidelines.

## Git Workflow (GitHub Flow)

`main` is the only long-lived branch and always holds releasable code. Every
change, from a typo to a feature, reaches it through a pull request.

1. Fork the repository (or create a branch, if you have write access) and
   branch from the latest `main`:
   ```bash
   git checkout main
   git pull origin main
   git checkout -b fix/issue-123
   ```
2. Make the change with tests. For a bug fix, add a test that fails on the
   current code first.
3. Run the local checks (`make ci`, see below).
4. Push the branch and open a pull request against `main`. Reference the issue
   it closes (`Fixes #123`).
5. CI must be green before a merge. The required checks on `main` are the unit
   tests on Linux, macOS and Windows, `Lint`, `Code Formatting` and
   `Cross-Build 32-bit`; the branch must be up to date with `main`.
6. A maintainer reviews and merges. Releases are tagged from `main` by the
   maintainers.

Branch names are free-form; `feat/…`, `fix/…`, `docs/…` are a good habit.

## Commit Message Guidelines

Follow [Conventional Commits](https://www.conventionalcommits.org/) specification:

```
<type>(<scope>): <description>

[optional body]

[optional footer]
```

### Types

- **feat**: New feature
- **fix**: Bug fix
- **docs**: Documentation changes
- **style**: Code style changes (formatting, etc.)
- **refactor**: Code refactoring
- **test**: Adding or updating tests
- **chore**: Maintenance tasks (build, dependencies, etc.)
- **perf**: Performance improvements

### Examples

```bash
feat: add support for sparse matrices in v5 format
fix: correct endianness handling in header parsing
docs: update README with compressed data limitations
refactor: simplify variable type conversion logic
test: add tests for complex number arrays
chore: update go.mod dependencies
```

## Code Quality Standards

### Before Committing

1. **Format code**:
   ```bash
   make fmt
   ```

2. **Run linter**:
   ```bash
   make lint
   ```

3. **Run tests**:
   ```bash
   make test
   ```

4. **All-in-one** (formatting, tests, lint and 32-bit builds; CI also runs
   `go vet` and the tests with the race detector):
   ```bash
   make ci
   ```
   `make pre-commit` runs the same checks but formats the code instead of
   only checking it.

### Pull Request Requirements

- [ ] Code is formatted (`make fmt`)
- [ ] Linter passes (`make lint`)
- [ ] All tests pass (`make test`)
- [ ] New code has tests (minimum 70% coverage)
- [ ] Documentation updated (if applicable)
- [ ] Commit messages follow conventions
- [ ] No sensitive data (credentials, tokens, etc.)

## Development Setup

### Prerequisites

- Go 1.25 or later
- golangci-lint v2 (`make install-lint`)
- A C compiler only for `make test-race` (the race detector needs cgo)

The HDF5 dependency (`github.com/scigolib/hdf5`) is an ordinary Go module and
is fetched by `go` like any other.

### Install Dependencies

```bash
git clone https://github.com/scigolib/matlab.git
cd matlab
make install-lint
```

To develop against a local checkout of the HDF5 library, add a `go.work` file
next to both clones (`go work init ./matlab ./hdf5`). Keep it out of commits:
CI builds with the version in `go.mod`.

### Running Tests

```bash
# Run all tests
make test

# Run with coverage
make test-coverage

# Run with race detector
make test-race

# Run benchmarks
make benchmark
```

### Running Linter

```bash
# Run linter
make lint

# Save linter report
make lint-report
```

## Project Structure

```
matlab/
├── matfile.go           # Public API - Reader (Open)
├── matfile_write.go     # Public API - Writer (Create)
├── options.go           # Public API - writer options (WithEndianness, ...)
├── example_test.go      # Testable examples shown on pkg.go.dev
├── types/               # Common data structures (PUBLIC)
│   ├── array.go        # Array types
│   └── variable.go     # Variable representation
├── internal/            # Private implementation details
│   ├── v5/             # MATLAB v5-v7.2 reader/writer
│   │   ├── parser.go       # Reader parser
│   │   ├── bounds.go       # Payload size checks and bounded reads
│   │   ├── header.go       # Header parsing
│   │   ├── data_tag.go     # Tag-length-value parsing
│   │   ├── matrix.go       # Matrix element parsing
│   │   ├── types.go        # Type constants and conversions
│   │   ├── compressed.go   # Compressed element (zlib) reading
│   │   └── writer.go       # v5 writer
│   └── v73/            # MATLAB v7.3+ reader/writer (HDF5-based)
│       ├── parser.go       # Reader parser
│       ├── adapter.go      # HDF5 to MATLAB adapter
│       └── writer.go       # v7.3 writer
├── cmd/example/         # Example program
├── examples/            # Usage examples
├── scripts/             # Test data generation and round-trip check
├── testdata/            # Test files (generated, legacy, SciPy reference)
└── Makefile             # Development commands
```

## Adding New Features

1. Check if issue exists, if not create one
2. Discuss approach in the issue
3. Create a branch from `main`
4. Implement feature with tests
5. Update documentation
6. Run quality checks (`make ci`)
7. Create a pull request to `main`
8. Wait for code review
9. Address feedback
10. Merge when approved

## Code Style Guidelines

### General Principles

- Follow Go conventions and idioms
- Write self-documenting code
- Add comments for complex logic (especially binary format parsing)
- Keep functions small and focused
- Use meaningful variable names

### Naming Conventions

- **Public types/functions**: `PascalCase` (e.g., `ParseMatrix`)
- **Private types/functions**: `camelCase` (e.g., `readTag`)
- **Constants**: `PascalCase` with context prefix (e.g., `miMATRIX`, `mxDOUBLE_CLASS`)
- **Test functions**: `Test*` (e.g., `TestParseHeader`)

### Error Handling

- Always check and handle errors
- Use descriptive error variables (`ErrUnsupportedVersion`, `ErrInvalidFormat`)
- Return errors immediately, don't wrap unnecessarily
- Validate inputs before processing

### Testing

- Use table-driven tests when appropriate
- Test both success and error cases
- Include test data files in `testdata/`
- Test both endianness variants (little/big)
- Mock external dependencies when needed

## Parser Implementation Patterns

### Streaming I/O
- Work with `io.Reader` for memory efficiency
- Use `io.ReadFull()` for exact byte reads
- Never load entire file into memory
- Track position for debugging

### Byte Order Handling
- Store `binary.ByteOrder` in parser state
- Detect endianness from header
- Use `Order.Uint32()`, `Order.Uint64()` consistently

### Padding Alignment
```go
// Always align to 8-byte boundaries in v5 format
padding := (8 - size%8) % 8
```

## Getting Help

- Check existing [issues](https://github.com/scigolib/matlab/issues) and
  [discussions](https://github.com/scigolib/matlab/discussions)
- The package documentation and examples are on
  [pkg.go.dev](https://pkg.go.dev/github.com/scigolib/matlab)
- The MAT-file format is described in MathWorks'
  [MAT-File Format](https://www.mathworks.com/help/pdf_doc/matlab/matfile_format.pdf)
- Ask questions in GitHub Issues or Discussions

## License

By contributing, you agree that your contributions will be licensed under the MIT License.

---

**Thank you for contributing to the MATLAB File Reader/Writer!** 🎉
