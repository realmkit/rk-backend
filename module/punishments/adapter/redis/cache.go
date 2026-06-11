// Package redis caches punishment restriction checks.
package redis

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/niflaot/gamehub-go/module/punishments/domain"
	goredis "github.com/redis/go-redis/v9"
)

// Cache stores restriction check results in Redis.
type Cache struct {
	client *goredis.Client
}

// NewCache creates a restriction cache.
func NewCache(client *goredis.Client) Cache {
	return Cache{client: client}
}

// Get returns a cached result.
func (cache Cache) Get(ctx context.Context, userID uuid.UUID, actionKey string) (domain.CheckResult, bool, error) {
	body, err := cache.client.Get(ctx, key(userID, actionKey)).Bytes()
	if err == goredis.Nil {
		return domain.CheckResult{}, false, nil
	}
	if err != nil {
		return domain.CheckResult{}, false, err
	}
	var result domain.CheckResult
	if err := json.Unmarshal(body, &result); err != nil {
		return domain.CheckResult{}, false, err
	}
	return result, true, nil
}

// Set stores a cached result.
func (cache Cache) Set(ctx context.Context, userID uuid.UUID, actionKey string, result domain.CheckResult, ttl time.Duration) error {
	body, err := json.Marshal(result)
	if err != nil {
		return err
	}
	return cache.client.Set(ctx, key(userID, actionKey), body, ttl).Err()
}

// ClearUser clears all restriction checks for one user.
func (cache Cache) ClearUser(ctx context.Context, userID uuid.UUID) error {
	pattern := "punishments:restrictions:v1:user:" + userID.String() + ":*"
	keys, err := cache.client.Keys(ctx, pattern).Result()
	if err != nil || len(keys) == 0 {
		return err
	}
	return cache.client.Del(ctx, keys...).Err()
}

// ClearAll clears all restriction caches.
func (cache Cache) ClearAll(ctx context.Context) error {
	keys, err := cache.client.Keys(ctx, "punishments:restrictions:v1:*").Result()
	if err != nil || len(keys) == 0 {
		return err
	}
	return cache.client.Del(ctx, keys...).Err()
}

// key returns the restriction cache key for a user action.
func key(userID uuid.UUID, actionKey string) string {
	return "punishments:restrictions:v1:user:" + userID.String() + ":action:" + actionKey
}
