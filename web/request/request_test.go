package request

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang-must/must"
)

func TestRequest(t *testing.T) {
	testCases := []struct {
		name string
		url  string
	}{
		{
			name: "positive/basic request",
			url:  "http://example.com/foo",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			raw := httptest.NewRequest(http.MethodGet, tc.url, nil)
			req := New(raw)

			must.NotNil(t, req)
			must.Equal(t, raw, req.ToHttp())
		})
	}
}
