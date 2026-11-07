package gowok

import (
	"github.com/gowok/gowok/async"
)

type _async struct{}

var Async = &_async{}

func (p *_async) All(tasks ...func() (any, error)) ([]any, error) {
	return async.All(tasks...)
}

func (p *_async) AllResults(tasks ...func() (any, error)) ([]any, error) {
	return async.AllResults(tasks...)
}

func (p *_async) AllLimit(limit int, tasks ...func() (any, error)) ([]any, error) {
	return async.AllLimit(limit, tasks...)
}

func (p *_async) AllLimitResults(limit int, tasks ...func() (any, error)) ([]any, error) {
	return async.AllLimitResults(limit, tasks...)
}
