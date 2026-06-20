package redis

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	goredis "github.com/redis/go-redis/v9"
)

// TestConfigDefaults verifies Redis timeout defaults are bounded.
func TestConfigDefaults(t *testing.T) {
	cfg := Config{}.Defaults()
	if cfg.DialTimeout != 5*time.Second {
		t.Fatalf("DialTimeout = %s, want 5s", cfg.DialTimeout)
	}
	if cfg.ReadTimeout != 3*time.Second || cfg.WriteTimeout != 3*time.Second {
		t.Fatalf("timeouts = %s/%s, want 3s/3s", cfg.ReadTimeout, cfg.WriteTimeout)
	}
}

// TestOpenConnectsToRedis verifies Open creates a healthy client.
func TestOpenConnectsToRedis(t *testing.T) {
	server := miniredis.RunT(t)

	client, err := Open(context.Background(), Config{Address: server.Addr()})
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer closeClient(t, client)

	if err := Health(context.Background(), client); err != nil {
		t.Fatalf("Health() error = %v", err)
	}
}

// TestOpenRejectsNilInjectedClient verifies injected clients must exist.
func TestOpenRejectsNilInjectedClient(t *testing.T) {
	if _, err := Open(context.Background(), Config{}, WithClient(nil)); err == nil {
		t.Fatalf("Open() error = nil, want error")
	}
}

// TestHealthRejectsNil verifies nil clients fail health checks.
func TestHealthRejectsNil(t *testing.T) {
	if err := Health(context.Background(), nil); err == nil {
		t.Fatalf("Health() error = nil, want error")
	}
}

// TestCloseAcceptsNil verifies nil clients can be closed.
func TestCloseAcceptsNil(t *testing.T) {
	if err := Close(nil); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
}

// closeClient closes a Redis client for tests.
func closeClient(t *testing.T, client *goredis.Client) {
	t.Helper()
	if err := Close(client); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
}
