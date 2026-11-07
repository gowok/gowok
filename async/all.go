package async

import (
	"sync"

	"github.com/gowok/fp/slices"
)

// All runs all tasks and returns their results in the same order as tasks.
// Nil tasks are skipped and leave a nil result at their index.
// If any task fails, All discards all results and returns the first error.
func All(tasks ...func() (any, error)) ([]any, error) {
	results, err := AllResults(tasks...)
	if err != nil {
		return nil, err
	}

	return results, nil
}

// AllLimit is like All, but runs at most limit tasks concurrently.
// A limit <= 0 means no limit and is equivalent to All.
func AllLimit(limit int, tasks ...func() (any, error)) ([]any, error) {
	results, err := AllLimitResults(limit, tasks...)
	if err != nil {
		return nil, err
	}

	return results, nil
}

// AllResults runs all tasks and returns their results in the same order as
// tasks. Unlike All, it keeps the results of the tasks that succeeded when one
// or more tasks fail. Failed and nil tasks leave a nil result at their index,
// so the returned slice always has the same length as tasks.
func AllResults(tasks ...func() (any, error)) ([]any, error) {
	return allResults(0, tasks...)
}

// AllLimitResults is like AllResults, but runs at most limit tasks
// concurrently. A limit <= 0 means no limit and is equivalent to AllResults.
func AllLimitResults(limit int, tasks ...func() (any, error)) ([]any, error) {
	return allResults(limit, tasks...)
}

// allResults runs tasks with a concurrency ceiling of limit. A limit <= 0, or a
// limit that already covers every task, runs them without a ceiling.
func allResults(limit int, tasks ...func() (any, error)) ([]any, error) {
	if limit >= len(tasks) {
		limit = 0
	}

	results := make([]any, len(tasks))
	var gerr error
	var gerrOnce sync.Once
	var sem chan struct{}
	if limit > 0 {
		sem = make(chan struct{}, limit)
	}

	slices.GoForEach(tasks, func(task func() (any, error), i int) {
		if task == nil {
			return
		}

		if sem != nil {
			sem <- struct{}{}
			defer func() { <-sem }()
		}

		data, err := task()
		if err != nil {
			gerrOnce.Do(func() {
				gerr = err
			})
			return
		}

		results[i] = data
	})

	return results, gerr
}
