package ratelimit

import (
	"context"
	"strconv"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/realmkit/rk-backend/pkg/api/headers"
	"github.com/realmkit/rk-backend/pkg/api/problem"
)

// Policy defines rate limit behavior.
type Policy struct {
	// Limit is the number of allowed requests per window.
	Limit int

	// Window is the rate limit window duration.
	Window time.Duration
}

// Decision describes one rate limit decision.
type Decision struct {
	// Allowed reports whether the request may continue.
	Allowed bool

	// Limit is the configured window limit.
	Limit int

	// Remaining is the remaining request count.
	Remaining int

	// ResetAt is the window reset time.
	ResetAt time.Time
}

// Store records rate limit usage.
type Store interface {
	// Allow records one hit for key and returns the decision.
	Allow(ctx context.Context, key string, policy Policy) (Decision, error)
}

// MemoryStore stores rate limits in process memory.
type MemoryStore struct {
	mu      sync.Mutex
	windows map[string]window
	now     func() time.Time
}

// window contains rate limit state for one key.
type window struct {
	count   int
	resetAt time.Time
}

// NewMemoryStore creates an in-memory rate limit store.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		windows: make(map[string]window),
		now:     time.Now,
	}
}

// Allow records one hit for key and returns the decision.
func (store *MemoryStore) Allow(_ context.Context, key string, policy Policy) (Decision, error) {
	store.mu.Lock()
	defer store.mu.Unlock()

	current := store.windows[key]
	now := store.now()
	if current.resetAt.IsZero() || !current.resetAt.After(now) {
		current = window{resetAt: now.Add(policy.Window)}
	}
	current.count++
	store.windows[key] = current

	remaining := policy.Limit - current.count
	if remaining < 0 {
		remaining = 0
	}

	return Decision{
		Allowed:   current.count <= policy.Limit,
		Limit:     policy.Limit,
		Remaining: remaining,
		ResetAt:   current.resetAt,
	}, nil
}

// Middleware returns Fiber rate limiting middleware.
func Middleware(store Store, policy Policy) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		decision, err := store.Allow(ctx.UserContext(), keyFor(ctx), policy)
		if err != nil {
			return err
		}

		writeHeaders(ctx, decision)
		if !decision.Allowed {
			ctx.Set(headers.RetryAfter, retryAfter(decision))
			return problem.Write(ctx, problem.New(fiber.StatusTooManyRequests, "rate_limited", "Rate limit exceeded."))
		}

		return ctx.Next()
	}
}

// keyFor returns the rate limit key for ctx.
func keyFor(ctx *fiber.Ctx) string {
	return ctx.IP() + ":" + ctx.Method() + ":" + ctx.Path()
}

// writeHeaders writes standard rate limit headers.
func writeHeaders(ctx *fiber.Ctx, decision Decision) {
	ctx.Set(headers.RateLimitLimit, strconv.Itoa(decision.Limit))
	ctx.Set(headers.RateLimitRemaining, strconv.Itoa(decision.Remaining))
	ctx.Set(headers.RateLimitReset, strconv.FormatInt(decision.ResetAt.Unix(), 10))
}

// retryAfter returns retry delay seconds.
func retryAfter(decision Decision) string {
	seconds := int(time.Until(decision.ResetAt).Seconds())
	if seconds < 1 {
		seconds = 1
	}
	return strconv.Itoa(seconds)
}
