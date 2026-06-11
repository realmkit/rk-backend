package ratelimit

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

const (
	// DefaultRedisScope is the default rate limit key namespace.
	DefaultRedisScope = "global"

	// redisKeyPrefix is the stable Redis rate limit namespace.
	redisKeyPrefix = "realmkit:ratelimit"
)

// RedisStore stores rate limit windows in Redis.
type RedisStore struct {
	client *goredis.Client
	scope  string
	now    func() time.Time
}

// RedisOption changes RedisStore behavior.
type RedisOption func(*RedisStore)

// WithRedisScope configures the Redis rate limit namespace.
func WithRedisScope(scope string) RedisOption {
	return func(store *RedisStore) {
		store.scope = normalizedScope(scope)
	}
}

// NewRedisStore creates a Redis-backed rate limit store.
func NewRedisStore(client *goredis.Client, options ...RedisOption) RedisStore {
	store := RedisStore{
		client: client,
		scope:  DefaultRedisScope,
		now:    time.Now,
	}
	for _, option := range options {
		option(&store)
	}
	if store.scope == "" {
		store.scope = DefaultRedisScope
	}
	return store
}

// Allow records one hit for key and returns the decision.
func (store RedisStore) Allow(ctx context.Context, key string, policy Policy) (Decision, error) {
	result, err := redisAllowScript.Run(ctx, store.client, []string{store.key(key)}, policy.Window.Milliseconds()).Result()
	if err != nil {
		return Decision{}, fmt.Errorf("allow rate limit hit: %w", err)
	}
	count, ttl, err := allowResult(result)
	if err != nil {
		return Decision{}, err
	}
	remaining := policy.Limit - count
	if remaining < 0 {
		remaining = 0
	}
	return Decision{
		Allowed:   count <= policy.Limit,
		Limit:     policy.Limit,
		Remaining: remaining,
		ResetAt:   store.now().Add(time.Duration(ttl) * time.Millisecond),
	}, nil
}

// key returns the Redis key for a rate limit subject.
func (store RedisStore) key(key string) string {
	return strings.Join([]string{redisKeyPrefix, store.scope, key}, ":")
}

// normalizedScope returns a safe Redis scope segment.
func normalizedScope(scope string) string {
	scope = strings.TrimSpace(scope)
	scope = strings.ReplaceAll(scope, ":", "_")
	return scope
}

// allowResult parses the Redis rate limit script result.
func allowResult(result any) (int, int64, error) {
	values, ok := result.([]any)
	if !ok || len(values) != 2 {
		return 0, 0, fmt.Errorf("invalid rate limit result")
	}
	count, err := strconv.Atoi(fmt.Sprint(values[0]))
	if err != nil {
		return 0, 0, fmt.Errorf("decode rate limit count: %w", err)
	}
	ttl, err := strconv.ParseInt(fmt.Sprint(values[1]), 10, 64)
	if err != nil {
		return 0, 0, fmt.Errorf("decode rate limit ttl: %w", err)
	}
	if ttl < 1 {
		ttl = 1
	}
	return count, ttl, nil
}

// redisAllowScript atomically increments one fixed rate limit window.
var redisAllowScript = goredis.NewScript(`
local count = redis.call("INCR", KEYS[1])
if count == 1 then
  redis.call("PEXPIRE", KEYS[1], ARGV[1])
end
local ttl = redis.call("PTTL", KEYS[1])
return {count, ttl}
`)
