package server

import (
	"net"
	"strconv"
	"time"
)

// Config contains Fiber HTTP server settings.
type Config struct {
	// Host is the network host the backend binds to.
	Host string `mapstructure:"host" default:"0.0.0.0"`

	// Port is the network port the backend binds to.
	Port int `mapstructure:"port" default:"8080"`

	// ReadTimeout bounds reading HTTP requests.
	ReadTimeout time.Duration `mapstructure:"read_timeout" default:"10s"`

	// WriteTimeout bounds writing HTTP responses.
	WriteTimeout time.Duration `mapstructure:"write_timeout" default:"45s"`

	// IdleTimeout bounds idle keep-alive connections.
	IdleTimeout time.Duration `mapstructure:"idle_timeout" default:"120s"`

	// StartupTimeout bounds dependency checks during server boot.
	StartupTimeout time.Duration `mapstructure:"startup_timeout" default:"30s"`

	// ShutdownTimeout bounds graceful HTTP server shutdown.
	ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout" default:"15s"`

	// RequestTimeout bounds ordinary HTTP request handling.
	RequestTimeout time.Duration `mapstructure:"request_timeout" default:"15s"`

	// AdminRequestTimeout bounds heavier administrative requests.
	AdminRequestTimeout time.Duration `mapstructure:"admin_request_timeout" default:"30s"`

	// UploadRequestTimeout bounds upload and import request handling.
	UploadRequestTimeout time.Duration `mapstructure:"upload_request_timeout" default:"120s"`
}

// Address returns the network address used by the Fiber server.
func (cfg Config) Address() string {
	return net.JoinHostPort(cfg.Host, strconv.Itoa(cfg.Port))
}

// Defaults returns cfg with internal runtime defaults applied.
func (cfg Config) Defaults() Config {
	if cfg.Host == "" {
		cfg.Host = "0.0.0.0"
	}
	if cfg.Port == 0 {
		cfg.Port = 8080
	}
	if cfg.ReadTimeout <= 0 {
		cfg.ReadTimeout = 10 * time.Second
	}
	if cfg.WriteTimeout <= 0 {
		cfg.WriteTimeout = 45 * time.Second
	}
	if cfg.IdleTimeout <= 0 {
		cfg.IdleTimeout = 120 * time.Second
	}
	if cfg.StartupTimeout <= 0 {
		cfg.StartupTimeout = 30 * time.Second
	}
	if cfg.ShutdownTimeout <= 0 {
		cfg.ShutdownTimeout = 15 * time.Second
	}
	if cfg.RequestTimeout <= 0 {
		cfg.RequestTimeout = 15 * time.Second
	}
	if cfg.AdminRequestTimeout <= 0 {
		cfg.AdminRequestTimeout = 30 * time.Second
	}
	if cfg.UploadRequestTimeout <= 0 {
		cfg.UploadRequestTimeout = 120 * time.Second
	}
	return cfg
}
