package runtime

import (
	"context"
	"errors"
	"os"
	"sync"
	"testing"

	"github.com/golang-must/must"
)

func TestRuntime_Context(t *testing.T) {
	t.Run("positive/new runtime has a live context", func(t *testing.T) {
		r := New()

		ctx := r.Context()
		must.NotNil(t, ctx)
		must.True(t, ctx.Err() == nil)
	})

	t.Run("positive/zero value returns a usable context", func(t *testing.T) {
		r := &Runtime{}

		ctx := r.Context()
		must.NotNil(t, ctx)
		must.Nil(t, ctx.Err())
	})
}

func TestRuntime_Cancel(t *testing.T) {
	t.Run("positive/cancel closes context", func(t *testing.T) {
		r := New()

		r.Cancel()
		must.True(t, errors.Is(r.Context().Err(), context.Canceled))
	})

	t.Run("positive/cancel is idempotent", func(t *testing.T) {
		r := New()

		r.Cancel()
		r.Cancel()
	})

	t.Run("negative/zero value does not panic", func(t *testing.T) {
		r := &Runtime{}
		r.Cancel()
	})

	t.Run("positive/run cancels context after returning", func(t *testing.T) {
		r := New()
		r.runFns = []func(){}

		r.Run(false)
		must.True(t, errors.Is(r.Context().Err(), context.Canceled))
	})
}

func TestRuntime(t *testing.T) {
	testCases := []struct {
		name   string
		action func(t *testing.T)
	}{
		{
			name: "positive/New with options",
			action: func(t *testing.T) {
				r := New(WithNumCPU(4))
				must.NotNil(t, r)
				must.Equal(t, 4, r.numCPU)
			},
		},
		{
			name: "negative/New with invalid CPU",
			action: func(t *testing.T) {
				r := New(WithNumCPU(0))
				must.NotNil(t, r)
				must.Equal(t, 0, r.numCPU)
			},
		},
		{
			name: "positive/AddRunFunc",
			action: func(t *testing.T) {
				r := New()
				initialLen := len(r.runFns)
				r.AddRunFunc(func() {})
				must.Equal(t, initialLen+1, len(r.runFns))
			},
		},
		{
			name: "positive/Run multiple functions",
			action: func(t *testing.T) {
				var mu sync.Mutex
				calledCount := 0

				r := New()
				r.AddRunFunc(func() {
					mu.Lock()
					defer mu.Unlock()
					calledCount++
				})
				r.AddRunFunc(func() {
					mu.Lock()
					defer mu.Unlock()
					calledCount++
				})

				r.Run(false)

				must.Equal(t, 2, calledCount)
			},
		},
		{
			name: "positive/Shutdown sends signal",
			action: func(t *testing.T) {
				r := New()
				r.Shutdown()
				sig := <-r.gracefulStop
				must.Equal(t, os.Kill, sig)
			},
		},
		{
			name: "positive/Run forever mode",
			action: func(t *testing.T) {
				oldExit := exit
				defer func() { exit = oldExit }()

				exitCalled := false
				exit = func(code int) {
					exitCalled = true
				}

				r := New()
				r.Shutdown()
				r.Run(true)

				must.True(t, exitCalled)
			},
		},
		{
			name: "negative/Run when runFns is nil",
			action: func(t *testing.T) {
				r := &Runtime{runFns: nil}
				r.Run()
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.action(t)
		})
	}
}
