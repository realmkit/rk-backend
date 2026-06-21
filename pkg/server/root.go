package server

import (
	"github.com/gofiber/fiber/v2"
	"github.com/realmkit/rk-backend/pkg/api/swagger"
)

// root returns the service browser entrypoint.
func root(development bool) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		if development {
			return ctx.Redirect(swagger.DocsPath, fiber.StatusFound)
		}
		return ctx.SendStatus(fiber.StatusNoContent)
	}
}
