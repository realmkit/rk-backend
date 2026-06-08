package redis

// Config contains Redis connection settings.
type Config struct {
	// Address is the Redis server address.
	Address string `mapstructure:"address" default:"localhost:6379"`

	// Password is the Redis authentication password.
	Password string `mapstructure:"password" default:""`

	// Database is the Redis logical database number.
	Database int `mapstructure:"database" default:"0"`
}
