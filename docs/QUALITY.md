# Engineering quality

What this library promises and how each promise is checked. This is the standard a change is judged by;
[CONTRIBUTING.md](../CONTRIBUTING.md) is the workflow and the commands.

## Correctness

The library reads and writes MATLAB's v5 and v7.3 formats. A value is correct when it matches what MATLAB writes and
what the reference implementations read — not merely when our own reader accepts what our own writer produced.

- **Checked against references.** Format behaviour is compared with matio, SciPy, the HDF5 C library and MathWorks'
  format documents. The exact versions we follow, and every known difference from them, are in
  [UPSTREAM.md](../UPSTREAM.md).
- **Where a spec defines a value, the spec decides.** A difference from a reference implementation is recorded in
  UPSTREAM.md, never silently copied.
- **Reference files are kept.** `testdata/scipy/` holds files written by MATLAB and SciPy and is read by
  `scipy_test.go`. Header detection is tested with MATLAB-style 512-byte user blocks in both byte orders.
- **What is not verified is said so.** MATLAB itself is not available to the maintainers, so loading our files in
  MATLAB is unverified; UPSTREAM.md says which parts are covered and which are not.

## Code

- **Errors are never swallowed.** Every error is returned with context (`%w`) or handled where the caller can see
  it. `_ = f()` on a returned error and `if err != nil { return nil }` are review failures.
- **Every size read from a file is validated against the remaining input before allocation.** A corrupt file
  produces an error, never a multi-gigabyte allocation or a panic.
- **Parsers stream.** They read from `io.Reader`, never load the file into memory, and take the byte order from the
  header instead of assuming one.
- **The library builds everywhere it claims:** Linux, macOS and Windows, and the 32-bit targets `linux/386`,
  `linux/arm` (GOARM=6) and `windows/386`.

## Tests

- **A bug fix starts with a test that fails before the fix.** The pull request says how the failure was seen.
- **New behaviour comes with tests**: table-driven where it fits, success and error paths, both byte orders for v5.
- **Corruption and security tests use sizes that cannot exhaust memory.** They claim a few megabytes against a few
  dozen bytes of input and assert the error.
- **Coverage floor: 70% overall.** Parsing and conversion code is the business logic of this library and is
  expected to carry tests for its success and error paths. Current figures come from `go test -cover` (README,
  "Current Status") and Codecov.
- **A test is never weakened, skipped or deleted to go green.** If intended behaviour changes, the assertion is
  updated on purpose and the pull request says which test and why.

## Architecture

- **Two parsers behind one public API.** `Open` detects the format from the header and routes to `internal/v5` or
  `internal/v73`; `Create` selects the writer for the requested version. Users never need to know the version of a
  file.
- **The public API follows semantic versioning.** A breaking change is a major release, a new feature a minor one, a
  fix a patch.
- **The v7.3 layer builds on `github.com/scigolib/hdf5`.** Format fixes there follow that module's releases, and the
  version in `go.mod` is part of what is tested.

## Review and release

- **Every pull request is reviewed by a maintainer before merge.** `main` accepts changes only through a pull
  request, and CI must be green: the `Required checks` job at the end of a run fails when any gating job failed.
- **Releases follow the maintainers' release procedure:** pre-flight checks, a pull request, green CI on `main`, a
  tag, and a GitHub release with notes from [CHANGELOG.md](../CHANGELOG.md). A tag without a GitHub release is an
  unfinished release.

## Enforcement

| Promise | Enforced by |
|---|---|
| Tests pass, race-clean, on Linux, macOS and Windows | CI: `Unit Tests - <os> - Go <version>` |
| The `go.mod` minimum builds and passes its tests | CI: `Test - ubuntu-latest - Go minimum` |
| 32-bit targets build, vet and pass their tests | CI: `Cross-Build 32-bit` |
| Lint is clean on Linux, Windows and macOS targets | CI: `Lint`, `Lint - <goos>/<goarch>` |
| Code is `gofmt`-ed and `go.mod`/`go.sum` are tidy | CI: `Code Formatting` |
| A dependency vulnerability is visible | CI: `Vulncheck` (informational; never blocks an unrelated pull request) |
| A red job cannot be merged as green | CI: `Required checks` (the gate) |
| Changes reach `main` only through a reviewed pull request | The `main-protection` ruleset and maintainer review |
