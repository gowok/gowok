package gowok

import (
	"net/http"

	"github.com/gowok/gowok/config"
	"github.com/ngamux/middleware/cors"
	"github.com/ngamux/middleware/log"
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

	server.Mux.Use(log.New(c.GetLog()))
	server.Mux.Use(cors.New(c.GetCors()))

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
	server.Mux.HandlerFunc(http.MethodGet, sc.Prefix, func(rw http.ResponseWriter, r *http.Request) error {
		http.StripPrefix(sc.Prefix, fs).ServeHTTP(rw, r)
		return nil
	})
}

func configureHttpViews(server *HttpMux, c *config.Web) {
	vc := c.GetViews()
	if !vc.Enabled {
		return
	}

	// TODO: make support global view and rendering function
}

func HttpBadRequest(rw http.ResponseWriter, body any) {
	res := ngamux.Res(rw).Status(http.StatusBadRequest)
	switch body.(type) {
	case string:
		res.Text(body.(string))
	case error:
		res.Text(body.(error).Error())
	default:
		res.JSON(body)
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
	switch body.(type) {
	case string:
		res.Text(body.(string))
	default:
		res.JSON(body)
	}
}
