# Contributing to fakeproto

Contributions are welcome - whether it's a bug report, a feature request, or a pull request. Opening an issue or a PR on GitHub is all it takes to get started.

## Getting started

```bash
git clone https://github.com/psyhatter/fakeproto
cd fakeproto
go mod download
```

Regenerate protobuf stubs after editing `.proto` files:

```bash
bash internal/test/generate.sh
```

## Running tests

```bash
# All tests with race detector (required before submitting a PR)
go test -race ./...

# Short fuzz pass
go test -fuzz=. -fuzztime=30s ./...
```

## Code style

- All comments must be written in **English**.
- Every test function and subtest must call `t.Parallel()` as its first line.
- Keep the public API minimal - discuss new options or behaviors in an issue first.

## Submitting a pull request

1. Fork the repository and create a feature branch.
2. Make your changes and add or update tests.
3. Run `go test -race ./...` - all tests must pass.
4. Open a PR against `main` with a clear description of what and why.

There is no formal CLA. By submitting a PR you agree to license your contribution under the [MIT License](LICENSE).
