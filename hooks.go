package gowok

import "github.com/gowok/gowok/some"

type Hook func()

type Hooks struct {
	onStarting Hook
	onStarted  Hook
	onStopped  some.Some[Hook]
}

func (h *Hooks) OnStarting(hook Hook) {
	h.onStarting = hook
}

func (h *Hooks) OnStarted(hook Hook) {
	h.onStarted = hook
}

func (h *Hooks) OnStopped(hook Hook) {
	h.onStopped = some.Of(&hook)
}
