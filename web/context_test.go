package web

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-must/must"
)

func TestCtx(t *testing.T) {
	t.Run("positive/NewCtx and getters", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		ctx := NewCtx(r.Context(), w, r)

		must.NotNil(t, ctx.Req())
		must.NotNil(t, ctx.Res())
		must.Equal(t, r, ctx.Req().ToHttp())
	})

	t.Run("positive/Write", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		ctx := NewCtx(r.Context(), w, r)

		content := []byte("hello")
		n, err := ctx.Write(content)
		must.Nil(t, err)
		must.Equal(t, len(content), n)
		must.Equal(t, "hello", w.Body.String())
	})

	t.Run("positive/Context manipulation", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		ctx := NewCtx(r.Context(), w, r)

		type key string
		k := key("foo")
		ctx.SetValue(k, "bar")
		must.Equal(t, "bar", ctx.Value(k))

		newCtx := context.WithValue(context.Background(), k, "baz")
		ctx.SetContext(newCtx)
		must.Equal(t, "baz", ctx.Value(k))
	})

	t.Run("positive/Context interface methods", func(t *testing.T) {
		ctxWithTimeout, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/", nil).WithContext(ctxWithTimeout)
		ctx := NewCtx(ctxWithTimeout, w, r)

		deadline, ok := ctx.Deadline()
		must.True(t, ok)
		must.NotNil(t, deadline)

		must.NotNil(t, ctx.Done())
		must.Nil(t, ctx.Err())
	})
}

func TestCtxSse(t *testing.T) {
	t.Run("positive/NewCtxSse success", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		ctx := NewCtx(r.Context(), w, r)

		sse, err := NewCtxSse(ctx)
		must.Nil(t, err)
		must.NotNil(t, sse)
	})

	t.Run("negative/NewCtxSse failure (no flusher)", func(t *testing.T) {
		w := &mockResponseWriterNoFlusher{httptest.NewRecorder()}
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		ctx := NewCtx(r.Context(), w, r)

		sse, err := NewCtxSse(ctx)
		must.NotNil(t, err)
		must.Nil(t, sse)
	})

	t.Run("positive/Publish, Emit, Flush", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		ctx := NewCtx(r.Context(), w, r)

		sse, _ := NewCtxSse(ctx)

		err := sse.Publish([]byte("hello"))
		must.Nil(t, err)
		must.Equal(t, "data: hello\n\n", w.Body.String())

		w.Body.Reset()
		err = sse.Emit("ping", []byte("pong"))
		must.Nil(t, err)
		must.Equal(t, "event: ping\ndata: pong\n\n", w.Body.String())

		w.Body.Reset()
		err = sse.PublishRaw("custom: %s", "value")
		must.Nil(t, err)
		must.Equal(t, "custom: value", w.Body.String())

		sse.Flush()
	})
}
