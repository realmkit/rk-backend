package postgres

import (
	"context"
	"fmt"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// DefaultMaxIdleConns is the internal PostgreSQL idle connection default.
const DefaultMaxIdleConns = 10

// DefaultMaxOpenConns is the internal PostgreSQL open connection default.
const DefaultMaxOpenConns = 25

// DefaultConnMaxLifetime is the internal PostgreSQL connection lifetime default.
const DefaultConnMaxLifetime = time.Hour

// Option changes how PostgreSQL connections are opened.
type Option func(*settings)

// settings contains PostgreSQL open settings.
type settings struct {
	dialector  gorm.Dialector
	gormConfig *gorm.Config
}

// WithDialector overrides the GORM dialector used by Open.
func WithDialector(dialector gorm.Dialector) Option {
	return func(settings *settings) {
		settings.dialector = dialector
	}
}

// WithGormConfig overrides the GORM configuration used by Open.
func WithGormConfig(config *gorm.Config) Option {
	return func(settings *settings) {
		settings.gormConfig = config
	}
}

// Open creates a GORM PostgreSQL connection and verifies it with Ping.
func Open(ctx context.Context, cfg Config, options ...Option) (*gorm.DB, error) {
	settings := settings{
		dialector:  postgres.Open(cfg.DSN()),
		gormConfig: &gorm.Config{},
	}
	for _, option := range options {
		option(&settings)
	}

	db, err := gorm.Open(settings.dialector, settings.gormConfig)
	if err != nil {
		return nil, fmt.Errorf("open postgres: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("read postgres handle: %w", err)
	}
	sqlDB.SetMaxIdleConns(DefaultMaxIdleConns)
	sqlDB.SetMaxOpenConns(DefaultMaxOpenConns)
	sqlDB.SetConnMaxLifetime(DefaultConnMaxLifetime)

	if err := sqlDB.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("ping postgres: %w", err)
	}

	return db, nil
}

// Health verifies the PostgreSQL connection is reachable.
func Health(ctx context.Context, db *gorm.DB) error {
	if db == nil {
		return fmt.Errorf("postgres database is nil")
	}

	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("read postgres handle: %w", err)
	}
	if err := sqlDB.PingContext(ctx); err != nil {
		return fmt.Errorf("ping postgres: %w", err)
	}

	return nil
}

// Close closes the PostgreSQL connection pool.
func Close(db *gorm.DB) error {
	if db == nil {
		return nil
	}

	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("read postgres handle: %w", err)
	}
	if err := sqlDB.Close(); err != nil {
		return fmt.Errorf("close postgres: %w", err)
	}

	return nil
}
