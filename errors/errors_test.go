package errors

import (
	"encoding/json"
	stderrors "errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang-must/must"
)

func TestNew(t *testing.T) {
	testCases := []struct {
		name         string
		text         string
		opts         []Option
		expectedCode int
	}{
		{
			name:         "positive/new error without code",
			text:         "test error",
			expectedCode: 0,
		},
		{
			name:         "positive/new error with code",
			text:         "error with code",
			opts:         []Option{WithCode(http.StatusBadRequest)},
			expectedCode: http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := New(tc.text, tc.opts...)
			must.Equal(t, tc.text, err.Error())
			must.Equal(t, tc.expectedCode, err.code)
		})
	}
}

func TestError_MarshalJSON(t *testing.T) {
	testCases := []struct {
		name     string
		err      Error
		expected string
	}{
		{
			name:     "positive/marshal error without code",
			err:      New("simple error"),
			expected: `{"error":"simple error"}`,
		},
		{
			name:     "positive/marshal error with code",
			err:      New("error with code", WithCode(404)),
			expected: `{"code":404,"error":"error with code"}`,
		},
		{
			name:     "positive/marshal nil-wrapped error",
			err:      Wrap(nil),
			expected: `{"error":""}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			b, err := tc.err.MarshalJSON()
			must.Nil(t, err)

			var expectedMap, actualMap map[string]any
			must.Nil(t, json.Unmarshal([]byte(tc.expected), &expectedMap))
			must.Nil(t, json.Unmarshal(b, &actualMap))
			must.Equal(t, expectedMap, actualMap)
		})
	}
}

func TestError_WriteResponse(t *testing.T) {
	testCases := []struct {
		name           string
		err            Error
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "positive/write response without code",
			err:            New("unauthorized"),
			expectedStatus: http.StatusOK, // ngamux default status if code is 0 maybe?
			expectedBody:   "unauthorized",
		},
		{
			name:           "positive/write response with code",
			err:            New("not found", WithCode(http.StatusNotFound)),
			expectedStatus: http.StatusNotFound,
			expectedBody:   "not found",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rw := httptest.NewRecorder()
			tc.err.WriteResponse(rw)

			must.Equal(t, tc.expectedStatus, rw.Code)
			must.Equal(t, tc.expectedBody, rw.Body.String())
		})
	}
}

func TestWrap(t *testing.T) {
	t.Run("positive/wrap plain error", func(t *testing.T) {
		inner := stderrors.New("inner error")
		err := Wrap(inner)

		must.Equal(t, "inner error", err.Error())
		must.Equal(t, 0, err.Code())
		must.True(t, stderrors.Is(err, inner))
		must.Equal(t, inner, err.Unwrap())
	})

	t.Run("positive/wrap with code", func(t *testing.T) {
		err := Wrap(stderrors.New("not found"), WithCode(http.StatusNotFound))

		must.Equal(t, "not found", err.Error())
		must.Equal(t, http.StatusNotFound, err.Code())
	})

	t.Run("positive/wrap preserves code from inner Error", func(t *testing.T) {
		inner := New("forbidden", WithCode(http.StatusForbidden))
		err := Wrap(inner)

		must.Equal(t, http.StatusForbidden, err.Code())
		must.True(t, stderrors.Is(err, inner))
	})

	t.Run("positive/wrap option overrides preserved code", func(t *testing.T) {
		inner := New("forbidden", WithCode(http.StatusForbidden))
		err := Wrap(inner, WithCode(http.StatusUnauthorized))

		must.Equal(t, http.StatusUnauthorized, err.Code())
	})

	t.Run("positive/wrap nil error does not panic", func(t *testing.T) {
		err := Wrap(nil)

		must.Equal(t, "", err.Error())
		must.Equal(t, 0, err.Code())
	})

	t.Run("positive/unwrap returns underlying error", func(t *testing.T) {
		must.Equal(t, "standalone", New("standalone").Unwrap().Error())
	})
}

func TestDynamicErrors(t *testing.T) {
	t.Run("ErrConfigDecoding", func(t *testing.T) {
		inner := New("inner error")
		err := ErrConfigDecoding(inner)
		must.Equal(t, "config decoding failed: inner error", err.Error())
	})

	t.Run("ErrNotConfigured", func(t *testing.T) {
		err := ErrNotConfigured("database")
		must.Equal(t, "database not configured", err.Error())
	})
}
