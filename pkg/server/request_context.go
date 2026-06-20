package server

import (
	"github.com/gofiber/fiber/v2"
	"github.com/realmkit/rk-backend/pkg/api/requestctx"
)

// requestContextMiddleware creates request deadline middleware.
func requestContextMiddleware(config Config) fiber.Handler {
	return requestctx.Middleware(
		config.RequestTimeout,
		requestctx.WithRouteProfile(fiber.MethodGet, "/events/ws", requestctx.Profile{Name: "stream", Timeout: 0}),
		requestctx.WithPathPrefixProfile("", "/events", requestctx.Profile{Name: "admin", Timeout: config.AdminRequestTimeout}),
		requestctx.WithPathPrefixProfile("", "/cronjobs", requestctx.Profile{Name: "admin", Timeout: config.AdminRequestTimeout}),
		requestctx.WithPathPrefixProfile(fiber.MethodPost, "/assets/upload-intents", requestctx.Profile{Name: "upload", Timeout: config.UploadRequestTimeout}),
		requestctx.WithPathPrefixProfile(fiber.MethodPost, "/assets/", requestctx.Profile{Name: "upload", Timeout: config.UploadRequestTimeout}),
		requestctx.WithSkipper(skipContextTimeout),
	)
}

// skipContextTimeout reports whether a request should not receive a deadline.
func skipContextTimeout(ctx *fiber.Ctx) bool {
	return ctx.Path() == "/health"
}
