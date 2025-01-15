package router

import (
	"net/http"

	"github.com/gowok/gowok/config"
	"github.com/ngamux/middleware/cors"
	"github.com/ngamux/middleware/log"
	"github.com/ngamux/middleware/pprof"
	"github.com/ngamux/ngamux"
)

var mux = &httpMux{}

func Router() *ngamux.HttpServeMux {
	return mux.mux
}

func Server() *http.Server {
	return mux.Server
}

type httpMux struct {
	mux    *ngamux.HttpServeMux
	Server *http.Server
}

func Configure(c *config.Web) {
	// conf := ngamux.Config{
	// 	ProxyHeader:           fiber.HeaderXForwardedFor,
	// }

	mm := ngamux.NewHttpServeMux()
	server := &httpMux{
		Server: &http.Server{
			Addr:    c.Host,
			Handler: mm,
		},
		mux: mm,
	}
	// configureHttpViews(server, c)
	configureHttpStatic(server, c)

	c.Log.IfPresent(func(ll config.WebLog) {
		if ll.Enabled {
			server.mux.Use(log.New())
		}
	})
	c.Cors.IfPresent(func(ll config.WebCors) {
		if ll.Enabled {
			server.mux.Use(cors.New(c.GetCors()))
		}
	})
	c.Pprof.IfPresent(func(ll config.WebPprof) {
		if ll.Enabled {
			server.mux.Use(pprof.New(c.GetPprof()))
		}
	})

	mux = server
}

func configureHttpStatic(server *httpMux, c *config.Web) {
	sc := c.GetStatic()
	if !sc.Enabled {
		return
	}

	fs := http.FileServer(http.Dir(sc.Dir))
	server.mux.HandleFunc(http.MethodGet, sc.Prefix, func(rw http.ResponseWriter, r *http.Request) {
		http.StripPrefix(sc.Prefix, fs).ServeHTTP(rw, r)
	})
}

// func configureHttpViews(server *HttpMux, c *config.Web) {
// 	vc := c.GetViews()
// 	if !vc.Enabled {
// 		return
// 	}
//
// 	// TODO: make support global view and rendering function
// }

func Group(path string) *ngamux.HttpServeMux {
	return mux.mux.Group(path)
}

func HandleFunc(method, path string, handlerFunc http.HandlerFunc, middleware ...ngamux.MiddlewareFunc) {
	mux.mux.HandleFunc(method, path, ngamux.WithMiddlewares(middleware...)(handlerFunc))
}

func Get(path string, handlerFunc http.HandlerFunc, middleware ...ngamux.MiddlewareFunc) {
	mux.mux.Get(path, ngamux.WithMiddlewares(middleware...)(handlerFunc))
}

func Post(path string, handlerFunc http.HandlerFunc, middleware ...ngamux.MiddlewareFunc) {
	mux.mux.Post(path, ngamux.WithMiddlewares(middleware...)(handlerFunc))
}

func Patch(path string, handlerFunc http.HandlerFunc, middleware ...ngamux.MiddlewareFunc) {
	mux.mux.Patch(path, ngamux.WithMiddlewares(middleware...)(handlerFunc))
}

func Put(path string, handlerFunc http.HandlerFunc, middleware ...ngamux.MiddlewareFunc) {
	mux.mux.Put(path, ngamux.WithMiddlewares(middleware...)(handlerFunc))
}

func Delete(path string, handlerFunc http.HandlerFunc, middleware ...ngamux.MiddlewareFunc) {
	mux.mux.Delete(path, ngamux.WithMiddlewares(middleware...)(handlerFunc))
}

func Use(middleware ...ngamux.MiddlewareFunc) {
	mux.mux.Use(middleware...)
}

func Annotate(annotators ...ngamux.Annotator) *ngamux.Annotation {
	return mux.mux.Annotate(annotators...)
}
