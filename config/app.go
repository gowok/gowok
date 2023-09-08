package config

import (
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

type App struct {
	Key  string
	Rest Rest
	Grpc Grpc
}

type Rest struct {
	Enabled bool
	Host    string
	Log     *struct {
		Format        string `yaml:"format"`
		TimeZone      string `yaml:"time_zone"`
		TimeFormat    string `yaml:"time_format"`
		DisableColors bool   `yaml:"disable_colors"`
	} `yaml:"log"`
	Cors *struct {
		AllowOrigins     string `yaml:"allow_origins"`
		AllowCredentials bool   `yaml:"allow_credentials"`
		AllowMethods     string `yaml:"allow_methods"`
		AllowHeaders     string `yaml:"allow_headers"`
		MaxAge           int    `yaml:"max_age"`
		ExposeHeaders    string `yaml:"expose_headers"`
	} `yaml:"cors"`
}

func (r Rest) GetLog() logger.Config {
	c := logger.ConfigDefault
	if r.Log == nil {
		return c
	}
	if r.Log.Format != "" {
		c.Format = r.Log.Format
	}
	if r.Log.TimeZone != "" {
		c.TimeZone = r.Log.TimeZone
	}
	if r.Log.TimeFormat != "" {
		c.TimeFormat = r.Log.TimeFormat
	}
	c.DisableColors = r.Log.DisableColors
	return c
}

func (r Rest) GetCors() cors.Config {
	c := cors.ConfigDefault
	if r.Cors == nil {
		return c
	}
	if r.Cors.AllowOrigins != "" {
		c.AllowOrigins = r.Cors.AllowOrigins
	}
	if r.Cors.AllowMethods != "" {
		c.AllowMethods = r.Cors.AllowMethods
	}
	if r.Cors.AllowHeaders != "" {
		c.AllowHeaders = r.Cors.AllowHeaders
	}
	if r.Cors.ExposeHeaders != "" {
		c.ExposeHeaders = r.Cors.ExposeHeaders
	}
	if r.Cors.AllowCredentials != false {
		c.AllowCredentials = r.Cors.AllowCredentials
	}
	if r.Cors.MaxAge != 0 {
		c.MaxAge = r.Cors.MaxAge
	}
	return c
}

type Grpc struct {
	Enabled bool
	Host    string
}
