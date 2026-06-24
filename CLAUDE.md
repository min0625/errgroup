# CLAUDE.md

This file provides instructions and conventions for AI coding agents working in this repository.

## Project Overview

`github.com/min0625/errgroup` is a Go library that extends [`golang.org/x/sync/errgroup`](https://pkg.go.dev/golang.org/x/sync/errgroup) with panic recovery. Panics occurring in goroutines started by `Go` or `TryGo` are caught and re-panicked inside `Wait`, wrapped as `PanicError` or `PanicValue`.

### Packages

A single Go module (`github.com/min0625/errgroup`) holds two packages:

| Import path | Description |
|-------------|-------------|
| `github.com/min0625/errgroup` | Drop-in replacement for `golang.org/x/sync/errgroup` |
| `github.com/min0625/errgroup/x/errgroup` | Context-aware variant — passes `context.Context` into each goroutine function |

## Repository Structure

```
errgroup.go          # Core Group type and methods (WithContext, Go, TryGo, Wait, SetLimit)
panic.go             # PanicError, PanicValue types and the exception() helper
errgroup_test.go     # Tests for the root package
example_test.go      # Runnable examples for the root package
README.md            # Root package docs
go.mod / go.sum      # Module definition
Makefile             # Developer commands
mise.toml            # Pinned tool versions (Go, golangci-lint)
.golangci.yaml       # Linter configuration
x/errgroup/
    errgroup.go      # Context-aware Group type (New constructor instead of WithContext)
    panic.go         # Type-alias re-exports of PanicError / PanicValue from the root package
    errgroup_test.go # Tests for the x/errgroup package
    example_test.go  # Runnable examples for the x/errgroup package
    README.md        # x/errgroup package docs
```

## Tool Versions

Tool versions are pinned in `mise.toml`. Install them with:

```sh
mise install
```

| Tool | Version |
|------|---------|
| Go | 1.26.x (module requires go 1.25) |
| golangci-lint | 2.x |

## Common Commands

All commands should be run from the repository root.

| Command | Description |
|---------|-------------|
| `make lint` | Verify lint config, then run `golangci-lint run` (new issues vs `HEAD`) |
| `make fix` | `go mod tidy`, then `golangci-lint run --fix` (new issues vs `HEAD`) |
| `make test` | Run all tests with race detector (`go test -race -failfast ./...`) |
| `make check-tidy` | Verify `go.mod` is tidy (`go mod tidy -diff`) |
| `make cover` | Run tests with coverage and print the per-function summary |
| `make check` | Run `check-tidy`, `lint`, then `test` (full CI gate) |

`lint`/`fix` compare against `NEW_FROM_REV` (default `HEAD`); override it to widen the range, e.g. `make lint NEW_FROM_REV=main`.

Always run `make check` before considering a change complete.

## Development Guidelines

### Code Style

- Follow standard Go conventions (`gofmt`, `goimports`).
- Exported symbols must have GoDoc comments. Keep them concise and consistent with the existing style.
- Internal helpers (unexported) should be commented only when non-obvious.

### Testing

- All new behaviour must have corresponding tests.
- Tests must be parallel where possible (`t.Parallel()`).
- Use [`github.com/stretchr/testify`](https://pkg.go.dev/github.com/stretchr/testify) (`assert`, `require`) — consistent with the existing test suite.
- Run tests with the race detector: `make test` (uses `-race` flag).
- Place tests in `_test` package (external test package), e.g. `package errgroup_test`.

### Panic Handling

- `PanicError` wraps recovered values that implement `error`.
- `PanicValue` wraps all other recovered values.
- Both expose a `Stack []byte` field with the captured stack trace.
- When multiple goroutines panic concurrently, only the first panic is propagated; the rest are silently discarded. Preserve this invariant.

### Dual-Package Consistency

Changes to the root package (`errgroup.go`, `panic.go`) that affect the public API or panic-recovery behaviour must be reflected in `x/errgroup/` where applicable, and vice versa. The `x/errgroup` panic types are type aliases of the root types, so they stay in sync automatically.

### Dependencies

- Keep dependencies minimal. The only non-test runtime dependency is `golang.org/x/sync`.
- Run `go mod tidy` after adding or removing imports.

### Linting

- The project uses `golangci-lint`, configured in `.golangci.yaml`.
- Suppress a lint warning with `//nolint:<linter> // <reason>` only when genuinely necessary, not to silence legitimate issues.
</content>
</invoke>
