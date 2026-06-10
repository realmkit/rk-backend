package cli

import (
	"context"

	"github.com/gofiber/fiber/v2"
	assetspostgres "github.com/niflaot/gamehub-go/module/assets/adapter/postgres"
	assetsapp "github.com/niflaot/gamehub-go/module/assets/application"
	"github.com/niflaot/gamehub-go/pkg/api/idempotency"
	"github.com/niflaot/gamehub-go/pkg/api/ratelimit"
	"github.com/niflaot/gamehub-go/pkg/config"
	"github.com/niflaot/gamehub-go/pkg/logger"
	"github.com/niflaot/gamehub-go/pkg/orm"
	"github.com/niflaot/gamehub-go/pkg/postgres"
	"github.com/niflaot/gamehub-go/pkg/postgres/migrations"
	gamehubredis "github.com/niflaot/gamehub-go/pkg/redis"
	"github.com/niflaot/gamehub-go/pkg/server"
	"github.com/niflaot/gamehub-go/pkg/storage"
	goredis "github.com/redis/go-redis/v9"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// commandDeps contains root command dependencies.
type commandDeps struct {
	loadConfig    func() (config.Config, error)
	newLogger     func(logger.Config) (*zap.Logger, error)
	newServer     func(*zap.Logger, bool, ...server.Option) *fiber.App
	listenServer  func(*fiber.App, string) error
	openPostgres  func(context.Context, postgres.Config) (*gorm.DB, error)
	closePostgres func(*gorm.DB) error
	openRedis     func(context.Context, gamehubredis.Config) (*goredis.Client, error)
	closeRedis    func(*goredis.Client) error
	newStorage    func(context.Context, storage.Config) (storage.Store, error)
	newRunner     func(*gorm.DB, *zap.Logger) migrations.Runner
}

// Run executes the GameHub CLI.
func Run(args []string, activeLogger **zap.Logger) error {
	return execute(activeLogger, args, defaultCommandDeps())
}

// defaultCommandDeps returns production command dependencies.
func defaultCommandDeps() commandDeps {
	return commandDeps{
		loadConfig: func() (config.Config, error) {
			return config.Load()
		},
		newLogger: func(cfg logger.Config) (*zap.Logger, error) {
			return logger.New(cfg)
		},
		newServer: func(log *zap.Logger, development bool, options ...server.Option) *fiber.App {
			return server.New(log, development, options...)
		},
		listenServer: listen,
		openPostgres: func(ctx context.Context, cfg postgres.Config) (*gorm.DB, error) {
			return postgres.Open(ctx, cfg)
		},
		closePostgres: postgres.Close,
		openRedis: func(ctx context.Context, cfg gamehubredis.Config) (*goredis.Client, error) {
			return gamehubredis.Open(ctx, cfg)
		},
		closeRedis: gamehubredis.Close,
		newStorage: func(ctx context.Context, cfg storage.Config) (storage.Store, error) {
			store, err := storage.NewS3Store(ctx, cfg)
			return store, err
		},
		newRunner: func(db *gorm.DB, log *zap.Logger) migrations.Runner {
			return migrations.NewRunner(db, migrations.DefaultSource(), migrations.WithLogger(log), migrations.WithExecutor("gamehub-cli"))
		},
	}
}

// execute executes the root command with dependencies.
func execute(activeLogger **zap.Logger, args []string, deps commandDeps) error {
	cmd := newRootCommand(activeLogger, deps)
	cmd.SetArgs(args)
	return cmd.Execute()
}

// newRootCommand creates the GameHub CLI root command.
func newRootCommand(activeLogger **zap.Logger, deps commandDeps) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "gamehub",
		Short:         "GameHub backend",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return cmd.Help()
		},
	}
	cmd.AddCommand(newStartCommand(activeLogger, deps))
	cmd.AddCommand(newMigrateCommand(activeLogger, deps))
	cmd.AddCommand(newEventsCommand(activeLogger, deps))
	cmd.AddCommand(newCronCommand(activeLogger, deps))
	cmd.AddCommand(newForumsCommand(activeLogger, deps))
	return cmd
}

// newStartCommand creates the start command.
func newStartCommand(activeLogger **zap.Logger, deps commandDeps) *cobra.Command {
	return &cobra.Command{
		Use:           "start",
		Aliases:       []string{"serve"},
		Short:         "Start the HTTP API server",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runStart(cmd.Context(), activeLogger, deps)
		},
	}
}

// runStart starts the HTTP API server.
func runStart(ctx context.Context, activeLogger **zap.Logger, deps commandDeps) error {
	cfg, log, err := runtime(ctx, activeLogger, deps)
	if err != nil {
		return err
	}
	options, closeRuntime, err := runtimeServerOptions(ctx, cfg, log, deps)
	if err != nil {
		return err
	}
	defer closeRuntime(log)
	development := cfg.Runtime.IsDevelopment()
	app := deps.newServer(log, development, options...)
	address := cfg.Server.Address()
	log.Info("starting gamehub backend", zap.String("address", address))
	return deps.listenServer(app, address)
}

// runtimeServerOptions creates server options from runtime dependencies.
func runtimeServerOptions(ctx context.Context, cfg config.Config, log *zap.Logger, deps commandDeps) ([]server.Option, func(*zap.Logger), error) {
	if log == nil {
		log = zap.NewNop()
	}
	options := []server.Option{server.WithCORS(cfg.CORS)}
	client, err := deps.openRedis(ctx, cfg.Redis)
	if err != nil {
		return nil, nil, err
	}
	logDevelopmentConnection(cfg, log, "redis connection established", zap.String("address", cfg.Redis.Address), zap.Int("database", cfg.Redis.Database))
	db, err := deps.openPostgres(ctx, cfg.Postgres)
	if err != nil {
		deps.closeRedis(client)
		return nil, nil, err
	}
	logDevelopmentConnection(cfg, log, "postgres connection established", zap.String("host", cfg.Postgres.Host), zap.Int("port", cfg.Postgres.Port), zap.String("database", cfg.Postgres.Database))
	assetStorage, err := deps.newStorage(ctx, cfg.Storage)
	if err != nil {
		closeDatabase(zap.NewNop(), deps.closePostgres, db)
		deps.closeRedis(client)
		return nil, nil, err
	}
	if err := assetStorage.Health(ctx); err != nil {
		closeDatabase(zap.NewNop(), deps.closePostgres, db)
		deps.closeRedis(client)
		return nil, nil, err
	}
	logDevelopmentConnection(cfg, log, "s3 storage connection established", zap.String("bucket", cfg.Storage.Bucket), zap.String("endpoint", cfg.Storage.Endpoint))
	assetRepository := assetspostgres.NewAssetRepository(orm.NewStore(db))
	eventService := eventsService(db, client)
	assetService := assetsapp.NewService(assetRepository, assetStorage, cfg.Storage.Bucket).WithEvents(eventService)
	groupService := groupsService(db, eventService)
	forumService := forumsService(db, client, assetService, eventService)
	userService := usersService(db, cfg, eventService)
	metadataService := metadataService(db, eventService)
	infraOptions, err := infrastructureOptions(ctx, db, eventService, forumService)
	if err != nil {
		closeDatabase(zap.NewNop(), deps.closePostgres, db)
		deps.closeRedis(client)
		return nil, nil, err
	}
	options = append(options,
		server.WithIdempotencyStore(idempotency.NewRedisStore(client)),
		server.WithRateLimitStore(ratelimit.NewRedisStore(client)),
		server.WithAuth(cfg.Auth, userService),
		server.WithAssets(assetshttpServices(assetService)),
		server.WithGroups(groupshttpServices(groupService)),
		server.WithForums(forumshttpServices(forumService)),
		server.WithMetadata(metadatahttpServices(metadataService)),
		server.WithUsers(usershttpServices(userService, groupService)),
	)
	options = append(options, infraOptions...)
	return options, func(log *zap.Logger) {
		closeDatabase(log, deps.closePostgres, db)
		if err := deps.closeRedis(client); err != nil {
			log.Error("close redis failed", zap.Error(err))
		}
	}, nil
}

// logDevelopmentConnection logs dependency startup success in development.
func logDevelopmentConnection(cfg config.Config, log *zap.Logger, message string, fields ...zap.Field) {
	if cfg.Runtime.IsDevelopment() {
		log.Info(message, fields...)
	}
}

// runtime loads configuration and creates the active logger.
func runtime(_ context.Context, activeLogger **zap.Logger, deps commandDeps) (config.Config, *zap.Logger, error) {
	cfg, err := deps.loadConfig()
	if err != nil {
		return config.Config{}, nil, err
	}
	log, err := deps.newLogger(cfg.Logging)
	if err != nil {
		return config.Config{}, nil, err
	}
	*activeLogger = log
	return cfg, log, nil
}

// closeDatabase closes a database and logs failures.
func closeDatabase(log *zap.Logger, closePostgres func(*gorm.DB) error, db *gorm.DB) {
	if err := closePostgres(db); err != nil {
		log.Error("close postgres failed", zap.Error(err))
	}
}

// listen starts the Fiber application on the configured address.
func listen(app *fiber.App, address string) error {
	return app.Listen(address)
}
