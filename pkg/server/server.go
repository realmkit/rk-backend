package server

import (
	"time"

	"github.com/gofiber/contrib/fiberzap/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	assetshttp "github.com/realmkit/rk-backend/module/assets/adapter/http"
	forumshttp "github.com/realmkit/rk-backend/module/forums/adapter/http"
	groupshttp "github.com/realmkit/rk-backend/module/groups/adapter/http"
	metadatahttp "github.com/realmkit/rk-backend/module/metadata/adapter/http"
	punishmentshttp "github.com/realmkit/rk-backend/module/punishments/adapter/http"
	ticketshttp "github.com/realmkit/rk-backend/module/tickets/adapter/http"
	userhttp "github.com/realmkit/rk-backend/module/user/adapter/http"
	"github.com/realmkit/rk-backend/pkg/api/auth"
	realmkitcors "github.com/realmkit/rk-backend/pkg/api/cors"
	"github.com/realmkit/rk-backend/pkg/api/headers"
	"github.com/realmkit/rk-backend/pkg/api/idempotency"
	"github.com/realmkit/rk-backend/pkg/api/problem"
	"github.com/realmkit/rk-backend/pkg/api/ratelimit"
	"github.com/realmkit/rk-backend/pkg/api/swagger"
	cronhttp "github.com/realmkit/rk-backend/pkg/cronjob/adapter/http"
	eventshttp "github.com/realmkit/rk-backend/pkg/events/adapter/http"
	"go.uber.org/zap"
)

// DefaultRateLimit is the server's initial local rate limit policy.
var DefaultRateLimit = ratelimit.Policy{Limit: 1000, Window: time.Minute}

// DefaultReadBufferSize allows local auth cookies and provider headers.
const DefaultReadBufferSize = 64 * 1024

// Option configures the HTTP server.
type Option func(*options)

// options contains optional server modules.
type options struct {
	cors                  realmkitcors.Config
	idempotencyConfigured bool
	idempotencyStore      idempotency.Store
	assets                *assetshttp.Services
	auth                  *auth.Config
	authProvisioner       auth.Provisioner
	cron                  *cronhttp.Services
	events                *eventshttp.Services
	forums                *forumshttp.Services
	groups                *groupshttp.Services
	metadata              *metadatahttp.Services
	punishments           *punishmentshttp.Services
	rateLimitStore        ratelimit.Store
	tickets               *ticketshttp.Services
	users                 *userhttp.Services
}

// WithEvents registers event routes with services.
func WithEvents(services eventshttp.Services) Option {
	return func(options *options) {
		options.events = &services
	}
}

// WithCron registers cron job routes with services.
func WithCron(services cronhttp.Services) Option {
	return func(options *options) {
		options.cron = &services
	}
}

// WithCORS configures browser cross-origin middleware.
func WithCORS(cfg realmkitcors.Config) Option {
	return func(options *options) {
		options.cors = cfg
	}
}

// WithIdempotencyStore configures the Redis idempotency store.
func WithIdempotencyStore(store idempotency.RedisStore) Option {
	return func(options *options) {
		options.idempotencyConfigured = true
		options.idempotencyStore = store
	}
}

// WithAssets registers assets routes with services.
func WithAssets(services assetshttp.Services) Option {
	return func(options *options) {
		options.assets = &services
	}
}

// WithForums registers forum routes with services.
func WithForums(services forumshttp.Services) Option {
	return func(options *options) {
		options.forums = &services
	}
}

// WithAuth configures public auth routes and protected-route middleware.
func WithAuth(config auth.Config, provisioner auth.Provisioner) Option {
	return func(options *options) {
		options.auth = &config
		options.authProvisioner = provisioner
	}
}

// WithGroups registers groups routes with services.
func WithGroups(services groupshttp.Services) Option {
	return func(options *options) {
		options.groups = &services
	}
}

// WithUsers registers user routes with services.
func WithUsers(services userhttp.Services) Option {
	return func(options *options) {
		options.users = &services
	}
}

// WithMetadata registers metadata routes with services.
func WithMetadata(services metadatahttp.Services) Option {
	return func(options *options) {
		options.metadata = &services
	}
}

// WithPunishments registers punishment routes with services.
func WithPunishments(services punishmentshttp.Services) Option {
	return func(options *options) {
		options.punishments = &services
	}
}

// WithTickets registers ticket routes with services.
func WithTickets(services ticketshttp.Services) Option {
	return func(options *options) {
		options.tickets = &services
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
		ReadBufferSize:        DefaultReadBufferSize,
	})
	app.Use(recover.New())
	app.Use(headers.Middleware())
	app.Use(realmkitcors.New(options.cors))
	app.Use(fiberzap.New(fiberzap.Config{
		Logger: log,
	}))
	app.Use(ratelimit.Middleware(options.rateLimitStore, DefaultRateLimit))
	app.Use(idempotency.Middleware(options.idempotencyStore, idempotency.WithLogger(log)))
	app.Get("/health", health)
	swagger.Register(app, development)

	api := app.Group("", headers.RequireJSON())
	auths := options.authHandlers(log, development)
	moduleAPI := api
	if auths.configured {
		auth.Register(api, *options.auth)
		moduleAPI = api.Group("", auths.optional)
	}
	if options.assets != nil {
		assetshttp.Register(moduleAPI, *options.assets)
	}
	if options.events != nil {
		eventshttp.Register(moduleAPI, *options.events)
	}
	if options.cron != nil {
		cronhttp.Register(moduleAPI, *options.cron)
	}
	if options.groups != nil {
		groupshttp.Register(moduleAPI, *options.groups)
	}
	if options.forums != nil {
		forumshttp.Register(moduleAPI, *options.forums)
	}
	if options.users != nil && auths.configured {
		userhttp.Register(api, *options.users, auths.required)
	}
	if options.metadata != nil {
		metadatahttp.Register(moduleAPI, *options.metadata)
	}
	if options.punishments != nil {
		punishmentshttp.Register(moduleAPI, *options.punishments)
	}
	if options.tickets != nil {
		ticketshttp.Register(moduleAPI, *options.tickets)
	}
	return app
}

// health returns the backend health response.
func health(ctx *fiber.Ctx) error {
	return ctx.SendStatus(fiber.StatusNoContent)
}

// optionsFrom applies server options.
func optionsFrom(opts []Option) options {
	options := options{
		cors:           realmkitcors.Config{Enabled: true, AllowOrigins: "http://localhost:3000,http://127.0.0.1:3000"},
		rateLimitStore: ratelimit.NewMemoryStore(),
	}
	for _, opt := range opts {
		opt(&options)
	}
	if !options.idempotencyConfigured {
		options.idempotencyStore = idempotency.NewMemoryStore()
	}
	if options.rateLimitStore == nil {
		options.rateLimitStore = ratelimit.NewMemoryStore()
	}
	return options
}
