package runtime

import "github.com/gowok/gowok/some"

type option func(*Runtime)

func WithRunFunc(runFunc func()) option {
	return func(runner *Runtime) {
		runner.runFns = []func(){runFunc}
	}
}

func WithGracefulStopFunc(gracefulStopFunc func()) option {
	return func(runner *Runtime) {
		runner.gracefulStopFunc = some.Of(gracefulStopFunc)
	}
}

func WithNumCPU(numCPU int) option {
	return func(runner *Runtime) {
		runner.numCPU = numCPU
	}
}

func WithRLimitEnabled() option {
	return func(runner *Runtime) {
		runner.rLimitEnable = true
	}
}
