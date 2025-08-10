package router

import (
	"net/http"

	"github.com/ngamux/ngamux"
)

type ResourceHandler interface {
	Index(http.ResponseWriter, *http.Request)
	Show(http.ResponseWriter, *http.Request)
	Store(http.ResponseWriter, *http.Request)
	Update(http.ResponseWriter, *http.Request)
	Destroy(http.ResponseWriter, *http.Request)
}

func Resource(path string, resource ResourceHandler, opts ...func(*ngamux.HttpServeMux)) {
	g := Router().Group(path)
	for _, opt := range opts {
		opt(g)
	}
	g.Get("", resource.Index)
	g.Post("", resource.Store)
	g.Get("/{id}", resource.Show)
	g.Put("/{id}", resource.Update)
	g.Delete("/{id}", resource.Destroy)
}

func WithResourceMiddleware(middlewares ...ngamux.MiddlewareFunc) func(*ngamux.HttpServeMux) {
	return func(mux *ngamux.HttpServeMux) {
		mux.Use(middlewares...)
	}
}
