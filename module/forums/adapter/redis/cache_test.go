package redis

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/google/uuid"
	"github.com/niflaot/gamehub-go/module/forums/domain"
	"github.com/niflaot/gamehub-go/pkg/pagination"
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

// TestTreeCacheStoresWidgetPages verifies Redis widget cache lifecycle.
func TestTreeCacheStoresWidgetPages(t *testing.T) {
	server := miniredis.RunT(t)
	client := goredis.NewClient(&goredis.Options{Addr: server.Addr()})
	defer client.Close()
	cache := NewTreeCache(client)
	latestKey := "forums:latest:v1:global:all:anonymous::20"
	mostLikedKey := "forums:most-liked:v1:" + uuid.NewString() + ":all:anonymous::20"
	latest := pagination.Result[domain.LatestPostSummary]{Items: []domain.LatestPostSummary{{ForumID: uuid.New(), ThreadID: uuid.New(), PostID: uuid.New(), AuthorUserID: uuid.New(), Sequence: 1, ThreadTitle: "Latest"}}}
	mostLiked := pagination.Result[domain.MostLikedPost]{Items: []domain.MostLikedPost{{ForumID: uuid.New(), ThreadID: uuid.New(), PostID: uuid.New(), AuthorUserID: uuid.New(), Sequence: 1, ThreadTitle: "Popular", LikeCount: 5}}}

	if err := cache.SetLatestPosts(context.Background(), latestKey, latest, time.Minute); err != nil {
		t.Fatalf("SetLatestPosts() error = %v", err)
	}
	if err := cache.SetMostLikedPosts(context.Background(), mostLikedKey, mostLiked, time.Minute); err != nil {
		t.Fatalf("SetMostLikedPosts() error = %v", err)
	}
	gotLatest, latestOK, err := cache.GetLatestPosts(context.Background(), latestKey)
	if err != nil || !latestOK || len(gotLatest.Items) != 1 {
		t.Fatalf("GetLatestPosts() = (%+v, %v, %v), want hit", gotLatest, latestOK, err)
	}
	gotMostLiked, mostLikedOK, err := cache.GetMostLikedPosts(context.Background(), mostLikedKey)
	if err != nil || !mostLikedOK || len(gotMostLiked.Items) != 1 {
		t.Fatalf("GetMostLikedPosts() = (%+v, %v, %v), want hit", gotMostLiked, mostLikedOK, err)
	}
	if err := cache.ClearLatestPosts(context.Background()); err != nil {
		t.Fatalf("ClearLatestPosts() error = %v", err)
	}
	if err := cache.ClearMostLikedPosts(context.Background()); err != nil {
		t.Fatalf("ClearMostLikedPosts() error = %v", err)
	}
	if _, ok, err := cache.GetLatestPosts(context.Background(), latestKey); err != nil || ok {
		t.Fatalf("GetLatestPosts() after clear ok=%v err=%v, want miss", ok, err)
	}
	if _, ok, err := cache.GetMostLikedPosts(context.Background(), mostLikedKey); err != nil || ok {
		t.Fatalf("GetMostLikedPosts() after clear ok=%v err=%v, want miss", ok, err)
	}
}
