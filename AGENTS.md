# AGENTS.md

This file provides instructions and conventions for AI coding agents working in this repository.

## Project Overview

`github.com/min0625/errgroup` is a Go library that extends [`golang.org/x/sync/errgroup`](https://pkg.go.dev/golang.org/x/sync/errgroup) with panic recovery. Panics occurring in goroutines started by `Go` or `TryGo` are caught and re-panicked inside `Wait`, wrapped as `PanicError` or `PanicValue`.

### Packages

| Path | Module path | Description |
|------|-------------|-------------|
| `/` (root) | `github.com/min0625/errgroup` | Drop-in replacement for `golang.org/x/sync/errgroup` |
| `x/errgroup/` | `github.com/min0625/errgroup/x/errgroup` | Context-aware variant — passes `context.Context` into each goroutine function |

## Repository Structure

```
errgroup.go          # Core Group type and methods
panic.go             # PanicError, PanicValue types and exception() helper
errgroup_test.go     # Tests for the root package
example_test.go      # Runnable examples for the root package
go.mod               # Module definition
Makefile             # Developer commands
mise.toml            # Pinned tool versions (Go, golangci-lint)
x/errgroup/
    errgroup.go      # Context-aware Group type
    panic.go         # Re-exports panic types from the root package (if any)
    errgroup_test.go # Tests for the x/errgroup package
    example_test.go  # Runnable examples for the x/errgroup package
```

## Tool Versions

Tool versions are pinned in `mise.toml`. Install them with:

```sh
mise install
```

| Tool | Version |
|------|---------|
| Go | 1.24.x |
| golangci-lint | 2.x |

## Common Commands

All commands should be run from the repository root.

| Command | Description |
|---------|-------------|
| `make lint` | Run linter (`golangci-lint run`) |
| `make fix` | Run linter with auto-fix (`golangci-lint run --fix`) |
| `make test` | Run all tests with race detector (`go test -v -race -failfast ./...`) |
| `make check` | Run `lint` then `test` (full CI gate) |

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

Changes to the root package (`errgroup.go`, `panic.go`) that affect the public API or panic-recovery behaviour must be reflected in `x/errgroup/` where applicable, and vice versa.

### Dependencies

- Keep dependencies minimal. The only non-test runtime dependency is `golang.org/x/sync`.
- Run `go mod tidy` after adding or removing imports.

### Linting

- The project uses `golangci-lint`. Lint rules are configured in the repository (check for `.golangci.yml` or inline `//nolint` directives).
- Suppress a lint warning with `//nolint:<linter> // <reason>` only when genuinely necessary, not to silence legitimate issues.
