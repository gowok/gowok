package gowok

import (
	"reflect"

	"github.com/ngamux/ngamux"
)

type Controller interface {
	Route(mux *ngamux.Ngamux)
	SetConfig(config *Config)
	SetModels(models Models)
}

type Controllers map[string]Controller

func (cs *Controllers) Add(c Controller) {
	t := reflect.TypeOf(c)
	(*cs)[t.Name()] = c
}

func (cs Controllers) Get(c Controller) (controller Controller, ok bool) {
	t := reflect.TypeOf(c)
	controller, ok = cs[t.Name()]
	return
}
