package config

import (
	"strings"

	"github.com/realmkit/rk-backend/pkg/api/auth"
	"github.com/realmkit/rk-backend/pkg/api/cors"
	"github.com/realmkit/rk-backend/pkg/logger"
	"github.com/realmkit/rk-backend/pkg/postgres"
	"github.com/realmkit/rk-backend/pkg/redis"
	"github.com/realmkit/rk-backend/pkg/server"
	"github.com/realmkit/rk-backend/pkg/storage"
)

// Config contains the RealmKit backend runtime configuration.
type Config struct {
	// Server contains Fiber HTTP server settings.
	Server server.Config `mapstructure:",squash"`

	// Runtime contains the root REALMKIT runtime settings.
	Runtime Runtime `mapstructure:",squash"`

	// Logging contains JSON logger settings.
	Logging logger.Config `mapstructure:"log"`

	// Postgres contains PostgreSQL connection settings.
	Postgres postgres.Config `mapstructure:"postgres"`

	// Redis contains Redis connection settings.
	Redis redis.Config `mapstructure:"redis"`

	// Storage contains S3-compatible object storage settings.
	Storage storage.Config `mapstructure:"storage"`

	// CORS contains browser cross-origin settings.
	CORS cors.Config `mapstructure:"cors"`

	// Auth contains OAuth and OIDC settings.
	Auth auth.Config `mapstructure:"auth"`
}

// Runtime contains the essential runtime settings required to start RealmKit.
type Runtime struct {
	// Environment is the named runtime environment.
	Environment string `mapstructure:"environment" default:"development"`
}

// IsDevelopment reports whether the backend is running in development mode.
func (runtime Runtime) IsDevelopment() bool {
	return strings.EqualFold(strings.TrimSpace(runtime.Environment), "development")
}
