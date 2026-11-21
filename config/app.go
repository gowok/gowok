package config

import (
	"io"
	"log/slog"
	"os"
	"time"

	"github.com/gowok/fp/some"
	"github.com/ngamux/middleware/cors"
	"github.com/ngamux/middleware/log"
)

type Web struct {
	Enabled   bool                    `json:"enabled,omitempty"`
	Host      string                  `json:"host,omitempty"`
	Log       some.Some[WebLog]       `json:"log"`
	Cors      some.Some[WebCors]      `json:"cors"`
	Pprof     some.Some[WebPprof]     `json:"pprof"`
	RateLimit some.Some[WebRateLimit] `json:"rate_limit"`
	RequestID some.Some[WebRequestID] `json:"request_id"`
	Timeout   some.Some[WebTimeout]   `json:"timeout"`
	Views     WebViews                `json:"views"`
	Static    WebStatic               `json:"static"`
}

type WebRequestID struct {
	Enabled bool   `json:"enabled,omitempty"`
	Header  string `json:"header,omitempty"`
}

type WebTimeout struct {
	Enabled  bool   `json:"enabled,omitempty"`
	Duration string `json:"duration,omitempty"`
}

func (r WebTimeout) GetDuration() time.Duration {
	if r.Duration == "" {
		return 0
	}

	duration, err := time.ParseDuration(r.Duration)
	if err != nil || duration <= 0 {
		return 0
	}

	return duration
}

type WebRateLimit struct {
	Enabled  bool     `json:"enabled,omitempty"`
	Requests int      `json:"requests,omitempty"`
	Window   string   `json:"window,omitempty"`
	Burst    int      `json:"burst,omitempty"`
	Exclude  []string `json:"exclude,omitempty"`
}

const defaultRateLimitWindow = time.Minute

func (r WebRateLimit) GetWindow() time.Duration {
	if r.Window == "" {
		return defaultRateLimitWindow
	}

	window, err := time.ParseDuration(r.Window)
	if err != nil || window <= 0 {
		return defaultRateLimitWindow
	}

	return window
}

type WebLog struct {
	Enabled bool `json:"enabled,omitempty"`
}

type WebCors struct {
	Enabled          bool   `json:"enabled,omitempty"`
	AllowOrigins     string `json:"allow_origins,omitempty"`
	AllowCredentials bool   `json:"allow_credentials,omitempty"`
	AllowMethods     string `json:"allow_methods,omitempty"`
	AllowHeaders     string `json:"allow_headers,omitempty"`
	MaxAge           int    `json:"max_age,omitempty"`
	ExposeHeaders    string `json:"expose_headers,omitempty"`
}

type WebPprof struct {
	Enabled bool   `json:"enabled,omitempty"`
	Prefix  string `json:"prefix,omitempty"`
}

type WebViews struct {
	Enabled bool   `json:"enabled,omitempty"`
	Dir     string `json:"dir,omitempty"`
	Layout  string `json:"layout,omitempty"`
}

type WebStatic struct {
	Enabled bool   `json:"enabled,omitempty"`
	Prefix  string `json:"prefix,omitempty"`
	Dir     string `json:"dir,omitempty"`
}

func (r Web) GetLog() log.Config {
	c := log.Config{
		Handler: slog.NewTextHandler(io.Discard, nil),
	}
	if !r.Log.IsPresent() {
		return c
	}

	if cc, ok := r.Log.Get(); ok && cc.Enabled {
		c.Handler = slog.NewJSONHandler(os.Stdout, nil)
	}
	return c
}

func (r Web) GetCors() cors.Config {
	c := cors.Config{}
	cc, ok := r.Cors.Get()
	if !ok {
		return c
	}

	if cc.AllowOrigins != "" {
		c.AllowOrigins = cc.AllowOrigins
	}
	if cc.AllowMethods != "" {
		c.AllowMethods = cc.AllowMethods
	}
	if cc.AllowHeaders != "" {
		c.AllowHeaders = cc.AllowHeaders
	}
	// if r.Cors.ExposeHeaders != "" {
	// 	c.ExposeHeaders = r.Cors.ExposeHeaders
	// }
	// if r.Cors.AllowCredentials != false {
	// 	c.AllowCredentials = r.Cors.AllowCredentials
	// }
	// if r.Cors.MaxAge != 0 {
	// 	c.MaxAge = r.Cors.MaxAge
	// }
	return c
}

func (r Web) GetViews() WebViews {
	v := WebViews{
		Enabled: r.Views.Enabled,
		Layout:  r.Views.Layout,
		Dir:     r.Views.Dir,
	}
	if !v.Enabled {
		return v
	}
	if r.Views.Dir == "" {
		v.Dir = "./views"
	}
	return v
}

func (r Web) GetStatic() WebStatic {
	v := WebStatic{
		Enabled: r.Static.Enabled,
		Dir:     r.Static.Dir,
		Prefix:  "/public",
	}
	if !v.Enabled {
		return v
	}
	if r.Static.Dir == "" {
		v.Dir = "./public"
	}
	if r.Static.Prefix != "" {
		v.Prefix = r.Static.Prefix
	}
	return v
}

type Grpc struct {
	Enabled bool
	Host    string
}
