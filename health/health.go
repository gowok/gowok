package health

import (
	"net/http"

	"github.com/gowok/gowok/web"
	"github.com/ngamux/ngamux"
)

var healths = map[string]func() any{}

func Configure() {
	r := web.Router().Group("/health")
	r.Get("", func(w http.ResponseWriter, r *http.Request) {
		urls := make(map[string]string, len(healths))
		for k := range healths {
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

		healthFunc, ok := healths[name]
		if !ok {
			http.NotFound(w, r)
			return
		}

		ngamux.Res(w).JSON(healthFunc())
	})
}

func Add(name string, healthFunc func() any) {
	switch {
	case name == "":
		return
	case healthFunc == nil:
		return
	}

	healths[name] = healthFunc
}
