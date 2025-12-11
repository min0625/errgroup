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
	assert.True(t, jobIsCanceled, true)
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
