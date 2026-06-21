// Package harness provides shared fixtures for RealmKit e2e tests.
package harness

import (
	"bytes"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/realmkit/rk-backend/pkg/api/idempotency"
	"github.com/realmkit/rk-backend/pkg/logger"
	"github.com/realmkit/rk-backend/pkg/server"
	goredis "github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// Ecosystem owns the in-process services used by e2e tests.
type Ecosystem struct {
	// App is the Fiber server under test.
	App *fiber.App

	// Database is the migrated local database fixture.
	Database *Database

	// Log is the structured logger passed to the server.
	Log *zap.Logger

	// LogBuffer captures structured logs for assertions.
	LogBuffer *bytes.Buffer

	// Redis is the isolated Redis fixture backing Redis-dependent middleware.
	Redis *miniredis.Miniredis

	// RedisClient is the client connected to Redis.
	RedisClient *goredis.Client

	// Storage is the in-memory S3-compatible object store.
	Storage *MemoryStorage

	// StorageBucket is the bucket name used by storage-backed services.
	StorageBucket string
}

// Option changes the e2e ecosystem bootstrap.
type Option func(*options)

// WithDatabase replaces the default migrated SQLite fixture.
func WithDatabase(database *Database) Option {
	return func(options *options) {
		options.database = database
	}
}

// WithDevelopment controls development-only server behavior.
func WithDevelopment(development bool) Option {
	return func(options *options) {
		options.development = development
	}
}

// WithLogger replaces the default captured JSON logger.
func WithLogger(log *zap.Logger) Option {
	return func(options *options) {
		options.log = log
		options.logBuffer = nil
	}
}

// WithServerOptions adds server wiring such as module routes.
func WithServerOptions(serverOptions ...server.Option) Option {
	return func(options *options) {
		options.serverOptions = append(options.serverOptions, serverOptions...)
	}
}

// WithStorage replaces the default in-memory storage fixture.
func WithStorage(store *MemoryStorage, bucket string) Option {
	return func(options *options) {
		options.storage = store
		options.storageBucket = bucket
	}
}

// New starts an isolated RealmKit server for e2e tests.
func New(t *testing.T, opts ...Option) *Ecosystem {
	t.Helper()

	options := newOptions(t)
	for _, option := range opts {
		option(options)
	}

	redisServer := miniredis.RunT(t)
	redisClient := goredis.NewClient(&goredis.Options{Addr: redisServer.Addr()})
	t.Cleanup(func() {
		if err := redisClient.Close(); err != nil {
			t.Fatalf("Close() error = %v", err)
		}
		redisServer.Close()
	})

	serverOptions := []server.Option{
		server.WithIdempotencyStore(idempotency.NewRedisStore(
			redisClient,
			idempotency.WithRedisScope("e2e-idempotency"),
		)),
	}
	serverOptions = append(serverOptions, options.serverOptions...)

	return &Ecosystem{
		App:           server.New(options.log, options.development, serverOptions...),
		Database:      options.database,
		Log:           options.log,
		LogBuffer:     options.logBuffer,
		Redis:         redisServer,
		RedisClient:   redisClient,
		Storage:       options.storage,
		StorageBucket: options.storageBucket,
	}
}

// options contains bootstrap settings before dependencies are created.
type options struct {
	database      *Database       // database stores the database value.
	development   bool            // development stores the development value.
	log           *zap.Logger     // log stores the log value.
	logBuffer     *bytes.Buffer   // logBuffer stores the log buffer value.
	serverOptions []server.Option // serverOptions stores the server options value.
	storage       *MemoryStorage  // storage stores the storage value.
	storageBucket string          // storageBucket stores the storage bucket value.
}

// newOptions creates default e2e bootstrap settings.
func newOptions(t *testing.T) *options {
	t.Helper()

	logBuffer := &bytes.Buffer{}
	log, err := logger.New(
		logger.Config{Level: "debug"},
		logger.WithOutput(logBuffer),
		logger.WithErrorOutput(logBuffer),
	)
	if err != nil {
		t.Fatalf("logger.New() error = %v", err)
	}

	return &options{
		database:      NewSQLiteDatabase(t),
		log:           log,
		logBuffer:     logBuffer,
		storage:       NewMemoryStorage(),
		storageBucket: "realmkit-e2e",
	}
}
