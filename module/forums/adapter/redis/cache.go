package redis

import (
	"context"
	"encoding/json"
	"time"

	"github.com/niflaot/gamehub-go/module/forums/domain"
	"github.com/niflaot/gamehub-go/module/forums/port"
	goredis "github.com/redis/go-redis/v9"
)

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
	keys, err := cache.client.Keys(ctx, "forums:tree:v1:*").Result()
	if err != nil {
		return err
	}
	if len(keys) == 0 {
		return nil
	}
	return cache.client.Del(ctx, keys...).Err()
}

// Ensure TreeCache implements port.TreeCache.
var _ port.TreeCache = TreeCache{}
