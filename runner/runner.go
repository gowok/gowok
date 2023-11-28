package runner

import (
	"os"
	"os/signal"
	"syscall"
)

type Runner struct {
	NumCPU           int
	RLimitEnable     bool
	RunFunc          func()
	GracefulStopFunc func()
}

func New(opts ...Option) *Runner {
	runner := &Runner{
		RunFunc:          func() {},
		GracefulStopFunc: func() {},
	}

	for _, opt := range opts {
		opt(runner)
	}
	return runner
}

func (r Runner) Run() {
	if r.RunFunc == nil {
		return
	}

	go r.RunFunc()
	if r.GracefulStopFunc != nil {
		r.gracefulStopRun(r.GracefulStopFunc)
	}
}

func (r Runner) gracefulStopRun(callback func()) {
	var gracefulStop = make(chan os.Signal)
	signal.Notify(gracefulStop, syscall.SIGTERM)
	signal.Notify(gracefulStop, syscall.SIGINT)

	func() {
		<-gracefulStop

		callback()

		os.Exit(0)
	}()
}
