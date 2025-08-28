package web

import (
	"net/http"

	"github.com/gowok/gowok/config"
	"github.com/gowok/gowok/errors"
	"github.com/gowok/gowok/some"
	"github.com/ngamux/middleware/cors"
	"github.com/ngamux/middleware/log"
	"github.com/ngamux/middleware/pprof"
	"github.com/ngamux/ngamux"
)

var mux = some.Empty[*httpMux]()

func Router() *ngamux.HttpServeMux {
	return mux.OrPanic(errors.ErrNotConfigured("web")).mux
}

func Server() *http.Server {
	return mux.OrPanic(errors.ErrNotConfigured("web")).Server
}

type httpMux struct {
	mux    *ngamux.HttpServeMux
	Server *http.Server
}

func Configure(c *config.Web) {
	if !c.Enabled {
		return
	}

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

	mux = some.Of(server)
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
