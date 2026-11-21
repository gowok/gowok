package gowok

import (
	"bytes"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/golang-must/must"
	"github.com/gowok/fp/some"
	"github.com/gowok/gowok/config"
	"github.com/gowok/gowok/web"
	"github.com/ngamux/middleware/ratelimit"
	"github.com/ngamux/ngamux"
)

type mockResource struct {
	called map[string]bool
}

func (m *mockResource) Index(w http.ResponseWriter, r *http.Request)   { m.called["Index"] = true }
func (m *mockResource) Show(w http.ResponseWriter, r *http.Request)    { m.called["Show"] = true }
func (m *mockResource) Store(w http.ResponseWriter, r *http.Request)   { m.called["Store"] = true }
func (m *mockResource) Update(w http.ResponseWriter, r *http.Request)  { m.called["Update"] = true }
func (m *mockResource) Destroy(w http.ResponseWriter, r *http.Request) { m.called["Destroy"] = true }

func TestWebHandler_Handler(t *testing.T) {
	handler := Web.Handler.Handler(func(ctx *web.Ctx) error {
		return nil
	})
	must.NotNil(t, handler)
}

func TestWebHandler_SSE(t *testing.T) {
	handler := Web.Handler.SSE(func(ctx *web.CtxSse) {
	})
	must.NotNil(t, handler)
}

func TestWebResource_New(t *testing.T) {
	oldMux := Web.HttpServeMux
	defer func() { Web.HttpServeMux = oldMux }()
	Web.HttpServeMux = ngamux.NewHttpServeMux()

	res := &mockResource{called: make(map[string]bool)}
	Web.Resource.New("/test", res)

	routes := []struct {
		method string
		path   string
		name   string
	}{
		{http.MethodGet, "/test", "Index"},
		{http.MethodPost, "/test", "Store"},
		{http.MethodGet, "/test/1", "Show"},
		{http.MethodPut, "/test/1", "Update"},
		{http.MethodDelete, "/test/1", "Destroy"},
	}

	for _, rt := range routes {
		t.Run(rt.method+" "+rt.path, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest(rt.method, rt.path, nil)
			Web.ServeHTTP(w, r)
			must.True(t, res.called[rt.name])
		})
	}
}

func TestWebResource_WithMiddleware(t *testing.T) {
	middlewareCalled := false
	var middleware ngamux.MiddlewareFunc = func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			middlewareCalled = true
			next(w, r)
		}
	}

	opt := Web.Resource.WithMiddleware(middleware)
	must.NotNil(t, opt)

	mux := ngamux.NewHttpServeMux()
	opt(mux)

	mux.Get("/test", func(w http.ResponseWriter, r *http.Request) {})
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/test", nil)
	mux.ServeHTTP(w, r)
	must.True(t, middlewareCalled)
}

func TestWebResponse_New(t *testing.T) {
	w := httptest.NewRecorder()
	resp := Web.Response.New(w)
	must.NotNil(t, resp)
}

func TestWebRequest_New(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	req := Web.Request.New(r)
	must.NotNil(t, req)
}

func TestWeb_Configure(t *testing.T) {
	oldMux := Web.HttpServeMux
	oldConfig := Config
	oldLogger := slog.Default()
	oldListenAndServe := webListenAndServe
	defer func() {
		Web.HttpServeMux = oldMux
		Config = oldConfig
		slog.SetDefault(oldLogger)
		webListenAndServe = oldListenAndServe
	}()

	tests := []struct {
		name          string
		webConfig     config.Web
		mockListenErr error
		expectedFatal bool
	}{
		{
			name: "positive/basic configuration",
			webConfig: config.Web{
				Host: ":8080",
			},
			mockListenErr: http.ErrServerClosed,
		},
		{
			name: "positive/static files enabled",
			webConfig: config.Web{
				Host: ":8080",
				Static: config.WebStatic{
					Enabled: true,
					Prefix:  "/static/",
					Dir:     "./public",
				},
			},
			mockListenErr: http.ErrServerClosed,
		},
		{
			name: "positive/middlewares enabled",
			webConfig: config.Web{
				Host:  ":8080",
				Log:   some.Of(config.WebLog{Enabled: true}),
				Cors:  some.Of(config.WebCors{Enabled: true}),
				Pprof: some.Of(config.WebPprof{Enabled: true}),
			},
			mockListenErr: http.ErrServerClosed,
		},
		{
			name: "negative/listen fails",
			webConfig: config.Web{
				Host: ":8080",
			},
			mockListenErr: errors.New("listen error"),
			expectedFatal: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Web.HttpServeMux = ngamux.NewHttpServeMux()
			Config = &config.Config{Web: tt.webConfig}

			var logBuf bytes.Buffer
			slog.SetDefault(slog.New(slog.NewTextHandler(&logBuf, nil)))
			webListenAndServe = func(s *http.Server) error {
				return tt.mockListenErr
			}

			Web.configure()
			Web.run()

			must.Equal(t, tt.expectedFatal, strings.Contains(logBuf.String(), "level=ERROR"))
			must.Equal(t, tt.webConfig.Host, Web.Server.Addr)
		})
	}
}

func TestWeb_RateLimit(t *testing.T) {
	oldMux := Web.HttpServeMux
	oldConfig := Config
	oldServer := Web.Server
	defer func() {
		Web.HttpServeMux = oldMux
		Config = oldConfig
		Web.Server = oldServer
	}()

	Web.HttpServeMux = ngamux.NewHttpServeMux()
	Config = &config.Config{Web: config.Web{
		Host: ":8080",
		RateLimit: some.Of(config.WebRateLimit{
			Enabled:  true,
			Requests: 1,
			Window:   "1m",
			Burst:    1,
			Exclude:  []string{"/health"},
		}),
	}}

	Web.configure()

	handled := 0
	Web.Get("/test", func(w http.ResponseWriter, r *http.Request) {
		handled++
		w.WriteHeader(http.StatusOK)
	})
	Web.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	do := func(path string) *httptest.ResponseRecorder {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, path, nil)
		Web.ServeHTTP(w, r)
		return w
	}

	first := do("/test")
	must.Equal(t, http.StatusOK, first.Code)
	must.Equal(t, "1", first.Header().Get(ratelimit.HeaderLimit))
	must.Equal(t, "0", first.Header().Get(ratelimit.HeaderRemaining))

	second := do("/test")
	must.Equal(t, http.StatusTooManyRequests, second.Code)
	must.NotEqual(t, "", second.Header().Get(ratelimit.HeaderRetryAfter))

	health := do("/health")
	must.Equal(t, http.StatusOK, health.Code)
	must.Equal(t, "", health.Header().Get(ratelimit.HeaderLimit))

	must.Equal(t, 1, handled)
}

func TestWeb_RequestID(t *testing.T) {
	oldMux := Web.HttpServeMux
	oldConfig := Config
	oldServer := Web.Server
	defer func() {
		Web.HttpServeMux = oldMux
		Config = oldConfig
		Web.Server = oldServer
	}()

	Web.HttpServeMux = ngamux.NewHttpServeMux()
	Config = &config.Config{Web: config.Web{
		Host:      ":8080",
		RequestID: some.Of(config.WebRequestID{Enabled: true}),
	}}

	Web.configure()

	seenID := ""
	Web.Get("/request-id", func(w http.ResponseWriter, r *http.Request) {
		seenID = r.Header.Get("X-Request-ID")
		w.WriteHeader(http.StatusOK)
	})

	do := func(path string, header ...string) *httptest.ResponseRecorder {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, path, nil)
		if len(header) == 2 {
			r.Header.Set(header[0], header[1])
		}
		Web.ServeHTTP(w, r)
		return w
	}

	t.Run("positive/generates request id", func(t *testing.T) {
		w := do("/request-id")

		must.Equal(t, http.StatusOK, w.Code)
		must.NotEqual(t, "", w.Header().Get("X-Request-ID"))
		must.Equal(t, w.Header().Get("X-Request-ID"), seenID)
	})

	t.Run("positive/keeps incoming request id", func(t *testing.T) {
		w := do("/request-id", "X-Request-ID", "request-123")

		must.Equal(t, "request-123", w.Header().Get("X-Request-ID"))
		must.Equal(t, "request-123", seenID)
	})
}

func TestWeb_RequestIDCustomHeader(t *testing.T) {
	oldMux := Web.HttpServeMux
	oldConfig := Config
	oldServer := Web.Server
	defer func() {
		Web.HttpServeMux = oldMux
		Config = oldConfig
		Web.Server = oldServer
	}()

	Web.HttpServeMux = ngamux.NewHttpServeMux()
	Config = &config.Config{Web: config.Web{
		Host:      ":8080",
		RequestID: some.Of(config.WebRequestID{Enabled: true, Header: "X-Trace-ID"}),
	}}

	Web.configure()

	seenID := ""
	Web.Get("/test", func(w http.ResponseWriter, r *http.Request) {
		seenID = r.Header.Get("X-Trace-ID")
		w.WriteHeader(http.StatusOK)
	})

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/test", nil)
	Web.ServeHTTP(w, r)

	must.NotEqual(t, "", w.Header().Get("X-Trace-ID"))
	must.Equal(t, w.Header().Get("X-Trace-ID"), seenID)
}
