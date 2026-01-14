package response

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
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

func TestDownload(t *testing.T) {
	cases := []struct {
		Name  string
		Input func() string
		Check func(*testing.T, string, *httptest.ResponseRecorder)
	}{
		{
			"positive",
			func() string {
				tmpFile, err := os.CreateTemp("", "testing-gowok-*.tmp")
				if err != nil {
					panic(err)
				}
				defer func() {
					_ = tmpFile.Close()
				}()

				_, _ = fmt.Fprint(tmpFile, "123")

				return tmpFile.Name()
			},
			func(t *testing.T, input string, r *httptest.ResponseRecorder) {
				header := r.Header()
				must.Equal(t, "application/octet-stream", header.Get("Content-Type"))
				must.Equal(t, fmt.Sprintf("attachment; filename=%s", path.Base(input)), header.Get("Content-Disposition"))

				body := r.Body.String()
				must.Equal(t, "123", body)
			},
		},
		{
			"negative filename empty",
			func() string {
				return ""
			},
			func(t *testing.T, input string, r *httptest.ResponseRecorder) {
				header := r.Header()
				must.Equal(t, "text/plain", header.Get("Content-Type"))
				must.Equal(t, "", header.Get("Content-Disposition"))

				must.Equal(t, http.StatusNotFound, r.Result().StatusCode)
			},
		},
		{
			"negative file not found",
			func() string {
				return "/gowok/not-found.tmp"
			},
			func(t *testing.T, input string, r *httptest.ResponseRecorder) {
				header := r.Header()
				must.Equal(t, "text/plain", header.Get("Content-Type"))
				must.Equal(t, "", header.Get("Content-Disposition"))

				must.Equal(t, http.StatusNotFound, r.Result().StatusCode)
			},
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			rraw := httptest.NewRecorder()
			r := New(rraw)
			file := c.Input()
			r.Download(file)
			c.Check(t, file, rraw)
		})
	}
}
