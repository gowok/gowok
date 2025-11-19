package async

import (
	"sync"
)

// All runs all tasks
func All(tasks ...func() (any, error)) ([]any, error) {
	var wg sync.WaitGroup
	results := make([]any, 0)
	var gerr error
	var gerrOnce sync.Once

	for _, task := range tasks {
		wg.Add(1)
		go func(t func() (any, error), eo *sync.Once) {
			defer wg.Done()
			data, err := t()
			if err != nil {
				gerrOnce.Do(func() {
					gerr = err
				})
			} else {
				results = append(results, data)
			}
		}(task, &gerrOnce)
	}

	wg.Wait()
	if gerr != nil {
		return nil, gerr
	}

	return results, nil
}
