package optional

type Optional[T any] struct {
	value     *T
	isPresent bool
}

func newOptional[T any](val *T) Optional[T] {
	isPresent := false
	if val != nil {
		isPresent = true
	}
	return Optional[T]{val, isPresent}
}

func Empty[T any]() Optional[T] {
	return newOptional[T](nil)
}

func Of[T any](val *T) Optional[T] {
	if val == nil {
		return Empty[T]()
	}
	return newOptional(val)
}

func (o Optional[T]) Get() (T, bool) {
	if !o.IsPresent() {
		var value T
		return value, false
	}
	return *o.value, true
}

func (o Optional[T]) IsPresent() bool {
	return o.isPresent
}

func (o Optional[T]) OrElse(val T) T {
	if !o.IsPresent() {
		return val
	}

	return *o.value
}

func (o Optional[T]) OrElseFunc(gen func() T) T {
	if !o.IsPresent() && gen != nil {
		return gen()
	}

	return *o.value
}

func (o Optional[T]) OrPanic(err error) T {
	if !o.IsPresent() {
		panic(err)
	}

	return *o.value
}
