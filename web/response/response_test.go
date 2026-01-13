package response

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang-must/must"
)

func TestHeader(t *testing.T) {
	cases := []struct {
		Name  string
		Input []string
		Check func(*testing.T, []string, http.Header)
	}{
		{
			"positive",
			[]string{
				"Authorization", "Bearer testing",
				"X-Server", "HTTP Gowok",
			},
			func(t *testing.T, input []string, h http.Header) {
				for i := 0; i <= len(input)/2; i += 2 {
					k := input[i]
					actual := h.Get(k)
					must.NotEqual(t, "", actual)
				}
			},
		},
		{
			"negative zero len",
			[]string{},
			func(t *testing.T, input []string, h http.Header) {
				must.Equal(t, 0, len(h))
			},
		},
		{
			"negative less than 2 len",
			[]string{"Authorization"},
			func(t *testing.T, input []string, h http.Header) {
				must.Equal(t, 0, len(h))
			},
		},
		{
			"negative odd len",
			[]string{
				"Authorization", "Bearer testing",
				"X-Server",
			},
			func(t *testing.T, input []string, h http.Header) {
				for i := 0; i <= len(input)/2; i += 2 {
					k := input[i]
					actual := h.Get(k)
					must.NotEqual(t, "", actual)
				}
			},
		},
		{
			"negative no key",
			[]string{"", "Bearer testing"},
			func(t *testing.T, input []string, h http.Header) {
				must.Equal(t, 0, len(h))
			},
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			r := New(httptest.NewRecorder())
			r.Header(c.Input...)

			c.Check(t, c.Input, r.ToHttp().Header())
		})
	}
}
