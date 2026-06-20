package redis

import "time"

// Config contains Redis connection settings.
type Config struct {
	// Address is the Redis server address.
	Address string `mapstructure:"address" default:"localhost:6379"`

	// Password is the Redis authentication password.
	Password string `mapstructure:"password" default:""`

	// Database is the Redis logical database number.
	Database int `mapstructure:"database" default:"0"`

	// DialTimeout bounds initial Redis connection setup.
	DialTimeout time.Duration `mapstructure:"dial_timeout" default:"5s"`

	// ReadTimeout bounds Redis read operations.
	ReadTimeout time.Duration `mapstructure:"read_timeout" default:"3s"`

	// WriteTimeout bounds Redis write operations.
	WriteTimeout time.Duration `mapstructure:"write_timeout" default:"3s"`
}

// Defaults returns config with package defaults applied.
func (config Config) Defaults() Config {
	if config.DialTimeout <= 0 {
		config.DialTimeout = 5 * time.Second
	}
	if config.ReadTimeout <= 0 {
		config.ReadTimeout = 3 * time.Second
	}
	if config.WriteTimeout <= 0 {
		config.WriteTimeout = 3 * time.Second
	}
	return config
}
