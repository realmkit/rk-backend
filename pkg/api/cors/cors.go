package cors

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	fibercors "github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/niflaot/gamehub-go/pkg/api/headers"
)

// New returns the configured CORS middleware.
func New(cfg Config) fiber.Handler {
	if !cfg.Enabled {
		return passthrough
	}
	return fibercors.New(fibercors.Config{
		AllowOrigins: normalizedOrigins(cfg.AllowOrigins),
		AllowMethods: strings.Join([]string{
			fiber.MethodGet,
			fiber.MethodPost,
			fiber.MethodPut,
			fiber.MethodPatch,
			fiber.MethodDelete,
			fiber.MethodOptions,
		}, ","),
		AllowHeaders: strings.Join([]string{
			headers.Accept,
			headers.Authorization,
			headers.ContentType,
			headers.CorrelationID,
			headers.IdempotencyKey,
			headers.IfMatch,
			headers.IfNoneMatch,
			headers.RequestID,
		}, ","),
		ExposeHeaders: strings.Join([]string{
			headers.CorrelationID,
			headers.ETag,
			headers.Location,
			headers.RateLimitLimit,
			headers.RateLimitRemaining,
			headers.RateLimitReset,
			headers.RequestID,
			headers.RetryAfter,
		}, ","),
	})
}

// passthrough skips CORS handling.
func passthrough(ctx *fiber.Ctx) error {
	return ctx.Next()
}

// normalizedOrigins trims comma-separated origin entries.
func normalizedOrigins(value string) string {
	parts := strings.Split(value, ",")
	origins := make([]string, 0, len(parts))
	for _, part := range parts {
		origin := strings.TrimSpace(part)
		if origin != "" {
			origins = append(origins, origin)
		}
	}
	return strings.Join(origins, ",")
}
