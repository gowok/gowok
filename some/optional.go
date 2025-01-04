package some

import (
	"errors"
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

	v := reflect.ValueOf(*val)
	k := v.Kind()

	if k == reflect.Invalid {
		return Empty[T]()
	}

	if (k == reflect.Func) && (v.IsZero() || v.IsNil()) {
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

var orPanicErr = errors.New("some: no value present")

func (o Some[T]) OrPanic(errs ...error) T {
	err := orPanicErr
	if len(errs) > 0 {
		err = errs[0]
	}

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
