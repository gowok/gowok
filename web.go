package gowok

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gowok/fp/some"
	"github.com/gowok/gowok/config"
	"github.com/gowok/gowok/web"
	"github.com/gowok/gowok/web/request"
	"github.com/gowok/gowok/web/response"
	"github.com/ngamux/middleware/auth"
	"github.com/ngamux/middleware/bodylimit"
	"github.com/ngamux/middleware/compress"
	"github.com/ngamux/middleware/cors"
	"github.com/ngamux/middleware/idempotency"
	middlewareLog "github.com/ngamux/middleware/log"
	"github.com/ngamux/middleware/pprof"
	"github.com/ngamux/middleware/ratelimit"
	"github.com/ngamux/middleware/realip"
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
	c.RealIP.IfPresent(func(ri config.WebRealIP) {
		if ri.Enabled {
			mux.Use(realip.New(realip.Config{
				TrustedProxies: ri.TrustedProxies,
			}))
		}
	})
	c.BodyLimit.IfPresent(func(bl config.WebBodyLimit) {
		if bl.Enabled {
			mux.Use(bodyLimitMiddleware(bl))
		}
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

	c.Idempotency.IfPresent(func(idem config.WebIdempotency) {
		if idem.Enabled {
			mux.Use(idempotencyMiddleware())
		}
	})

	Web.Server = server
}

func rateLimitMiddleware(rl config.WebRateLimit) ngamux.MiddlewareFunc {
	limited := ratelimit.New(ratelimit.Config{
		Requests: rl.Requests,
		Window:   rl.GetWindow(),
		Burst:    rl.Burst,
		KeyFunc:  rateLimitKeyFunc(rl),
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

func rateLimitKeyFunc(rl config.WebRateLimit) func(*http.Request) string {
	key := strings.TrimSpace(rl.Key)

	switch {
	case key == "" || strings.EqualFold(key, "client_ip"):
		return nil
	case strings.EqualFold(key, "xff"):
		return ratelimit.KeyFunc_ByClientIPFromXFF(rl.TrustedProxies...)
	case strings.HasPrefix(strings.ToLower(key), "header:"):
		header := strings.TrimSpace(key[len("header:"):])
		if header == "" {
			return nil
		}

		return func(r *http.Request) string {
			if value := strings.TrimSpace(r.Header.Get(header)); value != "" {
				return value
			}

			return ratelimit.KeyFunc_ByClientIP(r)
		}
	default:
		return nil
	}
}

func bodyLimitMiddleware(bl config.WebBodyLimit) ngamux.MiddlewareFunc {
	return bodylimit.New(bodylimit.Config{
		Limit: bl.Limit,
		ErrorHandler: func(rw http.ResponseWriter, r *http.Request, err error) {
			cfg, ok := Config.Web.BodyLimit.Get()
			if ok && cfg.ErrorHandler != nil {
				cfg.ErrorHandler(rw, r, err)
				return
			}

			ngamux.Res(rw).Status(http.StatusRequestEntityTooLarge).JSON(ngamux.Map{
				"error": err.Error(),
			})
		},
	})
}

func idempotencyMiddleware() ngamux.MiddlewareFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		var (
			once    sync.Once
			handler http.HandlerFunc
		)

		return func(rw http.ResponseWriter, r *http.Request) {
			once.Do(func() {
				idem, _ := Config.Web.Idempotency.Get()
				handler = idempotency.New(idempotency.Config{
					Header:       idem.Header,
					TTL:          idem.GetTTL(),
					Methods:      idem.Methods,
					Required:     idem.Required,
					Store:        idem.Store,
					ErrorHandler: idem.ErrorHandler,
				})(next)
			})

			handler(rw, r)
		}
	}
}

func authBearerMiddleware(c config.WebAuthBearer) ngamux.MiddlewareFunc {
	var tokensAuthorizer func(token string) (any, error)
	if len(c.Tokens) > 0 {
		tokensAuthorizer = auth.AuthorizerStatic(c.Tokens...)
	}

	var jwtAuthorizer func(token string) (any, error)
	if c.JWTSecret != "" {
		jwtAuthorizer = auth.AuthorizerJWT([]byte(c.JWTSecret))
	}

	bearer := auth.Bearer(auth.BearerConfig{
		Header:     c.Header,
		ContextKey: c.ContextKey,
		Authorizer: func(token string) (any, error) {
			ab, ok := Config.Web.AuthBearer.Get()
			if !ok {
				return nil, errors.New("auth bearer: authorizer is not configured")
			}

			if ab.Authorizer != nil {
				return ab.Authorizer(token)
			}
			if tokensAuthorizer != nil {
				return tokensAuthorizer(token)
			}
			if jwtAuthorizer != nil {
				return jwtAuthorizer(token)
			}

			return nil, errors.New("auth bearer: authorizer is not configured")
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
	if ab, ok := Config.Web.AuthBearer.Get(); ok && ab.Enabled &&
		ab.Authorizer == nil && len(ab.Tokens) == 0 && ab.JWTSecret == "" {
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

type WebConfigureIdempotencyInput struct {
	Store        idempotency.Store
	ErrorHandler func(rw http.ResponseWriter, r *http.Request, err error)
}

// WebConfigureIdempotency injects a custom store or error handler. Call it
// from Configures; the middleware resolves the store on the first request.
func WebConfigureIdempotency(input WebConfigureIdempotencyInput) ConfigureFunc {
	return func(context.Context) {
		idem, ok := Config.Web.Idempotency.Get()
		if !ok {
			slog.Warn("web: idempotency is not configured; store not applied")
			return
		}

		if input.Store != nil {
			idem.Store = input.Store
		}
		if input.ErrorHandler != nil {
			idem.ErrorHandler = input.ErrorHandler
		}

		Config.Web.Idempotency = some.Of(idem)
	}
}

type WebConfigureBodyLimitInput struct {
	ErrorHandler func(rw http.ResponseWriter, r *http.Request, err error)
}

// WebConfigureBodyLimit injects a custom body-limit error handler.
func WebConfigureBodyLimit(input WebConfigureBodyLimitInput) ConfigureFunc {
	return func(context.Context) {
		bl, ok := Config.Web.BodyLimit.Get()
		if !ok {
			slog.Warn("web: body_limit is not configured; error handler not applied")
			return
		}

		bl.ErrorHandler = input.ErrorHandler
		Config.Web.BodyLimit = some.Of(bl)
	}
}
