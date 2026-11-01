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

// AllResults runs all tasks and returns their results in the same order as
// tasks. Unlike All, it keeps the results of the tasks that succeeded when one
// or more tasks fail. Failed and nil tasks leave a nil result at their index,
// so the returned slice always has the same length as tasks.
func AllResults(tasks ...func() (any, error)) ([]any, error) {
	results := make([]any, len(tasks))
	var gerr error
	var gerrOnce sync.Once

	slices.GoForEach(tasks, func(task func() (any, error), i int) {
		if task == nil {
			return
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
