package gowok

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/golang-must/must"
	"github.com/gowok/fp/some"
	"github.com/gowok/gowok/config"
	"github.com/ngamux/middleware/idempotency"
)

func configureWeb(t *testing.T, webConfig config.Web) {
	t.Helper()

	resetIgniter()
	t.Cleanup(resetIgniter)

	Config = &config.Config{Web: webConfig}
	Web.configure()
}

func TestWeb_BodyLimit(t *testing.T) {
	configureWeb(t, config.Web{
		Enabled:   true,
		BodyLimit: some.Of(config.WebBodyLimit{Enabled: true, Limit: 5}),
	})

	Web.Post("/upload", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	do := func(body string) int {
		recorder := httptest.NewRecorder()
		Web.ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, "/upload", strings.NewReader(body)))

		return recorder.Code
	}

	must.Equal(t, http.StatusOK, do("12345"))
	must.Equal(t, http.StatusRequestEntityTooLarge, do("123456"))
}

func TestWeb_RealIP(t *testing.T) {
	configureWeb(t, config.Web{
		Enabled: true,
		RealIP: some.Of(config.WebRealIP{
			Enabled:        true,
			TrustedProxies: []string{"10.0.0.0/8"},
		}),
	})

	var seen string
	Web.Get("/ip", func(w http.ResponseWriter, r *http.Request) {
		seen = r.RemoteAddr
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/ip", nil)
	req.RemoteAddr = "10.0.0.1:1234"
	req.Header.Set("X-Forwarded-For", "203.0.113.9")
	Web.ServeHTTP(httptest.NewRecorder(), req)

	must.Equal(t, "203.0.113.9:1234", seen)
}

func TestWeb_Idempotency(t *testing.T) {
	configureWeb(t, config.Web{
		Enabled:     true,
		Idempotency: some.Of(config.WebIdempotency{Enabled: true}),
	})

	calls := 0
	Web.Post("/orders", func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte("created"))
	})

	do := func() *httptest.ResponseRecorder {
		recorder := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/orders", strings.NewReader(`{"value":1}`))
		req.Header.Set("Idempotency-Key", "key-1")
		Web.ServeHTTP(recorder, req)

		return recorder
	}

	first := do()
	second := do()

	must.Equal(t, 1, calls)
	must.Equal(t, http.StatusCreated, first.Code)
	must.Equal(t, http.StatusCreated, second.Code)
	must.Equal(t, "created", second.Body.String())
}

type countingStore struct {
	inner  idempotency.Store
	begins int
}

func (s *countingStore) Begin(key string) (*idempotency.StoredResponse, idempotency.State) {
	s.begins++
	return s.inner.Begin(key)
}

func (s *countingStore) Complete(key string, response *idempotency.StoredResponse) {
	s.inner.Complete(key, response)
}

func (s *countingStore) Abort(key string) {
	s.inner.Abort(key)
}

func TestWeb_IdempotencyConfigureStore(t *testing.T) {
	resetIgniter()
	defer resetIgniter()

	configure(config.Config{Web: config.Web{
		Enabled:     true,
		Idempotency: some.Of(config.WebIdempotency{Enabled: true}),
	}})

	store := &countingStore{inner: idempotency.NewMemoryStore(time.Minute)}
	Configures(WebConfigureIdempotency(WebConfigureIdempotencyInput{Store: store}))

	calls := 0
	Web.Post("/orders", func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.WriteHeader(http.StatusCreated)
	})

	do := func() {
		recorder := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/orders", nil)
		req.Header.Set("Idempotency-Key", "key-1")
		Web.ServeHTTP(recorder, req)
		must.Equal(t, http.StatusCreated, recorder.Code)
	}

	do()
	do()

	must.Equal(t, 1, calls)
	must.Equal(t, 2, store.begins)
}

func TestWeb_BodyLimitConfigureHandler(t *testing.T) {
	resetIgniter()
	defer resetIgniter()

	configure(config.Config{Web: config.Web{
		Enabled:   true,
		BodyLimit: some.Of(config.WebBodyLimit{Enabled: true, Limit: 1}),
	}})

	Configures(WebConfigureBodyLimit(WebConfigureBodyLimitInput{
		ErrorHandler: func(rw http.ResponseWriter, _ *http.Request, _ error) {
			rw.WriteHeader(http.StatusTeapot)
		},
	}))

	Web.Post("/upload", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	recorder := httptest.NewRecorder()
	Web.ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, "/upload", strings.NewReader("ab")))

	must.Equal(t, http.StatusTeapot, recorder.Code)
}

func TestWeb_RateLimitKeyHeader(t *testing.T) {
	configureWeb(t, config.Web{
		Enabled: true,
		RateLimit: some.Of(config.WebRateLimit{
			Enabled:  true,
			Requests: 1,
			Window:   "1h",
			Burst:    1,
			Key:      "header:X-API-Key",
		}),
	})

	Web.Get("/limited", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	do := func(apiKey string) int {
		recorder := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/limited", nil)
		req.RemoteAddr = "10.0.0.1:1234"
		req.Header.Set("X-API-Key", apiKey)
		Web.ServeHTTP(recorder, req)

		return recorder.Code
	}

	must.Equal(t, http.StatusOK, do("a"))
	must.Equal(t, http.StatusTooManyRequests, do("a"))
	must.Equal(t, http.StatusOK, do("b"))
}

func TestWeb_AuthBearerStaticTokens(t *testing.T) {
	configureWeb(t, config.Web{
		Enabled: true,
		AuthBearer: some.Of(config.WebAuthBearer{
			Enabled: true,
			Tokens:  []string{"secret-token"},
		}),
	})

	Web.Get("/private", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	do := func(token string) int {
		recorder := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/private", nil)
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
		Web.ServeHTTP(recorder, req)

		return recorder.Code
	}

	must.Equal(t, http.StatusOK, do("secret-token"))
	must.Equal(t, http.StatusUnauthorized, do("wrong"))
	must.Equal(t, http.StatusUnauthorized, do(""))
}
