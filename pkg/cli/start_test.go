package cli

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/realmkit/rk-backend/pkg/config"
	"github.com/realmkit/rk-backend/pkg/logger"
	"github.com/realmkit/rk-backend/pkg/postgres"
	realmkitredis "github.com/realmkit/rk-backend/pkg/redis"
	"github.com/realmkit/rk-backend/pkg/server"
	"github.com/realmkit/rk-backend/pkg/storage"
	goredis "github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

// TestExecuteReturnsStartErrors verifies the start command serves the API.
func TestExecuteReturnsStartErrors(t *testing.T) {
	activeLogger := zap.NewNop()
	want := errors.New("listen failed")
	deps := testCommandDeps(t)
	deps.serveServer = func(context.Context, *fiber.App, string, server.Config) error {
		return want
	}
	err := execute(context.Background(), &activeLogger, []string{"start"}, deps)
	if !errors.Is(err, want) {
		t.Fatalf("execute() error = %v, want %v", err, want)
	}
}

// TestRunStartLogsStartup verifies startup logging uses Zap in every environment.
func TestRunStartLogsStartup(t *testing.T) {
	var output bytes.Buffer
	activeLogger := zap.NewNop()
	cfg := config.Config{
		Server:  server.Config{Host: "127.0.0.1", Port: 9090},
		Runtime: config.Runtime{Environment: "development"},
		Logging: logger.Config{Level: "info"},
	}
	deps := testCommandDeps(t)
	deps.loadConfig = func() (config.Config, error) {
		return cfg, nil
	}
	deps.newLogger = func(cfg logger.Config) (*zap.Logger, error) {
		return logger.New(cfg, logger.WithOutput(&output))
	}
	deps.newServer = func(_ *zap.Logger, development bool, options ...server.Option) *fiber.App {
		if !development {
			t.Fatalf("development = false, want true")
		}
		if len(options) == 0 {
			t.Fatalf("server options = empty, want configured options")
		}
		return fiber.New()
	}
	deps.serveServer = func(_ context.Context, _ *fiber.App, address string, _ server.Config) error {
		if address != "127.0.0.1:9090" {
			t.Fatalf("address = %q, want %q", address, "127.0.0.1:9090")
		}
		return nil
	}
	if err := execute(context.Background(), &activeLogger, []string{"start"}, deps); err != nil {
		t.Fatalf("execute() error = %v", err)
	}
	if !bytes.Contains(output.Bytes(), []byte("starting realmkit backend")) {
		t.Fatalf("output = %q, want startup log", output.String())
	}
}

// TestRunStartPassesRootContextToServer verifies serving uses command context.
func TestRunStartPassesRootContextToServer(t *testing.T) {
	activeLogger := zap.NewNop()
	deps := testCommandDeps(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var served context.Context
	deps.serveServer = func(ctx context.Context, _ *fiber.App, _ string, _ server.Config) error {
		served = ctx
		return nil
	}

	if err := execute(ctx, &activeLogger, []string{"start"}, deps); err != nil {
		t.Fatalf("execute() error = %v", err)
	}
	if served != ctx {
		t.Fatalf("served context = %v, want command context", served)
	}
}

// TestRunStartUsesRedisStores verifies Redis-backed stores are wired for startup.
func TestRunStartUsesRedisStores(t *testing.T) {
	activeLogger := zap.NewNop()
	opened := false
	closed := false
	deps := testCommandDeps(t)
	deps.loadConfig = func() (config.Config, error) {
		return config.Config{Logging: logger.Config{Level: "info"}, Redis: realmkitredis.Config{Address: "localhost:6379"}}, nil
	}
	deps.openRedis = func(context.Context, realmkitredis.Config) (*goredis.Client, error) {
		opened = true
		return goredis.NewClient(&goredis.Options{Addr: "localhost:6379"}), nil
	}
	deps.closeRedis = func(*goredis.Client) error {
		closed = true
		return nil
	}
	deps.newServer = func(_ *zap.Logger, _ bool, options ...server.Option) *fiber.App {
		if len(options) < 3 {
			t.Fatalf("server options = %d, want CORS, idempotency, and rate limit options", len(options))
		}
		return fiber.New()
	}
	if err := execute(context.Background(), &activeLogger, []string{"start"}, deps); err != nil {
		t.Fatalf("execute() error = %v", err)
	}
	if !opened || !closed {
		t.Fatalf("redis opened=%v closed=%v, want both true", opened, closed)
	}
}

// TestRunStartReturnsRedisErrors verifies Redis is mandatory for startup.
func TestRunStartReturnsRedisErrors(t *testing.T) {
	activeLogger := zap.NewNop()
	want := errors.New("redis failed")
	deps := testCommandDeps(t)
	deps.openRedis = func(context.Context, realmkitredis.Config) (*goredis.Client, error) {
		return nil, want
	}
	deps.newServer = func(*zap.Logger, bool, ...server.Option) *fiber.App {
		t.Fatalf("newServer called after redis failure")
		return nil
	}
	err := execute(context.Background(), &activeLogger, []string{"start"}, deps)
	if !errors.Is(err, want) {
		t.Fatalf("execute() error = %v, want %v", err, want)
	}
}

// TestRunStartReturnsStorageHealthErrors verifies S3 health is required for startup.
func TestRunStartReturnsStorageHealthErrors(t *testing.T) {
	activeLogger := zap.NewNop()
	want := errors.New("s3 failed")
	deps := testCommandDeps(t)
	deps.newStorage = func(context.Context, storage.Config) (storage.Store, error) {
		return cliStore{healthError: want}, nil
	}
	deps.newServer = func(*zap.Logger, bool, ...server.Option) *fiber.App {
		t.Fatalf("newServer called after storage health failure")
		return nil
	}
	err := execute(context.Background(), &activeLogger, []string{"start"}, deps)
	if !errors.Is(err, want) {
		t.Fatalf("execute() error = %v, want %v", err, want)
	}
}

// TestRunStartLogsDependencyConnectionsInDevelopment verifies successful dependency logs in development.
func TestRunStartLogsDependencyConnectionsInDevelopment(t *testing.T) {
	var activeLogger *zap.Logger
	core, observed := observer.New(zap.InfoLevel)
	deps := testCommandDeps(t)
	deps.loadConfig = func() (config.Config, error) {
		return config.Config{
			Runtime:  config.Runtime{Environment: "development"},
			Logging:  logger.Config{Level: "info"},
			Postgres: postgres.Config{Host: "localhost", Port: 5432, Database: "realmkit"},
			Redis:    realmkitredis.Config{Address: "localhost:6379"},
			Storage:  storage.Config{Bucket: "realmkit-assets", Endpoint: "http://localhost:9000"},
		}, nil
	}
	deps.newLogger = func(logger.Config) (*zap.Logger, error) {
		return zap.New(core), nil
	}
	if err := execute(context.Background(), &activeLogger, []string{"start"}, deps); err != nil {
		t.Fatalf("execute() error = %v", err)
	}
	messages := observed.FilterMessageSnippet("connection established").All()
	if len(messages) != 3 {
		t.Fatalf("connection logs = %d, want 3", len(messages))
	}
}
