package runtime

import (
	"testing"

	"github.com/golang-must/must"
)

func TestOptions(t *testing.T) {
	testCases := []struct {
		name  string
		opt   option
		check func(t *testing.T, r *Runtime)
	}{
		{
			name: "positive/WithRunFunc",
			opt:  WithRunFunc(func() {}),
			check: func(t *testing.T, r *Runtime) {
				must.Equal(t, 1, len(r.runFns))
			},
		},
		{
			name: "positive/WithGracefulStopFunc",
			opt:  WithGracefulStopFunc(func() {}),
			check: func(t *testing.T, r *Runtime) {
				must.True(t, r.gracefulStopFunc.IsPresent())
			},
		},
		{
			name: "positive/WithNumCPU",
			opt:  WithNumCPU(2),
			check: func(t *testing.T, r *Runtime) {
				must.Equal(t, 2, r.numCPU)
			},
		},
		{
			name: "positive/WithRLimitEnabled",
			opt:  WithRLimitEnabled(),
			check: func(t *testing.T, r *Runtime) {
				must.True(t, r.rLimitEnable)
			},
		},
		{
			name: "negative/WithNumCPU zero",
			opt:  WithNumCPU(0),
			check: func(t *testing.T, r *Runtime) {
				must.Equal(t, 0, r.numCPU)
			},
		},
		{
			name: "negative/WithNumCPU negative",
			opt:  WithNumCPU(-1),
			check: func(t *testing.T, r *Runtime) {
				must.Equal(t, -1, r.numCPU)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			r := &Runtime{}
			tc.opt(r)
			tc.check(t, r)
		})
	}
}
