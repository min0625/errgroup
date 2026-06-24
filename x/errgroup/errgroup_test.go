package errgroup_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/min0625/errgroup/x/errgroup"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_Group(t *testing.T) {
	t.Parallel()

	var (
		jobXIsDone bool
		jobYIsDone bool
	)

	var g errgroup.Group

	g.Go(func(_ context.Context) error {
		time.Sleep(100 * time.Millisecond)

		jobXIsDone = true

		return nil
	})

	g.Go(func(_ context.Context) error {
		time.Sleep(100 * time.Millisecond)

		jobYIsDone = true

		return nil
	})

	require.NoError(t, g.Wait())
	assert.True(t, jobXIsDone)
	assert.True(t, jobYIsDone)
}

func Test_WithContext(t *testing.T) {
	t.Parallel()

	myErr := errors.New("oops")

	var jobIsCanceled bool

	g := errgroup.New(context.Background())

	g.Go(func(_ context.Context) error {
		return myErr
	})

	g.Go(func(ctx context.Context) error {
		select {
		case <-ctx.Done():
			jobIsCanceled = true
		case <-time.After(3 * time.Second):
			assert.Fail(t, "context should be canceled")
		}

		return nil
	})

	require.ErrorIs(t, g.Wait(), myErr)
	assert.True(t, jobIsCanceled)
}

func Test_WithContext_ZeroValue(t *testing.T) {
	t.Parallel()

	myErr := errors.New("oops")

	var jobIsCanceled bool

	var g errgroup.Group

	g.Go(func(_ context.Context) error {
		return myErr
	})

	g.Go(func(ctx context.Context) error {
		select {
		case <-ctx.Done():
			jobIsCanceled = true
		case <-time.After(3 * time.Second):
			assert.Fail(t, "context should be canceled")
		}

		return nil
	})

	require.ErrorIs(t, g.Wait(), myErr)
	assert.True(t, jobIsCanceled)
}

func Test_Group_Error(t *testing.T) {
	t.Parallel()

	myErr := errors.New("oops")

	var (
		jobXIsDone bool
		jobYIsDone bool
	)

	var g errgroup.Group

	g.Go(func(_ context.Context) error {
		time.Sleep(100 * time.Millisecond)

		jobXIsDone = true

		return nil
	})

	g.Go(func(_ context.Context) error {
		time.Sleep(100 * time.Millisecond)

		jobYIsDone = true

		return myErr
	})

	require.ErrorIs(t, g.Wait(), myErr)
	assert.True(t, jobXIsDone)
	assert.True(t, jobYIsDone)
}

func Test_Group_Panic(t *testing.T) {
	t.Parallel()

	var g errgroup.Group

	g.Go(func(_ context.Context) error {
		return nil
	})

	g.Go(func(_ context.Context) error {
		panic("oops")
	})

	assert.Panics(t, func() {
		_ = g.Wait()
	})
}

func Test_Group_PanicValue(t *testing.T) {
	t.Parallel()

	var g errgroup.Group

	g.Go(func(_ context.Context) error {
		return nil
	})

	g.Go(func(_ context.Context) error {
		panic("oops")
	})

	defer func() {
		p := recover()
		require.NotNil(t, p)

		pv := p.(errgroup.PanicValue)

		assert.Equal(t, "oops", pv.Recovered)
		assert.Condition(t, func() (success bool) {
			return len(pv.Stack) > 0
		})

		t.Log(pv.String())
		t.Log(pv.Recovered)
		t.Log(string(pv.Stack))
	}()

	_ = g.Wait()
}

func Test_Group_PanicError(t *testing.T) {
	t.Parallel()

	panicValue := errors.New("oops")

	var g errgroup.Group

	g.Go(func(_ context.Context) error {
		return nil
	})

	g.Go(func(_ context.Context) error {
		panic(panicValue)
	})

	defer func() {
		p := recover()
		require.NotNil(t, p)

		pe := p.(errgroup.PanicError)

		assert.Equal(t, panicValue, pe.Recovered)
		assert.Condition(t, func() (success bool) {
			return len(pe.Stack) > 0
		})

		t.Log(pe.Error())
		t.Log(pe.Recovered)
		t.Log(string(pe.Stack))
	}()

	_ = g.Wait()
}

func Test_Group_TryGo(t *testing.T) {
	t.Parallel()

	var g errgroup.Group

	g.SetLimit(1)

	started := make(chan struct{})
	release := make(chan struct{})

	// First goroutine: occupies the single slot.
	accepted := g.TryGo(func(_ context.Context) error {
		close(started)
		<-release

		return nil
	})

	require.True(t, accepted, "first TryGo should be accepted")

	// Wait until the first goroutine is running so the limit is truly reached.
	<-started

	// Second TryGo should be rejected because the limit is reached.
	rejected := g.TryGo(func(_ context.Context) error {
		return nil
	})

	assert.False(t, rejected, "TryGo should be rejected when limit is reached")

	// Release the first goroutine.
	close(release)

	require.NoError(t, g.Wait())
}

func Test_Group_TryGo_ContextPropagation(t *testing.T) {
	t.Parallel()

	myErr := errors.New("oops")

	var jobIsCanceled bool

	g := errgroup.New(context.Background())

	g.TryGo(func(_ context.Context) error {
		return myErr
	})

	accepted := g.TryGo(func(ctx context.Context) error {
		select {
		case <-ctx.Done():
			jobIsCanceled = true
		case <-time.After(3 * time.Second):
			assert.Fail(t, "context should be canceled")
		}

		return nil
	})

	require.True(t, accepted)
	require.ErrorIs(t, g.Wait(), myErr)
	assert.True(t, jobIsCanceled)
}

func Test_New_NilContext(t *testing.T) {
	t.Parallel()

	// New(nil) must fall back to context.Background() instead of panicking,
	// and the goroutine must receive a usable (non-nil, not-yet-canceled)
	// context.
	g := errgroup.New(nil) //nolint:staticcheck // SA1012: nil context is the case under test.

	// Check the context inside the goroutine: Wait cancels the derived context
	// when it returns, so it must be inspected before then.
	g.Go(func(ctx context.Context) error {
		require.NotNil(t, ctx)
		assert.NoError(t, ctx.Err())

		return nil
	})

	require.NoError(t, g.Wait())
}

func Test_Group_MultiplePanics(t *testing.T) {
	t.Parallel()

	// When several goroutines panic concurrently, Wait re-panics with a single
	// recovered value; the other panics are silently discarded rather than
	// collected. This exercises the throw drop path (channel already full) and
	// asserts Wait neither deadlocks nor races (run under -race).
	const n = 10

	var g errgroup.Group

	start := make(chan struct{})

	for i := range n {
		g.Go(func(_ context.Context) error {
			<-start

			panic(fmt.Sprintf("panic %d", i))
		})
	}

	close(start)

	defer func() {
		p := recover()
		require.NotNil(t, p)

		pv, ok := p.(errgroup.PanicValue)
		require.True(t, ok)
		assert.Contains(t, pv.Recovered, "panic ")
	}()

	_ = g.Wait()
}

func Test_Group_TryGo_Panic(t *testing.T) {
	t.Parallel()

	var g errgroup.Group

	accepted := g.TryGo(func(_ context.Context) error {
		panic("oops from TryGo")
	})

	require.True(t, accepted)

	defer func() {
		p := recover()
		require.NotNil(t, p)

		pv := p.(errgroup.PanicValue)

		assert.Equal(t, "oops from TryGo", pv.Recovered)
		assert.Condition(t, func() (success bool) {
			return len(pv.Stack) > 0
		})
	}()

	_ = g.Wait()
}
