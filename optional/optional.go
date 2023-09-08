package optional

import (
	"reflect"

	"github.com/gowok/gowok/exception"
)

type Optional[T any] struct {
	value *T
}

func New[T any](val *T) Optional[T] {
	return Optional[T]{val}
}

func Empty[T any]() Optional[T] {
	var val T
	return New(&val)
}

func Of[T any](val *T) (Optional[T], error) {
	if val == nil {
		return Empty[T](), exception.ErrNilPointerDeref
	}
	return New(val), nil
}

func (o Optional[T]) Get() (T, error) {
	if !o.IsPresent() {
		return *Empty[T]().value, exception.ErrGetOfNoValue
	}
	return *o.value, nil
}

func (o Optional[T]) IsPresent() bool {
	if o.value != nil {
		vOf := reflect.ValueOf(*o.value)
		if vOf.Kind() == reflect.Ptr && vOf.IsNil() {
			return false
		}
		return true
	}
	return false
}

func (o Optional[T]) OrElse(val T) T {
	if !o.IsPresent() {
		return val
	}

	return *o.value
}

func (o Optional[T]) OrPanic(err error) T {
	if !o.IsPresent() {
		panic(err)
	}

	return *o.value
}
