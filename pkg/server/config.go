package server

import (
	"net"
	"strconv"
)

// Config contains Fiber HTTP server settings.
type Config struct {
	// Host is the network host the backend binds to.
	Host string `mapstructure:"host" default:"0.0.0.0"`

	// Port is the network port the backend binds to.
	Port int `mapstructure:"port" default:"8080"`
}

// Address returns the network address used by the Fiber server.
func (cfg Config) Address() string {
	return net.JoinHostPort(cfg.Host, strconv.Itoa(cfg.Port))
}
