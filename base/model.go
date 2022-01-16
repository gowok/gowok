package base

import "reflect"

type Model interface{}
type Models map[string]Model

func (ms *Models) Add(m Model) {
	t := reflect.TypeOf(m)
	(*ms)[t.Name()] = m
}

func (ms Models) Get(m Model) (model Model, ok bool) {
	t := reflect.TypeOf(m)
	model, ok = ms[t.Name()]
	return
}
