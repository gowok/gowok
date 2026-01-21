package runtime

import (
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"

	"github.com/gowok/gowok/some"
)

var exit = os.Exit

type Runtime struct {
	numCPU           int
	rLimitEnable     bool
	runFns           []func()
	gracefulStop     chan os.Signal
	gracefulStopFunc some.Some[func()]
}

func New(opts ...option) *Runtime {
	runner := &Runtime{
		numCPU:           runtime.NumCPU(),
		runFns:           []func(){func() {}},
		gracefulStop:     make(chan os.Signal, 1),
		gracefulStopFunc: some.Empty[func()](),
	}

	for _, opt := range opts {
		opt(runner)
	}

	runtime.GOMAXPROCS(runner.numCPU)
	return runner
}

func (r *Runtime) AddRunFunc(runFunc func()) {
	r.runFns = append(r.runFns, runFunc)
}

func (r Runtime) Run(forever ...bool) {
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

func (r *Runtime) gracefulStopRun() {
	signal.Notify(r.gracefulStop, syscall.SIGTERM, syscall.SIGINT)

	func() {
		<-r.gracefulStop

		r.gracefulStopFunc.OrElse(func() {})()
		exit(0)
	}()
}

func (r *Runtime) Shutdown() {
	r.gracefulStop <- os.Kill
}
