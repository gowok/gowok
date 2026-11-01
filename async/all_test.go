package async

import (
	"errors"
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
