package runner

import (
	"github.com/gowok/gowok/some"
)

type Hooks struct {
	Init       some.Some[func()]
	onStarting some.Some[func()]
	onStarted  some.Some[func()]
	onStopped  some.Some[func()]
}

func (h *Hooks) SetInit(hook func()) {
	h.Init = some.Of(hook)
}

func (h *Hooks) SetOnStarting(hook func()) {
	h.onStarting = some.Of(hook)
}

func (h *Hooks) SetOnStarted(hook func()) {
	h.onStarted = some.Of(hook)
}

func (h *Hooks) SetOnStopped(hook func()) {
	h.onStopped = some.Of(hook)
}

// func (h *Hooks) Init() func() {
// 	return h.init.OrElse(func() {})
// }

func (h *Hooks) OnStarting() func() {
	return h.onStarting.OrElse(func() {})
}

func (h *Hooks) OnStarted() func() {
	return h.onStarted.OrElse(func() {})
}

func (h *Hooks) OnStopped() func() {
	return h.onStopped.OrElse(func() {})
}
