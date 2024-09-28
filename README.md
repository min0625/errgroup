# errgroup
[![Go Reference](https://pkg.go.dev/badge/github.com/min0625/errgroup.svg)](https://pkg.go.dev/github.com/min0625/errgroup)

A recoverable errgroup based on `golang.org/x/sync/errgroup` that can recover from panics. Panics are caught and re-panicked in the Wait function.

Ref: https://github.com/golang/go/issues/53757

## Installation
```sh
go get github.com/min0625/errgroup
```

## Example
```go

func Example() {
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
	}

	// Output: oops
}

```
