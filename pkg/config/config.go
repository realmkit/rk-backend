package config

import (
	"strings"

	"github.com/niflaot/gamehub/backend/pkg/logger"
	"github.com/niflaot/gamehub/backend/pkg/postgres"
	"github.com/niflaot/gamehub/backend/pkg/server"
)

// Config contains the GameHub backend runtime configuration.
type Config struct {
	// Server contains Fiber HTTP server settings.
	Server server.Config `mapstructure:",squash"`

	// Runtime contains the root GAMEHUB runtime settings.
	Runtime Runtime `mapstructure:",squash"`

	// Logging contains JSON logger settings.
	Logging logger.Config `mapstructure:"log"`

	// Postgres contains PostgreSQL connection settings.
	Postgres postgres.Config `mapstructure:"postgres"`
}

// Runtime contains the essential runtime settings required to start GameHub.
type Runtime struct {
	// Environment is the named runtime environment.
	Environment string `mapstructure:"environment" default:"development"`
}

// IsDevelopment reports whether the backend is running in development mode.
func (runtime Runtime) IsDevelopment() bool {
	return strings.EqualFold(strings.TrimSpace(runtime.Environment), "development")
}
