package idempotency

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

const (
	// DefaultRedisScope is the default idempotency key namespace.
	DefaultRedisScope = "global"

	// redisKeyPrefix is the stable Redis key namespace.
	redisKeyPrefix = "realmkit:idempotency"
)

// ErrEntryExpired reports that an idempotency entry expired before completion.
var ErrEntryExpired = errors.New("idempotency entry expired")

// ErrEntryConflict reports that a stored entry no longer matches the request.
var ErrEntryConflict = errors.New("idempotency entry fingerprint conflict")

// RedisStore stores idempotency records in Redis.
type RedisStore struct {
	client *goredis.Client
	scope  string
}

// RedisOption changes RedisStore behavior.
type RedisOption func(*RedisStore)

// WithRedisScope configures the Redis idempotency namespace.
func WithRedisScope(scope string) RedisOption {
	return func(store *RedisStore) {
		store.scope = normalizedScope(scope)
	}
}

// NewRedisStore creates a Redis-backed idempotency store.
func NewRedisStore(client *goredis.Client, options ...RedisOption) RedisStore {
	store := RedisStore{
		client: client,
		scope:  DefaultRedisScope,
	}
	for _, option := range options {
		option(&store)
	}
	if store.scope == "" {
		store.scope = DefaultRedisScope
	}
	return store
}

// Reserve reserves key for fingerprint or returns an existing entry.
func (store RedisStore) Reserve(ctx context.Context, key string, fingerprint string, ttl time.Duration) (Entry, bool, error) {
	entry := Entry{Fingerprint: fingerprint, Complete: false, ExpiresAt: time.Now().Add(ttl)}
	payload, err := json.Marshal(entry)
	if err != nil {
		return Entry{}, false, fmt.Errorf("encode idempotency entry: %w", err)
	}
	result, err := redisReserveScript.Run(ctx, store.client, []string{store.key(key)}, string(payload), ttl.Milliseconds()).Result()
	if err != nil {
		return Entry{}, false, fmt.Errorf("reserve idempotency key: %w", err)
	}
	exists, stored, err := reserveResult(result)
	if err != nil {
		return Entry{}, false, err
	}
	var reserved Entry
	if err := json.Unmarshal([]byte(stored), &reserved); err != nil {
		return Entry{}, false, fmt.Errorf("decode idempotency entry: %w", err)
	}
	return reserved, exists, nil
}

// Complete stores the response for key.
func (store RedisStore) Complete(ctx context.Context, key string, entry Entry) error {
	payload, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("encode idempotency entry: %w", err)
	}
	result, err := redisCompleteScript.Run(ctx, store.client, []string{store.key(key)}, entry.Fingerprint, string(payload)).Result()
	if err != nil {
		return fmt.Errorf("complete idempotency key: %w", err)
	}
	state, err := completeResult(result)
	if err != nil {
		return err
	}
	switch state {
	case "ok":
		return nil
	case "missing", "expired":
		return ErrEntryExpired
	case "conflict":
		return ErrEntryConflict
	default:
		return fmt.Errorf("unknown idempotency completion state %q", state)
	}
}

// Release removes an incomplete reservation for key and fingerprint.
func (store RedisStore) Release(ctx context.Context, key string, fingerprint string) error {
	result, err := redisReleaseScript.Run(ctx, store.client, []string{store.key(key)}, fingerprint).Result()
	if err != nil {
		return fmt.Errorf("release idempotency key: %w", err)
	}
	state, err := completeResult(result)
	if err != nil {
		return err
	}
	switch state {
	case "ok", "missing", "complete":
		return nil
	case "conflict":
		return ErrEntryConflict
	default:
		return fmt.Errorf("unknown idempotency release state %q", state)
	}
}

// key returns the Redis key for a client idempotency key.
func (store RedisStore) key(key string) string {
	return strings.Join([]string{redisKeyPrefix, store.scope, keyHash(key)}, ":")
}

// normalizedScope returns a safe Redis scope segment.
func normalizedScope(scope string) string {
	scope = strings.TrimSpace(scope)
	scope = strings.ReplaceAll(scope, ":", "_")
	return scope
}

// reserveResult parses the Redis reserve script result.
func reserveResult(result any) (bool, string, error) {
	values, ok := result.([]any)
	if !ok || len(values) != 2 {
		return false, "", fmt.Errorf("invalid idempotency reserve result")
	}
	exists, err := strconv.ParseBool(fmt.Sprint(values[0]))
	if err != nil {
		return false, "", fmt.Errorf("decode idempotency reserve state: %w", err)
	}
	return exists, fmt.Sprint(values[1]), nil
}

// completeResult parses the Redis complete script result.
func completeResult(result any) (string, error) {
	values, ok := result.([]any)
	if ok && len(values) > 0 {
		return fmt.Sprint(values[0]), nil
	}
	if value, ok := result.(string); ok {
		return value, nil
	}
	return "", fmt.Errorf("invalid idempotency complete result")
}

// redisReserveScript atomically reserves or reads an idempotency entry.
var redisReserveScript = goredis.NewScript(`
local existing = redis.call("GET", KEYS[1])
if existing then
  return {1, existing}
end
redis.call("SET", KEYS[1], ARGV[1], "PX", ARGV[2])
return {0, ARGV[1]}
`)

// redisCompleteScript atomically completes a matching idempotency entry.
var redisCompleteScript = goredis.NewScript(`
local existing = redis.call("GET", KEYS[1])
if not existing then
  return {"missing"}
end
local decoded = cjson.decode(existing)
if decoded["fingerprint"] ~= ARGV[1] then
  return {"conflict"}
end
local ttl = redis.call("PTTL", KEYS[1])
if ttl <= 0 then
  return {"expired"}
end
redis.call("SET", KEYS[1], ARGV[2], "PX", ttl)
return {"ok"}
`)

// redisReleaseScript atomically releases a matching incomplete idempotency entry.
var redisReleaseScript = goredis.NewScript(`
local existing = redis.call("GET", KEYS[1])
if not existing then
  return {"missing"}
end
local decoded = cjson.decode(existing)
if decoded["fingerprint"] ~= ARGV[1] then
  return {"conflict"}
end
if decoded["complete"] then
  return {"complete"}
end
redis.call("DEL", KEYS[1])
return {"ok"}
`)
