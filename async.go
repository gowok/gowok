package gowok

import (
	"github.com/gowok/gowok/async"
)

type _async struct{}

var Async = &_async{}

func (p *_async) All(tasks ...func() (any, error)) ([]any, error) {
	return async.All(tasks...)
}
