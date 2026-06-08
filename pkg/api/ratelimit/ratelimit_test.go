package ratelimit

import (
	"context"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/niflaot/gamehub-go/pkg/api/headers"
	"github.com/niflaot/gamehub-go/pkg/api/problem"
	goredis "github.com/redis/go-redis/v9"
)

// TestMiddlewareAllowsWithinLimit verifies requests pass while under limit.
func TestMiddlewareAllowsWithinLimit(t *testing.T) {
	app := fiber.New(fiber.Config{ErrorHandler: problem.Handler})
	app.Use(Middleware(NewMemoryStore(), Policy{Limit: 2, Window: time.Minute}))
	app.Get("/test", func(ctx *fiber.Ctx) error {
		return ctx.SendStatus(fiber.StatusNoContent)
	})

	res, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/test", nil), -1)
	if err != nil {
		t.Fatalf("Test() error = %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != fiber.StatusNoContent {
		t.Fatalf("StatusCode = %d, want %d", res.StatusCode, fiber.StatusNoContent)
	}
	if res.Header.Get(headers.RateLimitLimit) != "2" {
		t.Fatalf("RateLimit-Limit = %q, want %q", res.Header.Get(headers.RateLimitLimit), "2")
	}
}

// TestMiddlewareRejectsOverLimit verifies requests over limit fail with 429.
func TestMiddlewareRejectsOverLimit(t *testing.T) {
	app := fiber.New(fiber.Config{ErrorHandler: problem.Handler})
	app.Use(Middleware(NewMemoryStore(), Policy{Limit: 1, Window: time.Minute}))
	app.Get("/test", func(ctx *fiber.Ctx) error {
		return ctx.SendStatus(fiber.StatusNoContent)
	})

	first, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/test", nil), -1)
	if err != nil {
		t.Fatalf("Test() error = %v", err)
	}
	defer first.Body.Close()
	second, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/test", nil), -1)
	if err != nil {
		t.Fatalf("Test() error = %v", err)
	}
	defer second.Body.Close()

	if second.StatusCode != fiber.StatusTooManyRequests {
		t.Fatalf("StatusCode = %d, want %d", second.StatusCode, fiber.StatusTooManyRequests)
	}
	if second.Header.Get(headers.RetryAfter) == "" {
		t.Fatalf("Retry-After header = empty")
	}
}

// TestRedisStoreAllowsWithinLimit verifies Redis rate limits below the policy limit.
func TestRedisStoreAllowsWithinLimit(t *testing.T) {
	store, closeStore := newRedisStore(t)
	defer closeStore()

	decision, err := store.Allow(context.Background(), "client-1", Policy{Limit: 2, Window: time.Minute})
	if err != nil {
		t.Fatalf("Allow() error = %v", err)
	}

	if !decision.Allowed || decision.Remaining != 1 {
		t.Fatalf("Decision = %+v, want allowed with one remaining", decision)
	}
}

// TestRedisStoreRejectsOverLimit verifies Redis rejects requests beyond the policy limit.
func TestRedisStoreRejectsOverLimit(t *testing.T) {
	store, closeStore := newRedisStore(t)
	defer closeStore()
	policy := Policy{Limit: 1, Window: time.Minute}

	if _, err := store.Allow(context.Background(), "client-1", policy); err != nil {
		t.Fatalf("Allow() first error = %v", err)
	}
	decision, err := store.Allow(context.Background(), "client-1", policy)
	if err != nil {
		t.Fatalf("Allow() second error = %v", err)
	}

	if decision.Allowed || decision.Remaining != 0 {
		t.Fatalf("Decision = %+v, want rejected with zero remaining", decision)
	}
}

// newRedisStore creates a Redis rate limit store for tests.
func newRedisStore(t *testing.T) (RedisStore, func()) {
	t.Helper()
	server := miniredis.RunT(t)
	client := goredis.NewClient(&goredis.Options{Addr: server.Addr()})
	store := NewRedisStore(client, WithRedisScope("test"))
	return store, func() {
		if err := client.Close(); err != nil {
			t.Fatalf("Close() error = %v", err)
		}
	}
}
