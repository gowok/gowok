package async

import (
	"context"
	"errors"
	"sync"
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

func TestAllContext(t *testing.T) {
	t.Run("positive/all succeed and keep order", func(t *testing.T) {
		tasks := []func(context.Context) (any, error){
			func(context.Context) (any, error) { return "a", nil },
			func(context.Context) (any, error) { return 2, nil },
			func(context.Context) (any, error) { return true, nil },
		}

		results, err := AllContext(context.Background(), tasks...)
		must.Nil(t, err)
		must.Equal(t, 3, len(results))
		must.Equal(t, "a", results[0])
		must.Equal(t, 2, results[1])
		must.Equal(t, true, results[2])
	})

	t.Run("negative/task error", func(t *testing.T) {
		tasks := []func(context.Context) (any, error){
			func(context.Context) (any, error) { return nil, errors.New("failed") },
		}

		results, err := AllContext(context.Background(), tasks...)
		must.NotNil(t, err)
		must.Nil(t, results)
	})

	t.Run("negative/cancelled before start skips tasks", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		var calls atomic.Int32
		tasks := []func(context.Context) (any, error){
			func(context.Context) (any, error) { calls.Add(1); return 1, nil },
			func(context.Context) (any, error) { calls.Add(1); return 2, nil },
		}

		results, err := AllContext(ctx, tasks...)
		must.True(t, errors.Is(err, context.Canceled))
		must.Nil(t, results)
		must.Equal(t, int32(0), calls.Load())
	})

	t.Run("positive/nil task keeps index alignment", func(t *testing.T) {
		value := func(v any) func(context.Context) (any, error) {
			return func(context.Context) (any, error) { return v, nil }
		}

		results, err := AllContext(context.Background(), value(1), nil, value(3))
		must.Nil(t, err)
		must.Equal(t, 3, len(results))
		must.Equal(t, 1, results[0])
		must.Nil(t, results[1])
		must.Equal(t, 3, results[2])
	})
}

func TestAllLimitContext(t *testing.T) {
	t.Run("negative/cancelled while running skips queued tasks", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())

		started := make(chan struct{})
		var once sync.Once
		var calls atomic.Int32

		blocker := func(ctx context.Context) (any, error) {
			calls.Add(1)
			once.Do(func() { close(started) })
			<-ctx.Done()
			return nil, ctx.Err()
		}
		tasks := []func(context.Context) (any, error){blocker, blocker, blocker}

		done := make(chan error, 1)
		go func() {
			_, err := AllLimitContext(ctx, 1, tasks...)
			done <- err
		}()

		<-started
		cancel()

		err := <-done
		must.True(t, errors.Is(err, context.Canceled))
		must.Equal(t, int32(1), calls.Load())
	})

	t.Run("positive/bounded concurrency", func(t *testing.T) {
		const (
			limit = 3
			n     = 12
		)

		var current, peak atomic.Int32
		tasks := make([]func(context.Context) (any, error), 0, n)
		for range n {
			tasks = append(tasks, func(context.Context) (any, error) {
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

		results, err := AllLimitContext(context.Background(), limit, tasks...)
		must.Nil(t, err)
		must.Equal(t, n, len(results))
		must.True(t, peak.Load() <= limit)
		must.True(t, peak.Load() > 1)
	})

	t.Run("positive/limit larger than tasks", func(t *testing.T) {
		tasks := []func(context.Context) (any, error){
			func(context.Context) (any, error) { return "a", nil },
			func(context.Context) (any, error) { return "b", nil },
		}

		results, err := AllLimitContext(context.Background(), 100, tasks...)
		must.Nil(t, err)
		must.Equal(t, []any{"a", "b"}, results)
	})
}

func TestAllResultsContext(t *testing.T) {
	t.Run("negative/cancelled keeps partial results", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		var calls atomic.Int32
		tasks := []func(context.Context) (any, error){
			func(context.Context) (any, error) { calls.Add(1); return "a", nil },
			func(context.Context) (any, error) { calls.Add(1); return "b", nil },
		}

		results, err := AllResultsContext(ctx, tasks...)
		must.True(t, errors.Is(err, context.Canceled))
		must.Equal(t, 2, len(results))
		must.Nil(t, results[0])
		must.Nil(t, results[1])
		must.Equal(t, int32(0), calls.Load())
	})

	t.Run("negative/task error keeps successful results", func(t *testing.T) {
		tasks := []func(context.Context) (any, error){
			func(context.Context) (any, error) { return "a", nil },
			func(context.Context) (any, error) { return nil, errors.New("failed") },
			func(context.Context) (any, error) { return "c", nil },
		}

		results, err := AllResultsContext(context.Background(), tasks...)
		must.NotNil(t, err)
		must.Equal(t, 3, len(results))
		must.Equal(t, "a", results[0])
		must.Nil(t, results[1])
		must.Equal(t, "c", results[2])
	})
}

func TestAllLimitResultsContext(t *testing.T) {
	tasks := []func(context.Context) (any, error){
		func(context.Context) (any, error) { return 1, nil },
		nil,
		func(context.Context) (any, error) { return 3, nil },
	}

	results, err := AllLimitResultsContext(context.Background(), 2, tasks...)
	must.Nil(t, err)
	must.Equal(t, 3, len(results))
	must.Equal(t, 1, results[0])
	must.Nil(t, results[1])
	must.Equal(t, 3, results[2])
}
