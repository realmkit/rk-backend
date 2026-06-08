package postgres

import (
	"net"
	"net/url"
	"strconv"
)

// Config contains PostgreSQL connection settings.
type Config struct {
	// Host is the PostgreSQL server host.
	Host string `mapstructure:"host" default:"localhost"`

	// Port is the PostgreSQL server port.
	Port int `mapstructure:"port" default:"5432"`

	// Database is the PostgreSQL database name.
	Database string `mapstructure:"database"`

	// Username is the PostgreSQL login role.
	Username string `mapstructure:"username"`

	// Password is the PostgreSQL login password.
	Password string `mapstructure:"password"`

	// SSLMode is the PostgreSQL SSL mode.
	SSLMode string `mapstructure:"ssl_mode" default:"disable"`
}

// DSN returns a PostgreSQL connection string for the config.
func (cfg Config) DSN() string {
	dsn := url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(cfg.Username, cfg.Password),
		Host:   net.JoinHostPort(cfg.Host, strconv.Itoa(cfg.Port)),
		Path:   cfg.Database,
	}
	values := dsn.Query()
	values.Set("sslmode", cfg.SSLMode)
	dsn.RawQuery = values.Encode()
	return dsn.String()
}
