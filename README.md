# errgroup
[![Go Reference](https://pkg.go.dev/badge/github.com/min0625/errgroup.svg)](https://pkg.go.dev/github.com/min0625/errgroup)
[![codecov](https://codecov.io/gh/min0625/errgroup/branch/main/graph/badge.svg)](https://codecov.io/gh/min0625/errgroup)

**English** | [繁體中文](README.zh-TW.md)

A drop-in `golang.org/x/sync/errgroup` replacement that recovers panics in goroutines and re-panics them inside the `Wait` call.

Ref: https://github.com/golang/go/issues/53757

## Packages

| Package | Description |
|---|---|
| `github.com/min0625/errgroup` | Drop-in replacement for `golang.org/x/sync/errgroup` with panic recovery |
| [`github.com/min0625/errgroup/x/errgroup`](x/errgroup/README.md) | Context-aware variant — passes `context.Context` directly into each goroutine function |

## Installation
```sh
go get github.com/min0625/errgroup
```

## Example
```go
package main

import (
	"fmt"

	"github.com/min0625/errgroup"
)

func main() {
	// This case uses "github.com/min0625/errgroup" which will catch panics.
	// If you import "golang.org/x/sync/errgroup" instead, it won't catch panics.
	// You can try this in the Go Playground: https://go.dev/play/p/7pUX6uQ2mCH
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

	g.Go(func() error {
		// Do something
		return nil
	})

	g.Go(func() error {
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
