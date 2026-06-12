package server

import (
	"github.com/gofiber/fiber/v2"
	"github.com/realmkit/rk-backend/pkg/api/auth"
	"go.uber.org/zap"
)

// authHandlers contains required and optional auth middleware.
type authHandlers struct {
	required   fiber.Handler
	optional   fiber.Handler
	configured bool
}

// authHandlers returns shared auth middleware for server route composition.
func (options options) authHandlers(log *zap.Logger, development bool) authHandlers {
	if options.auth == nil || options.authProvisioner == nil {
		return authHandlers{}
	}
	validator := auth.NewValidator(*options.auth)
	settings := auth.MiddlewareConfig{Development: development, Log: log}
	return authHandlers{
		required:   auth.Middleware(*options.auth, validator, options.authProvisioner, settings),
		optional:   auth.OptionalMiddleware(*options.auth, validator, options.authProvisioner, settings),
		configured: true,
	}
}
