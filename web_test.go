package gowok

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang-must/must"
	"github.com/gowok/gowok/config"
	"github.com/gowok/gowok/some"
	"github.com/gowok/gowok/web"
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
	oldLogFatalln := webLogFatalln
	oldListenAndServe := webListenAndServe
	defer func() {
		Web.HttpServeMux = oldMux
		Config = oldConfig
		webLogFatalln = oldLogFatalln
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

			fatalCalled := false
			webLogFatalln = func(v ...any) {
				fatalCalled = true
			}
			webListenAndServe = func(s *http.Server) error {
				return tt.mockListenErr
			}

			Web.configure()

			must.Equal(t, tt.expectedFatal, fatalCalled)
			must.Equal(t, tt.webConfig.Host, Web.Server.Addr)
		})
	}
}
