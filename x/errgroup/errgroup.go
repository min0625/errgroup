// Package errgroup provides a more robust error group implementation
// that extends golang.org/x/sync/errgroup with panic recovery.
package errgroup

import (
	"context"
	"sync"

	"github.com/min0625/errgroup"
)

// A Group is a collection of goroutines working on subtasks that are part of
// the same overall task.
//
// A zero Group is valid, has no limit on the number of active goroutines,
// and uses context.Background() as its base context. To use a custom context
// that is canceled on error, create a Group with New instead.
//
// Unhandled panics from goroutines started by Go or TryGo are caught and
// re-panicked in the Wait call, wrapped as PanicError or PanicValue.
type Group struct {
	once sync.Once
	eg   *errgroup.Group
	ctx  context.Context
}

// New returns a new Group whose goroutines receive a Context derived from ctx.
//
// The derived Context is canceled the first time a function passed to Go
// returns a non-nil error or the first time Wait returns, whichever occurs
// first.
func New(ctx context.Context) *Group {
	g := &Group{}
	g.tryInit(ctx)

	return g
}

// SetLimit limits the number of active goroutines in this group to at most n.
// A negative value indicates no limit.
//
// Any subsequent call to the Go method will block until it can add an active
// goroutine without exceeding the configured limit.
//
// The limit must not be modified while any goroutines in the group are active.
//
// Calling SetLimit on a zero-value Group (i.e. not created with New) initializes
// the group with context.Background() as its base context.
func (g *Group) SetLimit(n int) {
	g.tryInit(context.Background())

	g.eg.SetLimit(n)
}

// Go calls the given function in a new goroutine, passing the group's derived
// Context as the argument. It blocks until the new goroutine can be added
// without the number of active goroutines in the group exceeding the configured
// limit.
//
// The first call to return a non-nil error cancels the group's context, if the
// group was created by calling New. The error will be returned by Wait.
//
// If f panics, the panic is caught and re-panicked in Wait, wrapped as a
// PanicError or PanicValue.
func (g *Group) Go(f func(context.Context) error) {
	g.tryInit(context.Background())

	g.eg.Go(func() (err error) {
		return f(g.ctx)
	})
}

// TryGo calls the given function in a new goroutine, passing the group's
// derived Context as the argument, only if the number of active goroutines in
// the group is currently below the configured limit.
//
// The return value reports whether the goroutine was started.
//
// If f panics, the panic is caught and re-panicked in Wait, wrapped as a
// PanicError or PanicValue.
func (g *Group) TryGo(f func(context.Context) error) bool {
	g.tryInit(context.Background())

	return g.eg.TryGo(func() (err error) {
		return f(g.ctx)
	})
}

// Wait blocks until all function calls from the Go method have returned, then
// returns the first non-nil error (if any) from them.
func (g *Group) Wait() error {
	g.tryInit(context.Background())

	return g.eg.Wait()
}

func (g *Group) tryInit(ctx context.Context) {
	g.once.Do(func() {
		initCtx := ctx
		if initCtx == nil {
			initCtx = context.Background()
		}

		eg, derivedCtx := errgroup.WithContext(initCtx)

		g.eg = eg
		g.ctx = derivedCtx
	})
}
