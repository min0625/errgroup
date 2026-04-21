package errgroup_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/min0625/errgroup"
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

	g.Go(func() error {
		time.Sleep(100 * time.Millisecond)

		jobXIsDone = true

		return nil
	})

	g.Go(func() error {
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

	g, ctx := errgroup.WithContext(context.Background())

	g.Go(func() error {
		return myErr
	})

	g.Go(func() error {
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

	g.Go(func() error {
		time.Sleep(100 * time.Millisecond)

		jobXIsDone = true

		return nil
	})

	g.Go(func() error {
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

	g.Go(func() error {
		return nil
	})

	g.Go(func() error {
		panic("oops")
	})

	assert.Panics(t, func() {
		_ = g.Wait()
	})
}

func Test_Group_PanicValue(t *testing.T) {
	t.Parallel()

	var g errgroup.Group

	g.Go(func() error {
		return nil
	})

	g.Go(func() error {
		panic("oops")
	})

	defer func() {
		if p := recover(); p != nil {
			pv := p.(errgroup.PanicValue)

			assert.Equal(t, "oops", pv.Recovered)
			assert.Condition(t, func() (success bool) {
				return len(pv.Stack) > 0
			})

			t.Log(pv.String())
			t.Log(pv.Recovered)
			t.Log(string(pv.Stack))
		}
	}()

	_ = g.Wait()
}

func Test_Group_PanicError(t *testing.T) {
	t.Parallel()

	panicValue := errors.New("oops")

	var g errgroup.Group

	g.Go(func() error {
		return nil
	})

	g.Go(func() error {
		panic(panicValue)
	})

	defer func() {
		if p := recover(); p != nil {
			pe := p.(errgroup.PanicError)

			assert.Equal(t, panicValue, pe.Recovered)
			assert.Condition(t, func() (success bool) {
				return len(pe.Stack) > 0
			})

			t.Log(pe.Error())
			t.Log(pe.Recovered)
			t.Log(string(pe.Stack))
		}
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
	accepted := g.TryGo(func() error {
		close(started)
		<-release

		return nil
	})

	require.True(t, accepted, "first TryGo should be accepted")

	// Wait until the first goroutine is running so the limit is truly reached.
	<-started

	// Second TryGo should be rejected because the limit is reached.
	rejected := g.TryGo(func() error {
		return nil
	})

	assert.False(t, rejected, "TryGo should be rejected when limit is reached")

	// Release the first goroutine.
	close(release)

	require.NoError(t, g.Wait())
}

func Test_Group_TryGo_Panic(t *testing.T) {
	t.Parallel()

	var g errgroup.Group

	accepted := g.TryGo(func() error {
		panic("oops from TryGo")
	})

	require.True(t, accepted)

	defer func() {
		if p := recover(); p != nil {
			pv := p.(errgroup.PanicValue)

			assert.Equal(t, "oops from TryGo", pv.Recovered)
			assert.Condition(t, func() (success bool) {
				return len(pv.Stack) > 0
			})
		}
	}()

	_ = g.Wait()
}
