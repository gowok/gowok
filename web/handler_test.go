package web

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang-must/must"
	gowok_errors "github.com/gowok/gowok/errors"
)

func TestHandler(t *testing.T) {
	testCases := []struct {
		name       string
		handler    func(ctx *Ctx) error
		wantStatus int
		wantBody   string
	}{
		{
			name: "positive/success",
			handler: func(ctx *Ctx) error {
				return ctx.Res().Ok("success")
			},
			wantStatus: http.StatusOK,
			wantBody:   "success",
		},
		{
			name: "positive/gowok error",
			handler: func(ctx *Ctx) error {
				return gowok_errors.New("bad request", gowok_errors.WithCode(http.StatusBadRequest))
			},
			wantStatus: http.StatusBadRequest,
			wantBody:   `{"code":400,"error":"bad request"}`,
		},
		{
			name: "positive/generic error",
			handler: func(ctx *Ctx) error {
				return errors.New("something went wrong")
			},
			wantStatus: http.StatusInternalServerError,
			wantBody:   "something went wrong",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, "/", nil)

			h := Handler(tc.handler)
			h.ServeHTTP(w, r)

			must.Equal(t, tc.wantStatus, w.Code)

			if tc.name == "positive/gowok error" {
				var res map[string]any
				err := json.Unmarshal(w.Body.Bytes(), &res)
				must.Nil(t, err)
				must.Equal(t, float64(400), res["code"])
				must.Equal(t, "bad request", res["error"])
			} else {
				must.Equal(t, tc.wantBody, w.Body.String())
			}
		})
	}
}

type mockResponseWriterNoFlusher struct {
	http.ResponseWriter
}

func TestHandlerSSE(t *testing.T) {
	testCases := []struct {
		name       string
		handler    func(ctx *CtxSse)
		writer     func() http.ResponseWriter
		wantStatus int
		wantHeader http.Header
	}{
		{
			name: "positive/success",
			handler: func(ctx *CtxSse) {
				_ = ctx.Publish([]byte("hello"))
			},
			writer: func() http.ResponseWriter {
				return httptest.NewRecorder()
			},
			wantStatus: http.StatusOK,
			wantHeader: http.Header{
				"Content-Type":  []string{"text/event-stream"},
				"Cache-Control": []string{"no-cache"},
				"Connection":    []string{"keep-alive"},
			},
		},
		{
			name:    "negative/unsupported streaming",
			handler: func(ctx *CtxSse) {},
			writer: func() http.ResponseWriter {
				return &mockResponseWriterNoFlusher{httptest.NewRecorder()}
			},
			wantStatus: http.StatusInternalServerError,
			wantHeader: http.Header{
				"Content-Type":  []string{"application/json"},
				"Cache-Control": []string{"no-cache"},
				"Connection":    []string{"keep-alive"},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			w := tc.writer()
			r := httptest.NewRequest(http.MethodGet, "/", nil)

			h := HandlerSSE(tc.handler)
			h.ServeHTTP(w, r)

			if rw, ok := w.(*httptest.ResponseRecorder); ok {
				must.Equal(t, tc.wantStatus, rw.Code)
				for k, v := range tc.wantHeader {
					must.Equal(t, v, rw.Header()[k])
				}
			} else if rw, ok := w.(*mockResponseWriterNoFlusher); ok {
				inner := rw.ResponseWriter.(*httptest.ResponseRecorder)
				must.Equal(t, tc.wantStatus, inner.Code)
				for k, v := range tc.wantHeader {
					must.Equal(t, v, inner.Header()[k])
				}
			}
		})
	}
}
