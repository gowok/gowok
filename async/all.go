package async

import (
	"sync"

	"github.com/gowok/fp/slices"
)

// All runs all tasks
func All(tasks ...func() (any, error)) ([]any, error) {
	var mu sync.Mutex
	results := make([]any, 0)
	var gerr error
	var gerrOnce sync.Once

	slices.GoForEach(tasks, func(task func() (any, error), _ int) {
		data, err := task()
		if err != nil {
			gerrOnce.Do(func() {
				gerr = err
			})
			return
		}

		mu.Lock()
		results = append(results, data)
		mu.Unlock()
	})

	if gerr != nil {
		return nil, gerr
	}

	return results, nil
}
