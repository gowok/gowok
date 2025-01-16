package gowok

import "github.com/gowok/gowok/some"

type Hooks struct {
	onStarting some.Some[func()]
	onStarted  some.Some[func()]
	onStopped  some.Some[func()]
}

func (h *Hooks) OnStarting(hook func()) {
	h.onStarting = some.Of(hook)
}

func (h *Hooks) OnStarted(hook func()) {
	h.onStarted = some.Of(hook)
}

func (h *Hooks) OnStopped(hook func()) {
	h.onStopped = some.Of(hook)
}
