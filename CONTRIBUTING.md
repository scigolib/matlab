# Contributing to MATLAB File Reader/Writer

Contributions are welcome: bug reports with a sample file, fixes, documentation, and MAT-files from your own field that
the library should read. This document describes the workflow and the checks.

## Workflow (GitHub Flow)

`main` is the only long-lived branch and always holds releasable code. Every change, from a typo to a feature,
reaches it through a pull request.

1. Fork the repository (or branch, if you have write access) from the latest `main`:
   ```bash
   git checkout main
   git pull origin main
   git checkout -b fix/issue-123
   ```
2. Make the change with tests (see Tests below). For a bug fix, write the test first and see it fail.
3. Run the local checks (see Local checks).
4. Push the branch and open a pull request against `main`. Reference the issue it closes (`Fixes #123`).
5. CI must be green. The required check on `main` is `Required checks`: the job at the end of every run that fails
   when any gating job failed — the unit tests on Linux, macOS and Windows, the `go.mod` minimum, lint on the
   Linux, Windows and macOS targets, formatting and the 32-bit cross-build. `Vulncheck` is informational and does
   not block a merge.
6. A maintainer reviews and approves every pull request before it is merged; `main` requires one approving
   review. Releases are tagged from `main` by the maintainers.

Branch names are free-form; `feat/...`, `fix/...`, `docs/...` are a good habit.

## Commit messages

[Conventional Commits](https://www.conventionalcommits.org/):

```
<type>(<scope>): <description>

[optional body]
```

Types: `feat`, `fix`, `docs`, `style`, `refactor`, `test`, `chore`, `perf`. Say what changed and why.

```
feat: add support for sparse matrices in v5 format
fix: correct endianness handling in header parsing
docs: update README with compressed data limitations
```

## Local checks

From the repository root, with Go 1.25 or later (`go.mod` says `go 1.25`) and golangci-lint v2
(`make install-lint`):

```bash
gofmt -l .              # must print nothing
go vet ./...
go test ./...
golangci-lint run ./... # 0 issues
```

The Makefile wraps them:

| Target | Does |
|---|---|
| `make build` | `go build ./...` |
| `make test` | `go test -v -coverprofile=coverage.out ./...` |
| `make test-race` | the same with `-race`; the race detector needs cgo, so a C compiler |
| `make lint` | `golangci-lint run --timeout=5m` |
| `make fmt` | `gofmt -w -s .` and `go mod tidy` |
| `make cross-build` | builds `linux/386`, `linux/arm` (GOARM=6) and `windows/386` |
| `make ci` | `fmt-check`, `test`, `lint`, `cross-build` |
| `make pre-commit` | `fmt`, `lint`, `test`, `cross-build` |
| `make examples` | builds `cmd/example` and `examples` |

CI runs the tests with `-race` on Linux, macOS and Windows, on Go `oldstable`, `stable` and the `go.mod` minimum;
`go vet` and `go mod verify`; golangci-lint on Linux and on the Windows and macOS targets; `gofmt` and `go mod tidy`
checks; and the 32-bit builds, with `go vet` and the tests for `linux/386`. The linter version is pinned in
`.github/workflows/test.yml`, so a local `latest` may differ from CI. Run the whole matrix locally only when you are
chasing a failure that CI reports.

Heavy runs: the Go test cache and temp directory can be placed on another disk with `GOTMPDIR`. Generated files
(coverage, logs, binaries) do not belong in the repository root; `.gitignore` covers the usual ones.

## Tests

- **A bug fix starts with a test that fails before the fix**, and the pull request says how you saw it fail.
- **New behaviour comes with tests**: table-driven where it fits, success and error cases, both byte orders for v5.
- **Corruption and security tests use sizes that cannot exhaust memory.** Claim a few megabytes against a few dozen
  bytes of input and assert the error; do not allocate gigabytes to prove a check.
- **Reference files** live in `testdata/`. Files this library writes are checked with `h5dump` (v7.3) on the
  maintainer's machine, and with matio or MATLAB when one is available. The matlab CI has no `h5dump` job, so run it
  yourself when you change a writer: `h5dump file.mat` must open the file without errors.
- `go run ./scripts/generate-testdata` regenerates `testdata/generated`; `go run ./scripts/verify-roundtrip` checks a
  write-and-read round trip.
- A test is never weakened to go green. If intended behaviour changed, update the assertion on purpose and say so.

## References

The behaviour follows matio, SciPy, the HDF5 C library and MathWorks' format documents. They are listed with exact
versions in [UPSTREAM.md](UPSTREAM.md). Where a spec defines a value, the spec decides; where we differ from a
reference, UPSTREAM.md records it.

## Documentation

- [README.md](README.md): features, quick start, current status. Every number has a command that produces it.
- [CHANGELOG.md](CHANGELOG.md): Keep a Changelog format with Added, Changed, Fixed and Security sections. Each release
  starts with a short intro, and a behaviour change carries an inline `Upgrading:` note.
- [docs/guides](docs/guides/): QUICKSTART and DATATYPES. [docs/architecture](docs/architecture/OVERVIEW.md): packages.
- [docs/QUALITY.md](docs/QUALITY.md): the engineering standard this library is held to, and how each part of it is
  enforced.
- [AGENTS.md](AGENTS.md) and [llms.txt](llms.txt): guidance for AI agents that use the library.
- [examples](examples/README.md): each example has a `main.go` and a README and must build:
  `go vet ./examples/...`.
- Code samples in docs compile against the current API and check every error, including the one from `Close`.

## Project structure

```
matlab/
├── matfile.go, matfile_write.go, options.go   # public API: Open, Create, options
├── example_test.go          # testable examples shown on pkg.go.dev
├── types/                   # Variable, DataType, NumericArray (public)
├── internal/v5/             # v5 reader and writer
├── internal/v73/            # v7.3 reader and writer, on the hdf5 module
├── cmd/example/             # example program
├── examples/                # numbered runnable examples
├── scripts/                 # generate-testdata, verify-roundtrip
├── testdata/                # scipy, generated and legacy test files
├── docs/                    # guides and architecture overview
├── UPSTREAM.md              # references and validation
└── Makefile
```

To develop against a local checkout of the HDF5 module, add a `go.work` next to both clones
(`go work init ./matlab ./hdf5`) and keep it out of commits: CI builds with the version in `go.mod`.

## Code style

- Follow Go conventions; comment the non-obvious, especially binary format details (padding, alignment, byte order).
- Parsers read from `io.Reader`, use `io.ReadFull` for exact reads, store the `binary.ByteOrder` found in the header
  and never assume one. Sizes taken from a file are checked against the input before any allocation.
- v5 elements are aligned to 8 bytes: `padding := (8 - size%8) % 8`.
- Check every error and wrap it with context using `%w`; callers match with `errors.Is`
  (`ErrInvalidFormat`, `ErrUnsupportedVersion`, `io.ErrUnexpectedEOF`). `Close` errors on written files are checked.
- Names follow Go: `PascalCase` exported, `camelCase` unexported, MATLAB's own constant names (`miMATRIX`,
  `mxDOUBLE_CLASS`) in `internal/v5`.

## Security

Report vulnerabilities privately, as described in [SECURITY.md](SECURITY.md), not in a public issue.

## Getting help

- [Issues](https://github.com/scigolib/matlab/issues) and [discussions](https://github.com/scigolib/matlab/discussions)
- [pkg.go.dev](https://pkg.go.dev/github.com/scigolib/matlab) for the API
- MathWorks' [MAT-File Format](https://www.mathworks.com/help/pdf_doc/matlab/matfile_format.pdf)

## License

By contributing, you agree that your contributions will be licensed under the MIT License.
