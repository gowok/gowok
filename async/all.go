package async

import (
	"context"
	"sync"

	"github.com/gowok/fp/slices"
)

// All runs all tasks and returns their results in the same order as tasks.
// Nil tasks are skipped and leave a nil result at their index.
// If any task fails, All discards all results and returns the first error.
func All(tasks ...func() (any, error)) ([]any, error) {
	return AllContext(context.Background(), withContext(tasks)...)
}

// AllLimit is like All, but runs at most limit tasks concurrently.
// A limit <= 0 means no limit and is equivalent to All.
func AllLimit(limit int, tasks ...func() (any, error)) ([]any, error) {
	return AllLimitContext(context.Background(), limit, withContext(tasks)...)
}

// AllResults runs all tasks and returns their results in the same order as
// tasks. Unlike All, it keeps the results of the tasks that succeeded when one
// or more tasks fail. Failed and nil tasks leave a nil result at their index,
// so the returned slice always has the same length as tasks.
func AllResults(tasks ...func() (any, error)) ([]any, error) {
	return AllResultsContext(context.Background(), withContext(tasks)...)
}

// AllLimitResults is like AllResults, but runs at most limit tasks
// concurrently. A limit <= 0 means no limit and is equivalent to AllResults.
func AllLimitResults(limit int, tasks ...func() (any, error)) ([]any, error) {
	return AllLimitResultsContext(context.Background(), limit, withContext(tasks)...)
}

func withContext(tasks []func() (any, error)) []func(context.Context) (any, error) {
	if len(tasks) == 0 {
		return nil
	}

	result := make([]func(context.Context) (any, error), len(tasks))
	for i, task := range tasks {
		if task == nil {
			continue
		}

		task := task
		result[i] = func(context.Context) (any, error) {
			return task()
		}
	}

	return result
}

// AllContext is like All, but receives a context and context-aware tasks.
// Tasks that have not started when the context is cancelled are skipped.
// If the context was cancelled, its error is returned instead of a task error.
func AllContext(ctx context.Context, tasks ...func(context.Context) (any, error)) ([]any, error) {
	results, err := AllResultsContext(ctx, tasks...)
	if err != nil {
		return nil, err
	}

	return results, nil
}

// AllLimitContext is like AllContext, but runs at most limit tasks
// concurrently. A limit <= 0 means no limit.
func AllLimitContext(ctx context.Context, limit int, tasks ...func(context.Context) (any, error)) ([]any, error) {
	results, err := AllLimitResultsContext(ctx, limit, tasks...)
	if err != nil {
		return nil, err
	}

	return results, nil
}

// AllResultsContext is like AllResults, but receives a context and
// context-aware tasks. Skipped tasks leave a nil result at their index.
func AllResultsContext(ctx context.Context, tasks ...func(context.Context) (any, error)) ([]any, error) {
	return allResultsContext(ctx, 0, tasks...)
}

// AllLimitResultsContext is like AllLimitResults, but receives a context and
// context-aware tasks. A limit <= 0 means no limit.
func AllLimitResultsContext(ctx context.Context, limit int, tasks ...func(context.Context) (any, error)) ([]any, error) {
	return allResultsContext(ctx, limit, tasks...)
}

func allResultsContext(ctx context.Context, limit int, tasks ...func(context.Context) (any, error)) ([]any, error) {
	if ctx == nil {
		ctx = context.Background()
	}
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

	slices.GoForEach(tasks, func(task func(context.Context) (any, error), i int) {
		if task == nil {
			return
		}

		if err := ctx.Err(); err != nil {
			return
		}

		if sem != nil {
			select {
			case sem <- struct{}{}:
				defer func() { <-sem }()
			case <-ctx.Done():
				return
			}
		}

		if err := ctx.Err(); err != nil {
			return
		}

		data, err := task(ctx)
		if err != nil {
			gerrOnce.Do(func() {
				gerr = err
			})
			return
		}

		results[i] = data
	})

	if err := ctx.Err(); err != nil {
		return results, err
	}

	return results, gerr
}
