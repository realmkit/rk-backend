package redis

import (
	"context"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/google/uuid"
	goredis "github.com/redis/go-redis/v9"
)

// TestCacheClearTicket verifies ticket cache invalidation.
func TestCacheClearTicket(t *testing.T) {
	server := miniredis.RunT(t)
	client := goredis.NewClient(&goredis.Options{Addr: server.Addr()})
	cache := NewCache(client)
	ticketID := uuid.New()
	ctx := context.Background()
	mustSet(t, client, "tickets:ticket:v1:"+ticketID.String()+":messages", "1")
	mustSet(t, client, "tickets:queue:v1:staff", "1")
	if err := cache.ClearTicket(ctx, ticketID); err != nil {
		t.Fatalf("ClearTicket() error = %v", err)
	}
	if server.Exists("tickets:ticket:v1:" + ticketID.String() + ":messages") {
		t.Fatalf("ticket key still exists")
	}
	if !server.Exists("tickets:queue:v1:staff") {
		t.Fatalf("queue key was unexpectedly cleared")
	}
}

// TestCacheClearAll verifies broad namespace invalidation.
func TestCacheClearAll(t *testing.T) {
	server := miniredis.RunT(t)
	client := goredis.NewClient(&goredis.Options{Addr: server.Addr()})
	cache := NewCache(client)
	ctx := context.Background()
	mustSet(t, client, "tickets:queue:v1:staff", "1")
	mustSet(t, client, "forums:tree:v1:anonymous", "1")
	if err := cache.ClearAll(ctx); err != nil {
		t.Fatalf("ClearAll() error = %v", err)
	}
	if server.Exists("tickets:queue:v1:staff") {
		t.Fatalf("ticket key still exists")
	}
	if !server.Exists("forums:tree:v1:anonymous") {
		t.Fatalf("non-ticket key was unexpectedly cleared")
	}
}

// mustSet stores a Redis key.
func mustSet(t *testing.T, client *goredis.Client, key string, value string) {
	t.Helper()
	if err := client.Set(context.Background(), key, value, 0).Err(); err != nil {
		t.Fatalf("Set(%q) error = %v", key, err)
	}
}
