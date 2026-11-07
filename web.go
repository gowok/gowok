package gowok

import (
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/gowok/gowok/config"
	"github.com/gowok/gowok/web"
	"github.com/gowok/gowok/web/request"
	"github.com/gowok/gowok/web/response"
	"github.com/ngamux/middleware/cors"
	middlewareLog "github.com/ngamux/middleware/log"
	"github.com/ngamux/middleware/pprof"
	"github.com/ngamux/ngamux"
)

var (
	webListenAndServe = func(s *http.Server) error {
		return s.ListenAndServe()
	}
)

type _web struct {
	*ngamux.HttpServeMux
	Server   *http.Server
	Handler  *_webHandler
	Response *_webResponse
	Request  *_webRequest
	Resource *_webResource
}

type _webHandler struct {
}

type _webResponse struct {
}

type _webRequest struct {
}

type _webResource struct {
}

var Web = &_web{
	HttpServeMux: ngamux.NewHttpServeMux(),
	Server:       &http.Server{},
	Handler:      &_webHandler{},
	Response:     &_webResponse{},
	Request:      &_webRequest{},
	Resource:     &_webResource{},
}

func (w *_webHandler) Handler(handler func(ctx *web.Ctx) error) http.HandlerFunc {
	return web.Handler(handler)
}

func (w *_webHandler) SSE(handler func(ctx *web.CtxSse)) http.HandlerFunc {
	return web.HandlerSSE(handler)
}

func (p *_web) configure() {
	c := Config.Web

	mux := Web.HttpServeMux
	server := &http.Server{
		Addr:              c.Host,
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	func() {
		sc := c.GetStatic()
		if !sc.Enabled {
			return
		}

		prefix := strings.TrimSuffix(sc.Prefix, "/")
		fs := http.FileServer(http.Dir(sc.Dir))
		mux.HandleFunc(http.MethodGet, prefix+"/", func(rw http.ResponseWriter, r *http.Request) {
			http.StripPrefix(prefix, fs).ServeHTTP(rw, r)
		})
	}()

	c.Log.IfPresent(func(ll config.WebLog) {
		if ll.Enabled {
			mux.Use(middlewareLog.New())
		}
	})
	c.Cors.IfPresent(func(ll config.WebCors) {
		if ll.Enabled {
			mux.Use(cors.New(c.GetCors()))
		}
	})
	c.Pprof.IfPresent(func(ll config.WebPprof) {
		if ll.Enabled {
			mux.Use(pprof.New())
		}
	})

	Web.Server = server
}

func (p *_web) run() {
	slog.Info("starting web", "host", Config.Web.Host)
	err := webListenAndServe(Web.Server)
	if err != nil {
		if errors.Is(err, http.ErrServerClosed) {
			return
		}
		slog.Error("web: failed to start", "error", err)
	}
}

func (p *_webResource) New(path string, resource web.ResourceHandler, opts ...func(*ngamux.HttpServeMux)) {
	g := Web.Group(path)
	for _, opt := range opts {
		opt(g)
	}
	g.Get("", resource.Index)
	g.Post("", resource.Store)
	g.Get("/{id}", resource.Show)
	g.Put("/{id}", resource.Update)
	g.Delete("/{id}", resource.Destroy)
}

func (p _webResource) WithMiddleware(middlewares ...ngamux.MiddlewareFunc) func(*ngamux.HttpServeMux) {
	return func(mux *ngamux.HttpServeMux) {
		mux.Use(middlewares...)
	}
}

func (p *_webResponse) New(w http.ResponseWriter) *response.Response {
	return response.New(w)
}

func (p *_webRequest) New(r *http.Request) *request.Request {
	return request.New(r)
}
