package runtime

import (
	"testing"

	"github.com/golang-must/must"
)

func TestHooks(t *testing.T) {
	testCases := []struct {
		name     string
		action   func(h *Hooks, hook func())
		wantCall bool
	}{
		{
			name:     "negative/Init not set",
			action:   func(h *Hooks, hook func()) { h.Init()() },
			wantCall: false,
		},
		{
			name: "positive/Init set",
			action: func(h *Hooks, hook func()) {
				h.SetInit(hook)
				h.Init()()
			},
			wantCall: true,
		},
		{
			name:     "negative/OnStarting not set",
			action:   func(h *Hooks, hook func()) { h.OnStarting()() },
			wantCall: false,
		},
		{
			name: "positive/OnStarting set",
			action: func(h *Hooks, hook func()) {
				h.SetOnStarting(hook)
				h.OnStarting()()
			},
			wantCall: true,
		},
		{
			name:     "negative/OnStarted not set",
			action:   func(h *Hooks, hook func()) { h.OnStarted()() },
			wantCall: false,
		},
		{
			name: "positive/OnStarted set",
			action: func(h *Hooks, hook func()) {
				h.SetOnStarted(hook)
				h.OnStarted()()
			},
			wantCall: true,
		},
		{
			name:     "negative/OnStopped not set",
			action:   func(h *Hooks, hook func()) { h.OnStopped()() },
			wantCall: false,
		},
		{
			name: "positive/OnStopped set",
			action: func(h *Hooks, hook func()) {
				h.SetOnStopped(hook)
				h.OnStopped()()
			},
			wantCall: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			hooks := &Hooks{}
			called := false
			hook := func() { called = true }

			tc.action(hooks, hook)
			must.Equal(t, tc.wantCall, called)
		})
	}
}
