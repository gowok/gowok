package config

import (
	"testing"

	"github.com/golang-must/must"
	"github.com/gowok/gowok/some"
)

func TestWeb_GetLog(t *testing.T) {
	testCases := []struct {
		name string
		web  Web
	}{
		{
			name: "positive/log is present and enabled",
			web: Web{
				Log: some.Of(WebLog{Enabled: true}),
			},
		},
		{
			name: "positive/log is present but disabled",
			web: Web{
				Log: some.Of(WebLog{Enabled: false}),
			},
		},
		{
			name: "positive/log is not present",
			web: Web{
				Log: some.Empty[WebLog](),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			res := tc.web.GetLog()
			must.NotNil(t, res.Handler)
		})
	}
}

func TestWeb_GetCors(t *testing.T) {
	testCases := []struct {
		name                 string
		web                  Web
		expectedAllowOrigins string
		expectedAllowHeaders string
		expectedAllowMethods string
	}{
		{
			name: "positive/cors is present and has allow origins",
			web: Web{
				Cors: some.Of(WebCors{Enabled: true, AllowOrigins: "*"}),
			},
			expectedAllowOrigins: "*",
		},
		{
			name: "positive/cors is not present",
			web: Web{
				Cors: some.Empty[WebCors](),
			},
			expectedAllowOrigins: "",
		},
		{
			name: "positive/cors with AllowMethods and AllowHeaders",
			web: Web{
				Cors: some.Of(WebCors{
					Enabled:      true,
					AllowOrigins: "http://localhost",
					AllowMethods: "GET,POST",
					AllowHeaders: "X-Header",
				}),
			},
			expectedAllowOrigins: "http://localhost",
			expectedAllowHeaders: "X-Header",
			expectedAllowMethods: "GET,POST",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			res := tc.web.GetCors()
			must.Equal(t, tc.expectedAllowOrigins, res.AllowOrigins)
			must.Equal(t, tc.expectedAllowMethods, res.AllowMethods)
			must.Equal(t, tc.expectedAllowHeaders, res.AllowHeaders)
		})
	}
}

func TestWeb_GetViews(t *testing.T) {
	testCases := []struct {
		name        string
		web         Web
		expectedDir string
	}{
		{
			name: "positive/views enabled with custom dir",
			web: Web{
				Views: WebViews{Enabled: true, Dir: "./custom-views"},
			},
			expectedDir: "./custom-views",
		},
		{
			name: "positive/views enabled with default dir",
			web: Web{
				Views: WebViews{Enabled: true, Dir: ""},
			},
			expectedDir: "./views",
		},
		{
			name: "positive/views disabled",
			web: Web{
				Views: WebViews{Enabled: false},
			},
			expectedDir: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			res := tc.web.GetViews()
			must.Equal(t, tc.expectedDir, res.Dir)
		})
	}
}

func TestWeb_GetStatic(t *testing.T) {
	testCases := []struct {
		name           string
		web            Web
		expectedDir    string
		expectedPrefix string
	}{
		{
			name: "positive/static enabled with custom dir and prefix",
			web: Web{
				Static: WebStatic{Enabled: true, Dir: "./assets", Prefix: "/assets"},
			},
			expectedDir:    "./assets",
			expectedPrefix: "/assets",
		},
		{
			name: "positive/static enabled with default dir and prefix",
			web: Web{
				Static: WebStatic{Enabled: true},
			},
			expectedDir:    "./public",
			expectedPrefix: "/public",
		},
		{
			name: "positive/static disabled",
			web: Web{
				Static: WebStatic{Enabled: false},
			},
			expectedDir:    "",
			expectedPrefix: "/public",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			res := tc.web.GetStatic()
			must.Equal(t, tc.expectedDir, res.Dir)
			must.Equal(t, tc.expectedPrefix, res.Prefix)
		})
	}
}
