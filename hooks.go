package gowok

import "github.com/gowok/gowok/some"

type Hook func()

type Hooks struct {
	onStarting some.Some[Hook]
	onStarted  some.Some[Hook]
	onStopped  some.Some[Hook]
}

func (h *Hooks) OnStarting(hook Hook) {
	h.onStarting = some.Of(hook)
}

func (h *Hooks) OnStarted(hook Hook) {
	h.onStarted = some.Of(hook)
}

func (h *Hooks) OnStopped(hook Hook) {
	h.onStopped = some.Of(hook)
}
