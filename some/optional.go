package some

import (
	"reflect"

	"gopkg.in/yaml.v3"
)

type Some[T any] struct {
	value     *T
	isPresent bool
}

func newOptional[T any](val *T) Some[T] {
	isPresent := false
	if val != nil {
		isPresent = true
	}
	return Some[T]{val, isPresent}
}

func Empty[T any]() Some[T] {
	return newOptional[T](nil)
}

func Of[T any](val *T) Some[T] {
	if val == nil {
		return Empty[T]()
	}

	if reflect.ValueOf(*val).IsNil() {
		return Empty[T]()
	}

	return newOptional(val)
}

func (o Some[T]) Get() (T, bool) {
	if !o.IsPresent() {
		var value T
		return value, false
	}
	return *o.value, true
}

func (o Some[T]) IsPresent() bool {
	return o.isPresent
}

func (o Some[T]) IfPresent(callback func(T)) {
	if val, ok := o.Get(); ok {
		callback(val)
	}
}

func (o Some[T]) OrElse(val T) T {
	if !o.IsPresent() {
		return val
	}

	return *o.value
}

func (o Some[T]) OrElseFunc(gen func() T) T {
	if !o.IsPresent() && gen != nil {
		return gen()
	}

	return *o.value
}

func (o Some[T]) OrPanic(err error) T {
	if !o.IsPresent() {
		panic(err)
	}

	return *o.value
}

func (o *Some[T]) UnmarshalYAML(value *yaml.Node) error {
	var v T
	if err := value.Decode(&v); err != nil {
		return err
	}
	o.value = &v
	o.isPresent = true
	return nil
}
