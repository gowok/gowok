package gowok

import (
	"bytes"
	"compress/gzip"
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

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
	oldMux := Web.Ngamux
	defer func() { Web.Ngamux = oldMux }()
	Web.Ngamux = ngamux.New()

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

	mux := ngamux.New()
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
	oldMux := Web.Ngamux
	oldConfig := Config
	oldLogger := slog.Default()
	oldListenAndServe := webListenAndServe
	defer func() {
		Web.Ngamux = oldMux
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
			Web.Ngamux = ngamux.New()
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
	oldMux := Web.Ngamux
	oldConfig := Config
	oldServer := Web.Server
	defer func() {
		Web.Ngamux = oldMux
		Config = oldConfig
		Web.Server = oldServer
	}()

	Web.Ngamux = ngamux.New()
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
	oldMux := Web.Ngamux
	oldConfig := Config
	oldServer := Web.Server
	defer func() {
		Web.Ngamux = oldMux
		Config = oldConfig
		Web.Server = oldServer
	}()

	Web.Ngamux = ngamux.New()
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
	oldMux := Web.Ngamux
	oldConfig := Config
	oldServer := Web.Server
	defer func() {
		Web.Ngamux = oldMux
		Config = oldConfig
		Web.Server = oldServer
	}()

	Web.Ngamux = ngamux.New()
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

func TestWeb_Timeout(t *testing.T) {
	oldMux := Web.Ngamux
	oldConfig := Config
	oldServer := Web.Server
	defer func() {
		Web.Ngamux = oldMux
		Config = oldConfig
		Web.Server = oldServer
	}()

	t.Run("positive/request gets deadline", func(t *testing.T) {
		Web.Ngamux = ngamux.New()
		Config = &config.Config{Web: config.Web{
			Host:    ":8080",
			Timeout: some.Of(config.WebTimeout{Enabled: true, Duration: "20ms"}),
		}}
		Web.configure()

		var deadline time.Duration
		var ctxErr error
		handlerDone := make(chan struct{})
		Web.Get("/test", func(w http.ResponseWriter, r *http.Request) {
			defer close(handlerDone)

			d, ok := r.Context().Deadline()
			must.True(t, ok)
			deadline = time.Until(d)

			<-r.Context().Done()
			ctxErr = r.Context().Err()
			w.WriteHeader(http.StatusOK)
		})

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/test", nil)
		Web.ServeHTTP(w, r)
		<-handlerDone

		must.True(t, deadline <= 20*time.Millisecond)
		must.True(t, errors.Is(ctxErr, context.DeadlineExceeded))
		must.Equal(t, http.StatusServiceUnavailable, w.Code)
		must.True(t, strings.Contains(w.Body.String(), "request timeout"))
	})

	t.Run("negative/disabled", func(t *testing.T) {
		Web.Ngamux = ngamux.New()
		Config = &config.Config{Web: config.Web{
			Host:    ":8080",
			Timeout: some.Of(config.WebTimeout{Enabled: false, Duration: "20ms"}),
		}}
		Web.configure()

		var hasDeadline bool
		Web.Get("/test", func(w http.ResponseWriter, r *http.Request) {
			_, hasDeadline = r.Context().Deadline()
			w.WriteHeader(http.StatusOK)
		})

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/test", nil)
		Web.ServeHTTP(w, r)

		must.False(t, hasDeadline)
	})
}

func TestWeb_Compress(t *testing.T) {
	oldMux := Web.Ngamux
	oldConfig := Config
	oldServer := Web.Server
	defer func() {
		Web.Ngamux = oldMux
		Config = oldConfig
		Web.Server = oldServer
	}()

	setup := func(enabled bool) {
		Web.Ngamux = ngamux.New()
		Config = &config.Config{Web: config.Web{
			Host:     ":8080",
			Compress: some.Of(config.WebCompress{Enabled: enabled}),
		}}
		Web.configure()

		Web.Get("/json", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"ok":true}`))
		})
		Web.Get("/empty", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		})
	}

	do := func(path, acceptEncoding string) *httptest.ResponseRecorder {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, path, nil)
		if acceptEncoding != "" {
			r.Header.Set("Accept-Encoding", acceptEncoding)
		}
		Web.ServeHTTP(w, r)
		return w
	}

	t.Run("positive/gzip", func(t *testing.T) {
		setup(true)
		w := do("/json", "gzip")

		must.Equal(t, "gzip", w.Header().Get("Content-Encoding"))
		must.True(t, strings.Contains(w.Header().Get("Vary"), "Accept-Encoding"))

		reader, err := gzip.NewReader(w.Body)
		must.Nil(t, err)
		defer func() { _ = reader.Close() }()

		body, err := io.ReadAll(reader)
		must.Nil(t, err)
		must.Equal(t, `{"ok":true}`, string(body))
	})

	t.Run("positive/no accept encoding", func(t *testing.T) {
		setup(true)
		w := do("/json", "")

		must.Equal(t, "", w.Header().Get("Content-Encoding"))
		must.Equal(t, `{"ok":true}`, w.Body.String())
	})

	t.Run("negative/disabled", func(t *testing.T) {
		setup(false)
		w := do("/json", "gzip")

		must.Equal(t, "", w.Header().Get("Content-Encoding"))
		must.Equal(t, `{"ok":true}`, w.Body.String())
	})

	t.Run("negative/no content body", func(t *testing.T) {
		setup(true)
		w := do("/empty", "gzip")

		must.Equal(t, http.StatusNoContent, w.Code)
		must.Equal(t, "", w.Header().Get("Content-Encoding"))
	})
}

func TestWeb_AuthBearer(t *testing.T) {
	oldMux := Web.Ngamux
	oldConfig := Config
	oldServer := Web.Server
	defer func() {
		Web.Ngamux = oldMux
		Config = oldConfig
		Web.Server = oldServer
	}()

	Web.Ngamux = ngamux.New()
	Config = &config.Config{
		Web: config.Web{
			Host: ":8080",
			AuthBearer: some.Of(config.WebAuthBearer{
				Enabled: true,
				Exclude: []string{"/health"},
				Realm:   "gowok",
				Authorizer: func(token string) (any, error) {
					if token != "secret-token" {
						return nil, errors.New("invalid token")
					}
					return token, nil
				},
			}),
		},
	}
	Web.configure()

	var gotToken any
	Web.Get("/private", func(w http.ResponseWriter, r *http.Request) {
		gotToken = r.Context().Value("token")
		w.WriteHeader(http.StatusOK)
	})
	Web.Get("/health", func(w http.ResponseWriter, r *http.Request) {
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

	testCases := []struct {
		name      string
		path      string
		header    []string
		wantCode  int
		wantToken any
	}{
		{
			name:      "positive/raw token",
			path:      "/private",
			header:    []string{"Authorization", "secret-token"},
			wantCode:  http.StatusOK,
			wantToken: "secret-token",
		},
		{
			name:      "positive/bearer token",
			path:      "/private",
			header:    []string{"Authorization", "Bearer secret-token"},
			wantCode:  http.StatusOK,
			wantToken: "secret-token",
		},
		{
			name:      "positive/bearer prefix is case insensitive",
			path:      "/private",
			header:    []string{"Authorization", "bEaReR secret-token"},
			wantCode:  http.StatusOK,
			wantToken: "secret-token",
		},
		{
			name:     "negative/invalid token",
			path:     "/private",
			header:   []string{"Authorization", "wrong-token"},
			wantCode: http.StatusUnauthorized,
		},
		{
			name:     "negative/missing token",
			path:     "/private",
			wantCode: http.StatusUnauthorized,
		},
		{
			name:     "positive/excluded path",
			path:     "/health",
			wantCode: http.StatusOK,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			gotToken = nil
			w := do(tc.path, tc.header...)

			must.Equal(t, tc.wantCode, w.Code)
			if tc.wantToken != nil {
				must.Equal(t, tc.wantToken, gotToken)
			}
			if tc.wantCode == http.StatusUnauthorized {
				must.Equal(t, "Bearer realm=\"gowok\"", w.Header().Get("WWW-Authenticate"))
			}
		})
	}
}

func TestWeb_AuthBearerWithoutAuthorizer(t *testing.T) {
	oldMux := Web.Ngamux
	oldConfig := Config
	oldServer := Web.Server
	defer func() {
		Web.Ngamux = oldMux
		Config = oldConfig
		Web.Server = oldServer
	}()

	Web.Ngamux = ngamux.New()
	Config = &config.Config{
		Web: config.Web{
			Host:       ":8080",
			AuthBearer: some.Of(config.WebAuthBearer{Enabled: true}),
		},
	}
	Web.configure()

	Web.Get("/private", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/private", nil)
	r.Header.Set("Authorization", "Bearer any-token")
	Web.ServeHTTP(w, r)

	must.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestWeb_AuthBearerRealm(t *testing.T) {
	oldMux := Web.Ngamux
	oldConfig := Config
	oldServer := Web.Server
	defer func() {
		Web.Ngamux = oldMux
		Config = oldConfig
		Web.Server = oldServer
	}()

	Web.Ngamux = ngamux.New()
	Config = &config.Config{
		Web: config.Web{
			Host: ":8080",
			AuthBearer: some.Of(config.WebAuthBearer{
				Enabled: true,
				Realm:   "gowok api",
			}),
		},
	}
	Web.configure()

	Web.Get("/private", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/private", nil)
	Web.ServeHTTP(w, r)

	must.Equal(t, http.StatusUnauthorized, w.Code)
	must.Equal(t, `Bearer realm="gowok api"`, w.Header().Get("WWW-Authenticate"))
}

func TestWeb_AuthBearerCustomHeader(t *testing.T) {
	oldMux := Web.Ngamux
	oldConfig := Config
	oldServer := Web.Server
	defer func() {
		Web.Ngamux = oldMux
		Config = oldConfig
		Web.Server = oldServer
	}()

	Web.Ngamux = ngamux.New()
	Config = &config.Config{
		Web: config.Web{
			Host: ":8080",
			AuthBearer: some.Of(config.WebAuthBearer{
				Enabled:    true,
				Header:     "X-API-Token",
				ContextKey: "api_token",
				Authorizer: func(token string) (any, error) {
					if token != "api-token" {
						return nil, errors.New("invalid token")
					}
					return token, nil
				},
			}),
		},
	}
	Web.configure()

	var gotToken any
	Web.Get("/private", func(w http.ResponseWriter, r *http.Request) {
		gotToken = r.Context().Value("api_token")
		w.WriteHeader(http.StatusOK)
	})

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/private", nil)
	r.Header.Set("X-API-Token", "api-token")
	Web.ServeHTTP(w, r)

	must.Equal(t, http.StatusOK, w.Code)
	must.Equal(t, "api-token", gotToken)
}

func TestWeb_AuthBearerConfigures(t *testing.T) {
	resetIgniter()
	defer resetIgniter()

	p := configure(config.Config{Web: config.Web{
		Enabled:    true,
		Host:       ":8080",
		AuthBearer: some.Of(config.WebAuthBearer{Enabled: true}),
	}})
	must.NotNil(t, p)

	Configures(WebConfigureAuthBearer(WebConfigureAuthBearerInput{
		Authorizer: func(token string) (any, error) {
			if token != "from-configures" {
				return nil, errors.New("invalid token")
			}
			return token, nil
		},
		ErrorHandler: func(rw http.ResponseWriter, err error) {
			rw.WriteHeader(http.StatusUnauthorized)
		},
	}))

	Web.Get("/private", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	do := func(token string) *httptest.ResponseRecorder {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/private", nil)
		r.Header.Set("Authorization", "Bearer "+token)
		Web.ServeHTTP(w, r)
		return w
	}

	must.Equal(t, http.StatusOK, do("from-configures").Code)
	must.Equal(t, http.StatusUnauthorized, do("wrong").Code)
}

func TestWeb_AuthBasic(t *testing.T) {
	oldMux := Web.Ngamux
	oldConfig := Config
	oldServer := Web.Server
	defer func() {
		Web.Ngamux = oldMux
		Config = oldConfig
		Web.Server = oldServer
	}()

	Web.Ngamux = ngamux.New()
	Config = &config.Config{
		Web: config.Web{
			Host: ":8080",
			AuthBasic: some.Of(config.WebAuthBasic{
				Enabled: true,
				Exclude: []string{"/health"},
				Authorizer: func(username, password string) bool {
					return username == "admin" && password == "secret"
				},
			}),
		},
	}
	Web.configure()

	Web.Get("/private", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	Web.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	do := func(path string, creds ...string) *httptest.ResponseRecorder {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, path, nil)
		if len(creds) == 2 {
			r.SetBasicAuth(creds[0], creds[1])
		}
		Web.ServeHTTP(w, r)
		return w
	}

	testCases := []struct {
		name     string
		path     string
		creds    []string
		wantCode int
	}{
		{
			name:     "positive/valid credentials",
			path:     "/private",
			creds:    []string{"admin", "secret"},
			wantCode: http.StatusOK,
		},
		{
			name:     "negative/invalid credentials",
			path:     "/private",
			creds:    []string{"admin", "wrong"},
			wantCode: http.StatusUnauthorized,
		},
		{
			name:     "negative/missing credentials",
			path:     "/private",
			wantCode: http.StatusUnauthorized,
		},
		{
			name:     "positive/excluded path",
			path:     "/health",
			wantCode: http.StatusOK,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			w := do(tc.path, tc.creds...)
			must.Equal(t, tc.wantCode, w.Code)
		})
	}
}

func TestWeb_AuthBasicWithoutAuthorizer(t *testing.T) {
	oldMux := Web.Ngamux
	oldConfig := Config
	oldServer := Web.Server
	defer func() {
		Web.Ngamux = oldMux
		Config = oldConfig
		Web.Server = oldServer
	}()

	Web.Ngamux = ngamux.New()
	Config = &config.Config{
		Web: config.Web{
			Host:      ":8080",
			AuthBasic: some.Of(config.WebAuthBasic{Enabled: true}),
		},
	}
	Web.configure()

	Web.Get("/private", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/private", nil)
	r.SetBasicAuth("root", "root")
	Web.ServeHTTP(w, r)

	must.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestWeb_AuthBasicRealm(t *testing.T) {
	oldMux := Web.Ngamux
	oldConfig := Config
	oldServer := Web.Server
	defer func() {
		Web.Ngamux = oldMux
		Config = oldConfig
		Web.Server = oldServer
	}()

	Web.Ngamux = ngamux.New()
	Config = &config.Config{
		Web: config.Web{
			Host: ":8080",
			AuthBasic: some.Of(config.WebAuthBasic{
				Enabled: true,
				Realm:   "gowok api",
			}),
		},
	}
	Web.configure()

	Web.Get("/private", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/private", nil)
	Web.ServeHTTP(w, r)

	must.Equal(t, http.StatusUnauthorized, w.Code)
	must.Equal(t, `Basic realm="gowok api"`, w.Header().Get("WWW-Authenticate"))
}

func TestWeb_AuthBasicConfigures(t *testing.T) {
	resetIgniter()
	defer resetIgniter()

	p := configure(config.Config{Web: config.Web{
		Enabled:   true,
		Host:      ":8080",
		AuthBasic: some.Of(config.WebAuthBasic{Enabled: true}),
	}})
	must.NotNil(t, p)

	Configures(WebConfigureAuthBasic(WebConfigureAuthBasicInput{
		Authorizer: func(username, password string) bool {
			return username == "from-configures" && password == "secret"
		},
		ErrorHandler: func(rw http.ResponseWriter, err error) {
			rw.WriteHeader(http.StatusUnauthorized)
		},
	}))

	Web.Get("/private", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	do := func(username, password string) *httptest.ResponseRecorder {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/private", nil)
		r.SetBasicAuth(username, password)
		Web.ServeHTTP(w, r)
		return w
	}

	must.Equal(t, http.StatusOK, do("from-configures", "secret").Code)
	must.Equal(t, http.StatusUnauthorized, do("from-configures", "wrong").Code)
}

func TestWebConfigureAuthBasic(t *testing.T) {
	oldConfig := Config
	defer func() { Config = oldConfig }()

	Config = &config.Config{Web: config.Web{
		AuthBasic: some.Of(config.WebAuthBasic{Enabled: true}),
	}}

	WebConfigureAuthBasic(WebConfigureAuthBasicInput{
		Authorizer: func(username, password string) bool {
			return username == password
		},
		ErrorHandler: func(rw http.ResponseWriter, err error) {
			rw.WriteHeader(http.StatusUnauthorized)
		},
	})(context.Background())

	ab, ok := Config.Web.AuthBasic.Get()
	must.True(t, ok)
	must.NotNil(t, ab.Authorizer)
	must.NotNil(t, ab.ErrorHandler)

	WebConfigureAuthBasic(WebConfigureAuthBasicInput{})(context.Background())

	ab, ok = Config.Web.AuthBasic.Get()
	must.True(t, ok)
	must.Nil(t, ab.Authorizer)
	must.Nil(t, ab.ErrorHandler)
}

func TestWebConfigureAuthBearer(t *testing.T) {
	oldConfig := Config
	defer func() { Config = oldConfig }()

	Config = &config.Config{Web: config.Web{
		AuthBearer: some.Of(config.WebAuthBearer{Enabled: true}),
	}}

	WebConfigureAuthBearer(WebConfigureAuthBearerInput{
		Authorizer: func(token string) (any, error) {
			return token, nil
		},
		ErrorHandler: func(rw http.ResponseWriter, err error) {
			rw.WriteHeader(http.StatusUnauthorized)
		},
	})(context.Background())

	ab, ok := Config.Web.AuthBearer.Get()
	must.True(t, ok)
	must.NotNil(t, ab.Authorizer)
	must.NotNil(t, ab.ErrorHandler)

	WebConfigureAuthBearer(WebConfigureAuthBearerInput{})(context.Background())

	ab, ok = Config.Web.AuthBearer.Get()
	must.True(t, ok)
	must.Nil(t, ab.Authorizer)
	must.Nil(t, ab.ErrorHandler)
}
