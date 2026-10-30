package singleton

import (
	"sync"
	"sync/atomic"
)

type SingletonFunc[T any] func() T

func New[T any](singletonFunc SingletonFunc[T]) func(...T) *T {
	var once sync.Once
	var value atomic.Pointer[T]

	return func(newValue ...T) *T {
		if len(newValue) > 0 {
			value.Store(&newValue[0])
			return value.Load()
		}

		once.Do(func() {
			if value.Load() != nil {
				return
			}
			create := singletonFunc()
			value.Store(&create)
		})

		return value.Load()
	}
}
