package gowok

type SingletonFunc[T any] func() T

func Singleton[T any](singletonFunc SingletonFunc[T]) func(...T) *T {
	var value *T

	return func(newValue ...T) *T {
		if len(newValue) > 0 {
			value = &newValue[0]
			return value
		}

		if value == nil {
			create := singletonFunc()
			value = &create
		}

		return value
	}
}
