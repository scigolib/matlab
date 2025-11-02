# Contributing to MATLAB File Reader/Writer

Thank you for considering contributing to the MATLAB File Reader/Writer! This document outlines the development workflow and guidelines.

## Git Workflow (Git-Flow)

This project uses Git-Flow branching model for development.

### Branch Structure

```
main                 # Production-ready code (tagged releases)
  └─ develop         # Integration branch for next release
       ├─ feature/*  # New features
       ├─ bugfix/*   # Bug fixes
       └─ hotfix/*   # Critical fixes from main
```

### Branch Purposes

- **main**: Production-ready code. Only releases are merged here.
- **develop**: Active development branch. All features merge here first.
- **feature/\***: New features. Branch from `develop`, merge back to `develop`.
- **bugfix/\***: Bug fixes. Branch from `develop`, merge back to `develop`.
- **hotfix/\***: Critical production fixes. Branch from `main`, merge to both `main` and `develop`.

### Workflow Commands

#### Starting a New Feature

```bash
# Create feature branch from develop
git checkout develop
git pull origin develop
git checkout -b feature/my-new-feature

# Work on your feature...
git add .
git commit -m "feat: add my new feature"

# When done, merge back to develop
git checkout develop
git merge --squash feature/my-new-feature  # Squash merge for clean history
git commit -m "feat: my new feature (squashed)"
git branch -d feature/my-new-feature
git push origin develop
```

#### Fixing a Bug

```bash
# Create bugfix branch from develop
git checkout develop
git pull origin develop
git checkout -b bugfix/fix-issue-123

# Fix the bug...
git add .
git commit -m "fix: resolve issue #123"

# Merge back to develop
git checkout develop
git merge --squash bugfix/fix-issue-123  # Squash merge for clean history
git commit -m "fix: resolve issue #123 (squashed)"
git branch -d bugfix/fix-issue-123
git push origin develop
```

#### Creating a Release

```bash
# Create release branch from develop
git checkout develop
git pull origin develop
git checkout -b release/v1.0.0

# Update version numbers, CHANGELOG, etc.
git add .
git commit -m "chore: prepare release v1.0.0"

# Merge to main and tag
git checkout main
git merge --no-ff release/v1.0.0
git tag -a v1.0.0 -m "Release v1.0.0"

# Merge back to develop
git checkout develop
git merge --no-ff release/v1.0.0

# Delete release branch
git branch -d release/v1.0.0

# Push everything
git push origin main develop --tags
```

#### Hotfix (Critical Production Bug)

```bash
# Create hotfix branch from main
git checkout main
git pull origin main
git checkout -b hotfix/critical-bug

# Fix the bug...
git add .
git commit -m "fix: critical production bug"

# Merge to main and tag
git checkout main
git merge --no-ff hotfix/critical-bug
git tag -a v1.0.1 -m "Hotfix v1.0.1"

# Merge to develop
git checkout develop
git merge --no-ff hotfix/critical-bug

# Delete hotfix branch
git branch -d hotfix/critical-bug

# Push everything
git push origin main develop --tags
```

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

4. **All-in-one**:
   ```bash
   make pre-commit
   ```

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
- golangci-lint
- Access to companion HDF5 library at `../hdf5`

### Install Dependencies

```bash
# Clone both repositories
cd D:\projects\scigolibs
git clone https://github.com/scigolib/hdf5.git
git clone https://github.com/scigolib/matlab.git

# Install golangci-lint
cd matlab
make install-lint
```

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
├── .claude/              # AI development configuration (private)
├── .gitignore           # Git ignore rules
├── Makefile             # Development commands
├── cmd/                 # Command-line utilities
│   └── example/        # Example program
├── examples/            # Usage examples
├── types/               # Common data structures (PUBLIC)
│   ├── array.go        # Array types
│   └── variable.go     # Variable representation
├── internal/            # Private implementation details
│   ├── v5/             # MATLAB v5-v7.2 reader/writer (PRIVATE)
│   │   ├── parser.go       # Reader parser
│   │   ├── header.go       # Header parsing
│   │   ├── data_tag.go     # Tag-length-value parsing
│   │   ├── matrix.go       # Matrix element parsing
│   │   ├── types.go        # Type constants and conversions
│   │   ├── compressed.go   # Compression support (planned)
│   │   └── writer.go       # v5 Writer (planned in TASK-011)
│   └── v73/            # MATLAB v7.3+ reader/writer (PRIVATE)
│       ├── parser.go       # Reader parser
│       ├── adapter.go      # HDF5 to MATLAB adapter
│       └── writer.go       # v7.3 Writer (HDF5-based)
├── matfile.go           # Public API - Reader (Open)
├── matfile_write.go     # Public API - Writer (Create)
├── LICENSE              # MIT License
└── README.md            # Main documentation
```

## Adding New Features

1. Check if issue exists, if not create one
2. Discuss approach in the issue
3. Create feature branch from `develop`
4. Implement feature with tests
5. Update documentation
6. Run quality checks (`make pre-commit`)
7. Create pull request to `develop`
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

- Check existing issues and discussions
- Read `.claude/CLAUDE.md` for architecture insights
- Review HDF5 library for reference implementation patterns
- Ask questions in GitHub Issues

## License

By contributing, you agree that your contributions will be licensed under the MIT License.

---

**Thank you for contributing to the MATLAB File Reader/Writer!** 🎉
