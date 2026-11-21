package gowok

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gowok/fp/some"
	"github.com/gowok/gowok/config"
	"github.com/gowok/gowok/web"
	"github.com/gowok/gowok/web/request"
	"github.com/gowok/gowok/web/response"
	"github.com/ngamux/middleware/auth"
	"github.com/ngamux/middleware/compress"
	"github.com/ngamux/middleware/cors"
	middlewareLog "github.com/ngamux/middleware/log"
	"github.com/ngamux/middleware/pprof"
	"github.com/ngamux/middleware/ratelimit"
	"github.com/ngamux/middleware/requestid"
	"github.com/ngamux/middleware/timeout"
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

	c.RequestID.IfPresent(func(rid config.WebRequestID) {
		if !rid.Enabled {
			return
		}

		if rid.Header == "" {
			mux.Use(requestid.New())
			return
		}

		header := rid.Header
		mux.Use(requestid.New(requestid.WithKeyHeader(func() string {
			return header
		})))
	})
	c.Compress.IfPresent(func(cc config.WebCompress) {
		if cc.Enabled {
			mux.Use(compress.New(compress.Config{
				Level:        cc.Level,
				BrotliLevel:  cc.BrotliLevel,
				ZstdLevel:    cc.ZstdLevel,
				ContentTypes: cc.ContentTypes,
			}))
		}
	})
	c.Timeout.IfPresent(func(t config.WebTimeout) {
		if t.Enabled {
			mux.Use(timeout.New(timeout.Config{
				Timeout: t.GetDuration(),
			}))
		}
	})
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
	c.RateLimit.IfPresent(func(rl config.WebRateLimit) {
		if rl.Enabled {
			mux.Use(rateLimitMiddleware(rl))
		}
	})

	c.AuthBearer.IfPresent(func(ab config.WebAuthBearer) {
		if ab.Enabled {
			mux.Use(authBearerMiddleware(ab))
		}
	})
	c.AuthBasic.IfPresent(func(ab config.WebAuthBasic) {
		if ab.Enabled {
			mux.Use(authBasicMiddleware(ab))
		}
	})

	Web.Server = server
}

func rateLimitMiddleware(rl config.WebRateLimit) ngamux.MiddlewareFunc {
	limited := ratelimit.New(ratelimit.Config{
		Requests: rl.Requests,
		Window:   rl.GetWindow(),
		Burst:    rl.Burst,
	})

	if len(rl.Exclude) == 0 {
		return limited
	}

	excluded := make(map[string]struct{}, len(rl.Exclude))
	for _, p := range rl.Exclude {
		excluded[p] = struct{}{}
	}

	return func(next http.HandlerFunc) http.HandlerFunc {
		handler := limited(next)
		return func(rw http.ResponseWriter, r *http.Request) {
			if _, ok := excluded[r.URL.Path]; ok {
				next(rw, r)
				return
			}

			handler(rw, r)
		}
	}
}

func authBearerMiddleware(c config.WebAuthBearer) ngamux.MiddlewareFunc {
	bearer := auth.Bearer(auth.BearerConfig{
		Header:     c.Header,
		ContextKey: c.ContextKey,
		Authorizer: func(token string) (any, error) {
			ab, ok := Config.Web.AuthBearer.Get()
			if !ok || ab.Authorizer == nil {
				return nil, errors.New("auth bearer: authorizer is not configured")
			}

			return ab.Authorizer(token)
		},
		ErrorHandler: func(rw http.ResponseWriter, err error) {
			ab, ok := Config.Web.AuthBearer.Get()
			if ok && ab.ErrorHandler != nil {
				ab.ErrorHandler(rw, err)
				return
			}

			if c.Realm != "" {
				rw.Header().Set("WWW-Authenticate", "Bearer realm="+strconv.Quote(c.Realm))
			}
			ngamux.Res(rw).Status(http.StatusUnauthorized).JSON(ngamux.Map{
				"error": "unauthorized",
			})
		},
	})

	if len(c.Exclude) == 0 {
		return bearer
	}

	excluded := make(map[string]struct{}, len(c.Exclude))
	for _, p := range c.Exclude {
		excluded[p] = struct{}{}
	}

	return func(next http.HandlerFunc) http.HandlerFunc {
		handler := bearer(next)
		return func(rw http.ResponseWriter, r *http.Request) {
			if _, ok := excluded[r.URL.Path]; ok {
				next(rw, r)
				return
			}

			handler(rw, r)
		}
	}
}

func authBasicMiddleware(c config.WebAuthBasic) ngamux.MiddlewareFunc {
	basic := auth.Basic(auth.BasicConfig{
		Realm: c.Realm,
		Authorizer: func(username string, password string) bool {
			ab, ok := Config.Web.AuthBasic.Get()
			if !ok || ab.Authorizer == nil {
				return false
			}

			return ab.Authorizer(username, password)
		},
		ErrorHandler: func(rw http.ResponseWriter, err error) {
			ab, ok := Config.Web.AuthBasic.Get()
			if ok && ab.ErrorHandler != nil {
				ab.ErrorHandler(rw, err)
				return
			}

			if c.Realm != "" {
				rw.Header().Set("WWW-Authenticate", "Basic realm="+strconv.Quote(c.Realm))
			}
			ngamux.Res(rw).Status(http.StatusUnauthorized).JSON(ngamux.Map{
				"error": "unauthorized",
			})
		},
	})

	if len(c.Exclude) == 0 {
		return basic
	}

	excluded := make(map[string]struct{}, len(c.Exclude))
	for _, p := range c.Exclude {
		excluded[p] = struct{}{}
	}

	return func(next http.HandlerFunc) http.HandlerFunc {
		handler := basic(next)
		return func(rw http.ResponseWriter, r *http.Request) {
			if _, ok := excluded[r.URL.Path]; ok {
				next(rw, r)
				return
			}

			handler(rw, r)
		}
	}
}

func (p *_web) run() {
	if ab, ok := Config.Web.AuthBearer.Get(); ok && ab.Enabled && ab.Authorizer == nil {
		slog.Warn("web: auth_bearer enabled without an Authorizer; all requests will be rejected")
	}
	if ab, ok := Config.Web.AuthBasic.Get(); ok && ab.Enabled && ab.Authorizer == nil {
		slog.Warn("web: auth_basic enabled without an Authorizer; all requests will be rejected")
	}

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

type WebConfigureAuthBearerInput struct {
	Authorizer   func(token string) (any, error)
	ErrorHandler func(rw http.ResponseWriter, err error)
}

func WebConfigureAuthBearer(input WebConfigureAuthBearerInput) ConfigureFunc {
	return func(context.Context) {
		ab, ok := Config.Web.AuthBearer.Get()
		if !ok {
			slog.Warn("web: auth_bearer is not configured; authorizer not applied")
			return
		}

		ab.Authorizer = input.Authorizer
		ab.ErrorHandler = input.ErrorHandler
		Config.Web.AuthBearer = some.Of(ab)
	}
}

type WebConfigureAuthBasicInput struct {
	Authorizer   func(username string, password string) bool
	ErrorHandler func(rw http.ResponseWriter, err error)
}

func WebConfigureAuthBasic(input WebConfigureAuthBasicInput) ConfigureFunc {
	return func(context.Context) {
		ab, ok := Config.Web.AuthBasic.Get()
		if !ok {
			slog.Warn("web: auth_basic is not configured; authorizer not applied")
			return
		}

		ab.Authorizer = input.Authorizer
		ab.ErrorHandler = input.ErrorHandler
		Config.Web.AuthBasic = some.Of(ab)
	}
}
