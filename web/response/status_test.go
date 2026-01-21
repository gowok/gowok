package response

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang-must/must"
)

func TestStatus(t *testing.T) {
	testCases := []struct {
		name       string
		method     func(ctx Response, body ...any) error
		wantStatus int
		body       []any
		wantBody   string
	}{
		{
			name: "positive/Ok",
			method: func(ctx Response, body ...any) error {
				return ctx.Ok(body...)
			},
			wantStatus: http.StatusOK,
			body:       []any{"ok"},
			wantBody:   "ok",
		},
		{
			name: "positive/BadRequest",
			method: func(ctx Response, body ...any) error {
				return ctx.BadRequest(body...)
			},
			wantStatus: http.StatusBadRequest,
			body:       []any{"bad"},
			wantBody:   "bad",
		},
		{
			name: "positive/Unauthorized",
			method: func(ctx Response, body ...any) error {
				return ctx.Unauthorized(body...)
			},
			wantStatus: http.StatusUnauthorized,
			body:       []any{"unauthorized"},
			wantBody:   "unauthorized",
		},
		{
			name: "positive/NotFound",
			method: func(ctx Response, body ...any) error {
				return ctx.NotFound(body...)
			},
			wantStatus: http.StatusNotFound,
			body:       []any{"not found"},
			wantBody:   "not found",
		},
		{
			name: "positive/InternalServerError",
			method: func(ctx Response, body ...any) error {
				return ctx.InternalServerError(body...)
			},
			wantStatus: http.StatusInternalServerError,
			body:       []any{"error"},
			wantBody:   "error",
		},
		{
			name: "positive/Created",
			method: func(ctx Response, body ...any) error {
				return ctx.Created(body...)
			},
			wantStatus: http.StatusCreated,
			body:       []any{"created"},
			wantBody:   "created",
		},
		{
			name: "positive/Forbidden",
			method: func(ctx Response, body ...any) error {
				return ctx.Forbidden(body...)
			},
			wantStatus: http.StatusForbidden,
			body:       []any{"forbidden"},
			wantBody:   "forbidden",
		},
		{
			name: "positive/Conflict",
			method: func(ctx Response, body ...any) error {
				return ctx.Conflict(body...)
			},
			wantStatus: http.StatusConflict,
			body:       []any{"conflict"},
			wantBody:   "conflict",
		},
		{
			name: "positive/NoContent",
			method: func(ctx Response, body ...any) error {
				return ctx.NoContent()
			},
			wantStatus: http.StatusNoContent,
			wantBody:   "",
		},
		{
			name: "positive/Accepted",
			method: func(ctx Response, body ...any) error {
				return ctx.Accepted(body...)
			},
			wantStatus: http.StatusAccepted,
			body:       []any{"accepted"},
			wantBody:   "accepted",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			ctx := New(w)
			err := tc.method(*ctx, tc.body...)
			must.Nil(t, err)
			must.Equal(t, tc.wantStatus, w.Code)
			must.Equal(t, tc.wantBody, w.Body.String())
		})
	}
}
