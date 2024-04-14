package gowok

type Hook func()

type Hooks struct {
	onStarting Hook
	onStarted  Hook
	onStopped  Hook
}

func (h *Hooks) OnStarting(hook Hook) {
	h.onStarting = hook
}

func (h *Hooks) OnStarted(hook Hook) {
	h.onStarted = hook
}

func (h *Hooks) OnStopped(hook Hook) {
	h.onStopped = hook
}
