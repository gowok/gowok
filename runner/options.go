package runner

import "github.com/gowok/gowok/some"

type option func(*Runner)

func WithRunFunc(runFunc func()) option {
	return func(runner *Runner) {
		runner.runFns = []func(){runFunc}
	}
}

func WithGracefulStopFunc(gracefulStopFunc func()) option {
	return func(runner *Runner) {
		runner.gracefulStopFunc = some.Of(gracefulStopFunc)
	}
}

func WithNumCPU(numCPU int) option {
	return func(runner *Runner) {
		runner.numCPU = numCPU
	}
}

func WithRLimitEnabled() option {
	return func(runner *Runner) {
		runner.rLimitEnable = true
	}
}
