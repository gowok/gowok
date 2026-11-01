package async

import (
	"sync"

	"github.com/gowok/fp/slices"
)

// All runs all tasks and returns their results in the same order as tasks.
func All(tasks ...func() (any, error)) ([]any, error) {
	results := make([]any, len(tasks))
	var gerr error
	var gerrOnce sync.Once

	slices.GoForEach(tasks, func(task func() (any, error), i int) {
		data, err := task()
		if err != nil {
			gerrOnce.Do(func() {
				gerr = err
			})
			return
		}

		results[i] = data
	})

	if gerr != nil {
		return nil, gerr
	}

	return results, nil
}
