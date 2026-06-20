package cli

import (
	"context"

	"github.com/gofiber/fiber/v2"
	assetspostgres "github.com/realmkit/rk-backend/module/assets/adapter/postgres"
	assetsapp "github.com/realmkit/rk-backend/module/assets/application"
	"github.com/realmkit/rk-backend/pkg/api/idempotency"
	"github.com/realmkit/rk-backend/pkg/api/ratelimit"
	"github.com/realmkit/rk-backend/pkg/cli/seedcmd"
	"github.com/realmkit/rk-backend/pkg/cli/serverrun"
	"github.com/realmkit/rk-backend/pkg/config"
	eventshttp "github.com/realmkit/rk-backend/pkg/events/adapter/http"
	"github.com/realmkit/rk-backend/pkg/logger"
	"github.com/realmkit/rk-backend/pkg/orm"
	"github.com/realmkit/rk-backend/pkg/postgres"
	"github.com/realmkit/rk-backend/pkg/postgres/migrations"
	realmkitredis "github.com/realmkit/rk-backend/pkg/redis"
	"github.com/realmkit/rk-backend/pkg/server"
	"github.com/realmkit/rk-backend/pkg/storage"
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
	serveServer   func(context.Context, *fiber.App, string, server.Config) error
	openPostgres  func(context.Context, postgres.Config) (*gorm.DB, error)
	closePostgres func(*gorm.DB) error
	openRedis     func(context.Context, realmkitredis.Config) (*goredis.Client, error)
	closeRedis    func(*goredis.Client) error
	newStorage    func(context.Context, storage.Config) (storage.Store, error)
	newRunner     func(*gorm.DB, *zap.Logger) migrations.Runner
}

// Run executes the RealmKit CLI.
func Run(ctx context.Context, args []string, activeLogger **zap.Logger) error {
	return execute(ctx, activeLogger, args, defaultCommandDeps())
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
		serveServer: serverrun.Serve,
		openPostgres: func(ctx context.Context, cfg postgres.Config) (*gorm.DB, error) {
			return postgres.Open(ctx, cfg)
		},
		closePostgres: postgres.Close,
		openRedis: func(ctx context.Context, cfg realmkitredis.Config) (*goredis.Client, error) {
			return realmkitredis.Open(ctx, cfg)
		},
		closeRedis: realmkitredis.Close,
		newStorage: func(ctx context.Context, cfg storage.Config) (storage.Store, error) {
			return storage.NewS3Store(ctx, cfg)
		},
		newRunner: func(db *gorm.DB, log *zap.Logger) migrations.Runner {
			return migrations.NewRunner(db, migrations.DefaultSource(), migrations.WithLogger(log), migrations.WithExecutor("realmkit-cli"))
		},
	}
}

// execute executes the root command with dependencies.
func execute(ctx context.Context, activeLogger **zap.Logger, args []string, deps commandDeps) error {
	cmd := newRootCommand(activeLogger, deps)
	cmd.SetContext(ctx)
	cmd.SetArgs(args)
	return cmd.ExecuteContext(ctx)
}

// newRootCommand creates the RealmKit CLI root command.
func newRootCommand(activeLogger **zap.Logger, deps commandDeps) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "realmkit",
		Short:         "RealmKit backend",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return cmd.Help()
		},
	}
	cmd.AddCommand(newStartCommand(activeLogger, deps))
	cmd.AddCommand(newMigrateCommand(activeLogger, deps))
	cmd.AddCommand(seedcmd.New(activeLogger, deps.loadConfig, deps.newLogger, deps.openPostgres, deps.closePostgres))
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
	cfg.Server = cfg.Server.Defaults()
	startupCtx, cancel := context.WithTimeout(ctx, cfg.Server.StartupTimeout)
	defer cancel()
	options, closeRuntime, err := runtimeServerOptions(startupCtx, cfg, log, deps)
	if err != nil {
		return err
	}
	defer closeRuntime(log)
	app := deps.newServer(log, cfg.Runtime.IsDevelopment(), append([]server.Option{server.WithConfig(cfg.Server)}, options...)...)
	log.Info("starting realmkit backend", zap.String("address", cfg.Server.Address()))
	return deps.serveServer(ctx, app, cfg.Server.Address(), cfg.Server)
}

// runtimeServerOptions creates server options from runtime dependencies.
func runtimeServerOptions(
	ctx context.Context,
	cfg config.Config,
	log *zap.Logger,
	deps commandDeps,
) ([]server.Option, func(*zap.Logger), error) {
	if log == nil {
		log = zap.NewNop()
	}
	options := []server.Option{server.WithCORS(cfg.CORS)}
	client, err := deps.openRedis(ctx, cfg.Redis)
	if err != nil {
		return nil, nil, err
	}
	logDevelopmentConnection(
		cfg,
		log,
		"redis connection established",
		zap.String("address", cfg.Redis.Address),
		zap.Int("database", cfg.Redis.Database),
	)
	db, err := deps.openPostgres(ctx, cfg.Postgres)
	if err != nil {
		deps.closeRedis(client)
		return nil, nil, err
	}
	logDevelopmentConnection(
		cfg,
		log,
		"postgres connection established",
		zap.String("host", cfg.Postgres.Host),
		zap.Int("port", cfg.Postgres.Port),
		zap.String("database", cfg.Postgres.Database),
	)
	if err := seedcmd.SeedThemeSigningKeys(ctx, db, cfg); err != nil {
		closeDatabase(zap.NewNop(), deps.closePostgres, db)
		deps.closeRedis(client)
		return nil, nil, err
	}
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
	logDevelopmentConnection(
		cfg,
		log,
		"s3 storage connection established",
		zap.String("bucket", cfg.Storage.Bucket),
		zap.String("endpoint", cfg.Storage.Endpoint),
	)
	eventHub := eventshttp.NewHub()
	eventService := eventsService(db, client, eventHub)
	assetService := assetsapp.NewService(assetspostgres.NewAssetRepository(orm.NewStore(db)), assetStorage, cfg.Storage.Bucket).WithEvents(eventService)
	groupService := groupsService(db, eventService)
	punishmentService := punishmentsService(db, client, eventService)
	forumService := forumsService(db, client, assetService, punishmentService, eventService)
	ticketService := ticketsService(db, client, assetService, punishmentService, eventService)
	userService := usersService(db, cfg, eventService)
	metadataService := metadataService(db, eventService)
	infraOptions, err := infrastructureOptions(ctx, db, eventService, eventHub, groupService, forumService, punishmentService, ticketService)
	if err != nil {
		closeDatabase(zap.NewNop(), deps.closePostgres, db)
		deps.closeRedis(client)
		return nil, nil, err
	}
	options = append(options,
		server.WithIdempotencyStore(idempotency.NewRedisStore(client)),
		server.WithRateLimitStore(ratelimit.NewRedisStore(client)),
		server.WithAuth(cfg.Auth, userService),
		server.WithAssets(assetshttpServices(assetService, groupService)),
		server.WithGroups(groupshttpServices(groupService, userService)),
		server.WithForums(forumshttpServices(forumService)),
		server.WithMetadata(metadatahttpServices(metadataService, groupService)),
		server.WithPunishments(punishmentshttpServices(punishmentService, groupService)),
		server.WithTickets(ticketshttpServices(ticketService, groupService)),
		server.WithUsers(usershttpServices(userService, groupService)),
	)
	return append(options, infraOptions...), func(log *zap.Logger) {
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
