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
//
// Unhandled panics from goroutines started by Go or TryGo are caught and
// re-panicked in the Wait call, wrapped as PanicError or PanicValue.
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
//
// If f panics, the panic is caught and re-panicked in Wait, wrapped as a
// PanicError or PanicValue.
func (g *Group) Go(f func() error) {
	g.eg.Go(g.do(f))
}

// TryGo calls the given function in a new goroutine only if the number of
// active goroutines in the group is currently below the configured limit.
//
// The return value reports whether the goroutine was started.
//
// If f panics, the panic is caught and re-panicked in Wait, wrapped as a
// PanicError or PanicValue.
func (g *Group) TryGo(f func() error) bool {
	return g.eg.TryGo(g.do(f))
}

// Wait blocks until all function calls from the Go method have returned, then
// returns the first non-nil error (if any) from them.
//
// If any goroutine panicked, Wait re-panics with the recovered value wrapped
// as a PanicError (if the panic value implements error) or PanicValue.
// If multiple goroutines panic, only the first panic is re-panicked; the rest
// are silently discarded.
//
// Implementation note: Wait runs g.eg.Wait() in a separate goroutine so that
// a panic inside eg.Wait itself (e.g. from the underlying errgroup) can also be
// caught and forwarded through the same panicValue channel.  The inner goroutine
// writes to a buffered channel of size 1, so it always completes and is never
// leaked, even when Wait returns early via a panic branch.
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
		// A goroutine panicked before eg.Wait returned.  The inner goroutine
		// will eventually write to waitError (buffered) and exit on its own.
		panic(panicValue)
	case err := <-waitError:
		// eg.Wait returned normally; do a non-blocking check in case a panic
		// was also sent (e.g. the goroutine panicked and the sentinel error
		// arrived first).
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
