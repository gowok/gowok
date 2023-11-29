package runner

type Option func(*Runner)

func WithRunFunc(runFunc func()) Option {
	return func(runner *Runner) {
		runner.runFns = []func(){runFunc}
	}
}

func WithGracefulStopFunc(gracefulStopFunc func()) Option {
	return func(runner *Runner) {
		runner.gracefulStopFunc = gracefulStopFunc
	}
}

func WithNumCPU(numCPU int) Option {
	return func(runner *Runner) {
		runner.numCPU = numCPU
	}
}

func WithRLimitEnable(rlimitEnable bool) Option {
	return func(runner *Runner) {
		runner.rLimitEnable = rlimitEnable
	}
}
