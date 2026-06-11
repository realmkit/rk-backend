// Package redis invalidates ticket read caches.
package redis

import (
	"context"

	"github.com/google/uuid"
	"github.com/niflaot/gamehub-go/module/tickets/port"
	goredis "github.com/redis/go-redis/v9"
)

// Cache clears ticket cache namespaces in Redis.
type Cache struct {
	client *goredis.Client
}

// NewCache creates a Redis ticket cache.
func NewCache(client *goredis.Client) Cache {
	return Cache{client: client}
}

// ClearTicket removes read caches for one ticket.
func (cache Cache) ClearTicket(ctx context.Context, ticketID uuid.UUID) error {
	return cache.clearPattern(ctx, "tickets:ticket:v1:"+ticketID.String()+":*")
}

// ClearQueues removes ticket queue caches.
func (cache Cache) ClearQueues(ctx context.Context) error {
	return cache.clearPattern(ctx, "tickets:queue:v1:*")
}

// ClearAll removes all ticket read caches.
func (cache Cache) ClearAll(ctx context.Context) error {
	return cache.clearPattern(ctx, "tickets:*")
}

// clearPattern deletes keys matching a Redis pattern.
func (cache Cache) clearPattern(ctx context.Context, pattern string) error {
	keys, err := cache.client.Keys(ctx, pattern).Result()
	if err != nil {
		return err
	}
	if len(keys) == 0 {
		return nil
	}
	return cache.client.Del(ctx, keys...).Err()
}

// Ensure Cache implements the ticket cache contract.
var _ port.Cache = Cache{}
