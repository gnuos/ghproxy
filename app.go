package main

import (
	_ "embed"
	"net/url"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/gofiber/fiber/v2/middleware/etag"
	"github.com/gofiber/fiber/v2/middleware/favicon"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

//go:embed favicon.ico
var icon []byte

//go:embed home.html
var homePage []byte

func startWeb() {
	setLog(cfg.LogLevel)

	app := fiber.New()

	app.Use(recover.New())
	app.Use(etag.New())

	app.Use(logger.New(logger.Config{
		TimeFormat: time.RFC3339Nano,
		TimeZone:   "Asia/Shanghai",
	}))

	app.Use(favicon.New(favicon.Config{
		Data: icon,
	}))

	app.Use("/*", func(c *fiber.Ctx) error {
		c.Locals("config", *cfg)

		return c.Next()
	})

	app.Head("*", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})
	app.Get("/", func(c *fiber.Ctx) error {
		q := c.Query("q")
		if q != "" {
			if _, err := url.Parse(q); err == nil {
				return c.Redirect(q)
			}
		}

		c.Set(fiber.HeaderContentType, fiber.MIMETextHTML)

		return c.Send(homePage)
	})

	app.Get("/*", GitHubProxyHandler)
	app.Post("/*", GitHubProxyHandler)

	log.Fatal(app.Listen(cfg.Listen))
}
