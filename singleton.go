package gowok

type SingletonFunc[T any] func() T

func Singleton[T any](singletonFunc SingletonFunc[T]) func() *T {
	var value *T

	return func() *T {
		if value == nil {
			create := singletonFunc()
			value = &create
		}

		return value
	}
}
