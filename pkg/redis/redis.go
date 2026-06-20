package redis

import (
	"context"
	"fmt"

	goredis "github.com/redis/go-redis/v9"
)

// Option changes how Redis connections are opened.
type Option func(*settings)

// settings contains Redis open settings.
type settings struct {
	client *goredis.Client
}

// WithClient overrides the Redis client used by Open.
func WithClient(client *goredis.Client) Option {
	return func(settings *settings) {
		settings.client = client
	}
}

// Open creates a Redis client and verifies it with Ping.
func Open(ctx context.Context, cfg Config, options ...Option) (*goredis.Client, error) {
	cfg = cfg.Defaults()
	settings := settings{
		client: goredis.NewClient(&goredis.Options{
			Addr:         cfg.Address,
			Password:     cfg.Password,
			DB:           cfg.Database,
			DialTimeout:  cfg.DialTimeout,
			ReadTimeout:  cfg.ReadTimeout,
			WriteTimeout: cfg.WriteTimeout,
		}),
	}
	for _, option := range options {
		option(&settings)
	}
	if settings.client == nil {
		return nil, fmt.Errorf("redis client is nil")
	}
	if err := Health(ctx, settings.client); err != nil {
		return nil, err
	}
	return settings.client, nil
}

// Close closes the Redis client.
func Close(client *goredis.Client) error {
	if client == nil {
		return nil
	}
	if err := client.Close(); err != nil {
		return fmt.Errorf("close redis: %w", err)
	}
	return nil
}
