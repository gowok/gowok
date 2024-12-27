package runner

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/gowok/gowok/some"
)

type Runner struct {
	numCPU           int
	rLimitEnable     bool
	runFns           []func()
	gracefulStopFunc some.Some[func()]
}

func New(opts ...Option) *Runner {
	runner := &Runner{
		runFns:           []func(){func() {}},
		gracefulStopFunc: some.Empty[func()](),
	}

	for _, opt := range opts {
		opt(runner)
	}
	return runner
}

func (r *Runner) AddRunFunc(runFunc func()) {
	r.runFns = append(r.runFns, runFunc)
}

func (r Runner) Run() {
	if r.runFns == nil {
		return
	}

	for i := len(r.runFns) - 1; i >= 0; i-- {
		go r.runFns[i]()
	}

	r.gracefulStopRun()
}

func (r Runner) gracefulStopRun() {
	var gracefulStop = make(chan os.Signal, 1)
	signal.Notify(gracefulStop, syscall.SIGTERM)
	signal.Notify(gracefulStop, syscall.SIGINT)

	func() {
		<-gracefulStop

		r.gracefulStopFunc.OrElse(func() {})()
		os.Exit(0)
	}()
}
