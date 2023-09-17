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
	conf = configureHttpViews(*c, conf)

	h := fiber.New(conf)
	h = configureHttpStatic(h, *c)

	h.Use(logger.New(c.GetLog()))
	h.Use(cors.New(c.GetCors()))

	if c.Pprof != nil && c.Pprof.Enabled {
		h.Use(pprof.New(c.GetPprof()))
	}

	return h
}

func configureHttpStatic(app *fiber.App, c config.Web) *fiber.App {
	sc := c.GetStatic()
	if !sc.Enabled {
		return app
	}
	app.Static(sc.Prefix, sc.Dir)
	return app
}

func configureHttpViews(c config.Web, fc fiber.Config) fiber.Config {
	vc := c.GetViews()
	if !vc.Enabled {
		return fc
	}

	v := html.New(vc.Dir, ".html")
	fc.Views = v
	if vc.Layout != "" {
		fc.ViewsLayout = vc.Layout
	}

	sc := c.GetStatic()
	root := "/public"
	v.AddFunc("public", func(path string) string {
		if sc.Enabled {
			root = sc.Prefix
		}
		return root + "/" + path
	})

	return fc
}
