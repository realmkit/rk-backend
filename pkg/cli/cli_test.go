package cli

import (
	"bytes"
	"context"
	"errors"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/realmkit/rk-backend/pkg/config"
	"github.com/realmkit/rk-backend/pkg/logger"
	"github.com/realmkit/rk-backend/pkg/postgres"
	"github.com/realmkit/rk-backend/pkg/postgres/migrations"
	"github.com/realmkit/rk-backend/pkg/postgres/seeding"
	realmkitredis "github.com/realmkit/rk-backend/pkg/redis"
	"github.com/realmkit/rk-backend/pkg/server"
	"github.com/realmkit/rk-backend/pkg/storage"
	goredis "github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// TestRootCommandShowsHelpByDefault verifies no-arg execution shows commands.
func TestRootCommandShowsHelpByDefault(t *testing.T) {
	activeLogger := zap.NewNop()
	cmd := newRootCommand(&activeLogger, testCommandDeps(t))
	var output bytes.Buffer
	cmd.SetOut(&output)
	cmd.SetErr(&output)
	cmd.SetArgs(nil)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !strings.Contains(output.String(), "start") || !strings.Contains(output.String(), "seed") ||
		!strings.Contains(output.String(), "forums") {
		t.Fatalf("output = %q, want start, seed, and forums commands", output.String())
	}
}

// TestRootCommandHelpPrintsUsage verifies help is the usage-printing path.
func TestRootCommandHelpPrintsUsage(t *testing.T) {
	activeLogger := zap.NewNop()
	cmd := newRootCommand(&activeLogger, testCommandDeps(t))
	var output bytes.Buffer
	cmd.SetOut(&output)
	cmd.SetErr(&output)
	cmd.SetArgs([]string{"help"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !strings.Contains(output.String(), "Usage:") {
		t.Fatalf("output = %q, want Usage", output.String())
	}
}

// TestRunPrintsHelp verifies the production entry point exposes command help.
func TestRunPrintsHelp(t *testing.T) {
	activeLogger := zap.NewNop()
	if err := Run([]string{"help"}, &activeLogger); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
}

// TestRootCommandErrorDoesNotPrintUsage verifies errors do not include usage output.
func TestRootCommandErrorDoesNotPrintUsage(t *testing.T) {
	activeLogger := zap.NewNop()
	cmd := newRootCommand(&activeLogger, testCommandDeps(t))
	var output bytes.Buffer
	cmd.SetOut(&output)
	cmd.SetErr(&output)
	cmd.SetArgs([]string{"migrate", "repair"})

	if err := cmd.Execute(); err == nil {
		t.Fatalf("Execute() error = nil, want error")
	}
	if strings.Contains(output.String(), "Usage:") {
		t.Fatalf("output = %q, want no Usage", output.String())
	}
}

// TestMigrateStatusReportsPendingMigration verifies migrate status uses the global runner.
func TestMigrateStatusReportsPendingMigration(t *testing.T) {
	activeLogger := zap.NewNop()
	cmd := newRootCommand(&activeLogger, testCommandDeps(t))
	var output bytes.Buffer
	cmd.SetOut(&output)
	cmd.SetErr(&output)
	cmd.SetArgs([]string{"migrate", "status"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !strings.Contains(output.String(), "pending=9") {
		t.Fatalf("output = %q, want pending=9", output.String())
	}
}

// TestMigrateDownRequiresDestructiveConfirmation verifies rollback is guarded.
func TestMigrateDownRequiresDestructiveConfirmation(t *testing.T) {
	activeLogger := zap.NewNop()
	cmd := newRootCommand(&activeLogger, testCommandDeps(t))
	var output bytes.Buffer
	cmd.SetOut(&output)
	cmd.SetErr(&output)
	cmd.SetArgs([]string{"migrate", "down"})

	if err := cmd.Execute(); err == nil {
		t.Fatalf("Execute() error = nil, want error")
	}
	if strings.Contains(output.String(), "Usage:") {
		t.Fatalf("output = %q, want no Usage", output.String())
	}
}

// TestMigrateCommandsApplyValidateAndReset verifies stateful migration commands.
func TestMigrateCommandsApplyValidateAndReset(t *testing.T) {
	db := newCommandDB(t)
	deps := testCommandDepsWithDB(t, db)

	output, err := executeCommand(t, []string{"migrate", "up"}, deps)
	if err != nil {
		t.Fatalf("up Execute() error = %v", err)
	}
	if !strings.Contains(output, "applied=9 pending=0") {
		t.Fatalf("up output = %q, want applied=9 pending=0", output)
	}

	output, err = executeCommand(t, []string{"migrate", "validate"}, deps)
	if err != nil {
		t.Fatalf("validate Execute() error = %v", err)
	}
	if !strings.Contains(output, "applied=9 pending=0") {
		t.Fatalf("validate output = %q, want applied=9 pending=0", output)
	}

	output, err = executeCommand(t, []string{"migrate", "reset", "--i-understand-this-can-destroy-data"}, deps)
	if err != nil {
		t.Fatalf("reset Execute() error = %v", err)
	}
	if !strings.Contains(output, "applied=0 pending=9") {
		t.Fatalf("reset output = %q, want applied=0 pending=9", output)
	}
}

// TestMigrateRepairRunsWithFlags verifies dirty migration repair wiring.
func TestMigrateRepairRunsWithFlags(t *testing.T) {
	db := newCommandDB(t)
	deps := testCommandDepsWithDB(t, db)
	runner := migrations.NewRunner(db, migrations.DefaultSource())
	loaded, err := migrations.Load(migrations.DefaultSource())
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if err := runner.Repair(context.Background(), loaded[0].Version, loaded[0].Checksum, "precondition"); err == nil {
		t.Fatalf("Repair() error = nil, want missing dirty record before setup")
	}
	store := migrations.NewStore(db)
	if err := store.Ensure(context.Background()); err != nil {
		t.Fatalf("Ensure() error = %v", err)
	}
	if err := store.Start(context.Background(), loaded[0], "tester", ""); err != nil {
		t.Fatalf("Start() error = %v", err)
	}

	_, err = executeCommand(
		t,
		[]string{"migrate", "repair", "--version", "1", "--checksum", loaded[0].Checksum, "--reason", "manual"},
		deps,
	)
	if err != nil {
		t.Fatalf("repair Execute() error = %v", err)
	}
}

// TestSeedCommandsApplyAndValidate verifies stateful data seed commands.
func TestSeedCommandsApplyAndValidate(t *testing.T) {
	db := newCommandDB(t)
	deps := testCommandDepsWithDB(t, db)
	if _, err := executeCommand(t, []string{"migrate", "up"}, deps); err != nil {
		t.Fatalf("migrate up error = %v", err)
	}

	output, err := executeCommand(t, []string{"seed", "up"}, deps)
	if err != nil {
		t.Fatalf("seed up error = %v", err)
	}
	if !strings.Contains(output, "applied=3 pending=0") {
		t.Fatalf("seed up output = %q, want applied=3 pending=0", output)
	}

	output, err = executeCommand(t, []string{"seed", "validate"}, deps)
	if err != nil {
		t.Fatalf("seed validate error = %v", err)
	}
	if !strings.Contains(output, "applied=3 pending=0") {
		t.Fatalf("seed validate output = %q, want applied=3 pending=0", output)
	}
}

// TestSeedDryRunReportsPendingSeeds verifies dry-run does not apply seeds.
func TestSeedDryRunReportsPendingSeeds(t *testing.T) {
	db := newCommandDB(t)
	deps := testCommandDepsWithDB(t, db)
	if _, err := executeCommand(t, []string{"migrate", "up"}, deps); err != nil {
		t.Fatalf("migrate up error = %v", err)
	}

	output, err := executeCommand(t, []string{"seed", "dry-run"}, deps)
	if err != nil {
		t.Fatalf("seed dry-run error = %v", err)
	}
	if !strings.Contains(output, "applied=0 pending=3") {
		t.Fatalf("dry-run output = %q, want pending seeds", output)
	}
}

// TestSeedGrantAdminCommandAssignsMembership verifies first-operator grants.
func TestSeedGrantAdminCommandAssignsMembership(t *testing.T) {
	db := newCommandDB(t)
	deps := testCommandDepsWithDB(t, db)
	userID := "00000000-0000-0000-0000-000000000901"
	if _, err := executeCommand(t, []string{"migrate", "up"}, deps); err != nil {
		t.Fatalf("migrate up error = %v", err)
	}
	if _, err := executeCommand(t, []string{"seed", "up"}, deps); err != nil {
		t.Fatalf("seed up error = %v", err)
	}
	insertUser := `
INSERT INTO users(id, status, avatar_asset_id, first_seen_at, last_seen_at, version, created_at, updated_at, deleted_at)
VALUES(?, 'active', NULL, CURRENT_TIMESTAMP, NULL, 1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, NULL)`
	if err := db.Exec(insertUser, userID).Error; err != nil {
		t.Fatalf("insert user error = %v", err)
	}

	output, err := executeCommand(t, []string{"seed", "grant-admin", "--user-id", userID}, deps)
	if err != nil {
		t.Fatalf("grant-admin error = %v", err)
	}
	if !strings.Contains(output, "created=true") {
		t.Fatalf("grant output = %q, want created=true", output)
	}
}

// TestForumStatsVerifyCommandReportsCleanCounters verifies operational stats command wiring.
func TestForumStatsVerifyCommandReportsCleanCounters(t *testing.T) {
	db := newCommandDB(t)
	deps := testCommandDepsWithDB(t, db)
	if _, err := executeCommand(t, []string{"migrate", "up"}, deps); err != nil {
		t.Fatalf("migrate up error = %v", err)
	}

	output, err := executeCommand(t, []string{"forums", "stats", "verify"}, deps)
	if err != nil {
		t.Fatalf("forums stats verify error = %v", err)
	}
	if !strings.Contains(output, "mismatches=0 repaired=false") {
		t.Fatalf("output = %q, want clean stats report", output)
	}
}

// TestForumCacheClearCommandUsesRedis verifies Redis-backed cache clearing command wiring.
func TestForumCacheClearCommandUsesRedis(t *testing.T) {
	server := miniredis.RunT(t)
	db := newCommandDB(t)
	deps := testCommandDepsWithDB(t, db)
	deps.openRedis = func(context.Context, realmkitredis.Config) (*goredis.Client, error) {
		return goredis.NewClient(&goredis.Options{Addr: server.Addr()}), nil
	}
	if _, err := executeCommand(t, []string{"migrate", "up"}, deps); err != nil {
		t.Fatalf("migrate up error = %v", err)
	}

	output, err := executeCommand(t, []string{"forums", "cache", "clear"}, deps)
	if err != nil {
		t.Fatalf("forums cache clear error = %v", err)
	}
	if !strings.Contains(output, "forum caches cleared") {
		t.Fatalf("output = %q, want cache clear message", output)
	}
}

// TestExecuteReturnsStartErrors verifies the start command serves the API.
func TestExecuteReturnsStartErrors(t *testing.T) {
	activeLogger := zap.NewNop()
	want := errors.New("listen failed")
	deps := testCommandDeps(t)
	deps.listenServer = func(*fiber.App, string) error {
		return want
	}

	err := execute(&activeLogger, []string{"start"}, deps)
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
	deps.listenServer = func(_ *fiber.App, address string) error {
		if address != "127.0.0.1:9090" {
			t.Fatalf("address = %q, want %q", address, "127.0.0.1:9090")
		}
		return nil
	}

	if err := execute(&activeLogger, []string{"start"}, deps); err != nil {
		t.Fatalf("execute() error = %v", err)
	}
	if !bytes.Contains(output.Bytes(), []byte("starting realmkit backend")) {
		t.Fatalf("output = %q, want startup log", output.String())
	}
}

// TestRunStartUsesRedisStores verifies Redis-backed stores are wired for startup.
func TestRunStartUsesRedisStores(t *testing.T) {
	activeLogger := zap.NewNop()
	opened := false
	closed := false
	deps := testCommandDeps(t)
	deps.loadConfig = func() (config.Config, error) {
		return config.Config{
			Logging: logger.Config{Level: "info"},
			Redis:   realmkitredis.Config{Address: "localhost:6379"},
		}, nil
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

	if err := execute(&activeLogger, []string{"start"}, deps); err != nil {
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

	err := execute(&activeLogger, []string{"start"}, deps)
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

	err := execute(&activeLogger, []string{"start"}, deps)
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

	if err := execute(&activeLogger, []string{"start"}, deps); err != nil {
		t.Fatalf("execute() error = %v", err)
	}
	messages := observed.FilterMessageSnippet("connection established").All()
	if len(messages) != 3 {
		t.Fatalf("connection logs = %d, want 3", len(messages))
	}
}

// testCommandDeps returns deterministic command dependencies.
func testCommandDeps(t *testing.T) commandDeps {
	t.Helper()
	return commandDeps{
		loadConfig: func() (config.Config, error) {
			return config.Config{Logging: logger.Config{Level: "info"}}, nil
		},
		newLogger: func(logger.Config) (*zap.Logger, error) {
			return zap.NewNop(), nil
		},
		newServer: func(*zap.Logger, bool, ...server.Option) *fiber.App {
			return fiber.New()
		},
		listenServer: func(*fiber.App, string) error {
			return nil
		},
		openPostgres: func(context.Context, postgres.Config) (*gorm.DB, error) {
			db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
			if err != nil {
				t.Fatalf("gorm.Open() error = %v", err)
			}
			return db, nil
		},
		closePostgres: func(*gorm.DB) error {
			return nil
		},
		openRedis: func(context.Context, realmkitredis.Config) (*goredis.Client, error) {
			return goredis.NewClient(&goredis.Options{Addr: "localhost:6379"}), nil
		},
		closeRedis: func(*goredis.Client) error {
			return nil
		},
		newStorage: func(context.Context, storage.Config) (storage.Store, error) {
			return cliStore{}, nil
		},
		newRunner: func(db *gorm.DB, log *zap.Logger) migrations.Runner {
			return migrations.NewRunner(db, migrations.DefaultSource(), migrations.WithLogger(log))
		},
		newSeedRunner: func(db *gorm.DB, log *zap.Logger) seeding.Runner {
			return seeding.NewRunner(db, seeding.DefaultSource(), seeding.WithLogger(log))
		},
	}
}

// cliStore is a fake object store for CLI tests.
type cliStore struct {
	healthError error
}

// Health verifies the storage backend is reachable.
func (store cliStore) Health(context.Context) error {
	return store.healthError
}

// Put stores object bytes.
func (cliStore) Put(context.Context, storage.Object, io.Reader) (storage.StoredObject, error) {
	return storage.StoredObject{}, nil
}

// Delete deletes an object by key.
func (cliStore) Delete(context.Context, string) error {
	return nil
}

// PresignPut creates a presigned upload request.
func (cliStore) PresignPut(context.Context, storage.PresignPutRequest) (storage.PresignedRequest, error) {
	return storage.PresignedRequest{}, nil
}

// PresignGet creates a presigned download URL.
func (cliStore) PresignGet(context.Context, string, time.Duration) (string, error) {
	return "", nil
}

// Head returns object metadata.
func (cliStore) Head(context.Context, string) (storage.ObjectInfo, error) {
	return storage.ObjectInfo{}, nil
}

// executeCommand executes a command and returns captured output.
func executeCommand(t *testing.T, args []string, deps commandDeps) (string, error) {
	t.Helper()
	activeLogger := zap.NewNop()
	cmd := newRootCommand(&activeLogger, deps)
	var output bytes.Buffer
	cmd.SetOut(&output)
	cmd.SetErr(&output)
	cmd.SetArgs(args)
	err := cmd.Execute()
	return output.String(), err
}

// testCommandDepsWithDB returns command dependencies bound to db.
func testCommandDepsWithDB(t *testing.T, db *gorm.DB) commandDeps {
	t.Helper()
	deps := testCommandDeps(t)
	deps.openPostgres = func(context.Context, postgres.Config) (*gorm.DB, error) {
		return db, nil
	}
	deps.closePostgres = func(*gorm.DB) error {
		return nil
	}
	return deps
}

// newCommandDB opens an in-memory command test database.
func newCommandDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("gorm.Open() error = %v", err)
	}
	return db
}
