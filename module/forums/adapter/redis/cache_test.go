package redis

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/google/uuid"
	"github.com/niflaot/gamehub-go/module/forums/domain"
	goredis "github.com/redis/go-redis/v9"
)

// TestTreeCacheStoresAndClearsTree verifies Redis tree cache lifecycle.
func TestTreeCacheStoresAndClearsTree(t *testing.T) {
	server := miniredis.RunT(t)
	client := goredis.NewClient(&goredis.Options{Addr: server.Addr()})
	defer client.Close()
	cache := NewTreeCache(client)
	key := "forums:tree:v1:anonymous"
	tree := domain.ForumTree{Categories: []domain.CategoryNode{{Category: domain.ForumCategory{ID: uuid.New(), Key: "official", Name: "Official"}}}}

	if err := cache.SetTree(context.Background(), key, tree, time.Minute); err != nil {
		t.Fatalf("SetTree() error = %v", err)
	}
	got, ok, err := cache.GetTree(context.Background(), key)
	if err != nil {
		t.Fatalf("GetTree() error = %v", err)
	}
	if !ok || len(got.Categories) != 1 {
		t.Fatalf("GetTree() = (%+v, %v), want cached tree", got, ok)
	}
	if err := cache.ClearTree(context.Background()); err != nil {
		t.Fatalf("ClearTree() error = %v", err)
	}
	if _, ok, err := cache.GetTree(context.Background(), key); err != nil || ok {
		t.Fatalf("GetTree() after clear ok=%v err=%v, want miss", ok, err)
	}
}
