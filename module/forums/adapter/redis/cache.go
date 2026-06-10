package redis

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"github.com/niflaot/gamehub-go/module/forums/domain"
	"github.com/niflaot/gamehub-go/module/forums/port"
	"github.com/niflaot/gamehub-go/pkg/pagination"
	goredis "github.com/redis/go-redis/v9"
)

// threadViewsKey stores buffered thread view increments.
const threadViewsKey = "forums:views:v1:threads"

// TreeCache stores visible forum trees in Redis.
type TreeCache struct {
	client *goredis.Client
}

// NewTreeCache creates a Redis tree cache.
func NewTreeCache(client *goredis.Client) TreeCache {
	return TreeCache{client: client}
}

// GetTree returns a cached tree when present.
func (cache TreeCache) GetTree(ctx context.Context, key string) (domain.ForumTree, bool, error) {
	value, err := cache.client.Get(ctx, key).Bytes()
	if err == goredis.Nil {
		return domain.ForumTree{}, false, nil
	}
	if err != nil {
		return domain.ForumTree{}, false, err
	}
	var tree domain.ForumTree
	if err := json.Unmarshal(value, &tree); err != nil {
		return domain.ForumTree{}, false, err
	}
	return tree, true, nil
}

// SetTree stores a tree for ttl.
func (cache TreeCache) SetTree(ctx context.Context, key string, tree domain.ForumTree, ttl time.Duration) error {
	value, err := json.Marshal(tree)
	if err != nil {
		return err
	}
	return cache.client.Set(ctx, key, value, ttl).Err()
}

// ClearTree removes forum tree cache entries.
func (cache TreeCache) ClearTree(ctx context.Context) error {
	return cache.clearPattern(ctx, "forums:tree:v1:*")
}

// GetLatestPosts returns a cached latest-post page when present.
func (cache TreeCache) GetLatestPosts(ctx context.Context, key string) (pagination.Result[domain.LatestPostSummary], bool, error) {
	var result pagination.Result[domain.LatestPostSummary]
	ok, err := cache.getJSON(ctx, key, &result)
	return result, ok, err
}

// SetLatestPosts stores a latest-post page for ttl.
func (cache TreeCache) SetLatestPosts(
	ctx context.Context,
	key string,
	result pagination.Result[domain.LatestPostSummary],
	ttl time.Duration,
) error {
	return cache.setJSON(ctx, key, result, ttl)
}

// ClearLatestPosts removes latest-post cache entries.
func (cache TreeCache) ClearLatestPosts(ctx context.Context) error {
	return cache.clearPattern(ctx, "forums:latest:v1:*")
}

// GetMostLikedPosts returns a cached most-liked page when present.
func (cache TreeCache) GetMostLikedPosts(ctx context.Context, key string) (pagination.Result[domain.MostLikedPost], bool, error) {
	var result pagination.Result[domain.MostLikedPost]
	ok, err := cache.getJSON(ctx, key, &result)
	return result, ok, err
}

// SetMostLikedPosts stores a most-liked page for ttl.
func (cache TreeCache) SetMostLikedPosts(
	ctx context.Context,
	key string,
	result pagination.Result[domain.MostLikedPost],
	ttl time.Duration,
) error {
	return cache.setJSON(ctx, key, result, ttl)
}

// ClearMostLikedPosts removes most-liked cache entries.
func (cache TreeCache) ClearMostLikedPosts(ctx context.Context) error {
	return cache.clearPattern(ctx, "forums:most-liked:v1:*")
}

// IncrementThreadView buffers one thread view.
func (cache TreeCache) IncrementThreadView(ctx context.Context, threadID string) error {
	return cache.client.HIncrBy(ctx, threadViewsKey, threadID, 1).Err()
}

// DrainThreadViews atomically returns and clears buffered thread views.
func (cache TreeCache) DrainThreadViews(ctx context.Context) (map[string]int64, error) {
	values, err := cache.client.HGetAll(ctx, threadViewsKey).Result()
	if err != nil {
		return nil, err
	}
	if len(values) == 0 {
		return map[string]int64{}, nil
	}
	pipe := cache.client.TxPipeline()
	pipe.Del(ctx, threadViewsKey)
	if _, err := pipe.Exec(ctx); err != nil {
		return nil, err
	}
	result := make(map[string]int64, len(values))
	for key, value := range values {
		parsed, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			continue
		}
		result[key] = parsed
	}
	return result, nil
}

// ClearAll removes all forum read-cache keys.
func (cache TreeCache) ClearAll(ctx context.Context) error {
	if err := cache.ClearTree(ctx); err != nil {
		return err
	}
	if err := cache.ClearLatestPosts(ctx); err != nil {
		return err
	}
	return cache.ClearMostLikedPosts(ctx)
}

// getJSON reads a JSON cache item.
func (cache TreeCache) getJSON(ctx context.Context, key string, target any) (bool, error) {
	value, err := cache.client.Get(ctx, key).Bytes()
	if err == goredis.Nil {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if err := json.Unmarshal(value, target); err != nil {
		return false, err
	}
	return true, nil
}

// setJSON stores a JSON cache item.
func (cache TreeCache) setJSON(ctx context.Context, key string, payload any, ttl time.Duration) error {
	value, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	return cache.client.Set(ctx, key, value, ttl).Err()
}

// clearPattern removes cache keys matching pattern.
func (cache TreeCache) clearPattern(ctx context.Context, pattern string) error {
	keys, err := cache.client.Keys(ctx, pattern).Result()
	if err != nil {
		return err
	}
	if len(keys) == 0 {
		return nil
	}
	return cache.client.Del(ctx, keys...).Err()
}

// Ensure TreeCache implements port.ReadCache.
var _ port.ReadCache = TreeCache{}
