package config

import (
	"io"
	"log/slog"
	"os"

	"github.com/gowok/gowok/some"
	"github.com/ngamux/middleware/cors"
	"github.com/ngamux/middleware/log"
	"github.com/ngamux/middleware/pprof"
)

type App struct {
	Key  string
	Web  Web
	Grpc Grpc
}

type Web struct {
	Enabled bool
	Host    string

	Log    some.Some[WebLog]   `yaml:"log"`
	Cors   some.Some[WebCors]  `yaml:"cors"`
	Pprof  some.Some[WebPprof] `yaml:"pprof"`
	Views  WebViews            `yaml:"views"`
	Static WebStatic           `yaml:"static"`
}

type WebLog struct {
	Enabled bool `yaml:"enabled"`
}

type WebCors struct {
	Enabled          bool   `yaml:"enabled"`
	AllowOrigins     string `yaml:"allow_origins"`
	AllowCredentials bool   `yaml:"allow_credentials"`
	AllowMethods     string `yaml:"allow_methods"`
	AllowHeaders     string `yaml:"allow_headers"`
	MaxAge           int    `yaml:"max_age"`
	ExposeHeaders    string `yaml:"expose_headers"`
}

type WebPprof struct {
	Enabled bool   `yaml:"enabled"`
	Prefix  string `yaml:"prefix"`
}

type WebViews struct {
	Enabled bool   `yaml:"enabled"`
	Dir     string `yaml:"dir"`
	Layout  string `yaml:"layout"`
}

type WebStatic struct {
	Enabled bool   `yaml:"enabled"`
	Prefix  string `yaml:"prefix"`
	Dir     string `yaml:"dir"`
}

func (r Web) GetLog() log.Config {
	c := log.Config{
		Handler: slog.NewTextHandler(io.Discard, nil),
	}
	if r.Log.IsPresent() {
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

func (r Web) GetPprof() pprof.Config {
	c := pprof.Config{}
	cc, ok := r.Pprof.Get()
	if !ok {
		return c
	}

	if cc.Prefix != "" {
		c.Prefix = cc.Prefix
	}
	return c
}

func (r Web) GetViews() WebViews {
	v := WebViews{
		Enabled: r.Views.Enabled,
		Layout:  r.Views.Layout,
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
