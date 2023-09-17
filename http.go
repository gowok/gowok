package gowok

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/pprof"
	"github.com/gowok/gowok/config"
)

func NewHTTP(c *config.Web) *fiber.App {
	h := fiber.New(fiber.Config{
		DisableStartupMessage: true,
	})

	h.Use(logger.New(c.GetLog()))
	h.Use(cors.New(c.GetCors()))

	if c.Pprof != nil && c.Pprof.Enable {
		h.Use(pprof.New(c.GetPprof()))
	}

	return h
}
