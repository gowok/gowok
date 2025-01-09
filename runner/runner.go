package runner

import (
	"os"
	"os/signal"
	"runtime"
	"sync"
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
		numCPU:           runtime.NumCPU(),
		runFns:           []func(){func() {}},
		gracefulStopFunc: some.Empty[func()](),
	}

	for _, opt := range opts {
		opt(runner)
	}

	runtime.GOMAXPROCS(runner.numCPU)
	return runner
}

func (r *Runner) AddRunFunc(runFunc func()) {
	r.runFns = append(r.runFns, runFunc)
}

func (r Runner) Run(forever ...bool) {
	if r.runFns == nil {
		return
	}

	var wg sync.WaitGroup
	for i := len(r.runFns) - 1; i >= 0; i-- {
		wg.Add(1)
		go func() {
			defer wg.Done()
			r.runFns[i]()
		}()
	}

	if len(forever) > 0 && forever[0] {
		r.gracefulStopRun()
	} else {
		wg.Wait()
	}
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
