package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestLoadUsesTaggedDefaults verifies Load returns tag defaults when no external configuration exists.
func TestLoadUsesTaggedDefaults(t *testing.T) {
	clearRealmKitEnv(t)
	setRequiredRootEnv(t)

	cfg, err := Load(WithEnvFile(filepath.Join(t.TempDir(), ".env")))
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Server.Host != "0.0.0.0" {
		t.Fatalf("Host = %q, want %q", cfg.Server.Host, "0.0.0.0")
	}
	if cfg.Server.Port != 8080 {
		t.Fatalf("Port = %d, want %d", cfg.Server.Port, 8080)
	}
	if cfg.Runtime.Environment != "development" {
		t.Fatalf("Environment = %q, want %q", cfg.Runtime.Environment, "development")
	}
	if cfg.Logging.Level != "info" {
		t.Fatalf("Level = %q, want %q", cfg.Logging.Level, "info")
	}
	if cfg.Postgres.Host != "localhost" {
		t.Fatalf("Postgres.Host = %q, want %q", cfg.Postgres.Host, "localhost")
	}
	if cfg.Postgres.Port != 5432 {
		t.Fatalf("Postgres.Port = %d, want %d", cfg.Postgres.Port, 5432)
	}
	if cfg.Postgres.Database != "realmkit" {
		t.Fatalf("Postgres.Database = %q, want %q", cfg.Postgres.Database, "realmkit")
	}
	if cfg.Redis.Address != "localhost:6379" {
		t.Fatalf("Redis.Address = %q, want %q", cfg.Redis.Address, "localhost:6379")
	}
	if cfg.Storage.Region != "auto" {
		t.Fatalf("Storage.Region = %q, want %q", cfg.Storage.Region, "auto")
	}
	if !cfg.CORS.Enabled {
		t.Fatalf("CORS.Enabled = false, want true")
	}
	if cfg.Auth.Provider != "generic_oidc" {
		t.Fatalf("Auth.Provider = %q, want generic_oidc", cfg.Auth.Provider)
	}
	if cfg.Server.RequestTimeout != 15*time.Second {
		t.Fatalf("RequestTimeout = %s, want 15s", cfg.Server.RequestTimeout)
	}
	if cfg.Redis.DialTimeout != 5*time.Second {
		t.Fatalf("Redis.DialTimeout = %s, want 5s", cfg.Redis.DialTimeout)
	}
	if cfg.Auth.HTTPTimeout != 5*time.Second {
		t.Fatalf("Auth.HTTPTimeout = %s, want 5s", cfg.Auth.HTTPTimeout)
	}
}

// TestLoadReadsRealmKitEnvFile verifies REALMKIT-prefixed values are loaded from .env files.
func TestLoadReadsRealmKitEnvFile(t *testing.T) {
	clearRealmKitEnv(t)

	path := filepath.Join(t.TempDir(), ".env")
	content := []byte(
		"REALMKIT_HOST=127.0.0.1\nREALMKIT_PORT=9090\nREALMKIT_ENVIRONMENT=test\nREALMKIT_LOG_LEVEL=debug\nREALMKIT_POSTGRES_HOST=db\nREALMKIT_POSTGRES_PORT=5433\nREALMKIT_POSTGRES_DATABASE=filedb\nREALMKIT_POSTGRES_USERNAME=fileuser\nREALMKIT_POSTGRES_PASSWORD=filepass\nREALMKIT_POSTGRES_SSL_MODE=require\nREALMKIT_REDIS_ADDRESS=redis:6379\nREALMKIT_STORAGE_BUCKET=file-bucket\nREALMKIT_STORAGE_ENDPOINT=http://storage:9000\nREALMKIT_STORAGE_ACCESS_KEY_ID=file-access\nREALMKIT_STORAGE_SECRET_ACCESS_KEY=file-secret\nREALMKIT_CORS_ALLOW_ORIGINS=https://admin.realmkit.test\nREALMKIT_AUTH_ISSUER_URL=https://auth.example.test\nREALMKIT_AUTH_AUDIENCE=file-api\nREALMKIT_AUTH_CLIENT_ID=file-frontend\n",
	)
	if err := os.WriteFile(path, content, 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	cfg, err := Load(WithEnvFile(path))
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Server.Host != "127.0.0.1" {
		t.Fatalf("Host = %q, want %q", cfg.Server.Host, "127.0.0.1")
	}
	if cfg.Server.Port != 9090 {
		t.Fatalf("Port = %d, want %d", cfg.Server.Port, 9090)
	}
	if cfg.Runtime.Environment != "test" {
		t.Fatalf("Environment = %q, want %q", cfg.Runtime.Environment, "test")
	}
	if cfg.Logging.Level != "debug" {
		t.Fatalf("Level = %q, want %q", cfg.Logging.Level, "debug")
	}
	if cfg.Postgres.Host != "db" {
		t.Fatalf("Postgres.Host = %q, want %q", cfg.Postgres.Host, "db")
	}
	if cfg.Postgres.Port != 5433 {
		t.Fatalf("Postgres.Port = %d, want %d", cfg.Postgres.Port, 5433)
	}
	if cfg.Postgres.SSLMode != "require" {
		t.Fatalf("Postgres.SSLMode = %q, want %q", cfg.Postgres.SSLMode, "require")
	}
	if cfg.Redis.Address != "redis:6379" {
		t.Fatalf("Redis.Address = %q, want %q", cfg.Redis.Address, "redis:6379")
	}
	if cfg.Storage.Bucket != "file-bucket" {
		t.Fatalf("Storage.Bucket = %q, want file-bucket", cfg.Storage.Bucket)
	}
	if cfg.CORS.AllowOrigins != "https://admin.realmkit.test" {
		t.Fatalf("CORS.AllowOrigins = %q, want configured origin", cfg.CORS.AllowOrigins)
	}
	if cfg.Auth.IssuerURL != "https://auth.example.test" {
		t.Fatalf("Auth.IssuerURL = %q, want configured issuer", cfg.Auth.IssuerURL)
	}
}

// TestLoadEnvironmentOverridesEnvFile verifies operating system environment values have highest precedence.
func TestLoadEnvironmentOverridesEnvFile(t *testing.T) {
	clearRealmKitEnv(t)

	path := filepath.Join(t.TempDir(), ".env")
	content := []byte(
		"REALMKIT_HOST=127.0.0.1\nREALMKIT_PORT=9090\nREALMKIT_ENVIRONMENT=file\nREALMKIT_LOG_LEVEL=info\nREALMKIT_POSTGRES_DATABASE=filedb\nREALMKIT_POSTGRES_USERNAME=fileuser\nREALMKIT_POSTGRES_PASSWORD=filepass\nREALMKIT_STORAGE_BUCKET=file-bucket\nREALMKIT_STORAGE_ENDPOINT=http://storage:9000\nREALMKIT_STORAGE_ACCESS_KEY_ID=file-access\nREALMKIT_STORAGE_SECRET_ACCESS_KEY=file-secret\nREALMKIT_AUTH_ISSUER_URL=https://file-auth.example.test\nREALMKIT_AUTH_AUDIENCE=file-api\nREALMKIT_AUTH_CLIENT_ID=file-frontend\n",
	)
	if err := os.WriteFile(path, content, 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	t.Setenv("REALMKIT_HOST", "localhost")
	t.Setenv("REALMKIT_PORT", "7070")
	t.Setenv("REALMKIT_ENVIRONMENT", "test")
	t.Setenv("REALMKIT_LOG_LEVEL", "warn")
	t.Setenv("REALMKIT_POSTGRES_DATABASE", "envdb")
	t.Setenv("REALMKIT_POSTGRES_USERNAME", "envuser")
	t.Setenv("REALMKIT_POSTGRES_PASSWORD", "envpass")
	t.Setenv("REALMKIT_REDIS_DATABASE", "3")
	t.Setenv("REALMKIT_STORAGE_BUCKET", "env-bucket")
	t.Setenv("REALMKIT_AUTH_AUDIENCE", "env-api")

	cfg, err := Load(WithEnvFile(path))
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Server.Host != "localhost" {
		t.Fatalf("Host = %q, want %q", cfg.Server.Host, "localhost")
	}
	if cfg.Server.Port != 7070 {
		t.Fatalf("Port = %d, want %d", cfg.Server.Port, 7070)
	}
	if cfg.Runtime.Environment != "test" {
		t.Fatalf("Environment = %q, want %q", cfg.Runtime.Environment, "test")
	}
	if cfg.Logging.Level != "warn" {
		t.Fatalf("Level = %q, want %q", cfg.Logging.Level, "warn")
	}
	if cfg.Postgres.Database != "envdb" {
		t.Fatalf("Postgres.Database = %q, want %q", cfg.Postgres.Database, "envdb")
	}
	if cfg.Redis.Database != 3 {
		t.Fatalf("Redis.Database = %d, want %d", cfg.Redis.Database, 3)
	}
	if cfg.Storage.Bucket != "env-bucket" {
		t.Fatalf("Storage.Bucket = %q, want env-bucket", cfg.Storage.Bucket)
	}
	if cfg.Auth.Audience != "env-api" {
		t.Fatalf("Auth.Audience = %q, want env-api", cfg.Auth.Audience)
	}
}

// TestLoadWithDisabledEnvFile verifies the loader can run without reading any .env file.
func TestLoadWithDisabledEnvFile(t *testing.T) {
	clearRealmKitEnv(t)
	setRequiredRootEnv(t)

	cfg, err := Load(WithEnvFile(""))
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Server.Port != 8080 {
		t.Fatalf("Port = %d, want %d", cfg.Server.Port, 8080)
	}
}

// TestLoadReadsUnprefixedEnvFileKeys verifies .env files may use config keys directly.
func TestLoadReadsUnprefixedEnvFileKeys(t *testing.T) {
	clearRealmKitEnv(t)

	path := filepath.Join(t.TempDir(), ".env")
	content := []byte(
		"host=127.0.0.2\nport=6060\nenvironment=local\nlog.level=error\npostgres.database=keydb\npostgres.username=keyuser\npostgres.password=keypass\nstorage.bucket=key-bucket\nstorage.endpoint=http://storage:9000\nstorage.access_key_id=key-access\nstorage.secret_access_key=key-secret\nauth.issuer_url=https://key-auth.example.test\nauth.audience=key-api\nauth.client_id=key-frontend\n",
	)
	if err := os.WriteFile(path, content, 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	cfg, err := Load(WithEnvFile(path))
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Server.Host != "127.0.0.2" {
		t.Fatalf("Host = %q, want %q", cfg.Server.Host, "127.0.0.2")
	}
	if cfg.Server.Port != 6060 {
		t.Fatalf("Port = %d, want %d", cfg.Server.Port, 6060)
	}
	if cfg.Runtime.Environment != "local" {
		t.Fatalf("Environment = %q, want %q", cfg.Runtime.Environment, "local")
	}
	if cfg.Logging.Level != "error" {
		t.Fatalf("Level = %q, want %q", cfg.Logging.Level, "error")
	}
	if cfg.Postgres.Username != "keyuser" {
		t.Fatalf("Postgres.Username = %q, want %q", cfg.Postgres.Username, "keyuser")
	}
	if cfg.Storage.Bucket != "key-bucket" {
		t.Fatalf("Storage.Bucket = %q, want key-bucket", cfg.Storage.Bucket)
	}
}

// TestLoadReturnsEnvFileReadErrors verifies invalid .env paths return useful errors.
func TestLoadReturnsEnvFileReadErrors(t *testing.T) {
	clearRealmKitEnv(t)
	setRequiredRootEnv(t)

	if _, err := Load(WithEnvFile(t.TempDir())); err == nil {
		t.Fatalf("Load() error = nil, want error")
	}
}
