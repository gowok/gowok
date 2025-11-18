package gowok

import (
	"github.com/gowok/gowok/runtime"
)

type _hooks struct {
	*runtime.Hooks
}

var Hooks = &_hooks{&runtime.Hooks{}}
