package some

import (
	"bytes"
	"reflect"

	"github.com/gowok/gowok/errors"
	"github.com/gowok/gowok/json"
)

type Some[T any] struct {
	value     *T
	isPresent bool
}

func newOptional[T any](val *T) Some[T] {
	isPresent := val != nil
	return Some[T]{val, isPresent}
}

func Empty[T any]() Some[T] {
	return newOptional[T](nil)
}

func Of[T any](val T) Some[T] {
	v := reflect.ValueOf(val)
	k := v.Kind()

	if k == reflect.Invalid {
		return Empty[T]()
	}

	if (k == reflect.Func || k == reflect.Pointer) && (v.IsZero() || v.IsNil()) {
		return Empty[T]()
	}

	return newOptional(&val)
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

func (o Some[T]) IfPresent(callback func(T), elseCallback ...func()) {
	if val, ok := o.Get(); ok {
		callback(val)
		return
	}

	if len(elseCallback) > 0 {
		elseCallback[0]()
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

func (o Some[T]) OrPanic(errs ...error) T {
	err := errors.ErrNoValuePresent
	if len(errs) > 0 {
		err = errors.New(errs[0].Error())
	}

	if !o.IsPresent() {
		panic(err)
	}

	return *o.value
}

func (o *Some[T]) UnmarshalJSON(b []byte) error {
	if bytes.Equal(b, []byte("{}")) {
		return nil
	}

	var v *T
	err := json.Unmarshal(b, &v)
	if err != nil {
		return err
	}

	if v != nil {
		o.value = v
		o.isPresent = true
	}
	return nil
}

func (o Some[T]) MarshalJSON() ([]byte, error) {
	if o.IsPresent() {
		return json.Marshal(o.value)
	}

	return json.Marshal(map[string]any{})
}
