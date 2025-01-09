package gowok

import "github.com/gowok/gowok/singleton"

// Singleton creates new object and gets if created.
// Deprecated: Use singleton.New instead.
func Singleton[T any](singletonFunc singleton.SingletonFunc[T]) func(...T) *T {
	return singleton.New(singletonFunc)
}
