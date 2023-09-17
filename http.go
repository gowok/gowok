package gowok

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/pprof"
	"github.com/gofiber/template/html/v2"
	"github.com/gowok/gowok/config"
)

func NewHTTP(c *config.Web) *fiber.App {
	conf := fiber.Config{
		DisableStartupMessage: true,
	}
	vc := c.GetViews()
	if vc.Enabled {
		v := html.New(vc.Dir, ".html")
		conf.Views = v
		conf.ViewsLayout = vc.Layout
	}
	h := fiber.New(conf)

	h.Use(logger.New(c.GetLog()))
	h.Use(cors.New(c.GetCors()))

	if c.Pprof != nil && c.Pprof.Enabled {
		h.Use(pprof.New(c.GetPprof()))
	}

	return h
}
