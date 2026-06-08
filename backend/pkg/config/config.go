package config

import "github.com/niflaot/gamehub/backend/pkg/logger"

// Config contains the GameHub backend runtime configuration.
type Config struct {
	// Runtime contains the root GAMEHUB runtime settings.
	Runtime Runtime `mapstructure:",squash"`

	// Logging contains JSON logger settings.
	Logging logger.Config `mapstructure:"log"`
}

// Runtime contains the essential runtime settings required to start GameHub.
type Runtime struct {
	// Host is the network host the backend binds to.
	Host string `mapstructure:"host" default:"0.0.0.0"`

	// Port is the network port the backend binds to.
	Port int `mapstructure:"port" default:"8080"`

	// Environment is the named runtime environment.
	Environment string `mapstructure:"environment" default:"development"`
}
