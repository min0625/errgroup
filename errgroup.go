// Package errgroup provides a more robust error group implementation
// that extends golang.org/x/sync/errgroup with panic recovery.
package errgroup

import (
	"context"
	"errors"
	"sync"

	"golang.org/x/sync/errgroup"
)

// A Group is a collection of goroutines working on subtasks that are part of
// the same overall task.
//
// A zero Group is valid, has no limit on the number of active goroutines,
// and does not cancel on error.
type Group struct {
	once       sync.Once
	panicValue chan any
	eg         errgroup.Group
}

// WithContext returns a new Group and an associated Context derived from ctx.
//
// The derived Context is canceled the first time a function passed to Go
// returns a non-nil error or the first time Wait returns, whichever occurs
// first.
func WithContext(ctx context.Context) (*Group, context.Context) {
	eg, ctx := errgroup.WithContext(ctx)

	return &Group{
		eg: *eg, //nolint:govet // copylocks: Copy before the first use.
	}, ctx
}

// SetLimit limits the number of active goroutines in this group to at most n.
// A negative value indicates no limit.
//
// Any subsequent call to the Go method will block until it can add an active
// goroutine without exceeding the configured limit.
//
// The limit must not be modified while any goroutines in the group are active.
func (g *Group) SetLimit(n int) {
	g.eg.SetLimit(n)
}

// Go calls the given function in a new goroutine.
// It blocks until the new goroutine can be added without the number of
// active goroutines in the group exceeding the configured limit.
//
// The first call to return a non-nil error cancels the group's context, if the
// group was created by calling WithContext. The error will be returned by Wait.
func (g *Group) Go(f func() error) {
	g.eg.Go(g.do(f))
}

// TryGo calls the given function in a new goroutine only if the number of
// active goroutines in the group is currently below the configured limit.
//
// The return value reports whether the goroutine was started.
func (g *Group) TryGo(f func() error) bool {
	return g.eg.TryGo(g.do(f))
}

// Wait blocks until all function calls from the Go method have returned, then
// returns the first non-nil error (if any) from them.
func (g *Group) Wait() error {
	g.init()

	waitError := make(chan error, 1)

	go func() {
		defer func() {
			if r := recover(); r != nil {
				g.throw(exception(r))
			}
		}()

		waitError <- g.eg.Wait()
	}()

	select {
	case panicValue := <-g.panicValue:
		panic(panicValue)
	case err := <-waitError:
		// Double check panic occurred
		select {
		case panicValue := <-g.panicValue:
			panic(panicValue)
		default:
			return err
		}
	}
}

func (g *Group) init() {
	g.once.Do(func() {
		g.panicValue = make(chan any, 1)
	})
}

func (g *Group) throw(p any) {
	g.init()

	select {
	case g.panicValue <- p:
	default:
	}
}

func (g *Group) do(f func() error) func() error {
	return func() (err error) {
		defer func() {
			if r := recover(); r != nil {
				// Return a non-nil error to ensure the context is canceled
				err = errPanic

				g.throw(exception(r))
			}
		}()

		return f()
	}
}

var errPanic = errors.New("panic in errgroup.Group")
