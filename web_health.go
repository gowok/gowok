package gowok

import (
	"net/http"

	"github.com/ngamux/ngamux"
)

type _health struct {
	healths map[string]func() any
}

var Health = _health{
	healths: map[string]func() any{},
}

func (p *_health) Configure() {
	r := Web.Group("/health")
	r.Get("", func(w http.ResponseWriter, r *http.Request) {
		urls := make(map[string]string, len(p.healths))
		for k := range p.healths {
			urls[k] = "/health/" + k
		}
		ngamux.Res(w).JSON(urls)
	})

	r.Get("/{name}", func(w http.ResponseWriter, r *http.Request) {
		name := r.PathValue("name")
		if name == "" {
			http.NotFound(w, r)
			return
		}

		healthFunc, ok := p.healths[name]
		if !ok {
			http.NotFound(w, r)
			return
		}

		ngamux.Res(w).JSON(healthFunc())
	})
}

func (p *_health) Add(name string, healthFunc func() any) {
	switch {
	case name == "":
		return
	case healthFunc == nil:
		return
	}

	p.healths[name] = healthFunc
}
