package errgroup_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/min0625/errgroup"
	"github.com/stretchr/testify/assert"
)

func Test_Group(t *testing.T) {
	t.Parallel()

	var (
		jobXIsDone bool
		jobYIsDone bool
	)

	var g errgroup.Group

	g.Go(func() error {
		jobXIsDone = true

		return nil
	})

	g.Go(func() error {
		jobYIsDone = true

		return nil
	})

	assert.NoError(t, g.Wait())
	assert.Equal(t, jobXIsDone, true)
	assert.Equal(t, jobYIsDone, true)
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

	assert.ErrorIs(t, g.Wait(), myErr)
	assert.Equal(t, jobIsCanceled, true)
}

func Test_Group_Error(t *testing.T) {
	t.Parallel()

	myErr := errors.New("oops")

	var g errgroup.Group

	g.Go(func() error {
		return nil
	})

	g.Go(func() error {
		return myErr
	})

	assert.ErrorIs(t, g.Wait(), myErr)
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
		}
	}()

	_ = g.Wait()
}
