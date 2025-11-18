package gowok

import (
	"net/http"

	"github.com/gowok/gowok/config"
	"github.com/gowok/gowok/errors"
	"github.com/gowok/gowok/web"
	"github.com/ngamux/middleware/cors"
	"github.com/ngamux/middleware/log"
	"github.com/ngamux/middleware/pprof"
	"github.com/ngamux/ngamux"
)

type _web struct {
	*ngamux.HttpServeMux
	Server  *http.Server
	Handler *_webHandler
}

type _webHandler struct {
}

var Web = &_web{
	HttpServeMux: ngamux.NewHttpServeMux(),
	Server:       &http.Server{},
	Handler:      &_webHandler{},
}

func (w *_webHandler) Handler(handler func(ctx *web.Ctx) error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := web.NewCtx(r.Context(), w, r)
		err := handler(ctx)
		if err != nil {
			switch e := err.(type) {
			case errors.Error:
				ctx.Res().JSON(e)
			default:
				ctx.Res().InternalServerError(err)
			}
		}
	}
}

func (w *_webHandler) SSE(handler func(ctx *web.CtxSse)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		ctx, err := web.NewCtxSse(web.NewCtx(r.Context(), w, r))
		if err != nil {
			ctx.Res().InternalServerError(err)
			return
		}

		handler(ctx)
	}
}

func (p *_web) Configure(c *config.Web) {
	mux := Web.HttpServeMux
	server := &http.Server{
		Addr:    c.Host,
		Handler: Web.HttpServeMux,
	}

	func() {
		sc := c.GetStatic()
		if !sc.Enabled {
			return
		}

		fs := http.FileServer(http.Dir(sc.Dir))
		mux.HandleFunc(http.MethodGet, sc.Prefix, func(rw http.ResponseWriter, r *http.Request) {
			http.StripPrefix(sc.Prefix, fs).ServeHTTP(rw, r)
		})
	}()

	c.Log.IfPresent(func(ll config.WebLog) {
		if ll.Enabled {
			mux.Use(log.New())
		}
	})
	c.Cors.IfPresent(func(ll config.WebCors) {
		if ll.Enabled {
			mux.Use(cors.New(c.GetCors()))
		}
	})
	c.Pprof.IfPresent(func(ll config.WebPprof) {
		if ll.Enabled {
			mux.Use(pprof.New(c.GetPprof()))
		}
	})

	Web.Server = server
}

func (w *_web) Resource(path string, resource web.ResourceHandler, opts ...func(*ngamux.HttpServeMux)) {
	g := w.Group(path)
	for _, opt := range opts {
		opt(g)
	}
	g.Get("", resource.Index)
	g.Post("", resource.Store)
	g.Get("/{id}", resource.Show)
	g.Put("/{id}", resource.Update)
	g.Delete("/{id}", resource.Destroy)
}

func (w _web) WithResourceMiddleware(middlewares ...ngamux.MiddlewareFunc) func(*ngamux.HttpServeMux) {
	return func(mux *ngamux.HttpServeMux) {
		mux.Use(middlewares...)
	}
}
