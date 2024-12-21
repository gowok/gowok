package gowok

import (
	"net/http"

	"github.com/gowok/gowok/config"
	"github.com/ngamux/middleware/cors"
	"github.com/ngamux/middleware/log"
	"github.com/ngamux/middleware/pprof"
	"github.com/ngamux/ngamux"
)

type HttpMux struct {
	Mux    *ngamux.HttpServeMux
	Server *http.Server
}

func NewHTTP(c *config.Web) *HttpMux {
	// conf := ngamux.Config{
	// 	ProxyHeader:           fiber.HeaderXForwardedFor,
	// }

	mux := ngamux.NewHttpServeMux()
	server := &HttpMux{
		Server: &http.Server{
			Addr:    c.Host,
			Handler: mux,
		},
		Mux: mux,
	}
	// configureHttpViews(server, c)
	configureHttpStatic(server, c)

	if c.Log != nil && c.Log.Enabled {
		server.Mux.Use(log.New(c.GetLog()))
	}
	if c.Cors != nil && c.Cors.Enabled {
		server.Mux.Use(cors.New(c.GetCors()))
	}
	if c.Pprof != nil && c.Pprof.Enabled {
		server.Mux.Use(pprof.New(c.GetPprof()))
	}

	// if c.Pprof != nil && c.Pprof.Enabled {
	// 	h.Use(pprof.New(c.GetPprof()))
	// }

	return server
}

func configureHttpStatic(server *HttpMux, c *config.Web) {
	sc := c.GetStatic()
	if !sc.Enabled {
		return
	}

	fs := http.FileServer(http.Dir(sc.Dir))
	server.Mux.HandleFunc(http.MethodGet, sc.Prefix, func(rw http.ResponseWriter, r *http.Request) {
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

func HttpBadRequest(rw http.ResponseWriter, body any) {
	res := ngamux.Res(rw).Status(http.StatusBadRequest)
	switch b := body.(type) {
	case string:
		res.Text(b)
	case error:
		res.Text(b.Error())
	default:
		res.JSON(b)
	}
}
func HttpUnauthorized(rw http.ResponseWriter) {
	ngamux.Res(rw).Status(http.StatusUnauthorized).Text("unauthorized")
}
func HttpNotFound(rw http.ResponseWriter) {
	ngamux.Res(rw).Status(http.StatusUnauthorized).Text("not found")
}
func HttpOk(rw http.ResponseWriter, body any) {
	res := ngamux.Res(rw).Status(http.StatusOK)
	switch b := body.(type) {
	case string:
		res.Text(b)
	default:
		res.JSON(b)
	}
}

func HttpInternalServerError(rw http.ResponseWriter, body any) {
	res := ngamux.Res(rw).Status(http.StatusInternalServerError)
	switch b := body.(type) {
	case string:
		res.Text(b)
	case error:
		res.Text(b.Error())
	default:
		res.JSON(b)
	}
}

func HttpCreated(rw http.ResponseWriter, body any) {
	res := ngamux.Res(rw).Status(http.StatusCreated)
	switch b := body.(type) {
	case string:
		res.Text(b)
	default:
		res.JSON(b)
	}
}

func HttpForbidden(rw http.ResponseWriter) {
	ngamux.Res(rw).Status(http.StatusForbidden).Text("forbidden")
}

func HttpConflict(rw http.ResponseWriter, body any) {
	res := ngamux.Res(rw).Status(http.StatusConflict)
	switch b := body.(type) {
	case string:
		res.Text(b)
	default:
		res.JSON(b)
	}
}

func HttpNoContent(rw http.ResponseWriter) {
	ngamux.Res(rw).Status(http.StatusNoContent).Text("")
}

func HttpAccepted(rw http.ResponseWriter, body any) {
	res := ngamux.Res(rw).Status(http.StatusAccepted)
	switch b := body.(type) {
	case string:
		res.Text(b)
	default:
		res.JSON(b)
	}
}
