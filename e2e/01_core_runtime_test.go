package e2e

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/realmkit/rk-backend/e2e/harness"
	"github.com/realmkit/rk-backend/pkg/cli"
	"github.com/realmkit/rk-backend/pkg/config"
	pkgredis "github.com/realmkit/rk-backend/pkg/redis"
	"go.uber.org/zap"
)

// TestCoreCLIHelpRuns verifies the root CLI can execute without runtime setup.
func TestCoreCLIHelpRuns(t *testing.T) {
	steps := harness.NewSteps(t)
	steps.Log("execute root CLI help command")
	active := zap.NewNop()
	if err := cli.Run(context.Background(), []string{"--help"}, &active); err != nil {
		t.Fatalf("cli.Run() error = %v", err)
	}
}

// TestCoreConfigLoadsRuntimeSettings verifies required config composes.
func TestCoreConfigLoadsRuntimeSettings(t *testing.T) {
	steps := harness.NewSteps(t)
	steps.Log("set required runtime environment")
	setRequiredConfig(t)

	steps.Log("load configuration without local env file")
	loaded, err := config.Load(config.WithEnvFile(filepath.Join(t.TempDir(), "missing.env")))
	if err != nil {
		t.Fatalf("config.Load() error = %v", err)
	}

	if loaded.Runtime.Environment != "test" {
		t.Fatalf("environment = %q, want test", loaded.Runtime.Environment)
	}
	if loaded.Postgres.Database != "realmkit_e2e" || loaded.Storage.Bucket != "realmkit-e2e" {
		t.Fatalf("config = %+v, want configured database and bucket", loaded)
	}
}

// TestCoreRedisHealth verifies Redis connectivity and failure reporting.
func TestCoreRedisHealth(t *testing.T) {
	steps := harness.NewSteps(t)
	steps.Log("start ecosystem with isolated Redis")
	ecosystem := harness.New(t)

	steps.Log("verify Redis health succeeds")
	if err := pkgredis.Health(context.Background(), ecosystem.RedisClient); err != nil {
		t.Fatalf("redis.Health() error = %v", err)
	}

	steps.Log("verify Redis health reports invalid clients cleanly")
	if err := pkgredis.Health(context.Background(), nil); err == nil {
		t.Fatalf("redis.Health() error = nil, want failure")
	}
}

// setRequiredConfig sets mandatory configuration values.
func setRequiredConfig(t *testing.T) {
	t.Helper()
	t.Setenv("REALMKIT_ENVIRONMENT", "test")
	t.Setenv("REALMKIT_POSTGRES_DATABASE", "realmkit_e2e")
	t.Setenv("REALMKIT_POSTGRES_USERNAME", "realmkit")
	t.Setenv("REALMKIT_POSTGRES_PASSWORD", "secret")
	t.Setenv("REALMKIT_REDIS_ADDRESS", "127.0.0.1:6379")
	t.Setenv("REALMKIT_STORAGE_BUCKET", "realmkit-e2e")
	t.Setenv("REALMKIT_STORAGE_ENDPOINT", "https://storage.e2e")
	t.Setenv("REALMKIT_STORAGE_ACCESS_KEY_ID", "access")
	t.Setenv("REALMKIT_STORAGE_SECRET_ACCESS_KEY", "secret")
	t.Setenv("REALMKIT_AUTH_ISSUER_URL", "https://auth.e2e")
	t.Setenv("REALMKIT_AUTH_AUDIENCE", "realmkit")
	t.Setenv("REALMKIT_AUTH_CLIENT_ID", "frontend")
}
