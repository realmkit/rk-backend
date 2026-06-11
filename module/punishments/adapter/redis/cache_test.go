package redis

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/module/punishments/domain"
	goredis "github.com/redis/go-redis/v9"
)

// TestCacheStoresAndClearsRestriction verifies restriction cache lifecycle.
func TestCacheStoresAndClearsRestriction(t *testing.T) {
	server := miniredis.RunT(t)
	client := goredis.NewClient(&goredis.Options{Addr: server.Addr()})
	defer client.Close()
	cache := NewCache(client)
	userID := uuid.New()
	result := domain.CheckResult{Allowed: false}

	if err := cache.Set(context.Background(), userID, domain.ActionForumsReply, result, time.Minute); err != nil {
		t.Fatalf("Set() error = %v", err)
	}
	got, ok, err := cache.Get(context.Background(), userID, domain.ActionForumsReply)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if !ok || got.Allowed {
		t.Fatalf("Get() = (%+v, %v), want denied hit", got, ok)
	}
	if err := cache.ClearUser(context.Background(), userID); err != nil {
		t.Fatalf("ClearUser() error = %v", err)
	}
	if _, ok, err := cache.Get(context.Background(), userID, domain.ActionForumsReply); err != nil || ok {
		t.Fatalf("Get() after ClearUser ok=%v err=%v, want miss", ok, err)
	}
}

// TestCacheClearAllRemovesEveryRestrictionKey verifies global invalidation.
func TestCacheClearAllRemovesEveryRestrictionKey(t *testing.T) {
	server := miniredis.RunT(t)
	client := goredis.NewClient(&goredis.Options{Addr: server.Addr()})
	defer client.Close()
	cache := NewCache(client)

	if err := cache.Set(context.Background(), uuid.New(), domain.ActionForumsReply, domain.CheckResult{Allowed: false}, time.Minute); err != nil {
		t.Fatalf("Set() error = %v", err)
	}
	if err := cache.ClearAll(context.Background()); err != nil {
		t.Fatalf("ClearAll() error = %v", err)
	}
	keys := server.Keys()
	if len(keys) != 0 {
		t.Fatalf("keys = %+v, want cleared cache", keys)
	}
}
