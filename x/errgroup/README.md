# errgroup/x/errgroup
[![Go Reference](https://pkg.go.dev/badge/github.com/min0625/errgroup/x/errgroup.svg)](https://pkg.go.dev/github.com/min0625/errgroup/x/errgroup)
[![codecov](https://codecov.io/gh/min0625/errgroup/branch/main/graph/badge.svg)](https://codecov.io/gh/min0625/errgroup)

A context-aware variant of [`github.com/min0625/errgroup`](../../README.md) that passes a derived `context.Context` directly into each goroutine function, eliminating the need to capture the context via closure.

## Differences from the root package

| | `github.com/min0625/errgroup` | `github.com/min0625/errgroup/x/errgroup` |
|---|---|---|
| Goroutine signature | `func() error` | `func(context.Context) error` |
| Context access | capture via closure | passed as argument |
| Constructor | `WithContext(ctx)` returns `(*Group, context.Context)` | `New(ctx)` returns `*Group` |
| Zero value | valid, no context cancellation | valid, uses `context.Background()` |

## Installation

```sh
go get github.com/min0625/errgroup/x/errgroup
```

## Example

```go
func Example() {
	// This case uses "github.com/min0625/errgroup/x/errgroup" which will catch panics.
	// If you import "golang.org/x/sync/errgroup" instead, it won't catch panics.
	var g errgroup.Group

	defer func() {
		// Will catch the panic.
		if p := recover(); p != nil {
			switch t := p.(type) {
			case errgroup.PanicValue:
				fmt.Println(t.Recovered)
			case errgroup.PanicError:
				fmt.Println(t.Recovered)
			}
		}
	}()

	g.Go(func(_ context.Context) error {
		// Do something
		return nil
	})

	g.Go(func(_ context.Context) error {
		panic("oops")
	})

	if err := g.Wait(); err != nil {
		// Handle error
		fmt.Println(err)
		return
	}

	// Output: oops
}
```

## Panic behaviour

Panics in goroutines started by `Go` or `TryGo` are caught and re-panicked inside `Wait`, wrapped as:

- `PanicError` — when the panicked value implements `error`
- `PanicValue` — for all other values

Both types expose a `Stack` field containing the stack trace captured at the point of the panic.

If multiple goroutines panic concurrently, only the first panic is propagated; the rest are silently discarded.

## Zero value caveat

A zero-value `Group` (not created with `New`) is valid and automatically initialises itself on the first call to `Go`, `TryGo`, `SetLimit`, or `Wait`. However, the base context will be fixed to `context.Background()`. To use a custom cancellable context, always create the group with `New(ctx)`.
