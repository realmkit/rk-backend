package idempotency

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/niflaot/gamehub-go/pkg/api/headers"
	"github.com/niflaot/gamehub-go/pkg/api/problem"
	goredis "github.com/redis/go-redis/v9"
)

// TestMiddlewareReplaysCompletedRequests verifies matching retries replay responses.
func TestMiddlewareReplaysCompletedRequests(t *testing.T) {
	store, closeStore := newRedisStore(t)
	defer closeStore()
	app := fiber.New(fiber.Config{ErrorHandler: problem.Handler})
	app.Use(Middleware(store))
	calls := 0
	app.Post("/create", func(ctx *fiber.Ctx) error {
		calls++
		return ctx.Status(fiber.StatusCreated).SendString("created")
	})

	first := requestWithKey("same", "value")
	second := requestWithKey("same", "value")
	firstRes, err := app.Test(first, -1)
	if err != nil {
		t.Fatalf("Test() error = %v", err)
	}
	defer firstRes.Body.Close()
	secondRes, err := app.Test(second, -1)
	if err != nil {
		t.Fatalf("Test() error = %v", err)
	}
	defer secondRes.Body.Close()

	if calls != 1 {
		t.Fatalf("calls = %d, want %d", calls, 1)
	}
	if secondRes.StatusCode != fiber.StatusCreated {
		t.Fatalf("StatusCode = %d, want %d", secondRes.StatusCode, fiber.StatusCreated)
	}
}

// TestMiddlewareRejectsConflictingRequests verifies key reuse with different bodies fails.
func TestMiddlewareRejectsConflictingRequests(t *testing.T) {
	store, closeStore := newRedisStore(t)
	defer closeStore()
	app := fiber.New(fiber.Config{ErrorHandler: problem.Handler})
	app.Use(Middleware(store))
	app.Post("/create", func(ctx *fiber.Ctx) error {
		return ctx.SendString("ok")
	})

	firstRes, err := app.Test(requestWithKey("same", "one"), -1)
	if err != nil {
		t.Fatalf("Test() error = %v", err)
	}
	defer firstRes.Body.Close()
	secondRes, err := app.Test(requestWithKey("same", "two"), -1)
	if err != nil {
		t.Fatalf("Test() error = %v", err)
	}
	defer secondRes.Body.Close()

	if secondRes.StatusCode != fiber.StatusConflict {
		t.Fatalf("StatusCode = %d, want %d", secondRes.StatusCode, fiber.StatusConflict)
	}
}

// TestMiddlewareIgnoresSafeMethods verifies GET requests bypass idempotency.
func TestMiddlewareIgnoresSafeMethods(t *testing.T) {
	store, closeStore := newRedisStore(t)
	defer closeStore()
	app := fiber.New()
	app.Use(Middleware(store))
	app.Get("/read", func(ctx *fiber.Ctx) error {
		return ctx.SendStatus(fiber.StatusNoContent)
	})

	res, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/read", nil), -1)
	if err != nil {
		t.Fatalf("Test() error = %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != fiber.StatusNoContent {
		t.Fatalf("StatusCode = %d, want %d", res.StatusCode, fiber.StatusNoContent)
	}
}

// TestMiddlewareRejectsLongKeys verifies client-provided keys are bounded.
func TestMiddlewareRejectsLongKeys(t *testing.T) {
	store, closeStore := newRedisStore(t)
	defer closeStore()
	app := fiber.New(fiber.Config{ErrorHandler: problem.Handler})
	app.Use(Middleware(store))
	app.Post("/create", func(ctx *fiber.Ctx) error {
		return ctx.SendString("ok")
	})

	req := requestWithKey(strings.Repeat("a", MaxKeyLength+1), "value")
	res, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("Test() error = %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != fiber.StatusBadRequest {
		t.Fatalf("StatusCode = %d, want %d", res.StatusCode, fiber.StatusBadRequest)
	}
}

// TestRedisStoreReplaysCompletedRequests verifies Redis preserves completed responses.
func TestRedisStoreReplaysCompletedRequests(t *testing.T) {
	store, closeStore := newRedisStore(t)
	defer closeStore()

	entry, exists, err := store.Reserve(context.Background(), "key-1", "fingerprint", time.Hour)
	if err != nil {
		t.Fatalf("Reserve() error = %v", err)
	}
	if exists || entry.Complete {
		t.Fatalf("Reserve() = (%+v, %v), want new incomplete entry", entry, exists)
	}

	err = store.Complete(context.Background(), "key-1", Entry{
		Fingerprint: "fingerprint",
		Status:      fiber.StatusCreated,
		Body:        []byte("created"),
		ContentType: "text/plain",
		Complete:    true,
		ExpiresAt:   time.Now().Add(time.Hour),
	})
	if err != nil {
		t.Fatalf("Complete() error = %v", err)
	}

	replayed, exists, err := store.Reserve(context.Background(), "key-1", "fingerprint", time.Hour)
	if err != nil {
		t.Fatalf("Reserve() replay error = %v", err)
	}
	if !exists || !replayed.Complete || string(replayed.Body) != "created" {
		t.Fatalf("Reserve() replay = (%+v, %v), want completed created response", replayed, exists)
	}
}

// TestRedisStoreReturnsExistingDifferentFingerprint verifies conflicts are visible to middleware.
func TestRedisStoreReturnsExistingDifferentFingerprint(t *testing.T) {
	store, closeStore := newRedisStore(t)
	defer closeStore()

	if _, _, err := store.Reserve(context.Background(), "key-1", "first", time.Hour); err != nil {
		t.Fatalf("Reserve() error = %v", err)
	}

	entry, exists, err := store.Reserve(context.Background(), "key-1", "second", time.Hour)
	if err != nil {
		t.Fatalf("Reserve() conflict error = %v", err)
	}
	if !exists || entry.Fingerprint != "first" {
		t.Fatalf("Reserve() = (%+v, %v), want existing first fingerprint", entry, exists)
	}
}

// TestRedisStoreCompleteFailsAfterExpiry verifies completion fails closed after expiry.
func TestRedisStoreCompleteFailsAfterExpiry(t *testing.T) {
	server := miniredis.RunT(t)
	client := goredis.NewClient(&goredis.Options{Addr: server.Addr()})
	defer closeRedisClient(t, client)
	store := NewRedisStore(client, WithRedisScope("test"))

	if _, _, err := store.Reserve(context.Background(), "key-1", "fingerprint", time.Second); err != nil {
		t.Fatalf("Reserve() error = %v", err)
	}
	server.FastForward(2 * time.Second)

	err := store.Complete(context.Background(), "key-1", Entry{Fingerprint: "fingerprint", Complete: true})
	if !errors.Is(err, ErrEntryExpired) {
		t.Fatalf("Complete() error = %v, want %v", err, ErrEntryExpired)
	}
}

// TestRedisStoreCompleteRejectsFingerprintMismatch verifies completion protects stored entries.
func TestRedisStoreCompleteRejectsFingerprintMismatch(t *testing.T) {
	store, closeStore := newRedisStore(t)
	defer closeStore()

	if _, _, err := store.Reserve(context.Background(), "key-1", "first", time.Hour); err != nil {
		t.Fatalf("Reserve() error = %v", err)
	}

	err := store.Complete(context.Background(), "key-1", Entry{Fingerprint: "second", Complete: true})
	if !errors.Is(err, ErrEntryConflict) {
		t.Fatalf("Complete() error = %v, want %v", err, ErrEntryConflict)
	}
}

// requestWithKey creates a POST request with an idempotency key.
func requestWithKey(key string, body string) *http.Request {
	req := httptest.NewRequest(fiber.MethodPost, "/create", strings.NewReader(body))
	req.Header.Set(headers.IdempotencyKey, key)
	return req
}

// newRedisStore creates a Redis idempotency store for tests.
func newRedisStore(t *testing.T) (RedisStore, func()) {
	t.Helper()
	server := miniredis.RunT(t)
	client := goredis.NewClient(&goredis.Options{Addr: server.Addr()})
	store := NewRedisStore(client, WithRedisScope("test"))
	return store, func() {
		closeRedisClient(t, client)
	}
}

// closeRedisClient closes a Redis client for tests.
func closeRedisClient(t *testing.T, client *goredis.Client) {
	t.Helper()
	if err := client.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
}
