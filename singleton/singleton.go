package singleton

import (
	"sync"
)

type SingletonFunc[T any] func() T

func New[T any](singletonFunc SingletonFunc[T]) func(...T) *T {
	var once sync.Once
	var value *T

	return func(newValue ...T) *T {
		if len(newValue) > 0 {
			value = &newValue[0]
			return value
		}

		if value == nil {
			once.Do(func() {
				create := singletonFunc()
				value = &create
			})
		}

		return value
	}
}
