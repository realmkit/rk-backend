package config

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

// TestLoadUsesTaggedDefaults verifies Load returns tag defaults when no external configuration exists.
func TestLoadUsesTaggedDefaults(t *testing.T) {
	clearGameHubEnv(t)
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
	if cfg.Postgres.Database != "gamehub" {
		t.Fatalf("Postgres.Database = %q, want %q", cfg.Postgres.Database, "gamehub")
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
}

// TestLoadReadsGameHubEnvFile verifies GAMEHUB-prefixed values are loaded from .env files.
func TestLoadReadsGameHubEnvFile(t *testing.T) {
	clearGameHubEnv(t)

	path := filepath.Join(t.TempDir(), ".env")
	content := []byte("GAMEHUB_HOST=127.0.0.1\nGAMEHUB_PORT=9090\nGAMEHUB_ENVIRONMENT=test\nGAMEHUB_LOG_LEVEL=debug\nGAMEHUB_POSTGRES_HOST=db\nGAMEHUB_POSTGRES_PORT=5433\nGAMEHUB_POSTGRES_DATABASE=filedb\nGAMEHUB_POSTGRES_USERNAME=fileuser\nGAMEHUB_POSTGRES_PASSWORD=filepass\nGAMEHUB_POSTGRES_SSL_MODE=require\nGAMEHUB_REDIS_ADDRESS=redis:6379\nGAMEHUB_STORAGE_BUCKET=file-bucket\nGAMEHUB_STORAGE_ENDPOINT=http://storage:9000\nGAMEHUB_STORAGE_ACCESS_KEY_ID=file-access\nGAMEHUB_STORAGE_SECRET_ACCESS_KEY=file-secret\nGAMEHUB_CORS_ALLOW_ORIGINS=https://admin.gamehub.test\n")
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
	if cfg.CORS.AllowOrigins != "https://admin.gamehub.test" {
		t.Fatalf("CORS.AllowOrigins = %q, want configured origin", cfg.CORS.AllowOrigins)
	}
}

// TestLoadEnvironmentOverridesEnvFile verifies operating system environment values have highest precedence.
func TestLoadEnvironmentOverridesEnvFile(t *testing.T) {
	clearGameHubEnv(t)

	path := filepath.Join(t.TempDir(), ".env")
	content := []byte("GAMEHUB_HOST=127.0.0.1\nGAMEHUB_PORT=9090\nGAMEHUB_ENVIRONMENT=file\nGAMEHUB_LOG_LEVEL=info\nGAMEHUB_POSTGRES_DATABASE=filedb\nGAMEHUB_POSTGRES_USERNAME=fileuser\nGAMEHUB_POSTGRES_PASSWORD=filepass\nGAMEHUB_STORAGE_BUCKET=file-bucket\nGAMEHUB_STORAGE_ENDPOINT=http://storage:9000\nGAMEHUB_STORAGE_ACCESS_KEY_ID=file-access\nGAMEHUB_STORAGE_SECRET_ACCESS_KEY=file-secret\n")
	if err := os.WriteFile(path, content, 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	t.Setenv("GAMEHUB_HOST", "localhost")
	t.Setenv("GAMEHUB_PORT", "7070")
	t.Setenv("GAMEHUB_ENVIRONMENT", "test")
	t.Setenv("GAMEHUB_LOG_LEVEL", "warn")
	t.Setenv("GAMEHUB_POSTGRES_DATABASE", "envdb")
	t.Setenv("GAMEHUB_POSTGRES_USERNAME", "envuser")
	t.Setenv("GAMEHUB_POSTGRES_PASSWORD", "envpass")
	t.Setenv("GAMEHUB_REDIS_DATABASE", "3")
	t.Setenv("GAMEHUB_STORAGE_BUCKET", "env-bucket")

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
}

// TestLoadWithDisabledEnvFile verifies the loader can run without reading any .env file.
func TestLoadWithDisabledEnvFile(t *testing.T) {
	clearGameHubEnv(t)
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
	clearGameHubEnv(t)

	path := filepath.Join(t.TempDir(), ".env")
	content := []byte("host=127.0.0.2\nport=6060\nenvironment=local\nlog.level=error\npostgres.database=keydb\npostgres.username=keyuser\npostgres.password=keypass\nstorage.bucket=key-bucket\nstorage.endpoint=http://storage:9000\nstorage.access_key_id=key-access\nstorage.secret_access_key=key-secret\n")
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
	clearGameHubEnv(t)
	setRequiredRootEnv(t)

	if _, err := Load(WithEnvFile(t.TempDir())); err == nil {
		t.Fatalf("Load() error = nil, want error")
	}
}

// TestSchemaRequiresFieldsWithoutDefaults verifies a field without a default tag is mandatory.
func TestSchemaRequiresFieldsWithoutDefaults(t *testing.T) {
	clearGameHubEnv(t)

	type requiredConfig struct {
		Token string `mapstructure:"token"`
	}

	fields, err := schemaFor(requiredConfig{})
	if err != nil {
		t.Fatalf("schemaFor() error = %v", err)
	}

	source := newViper(defaultPrefix)
	for _, field := range fields {
		if err := source.BindEnv(field.key, field.env(defaultPrefix)); err != nil {
			t.Fatalf("BindEnv() error = %v", err)
		}
	}

	if err := validateRequired(source, fields); err == nil {
		t.Fatalf("validateRequired() error = nil, want error")
	}
}

// TestValidateRequiredRejectsEmptyString verifies mandatory string settings cannot be blank.
func TestValidateRequiredRejectsEmptyString(t *testing.T) {
	fields := []fieldSpec{{key: "token"}}
	source := newViper(defaultPrefix)
	source.Set("token", " ")

	if err := validateRequired(source, fields); err == nil {
		t.Fatalf("validateRequired() error = nil, want error")
	}
}

// TestValidateRequiredAcceptsPresentValues verifies mandatory settings pass when configured.
func TestValidateRequiredAcceptsPresentValues(t *testing.T) {
	fields := []fieldSpec{{key: "enabled"}}
	source := newViper(defaultPrefix)
	source.Set("enabled", false)

	if err := validateRequired(source, fields); err != nil {
		t.Fatalf("validateRequired() error = %v", err)
	}
}

// clearGameHubEnv clears GAMEHUB variables that can affect loader tests.
func clearGameHubEnv(t *testing.T) {
	t.Helper()
	t.Setenv("GAMEHUB_HOST", "")
	t.Setenv("GAMEHUB_PORT", "")
	t.Setenv("GAMEHUB_ENVIRONMENT", "")
	t.Setenv("GAMEHUB_LOG_LEVEL", "")
	t.Setenv("GAMEHUB_TOKEN", "")
	t.Setenv("GAMEHUB_ENABLED", "")
	t.Setenv("GAMEHUB_POSTGRES_HOST", "")
	t.Setenv("GAMEHUB_POSTGRES_PORT", "")
	t.Setenv("GAMEHUB_POSTGRES_DATABASE", "")
	t.Setenv("GAMEHUB_POSTGRES_USERNAME", "")
	t.Setenv("GAMEHUB_POSTGRES_PASSWORD", "")
	t.Setenv("GAMEHUB_POSTGRES_SSL_MODE", "")
	t.Setenv("GAMEHUB_REDIS_ADDRESS", "")
	t.Setenv("GAMEHUB_REDIS_PASSWORD", "")
	t.Setenv("GAMEHUB_REDIS_DATABASE", "")
	t.Setenv("GAMEHUB_STORAGE_BUCKET", "")
	t.Setenv("GAMEHUB_STORAGE_REGION", "")
	t.Setenv("GAMEHUB_STORAGE_ENDPOINT", "")
	t.Setenv("GAMEHUB_STORAGE_ACCESS_KEY_ID", "")
	t.Setenv("GAMEHUB_STORAGE_SECRET_ACCESS_KEY", "")
	t.Setenv("GAMEHUB_STORAGE_PUBLIC_BASE_URL", "")
	t.Setenv("GAMEHUB_CORS_ENABLED", "")
	t.Setenv("GAMEHUB_CORS_ALLOW_ORIGINS", "")
}

// setRequiredRootEnv sets required root variables for root config tests.
func setRequiredRootEnv(t *testing.T) {
	t.Helper()
	t.Setenv("GAMEHUB_POSTGRES_DATABASE", "gamehub")
	t.Setenv("GAMEHUB_POSTGRES_USERNAME", "gamehub")
	t.Setenv("GAMEHUB_POSTGRES_PASSWORD", "gamehub")
	t.Setenv("GAMEHUB_STORAGE_BUCKET", "gamehub-assets")
	t.Setenv("GAMEHUB_STORAGE_ENDPOINT", "http://localhost:9000")
	t.Setenv("GAMEHUB_STORAGE_ACCESS_KEY_ID", "gamehub")
	t.Setenv("GAMEHUB_STORAGE_SECRET_ACCESS_KEY", "gamehub")
}

// TestSchemaCollectsSquashedFields verifies squashed structs expose root-level GAMEHUB variables.
func TestSchemaCollectsSquashedFields(t *testing.T) {
	fields, err := schemaFor(Config{})
	if err != nil {
		t.Fatalf("schemaFor() error = %v", err)
	}

	got := make([]string, 0, len(fields))
	for _, field := range fields {
		got = append(got, field.key)
	}

	want := []string{"host", "port", "environment", "log.level", "postgres.host", "postgres.port", "postgres.database", "postgres.username", "postgres.password", "postgres.ssl_mode", "redis.address", "redis.password", "redis.database", "storage.bucket", "storage.region", "storage.endpoint", "storage.access_key_id", "storage.secret_access_key", "storage.public_base_url", "cors.enabled", "cors.allow_origins"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("fields = %v, want %v", got, want)
	}
}

// TestRuntimeIsDevelopment verifies development environment matching is normalized.
func TestRuntimeIsDevelopment(t *testing.T) {
	if !(Runtime{Environment: " Development "}).IsDevelopment() {
		t.Fatalf("IsDevelopment() = false, want true")
	}
	if (Runtime{Environment: "production"}).IsDevelopment() {
		t.Fatalf("IsDevelopment() = true, want false")
	}
}

// TestSchemaRejectsNonStructs verifies only struct values can define configuration schemas.
func TestSchemaRejectsNonStructs(t *testing.T) {
	if _, err := schemaFor("invalid"); err == nil {
		t.Fatalf("schemaFor() error = nil, want error")
	}
}

// TestSchemaCollectsNestedAndSkippedFields verifies nested structs, skipped tags, and fallback names.
func TestSchemaCollectsNestedAndSkippedFields(t *testing.T) {
	type databaseConfig struct {
		URL     string `mapstructure:"url" default:"postgres://localhost/gamehub"`
		NoTag   string `default:"fallback"`
		Ignored string `mapstructure:"-"`
	}
	type appConfig struct {
		Database databaseConfig `mapstructure:"database"`
	}

	fields, err := schemaFor(appConfig{})
	if err != nil {
		t.Fatalf("schemaFor() error = %v", err)
	}

	got := make([]string, 0, len(fields))
	for _, field := range fields {
		got = append(got, field.key)
	}

	want := []string{"database.url", "database.no_tag"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("fields = %v, want %v", got, want)
	}
	if fields[0].env(defaultPrefix) != "GAMEHUB_DATABASE_URL" {
		t.Fatalf("env = %q, want %q", fields[0].env(defaultPrefix), "GAMEHUB_DATABASE_URL")
	}
}
