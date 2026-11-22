package testutil

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang-must/must"
)

func echoHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)

		w.Header().Set("X-Method", r.Method)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{
			"method": r.Method,
			"path":   r.URL.Path,
			"query":  r.URL.RawQuery,
			"body":   string(body),
			"type":   r.Header.Get("Content-Type"),
		})
	})
}

func TestClientVerbs(t *testing.T) {
	client := New(t, echoHandler())

	t.Run("positive/get", func(t *testing.T) {
		var body map[string]string
		client.Get("/users").IsStatusEquals(http.StatusOK).JSON(&body)

		must.Equal(t, http.MethodGet, body["method"])
		must.Equal(t, "/users", body["path"])
		must.Equal(t, "", body["body"])
	})

	t.Run("positive/post json body", func(t *testing.T) {
		var body map[string]string
		client.Post("/users", map[string]string{"name": "ada"}).
			IsStatusEquals(http.StatusOK).
			JSON(&body)

		must.Equal(t, http.MethodPost, body["method"])
		must.Equal(t, `{"name":"ada"}`, body["body"])
		must.Equal(t, "application/json", body["type"])
	})

	t.Run("positive/post raw body", func(t *testing.T) {
		var body map[string]string
		client.Post("/raw", "hello").IsStatusEquals(http.StatusOK).JSON(&body)

		must.Equal(t, "hello", body["body"])
		must.Equal(t, "", body["type"])
	})

	t.Run("positive/post bytes body", func(t *testing.T) {
		var body map[string]string
		client.Post("/bytes", []byte("raw-bytes")).IsStatusEquals(http.StatusOK).JSON(&body)

		must.Equal(t, "raw-bytes", body["body"])
	})

	t.Run("positive/put patch delete", func(t *testing.T) {
		for _, tc := range []struct {
			method string
			send   func() *Response
		}{
			{http.MethodPut, func() *Response { return client.Put("/item", map[string]string{"a": "b"}) }},
			{http.MethodPatch, func() *Response { return client.Patch("/item", map[string]string{"a": "b"}) }},
			{http.MethodDelete, func() *Response { return client.Delete("/item") }},
		} {
			var body map[string]string
			tc.send().IsStatusEquals(http.StatusOK).JSON(&body)
			must.Equal(t, tc.method, body["method"])
		}
	})
}

func TestClientDo(t *testing.T) {
	client := New(t, echoHandler())

	req := httptest.NewRequest(http.MethodGet, "/search?q=go", nil)
	req.Header.Set("X-Test", "value")

	var body map[string]string
	client.Do(req).IsHeaderEquals("X-Method", http.MethodGet).IsStatusEquals(http.StatusOK).JSON(&body)

	must.Equal(t, "q=go", body["query"])
}

func TestResponseAccessors(t *testing.T) {
	client := New(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte("created"))
	}))

	res := client.Get("/")

	must.Equal(t, http.StatusCreated, res.Status())
	must.Equal(t, "created", res.Text())
	must.Equal(t, []byte("created"), res.Bytes())
	res.IsStatusEquals(http.StatusCreated).IsBodyEquals("created").IsHeaderEquals("Content-Type", "")
}

func TestResponseIsStatus(t *testing.T) {
	t.Run("positive/equals", func(t *testing.T) {
		res, tb := newFakeResponse(t, http.StatusOK, "")
		res.IsStatusEquals(http.StatusOK)
		must.False(t, tb.failed)
	})

	t.Run("negative/equals", func(t *testing.T) {
		res, tb := newFakeResponse(t, http.StatusTeapot, "")
		res.IsStatusEquals(http.StatusOK)
		must.True(t, tb.failed)
	})

	t.Run("positive/not equals", func(t *testing.T) {
		res, tb := newFakeResponse(t, http.StatusTeapot, "")
		res.IsStatusNotEquals(http.StatusOK)
		must.False(t, tb.failed)
	})

	t.Run("negative/not equals", func(t *testing.T) {
		res, tb := newFakeResponse(t, http.StatusOK, "")
		res.IsStatusNotEquals(http.StatusOK)
		must.True(t, tb.failed)
	})
}

func TestResponseIsHeader(t *testing.T) {
	t.Run("positive/equals", func(t *testing.T) {
		res, tb := newFakeResponse(t, http.StatusOK, "")
		res.Header().Set("Content-Type", "text/plain")
		res.IsHeaderEquals("Content-Type", "text/plain")
		must.False(t, tb.failed)
	})

	t.Run("negative/equals", func(t *testing.T) {
		res, tb := newFakeResponse(t, http.StatusOK, "")
		res.IsHeaderEquals("Content-Type", "text/plain")
		must.True(t, tb.failed)
	})

	t.Run("positive/not equals", func(t *testing.T) {
		res, tb := newFakeResponse(t, http.StatusOK, "")
		res.IsHeaderNotEquals("Content-Type", "text/plain")
		must.False(t, tb.failed)
	})

	t.Run("negative/not equals", func(t *testing.T) {
		res, tb := newFakeResponse(t, http.StatusOK, "")
		res.Header().Set("Content-Type", "text/plain")
		res.IsHeaderNotEquals("Content-Type", "text/plain")
		must.True(t, tb.failed)
	})
}

func TestResponseIsBody(t *testing.T) {
	t.Run("positive/equals", func(t *testing.T) {
		res, tb := newFakeResponse(t, http.StatusOK, "ok")
		res.IsBodyEquals("ok")
		must.False(t, tb.failed)
	})

	t.Run("negative/equals", func(t *testing.T) {
		res, tb := newFakeResponse(t, http.StatusOK, "ok")
		res.IsBodyEquals("other")
		must.True(t, tb.failed)
	})

	t.Run("positive/not equals", func(t *testing.T) {
		res, tb := newFakeResponse(t, http.StatusOK, "ok")
		res.IsBodyNotEquals("other")
		must.False(t, tb.failed)
	})

	t.Run("negative/not equals", func(t *testing.T) {
		res, tb := newFakeResponse(t, http.StatusOK, "ok")
		res.IsBodyNotEquals("ok")
		must.True(t, tb.failed)
	})
}

func TestResponseIsJSON(t *testing.T) {
	t.Run("positive/equals", func(t *testing.T) {
		res, tb := newFakeResponse(t, http.StatusOK, `{"name":"ada","age":36}`)
		res.IsJSONEquals(map[string]any{"name": "ada", "age": 36})
		must.False(t, tb.failed)
	})

	t.Run("negative/equals", func(t *testing.T) {
		res, tb := newFakeResponse(t, http.StatusOK, `{"name":"ada"}`)
		res.IsJSONEquals(map[string]string{"name": "grace"})
		must.True(t, tb.failed)
	})

	t.Run("positive/not equals", func(t *testing.T) {
		res, tb := newFakeResponse(t, http.StatusOK, `{"name":"ada"}`)
		res.IsJSONNotEquals(map[string]string{"name": "grace"})
		must.False(t, tb.failed)
	})

	t.Run("negative/not equals", func(t *testing.T) {
		res, tb := newFakeResponse(t, http.StatusOK, `{"name":"ada"}`)
		res.IsJSONNotEquals(map[string]string{"name": "ada"})
		must.True(t, tb.failed)
	})

	t.Run("negative/not json", func(t *testing.T) {
		res, tb := newFakeResponse(t, http.StatusOK, "not-json")
		res.IsJSONEquals(map[string]string{"name": "ada"})
		must.True(t, tb.failed)
	})
}

func newFakeResponse(t *testing.T, code int, body string) (*Response, *fakeTB) {
	t.Helper()

	tb := &fakeTB{T: t}
	recorder := httptest.NewRecorder()
	recorder.Code = code
	_, _ = recorder.Body.WriteString(body)

	return &Response{t: tb, ResponseRecorder: recorder}, tb
}

// fakeTB implements testing.TB without failing the real test.
type fakeTB struct {
	*testing.T
	failed bool
}

func (f *fakeTB) Errorf(string, ...any) { f.failed = true }

func (f *fakeTB) Fatalf(string, ...any) { f.failed = true }
