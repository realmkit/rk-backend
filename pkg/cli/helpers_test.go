package cli

import (
	"bytes"
	"context"
	"io"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/realmkit/rk-backend/pkg/config"
	"github.com/realmkit/rk-backend/pkg/logger"
	"github.com/realmkit/rk-backend/pkg/postgres"
	"github.com/realmkit/rk-backend/pkg/postgres/migrations"
	realmkitredis "github.com/realmkit/rk-backend/pkg/redis"
	"github.com/realmkit/rk-backend/pkg/server"
	"github.com/realmkit/rk-backend/pkg/storage"
	goredis "github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

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
