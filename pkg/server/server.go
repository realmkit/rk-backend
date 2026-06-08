package server

import (
	"github.com/gofiber/contrib/fiberzap/v2"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

// New creates a Fiber application with Zap request logging.
func New(log *zap.Logger, development bool) *fiber.App {
	if log == nil {
		log = zap.NewNop()
	}

	app := fiber.New(fiber.Config{
		DisableStartupMessage: !development,
	})
	app.Use(fiberzap.New(fiberzap.Config{
		Logger: log,
	}))
	app.Get("/health", health)
	return app
}

// health returns the backend health response.
func health(ctx *fiber.Ctx) error {
	return ctx.SendStatus(fiber.StatusNoContent)
}
