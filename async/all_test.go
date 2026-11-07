package async

import (
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/golang-must/must"
)

func TestAll(t *testing.T) {
	testCases := []struct {
		name          string
		tasks         []func() (any, error)
		expectedLen   int
		expectError   bool
		errorContains string
	}{
		{
			name: "positive/multiple successful tasks",
			tasks: []func() (any, error){
				func() (any, error) { return "result1", nil },
				func() (any, error) { return 42, nil },
				func() (any, error) { return true, nil },
			},
			expectedLen: 3,
			expectError: false,
		},
		{
			name: "positive/single task",
			tasks: []func() (any, error){
				func() (any, error) { return "single result", nil },
			},
			expectedLen: 1,
			expectError: false,
		},
		{
			name:        "positive/empty tasks",
			tasks:       []func() (any, error){},
			expectedLen: 0,
			expectError: false,
		},
		{
			name: "positive/nil task is skipped",
			tasks: []func() (any, error){
				func() (any, error) { return "result1", nil },
				nil,
				func() (any, error) { return "result3", nil },
			},
			expectedLen: 3,
			expectError: false,
		},
		{
			name: "negative/single task with error",
			tasks: []func() (any, error){
				func() (any, error) { return nil, errors.New("task failed") },
			},
			expectedLen: 0,
			expectError: true,
		},
		{
			name: "negative/multiple tasks with one error",
			tasks: []func() (any, error){
				func() (any, error) { return "result1", nil },
				func() (any, error) { return nil, errors.New("second task failed") },
				func() (any, error) { return "result3", nil },
			},
			expectedLen: 0,
			expectError: true,
		},
		{
			name: "negative/all tasks with errors",
			tasks: []func() (any, error){
				func() (any, error) { return nil, errors.New("error 1") },
				func() (any, error) { return nil, errors.New("error 2") },
				func() (any, error) { return nil, errors.New("error 3") },
			},
			expectedLen: 0,
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			results, err := All(tc.tasks...)

			if tc.expectError {
				must.NotNil(t, err)
			} else {
				must.Nil(t, err)
			}

			must.Equal(t, tc.expectedLen, len(results))
		})
	}
}

func TestAll_NilTaskKeepsIndexAlignment(t *testing.T) {
	value := func(v any) func() (any, error) {
		return func() (any, error) { return v, nil }
	}

	results, err := All(value(1), nil, value(3))
	must.Nil(t, err)
	must.Equal(t, 3, len(results))
	must.Equal(t, 1, results[0].(int))
	must.Nil(t, results[1])
	must.Equal(t, 3, results[2].(int))
}

func TestAll_PreservesOrder(t *testing.T) {
	const n = 8

	tasks := make([]func() (any, error), 0, n)
	for i := 0; i < n; i++ {
		i := i
		tasks = append(tasks, func() (any, error) {
			time.Sleep(time.Duration(n-i) * 5 * time.Millisecond)
			return i, nil
		})
	}

	results, err := All(tasks...)
	must.Nil(t, err)
	must.Equal(t, n, len(results))
	for i, result := range results {
		must.Equal(t, i, result.(int))
	}
}

func TestAllResults(t *testing.T) {
	testCases := []struct {
		name        string
		tasks       []func() (any, error)
		expectError bool
		expectNil   []int
	}{
		{
			name: "positive/all succeed",
			tasks: []func() (any, error){
				func() (any, error) { return "a", nil },
				func() (any, error) { return 2, nil },
			},
			expectError: false,
		},
		{
			name:        "positive/empty tasks",
			tasks:       []func() (any, error){},
			expectError: false,
		},
		{
			name: "negative/one of many fails keeps the rest",
			tasks: []func() (any, error){
				func() (any, error) { return "a", nil },
				func() (any, error) { return nil, errors.New("task failed") },
				func() (any, error) { return "c", nil },
			},
			expectError: true,
			expectNil:   []int{1},
		},
		{
			name: "positive/nil task is skipped",
			tasks: []func() (any, error){
				func() (any, error) { return "a", nil },
				nil,
			},
			expectError: false,
			expectNil:   []int{1},
		},
		{
			name: "negative/all fail",
			tasks: []func() (any, error){
				func() (any, error) { return nil, errors.New("error 1") },
				func() (any, error) { return nil, errors.New("error 2") },
			},
			expectError: true,
			expectNil:   []int{0, 1},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			results, err := AllResults(tc.tasks...)

			if tc.expectError {
				must.NotNil(t, err)
			} else {
				must.Nil(t, err)
			}

			must.NotNil(t, results)
			must.Equal(t, len(tc.tasks), len(results))

			nilIndexes := map[int]bool{}
			for _, i := range tc.expectNil {
				nilIndexes[i] = true
			}
			for i, result := range results {
				if nilIndexes[i] {
					must.Nil(t, result)
					continue
				}
				must.NotNil(t, result)
			}
		})
	}
}

func TestAllLimit(t *testing.T) {
	testCases := []struct {
		name        string
		limit       int
		tasks       []func() (any, error)
		expectedLen int
		expectError bool
	}{
		{
			name:  "positive/multiple successful tasks",
			limit: 2,
			tasks: []func() (any, error){
				func() (any, error) { return "result1", nil },
				func() (any, error) { return 42, nil },
				func() (any, error) { return true, nil },
			},
			expectedLen: 3,
		},
		{
			name:  "positive/single task",
			limit: 4,
			tasks: []func() (any, error){
				func() (any, error) { return "single result", nil },
			},
			expectedLen: 1,
		},
		{
			name:        "positive/empty tasks",
			limit:       2,
			tasks:       []func() (any, error){},
			expectedLen: 0,
		},
		{
			name:  "positive/nil task is skipped",
			limit: 2,
			tasks: []func() (any, error){
				func() (any, error) { return "result1", nil },
				nil,
				func() (any, error) { return "result3", nil },
			},
			expectedLen: 3,
		},
		{
			name:  "positive/limit larger than tasks",
			limit: 100,
			tasks: []func() (any, error){
				func() (any, error) { return "a", nil },
				func() (any, error) { return "b", nil },
			},
			expectedLen: 2,
		},
		{
			name:  "positive/zero limit means unlimited",
			limit: 0,
			tasks: []func() (any, error){
				func() (any, error) { return "a", nil },
				func() (any, error) { return "b", nil },
			},
			expectedLen: 2,
		},
		{
			name:  "positive/negative limit means unlimited",
			limit: -1,
			tasks: []func() (any, error){
				func() (any, error) { return "a", nil },
				func() (any, error) { return "b", nil },
			},
			expectedLen: 2,
		},
		{
			name:  "negative/one of many fails",
			limit: 2,
			tasks: []func() (any, error){
				func() (any, error) { return "result1", nil },
				func() (any, error) { return nil, errors.New("second task failed") },
				func() (any, error) { return "result3", nil },
			},
			expectedLen: 0,
			expectError: true,
		},
		{
			name:  "negative/limit one fails",
			limit: 1,
			tasks: []func() (any, error){
				func() (any, error) { return nil, errors.New("error 1") },
				func() (any, error) { return nil, errors.New("error 2") },
			},
			expectedLen: 0,
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			results, err := AllLimit(tc.limit, tc.tasks...)

			if tc.expectError {
				must.NotNil(t, err)
			} else {
				must.Nil(t, err)
			}

			must.Equal(t, tc.expectedLen, len(results))
		})
	}
}

func TestAllLimitResults(t *testing.T) {
	testCases := []struct {
		name        string
		limit       int
		tasks       []func() (any, error)
		expectError bool
		expectNil   []int
	}{
		{
			name:  "positive/all succeed",
			limit: 2,
			tasks: []func() (any, error){
				func() (any, error) { return "a", nil },
				func() (any, error) { return 2, nil },
			},
		},
		{
			name:  "negative/one of many fails keeps the rest",
			limit: 1,
			tasks: []func() (any, error){
				func() (any, error) { return "a", nil },
				func() (any, error) { return nil, errors.New("task failed") },
				func() (any, error) { return "c", nil },
			},
			expectError: true,
			expectNil:   []int{1},
		},
		{
			name:  "positive/nil task is skipped",
			limit: 3,
			tasks: []func() (any, error){
				func() (any, error) { return "a", nil },
				nil,
			},
			expectNil: []int{1},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			results, err := AllLimitResults(tc.limit, tc.tasks...)

			if tc.expectError {
				must.NotNil(t, err)
			} else {
				must.Nil(t, err)
			}

			must.NotNil(t, results)
			must.Equal(t, len(tc.tasks), len(results))

			nilIndexes := map[int]bool{}
			for _, i := range tc.expectNil {
				nilIndexes[i] = true
			}
			for i, result := range results {
				if nilIndexes[i] {
					must.Nil(t, result)
					continue
				}
				must.NotNil(t, result)
			}
		})
	}
}

func TestAllLimit_NilTaskKeepsIndexAlignment(t *testing.T) {
	value := func(v any) func() (any, error) {
		return func() (any, error) { return v, nil }
	}

	results, err := AllLimit(1, value(1), nil, value(3))
	must.Nil(t, err)
	must.Equal(t, 3, len(results))
	must.Equal(t, 1, results[0].(int))
	must.Nil(t, results[1])
	must.Equal(t, 3, results[2].(int))
}

func TestAllLimit_PreservesOrder(t *testing.T) {
	const n = 8

	tasks := make([]func() (any, error), 0, n)
	for i := 0; i < n; i++ {
		i := i
		tasks = append(tasks, func() (any, error) {
			time.Sleep(time.Duration(n-i) * 5 * time.Millisecond)
			return i, nil
		})
	}

	results, err := AllLimit(2, tasks...)
	must.Nil(t, err)
	must.Equal(t, n, len(results))
	for i, result := range results {
		must.Equal(t, i, result.(int))
	}
}

func TestAllLimit_BoundedConcurrency(t *testing.T) {
	const (
		limit = 3
		n     = 12
	)

	var current, peak atomic.Int32

	tasks := make([]func() (any, error), 0, n)
	for i := 0; i < n; i++ {
		tasks = append(tasks, func() (any, error) {
			running := current.Add(1)
			for {
				seen := peak.Load()
				if running <= seen || peak.CompareAndSwap(seen, running) {
					break
				}
			}
			time.Sleep(5 * time.Millisecond)
			current.Add(-1)
			return nil, nil
		})
	}

	results, err := AllLimit(limit, tasks...)
	must.Nil(t, err)
	must.Equal(t, n, len(results))
	must.True(t, peak.Load() <= limit)
	must.True(t, peak.Load() > 1)
}

func TestAllLimit_LimitOneRunsOneAtATime(t *testing.T) {
	const n = 5

	var current, peak atomic.Int32
	tasks := make([]func() (any, error), 0, n)
	for i := 0; i < n; i++ {
		i := i
		tasks = append(tasks, func() (any, error) {
			running := current.Add(1)
			for {
				seen := peak.Load()
				if running <= seen || peak.CompareAndSwap(seen, running) {
					break
				}
			}
			time.Sleep(2 * time.Millisecond)
			current.Add(-1)
			return i, nil
		})
	}

	results, err := AllLimit(1, tasks...)
	must.Nil(t, err)
	must.Equal(t, n, len(results))
	must.Equal(t, int32(1), peak.Load())

	// Concurrency is bounded but execution order is not part of the contract.
	seen := map[int]bool{}
	for i, result := range results {
		must.Equal(t, i, result.(int))
		seen[result.(int)] = true
	}
	must.Equal(t, n, len(seen))
}
