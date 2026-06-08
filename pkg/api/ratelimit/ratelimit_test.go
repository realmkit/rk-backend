package ratelimit

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/niflaot/gamehub-go/pkg/api/headers"
	"github.com/niflaot/gamehub-go/pkg/api/problem"
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
