package runner

type Option func(*Runner)

func WithRunFunc(runFunc func()) Option {
	return func(runner *Runner) {
		runner.RunFunc = runFunc
	}
}

func WithGracefulStopFunc(gracefulStopFunc func()) Option {
	return func(runner *Runner) {
		runner.GracefulStopFunc = gracefulStopFunc
	}
}

func WithNumCPU(numCPU int) Option {
	return func(runner *Runner) {
		runner.NumCPU = numCPU
	}
}

func WithRLimitEnable(rlimitEnable bool) Option {
	return func(runner *Runner) {
		runner.RLimitEnable = rlimitEnable
	}
}
