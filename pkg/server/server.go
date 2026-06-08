package server

import (
	"time"

	"github.com/gofiber/contrib/fiberzap/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	metadatahttp "github.com/niflaot/gamehub-go/module/metadata/adapter/http"
	"github.com/niflaot/gamehub-go/pkg/api/headers"
	"github.com/niflaot/gamehub-go/pkg/api/idempotency"
	"github.com/niflaot/gamehub-go/pkg/api/problem"
	"github.com/niflaot/gamehub-go/pkg/api/ratelimit"
	"github.com/niflaot/gamehub-go/pkg/api/swagger"
	"github.com/niflaot/gamehub-go/pkg/api/versioning"
	"go.uber.org/zap"
)

// DefaultRateLimit is the server's initial local rate limit policy.
var DefaultRateLimit = ratelimit.Policy{Limit: 1000, Window: time.Minute}

// Option configures the HTTP server.
type Option func(*options)

// options contains optional server modules.
type options struct {
	metadata *metadatahttp.Services
}

// WithMetadata registers metadata routes with services.
func WithMetadata(services metadatahttp.Services) Option {
	return func(options *options) {
		options.metadata = &services
	}
}

// New creates a Fiber application with Zap request logging.
func New(log *zap.Logger, development bool, opts ...Option) *fiber.App {
	if log == nil {
		log = zap.NewNop()
	}
	options := optionsFrom(opts)

	app := fiber.New(fiber.Config{
		DisableStartupMessage: !development,
		ErrorHandler:          problem.Handler,
	})
	app.Use(recover.New())
	app.Use(headers.Middleware())
	app.Use(fiberzap.New(fiberzap.Config{
		Logger: log,
	}))
	app.Use(ratelimit.Middleware(ratelimit.NewMemoryStore(), DefaultRateLimit))
	app.Use(idempotency.Middleware(idempotency.NewMemoryStore()))
	app.Get("/health", health)
	v1 := versioning.V1.Group(app)
	v1.Use(headers.RequireJSON())
	v1.Get("/health", health)
	if options.metadata != nil {
		metadatahttp.Register(v1, *options.metadata)
	}
	swagger.Register(app, development)
	return app
}

// health returns the backend health response.
func health(ctx *fiber.Ctx) error {
	return ctx.SendStatus(fiber.StatusNoContent)
}

// optionsFrom applies server options.
func optionsFrom(opts []Option) options {
	var options options
	for _, opt := range opts {
		opt(&options)
	}
	return options
}
