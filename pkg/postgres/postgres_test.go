package postgres

import (
	"context"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// TestConfigDSN verifies PostgreSQL connection strings are encoded correctly.
func TestConfigDSN(t *testing.T) {
	cfg := Config{
		Host:     "db.local",
		Port:     5433,
		Database: "realmkit",
		Username: "realm kit",
		Password: "secret/value",
		SSLMode:  "require",
	}

	got := cfg.DSN()
	want := "postgres://realm%20kit:secret%2Fvalue@db.local:5433/realmkit?sslmode=require"
	if got != want {
		t.Fatalf("DSN() = %q, want %q", got, want)
	}
}

// TestOpenAppliesConnectionDefaults verifies Open configures a reachable GORM handle.
func TestOpenAppliesConnectionDefaults(t *testing.T) {
	db, err := Open(context.Background(), Config{}, WithDialector(sqlite.Open(":memory:")))
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer func() {
		if err := Close(db); err != nil {
			t.Fatalf("Close() error = %v", err)
		}
	}()

	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("DB() error = %v", err)
	}
	stats := sqlDB.Stats()
	if stats.MaxOpenConnections != DefaultMaxOpenConns {
		t.Fatalf("MaxOpenConnections = %d, want %d", stats.MaxOpenConnections, DefaultMaxOpenConns)
	}
}

// TestOpenAcceptsGormConfig verifies Open applies custom GORM settings.
func TestOpenAcceptsGormConfig(t *testing.T) {
	db, err := Open(
		context.Background(),
		Config{},
		WithDialector(sqlite.Open(":memory:")),
		WithGormConfig(&gorm.Config{SkipDefaultTransaction: true}),
	)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer func() {
		if err := Close(db); err != nil {
			t.Fatalf("Close() error = %v", err)
		}
	}()

	if !db.Config.SkipDefaultTransaction {
		t.Fatalf("SkipDefaultTransaction = false, want true")
	}
}

// TestOpenReturnsDialectorErrors verifies Open wraps GORM open failures.
func TestOpenReturnsDialectorErrors(t *testing.T) {
	if _, err := Open(context.Background(), Config{}, WithDialector(sqlite.Open("/missing/realmkit/test.db"))); err == nil {
		t.Fatalf("Open() error = nil, want error")
	}
}

// TestHealthRejectsNil verifies Health fails with a nil database handle.
func TestHealthRejectsNil(t *testing.T) {
	if err := Health(context.Background(), nil); err == nil {
		t.Fatalf("Health() error = nil, want error")
	}
}

// TestHealthUsesPing verifies Health accepts a reachable database handle.
func TestHealthUsesPing(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer func() {
		if err := Close(db); err != nil {
			t.Fatalf("Close() error = %v", err)
		}
	}()

	if err := Health(context.Background(), db); err != nil {
		t.Fatalf("Health() error = %v", err)
	}
}

// TestCloseAcceptsNil verifies nil database handles are safe to close.
func TestCloseAcceptsNil(t *testing.T) {
	if err := Close(nil); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
}
