package redis

import (
	"context"
	"testing"

	"github.com/alicebob/miniredis/v2"
	goredis "github.com/redis/go-redis/v9"
)

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
