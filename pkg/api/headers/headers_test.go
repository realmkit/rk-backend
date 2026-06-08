package headers

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
)

// TestMiddlewareGeneratesRequestHeaders verifies IDs are generated when absent.
func TestMiddlewareGeneratesRequestHeaders(t *testing.T) {
	app := fiber.New()
	app.Use(Middleware())
	app.Get("/test", func(ctx *fiber.Ctx) error {
		if CurrentRequestID(ctx) == "" {
			t.Fatalf("CurrentRequestID() = empty")
		}
		if CurrentCorrelationID(ctx) == "" {
			t.Fatalf("CurrentCorrelationID() = empty")
		}
		return ctx.SendStatus(fiber.StatusNoContent)
	})

	res, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/test", nil), -1)
	if err != nil {
		t.Fatalf("Test() error = %v", err)
	}
	defer res.Body.Close()

	if res.Header.Get(RequestID) == "" {
		t.Fatalf("%s header = empty", RequestID)
	}
	if res.Header.Get(CorrelationID) == "" {
		t.Fatalf("%s header = empty", CorrelationID)
	}
}

// TestMiddlewarePreservesRequestHeaders verifies client-provided IDs are preserved.
func TestMiddlewarePreservesRequestHeaders(t *testing.T) {
	app := fiber.New()
	app.Use(Middleware())
	app.Get("/test", func(ctx *fiber.Ctx) error {
		return ctx.SendStatus(fiber.StatusNoContent)
	})

	req := httptest.NewRequest(fiber.MethodGet, "/test", nil)
	req.Header.Set(RequestID, "req-1")
	req.Header.Set(CorrelationID, "corr-1")
	res, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("Test() error = %v", err)
	}
	defer res.Body.Close()

	if res.Header.Get(RequestID) != "req-1" {
		t.Fatalf("%s = %q, want %q", RequestID, res.Header.Get(RequestID), "req-1")
	}
	if res.Header.Get(CorrelationID) != "corr-1" {
		t.Fatalf("%s = %q, want %q", CorrelationID, res.Header.Get(CorrelationID), "corr-1")
	}
}

// TestRequireJSONRejectsUnsupportedAccept verifies unsupported response media types fail.
func TestRequireJSONRejectsUnsupportedAccept(t *testing.T) {
	app := fiber.New()
	app.Use(RequireJSON())
	app.Get("/test", func(ctx *fiber.Ctx) error {
		return ctx.SendStatus(fiber.StatusNoContent)
	})

	req := httptest.NewRequest(fiber.MethodGet, "/test", nil)
	req.Header.Set(Accept, "text/plain")
	res, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("Test() error = %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != fiber.StatusNotAcceptable {
		t.Fatalf("StatusCode = %d, want %d", res.StatusCode, fiber.StatusNotAcceptable)
	}
}

// TestRequireJSONRejectsUnsupportedContentType verifies non-JSON request bodies fail.
func TestRequireJSONRejectsUnsupportedContentType(t *testing.T) {
	app := fiber.New()
	app.Use(RequireJSON())
	app.Post("/test", func(ctx *fiber.Ctx) error {
		return ctx.SendStatus(fiber.StatusNoContent)
	})

	req := httptest.NewRequest(fiber.MethodPost, "/test", strings.NewReader("body"))
	req.Header.Set(ContentType, "text/plain")
	res, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("Test() error = %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != fiber.StatusUnsupportedMediaType {
		t.Fatalf("StatusCode = %d, want %d", res.StatusCode, fiber.StatusUnsupportedMediaType)
	}
}

// TestRequireJSONAllowsJSON verifies JSON-compatible headers pass.
func TestRequireJSONAllowsJSON(t *testing.T) {
	app := fiber.New()
	app.Use(RequireJSON())
	app.Post("/test", func(ctx *fiber.Ctx) error {
		return ctx.SendStatus(fiber.StatusNoContent)
	})

	req := httptest.NewRequest(fiber.MethodPost, "/test", strings.NewReader("{}"))
	req.Header.Set(Accept, "application/json")
	req.Header.Set(ContentType, "application/json; charset=utf-8")
	res, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("Test() error = %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != fiber.StatusNoContent {
		t.Fatalf("StatusCode = %d, want %d", res.StatusCode, fiber.StatusNoContent)
	}
}
