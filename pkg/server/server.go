package server

import (
	"time"

	"github.com/gofiber/contrib/fiberzap/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	assetshttp "github.com/niflaot/gamehub-go/module/assets/adapter/http"
	groupshttp "github.com/niflaot/gamehub-go/module/groups/adapter/http"
	metadatahttp "github.com/niflaot/gamehub-go/module/metadata/adapter/http"
	gamehubcors "github.com/niflaot/gamehub-go/pkg/api/cors"
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
	cors                  gamehubcors.Config
	idempotencyConfigured bool
	idempotencyRedisStore idempotency.RedisStore
	assets                *assetshttp.Services
	groups                *groupshttp.Services
	metadata              *metadatahttp.Services
	rateLimitStore        ratelimit.Store
}

// WithCORS configures browser cross-origin middleware.
func WithCORS(cfg gamehubcors.Config) Option {
	return func(options *options) {
		options.cors = cfg
	}
}

// WithIdempotencyStore configures the Redis idempotency store.
func WithIdempotencyStore(store idempotency.RedisStore) Option {
	return func(options *options) {
		options.idempotencyConfigured = true
		options.idempotencyRedisStore = store
	}
}

// WithAssets registers assets routes with services.
func WithAssets(services assetshttp.Services) Option {
	return func(options *options) {
		options.assets = &services
	}
}

// WithGroups registers groups routes with services.
func WithGroups(services groupshttp.Services) Option {
	return func(options *options) {
		options.groups = &services
	}
}

// WithMetadata registers metadata routes with services.
func WithMetadata(services metadatahttp.Services) Option {
	return func(options *options) {
		options.metadata = &services
	}
}

// WithRateLimitStore configures the rate limit store.
func WithRateLimitStore(store ratelimit.Store) Option {
	return func(options *options) {
		options.rateLimitStore = store
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
	app.Use(gamehubcors.New(options.cors))
	app.Use(fiberzap.New(fiberzap.Config{
		Logger: log,
	}))
	app.Use(ratelimit.Middleware(options.rateLimitStore, DefaultRateLimit))
	app.Use(idempotency.Middleware(options.idempotencyRedisStore, idempotency.WithLogger(log)))
	app.Get("/health", health)
	v1 := versioning.V1.Group(app)
	v1.Use(headers.RequireJSON())
	v1.Get("/health", health)
	if options.assets != nil {
		assetshttp.Register(v1, *options.assets)
	}
	if options.groups != nil {
		groupshttp.Register(v1, *options.groups)
	}
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
	options := options{
		cors:           gamehubcors.Config{Enabled: true, AllowOrigins: "http://localhost:3000,http://127.0.0.1:3000"},
		rateLimitStore: ratelimit.NewMemoryStore(),
	}
	for _, opt := range opts {
		opt(&options)
	}
	if !options.idempotencyConfigured {
		panic("redis idempotency store is required")
	}
	if options.rateLimitStore == nil {
		options.rateLimitStore = ratelimit.NewMemoryStore()
	}
	return options
}
