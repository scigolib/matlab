# Examples

Runnable programs, each in its own directory with a `main.go` and a README. Run them from the repository root.

| Example | What it shows |
|---|---|
| [01-read](01-read/) | Open a v5 or v7.3 file and list its variables |
| [02-write-complex](02-write-complex/) | Write a complex array to a v7.3 file |
| [03-write-v5](03-write-v5/) | Write a vector and a matrix to a big-endian v5 file, read them back |
| [04-write-v73](04-write-v73/) | Write double and int32 variables to a v7.3 file, read them back |

Check that they build with `go vet ./examples/...`. For the testable examples that appear on
pkg.go.dev, see `example_test.go` in the repository root.
