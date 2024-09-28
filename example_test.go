package errgroup_test

import (
	"fmt"

	"github.com/min0625/errgroup"
)

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
