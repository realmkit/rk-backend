package idempotency

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/niflaot/gamehub-go/pkg/api/headers"
	"github.com/niflaot/gamehub-go/pkg/api/problem"
)

// TestMiddlewareReplaysCompletedRequests verifies matching retries replay responses.
func TestMiddlewareReplaysCompletedRequests(t *testing.T) {
	app := fiber.New(fiber.Config{ErrorHandler: problem.Handler})
	app.Use(Middleware(NewMemoryStore()))
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
	app := fiber.New(fiber.Config{ErrorHandler: problem.Handler})
	app.Use(Middleware(NewMemoryStore()))
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
	app := fiber.New()
	app.Use(Middleware(NewMemoryStore()))
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

// requestWithKey creates a POST request with an idempotency key.
func requestWithKey(key string, body string) *http.Request {
	req := httptest.NewRequest(fiber.MethodPost, "/create", strings.NewReader(body))
	req.Header.Set(headers.IdempotencyKey, key)
	return req
}
