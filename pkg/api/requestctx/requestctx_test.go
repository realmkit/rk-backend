package requestctx

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
)

// TestMiddlewareInstallsDeadline verifies request handlers receive a deadline.
func TestMiddlewareInstallsDeadline(t *testing.T) {
	app := fiber.New()
	app.Use(Middleware(time.Minute))
	app.Get("/test", func(ctx *fiber.Ctx) error {
		if _, ok := ctx.UserContext().Deadline(); !ok {
			t.Fatalf("Deadline() ok = false, want true")
		}
		if profile, ok := CurrentProfile(ctx); !ok || profile != "standard" {
			t.Fatalf("CurrentProfile() = %q %v, want standard true", profile, ok)
		}
		if _, ok := CurrentDeadline(ctx); !ok {
			t.Fatalf("CurrentDeadline() ok = false, want true")
		}
		return ctx.SendStatus(fiber.StatusNoContent)
	})

	response, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/test", nil), -1)
	if err != nil {
		t.Fatalf("Test() error = %v", err)
	}
	defer response.Body.Close()
	if response.StatusCode != fiber.StatusNoContent {
		t.Fatalf("StatusCode = %d, want %d", response.StatusCode, fiber.StatusNoContent)
	}
}

// TestMiddlewareSkipsConfiguredRequests verifies skipped routes keep the parent context.
func TestMiddlewareSkipsConfiguredRequests(t *testing.T) {
	app := fiber.New()
	app.Use(Middleware(time.Minute, WithSkipper(func(ctx *fiber.Ctx) bool {
		return ctx.Path() == "/health"
	})))
	app.Get("/health", func(ctx *fiber.Ctx) error {
		if _, ok := ctx.UserContext().Deadline(); ok {
			t.Fatalf("Deadline() ok = true, want false")
		}
		return ctx.SendStatus(fiber.StatusNoContent)
	})

	response, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/health", nil), -1)
	if err != nil {
		t.Fatalf("Test() error = %v", err)
	}
	defer response.Body.Close()
	if response.StatusCode != fiber.StatusNoContent {
		t.Fatalf("StatusCode = %d, want %d", response.StatusCode, fiber.StatusNoContent)
	}
}

// TestMiddlewareUsesRouteProfile verifies exact route profiles override defaults.
func TestMiddlewareUsesRouteProfile(t *testing.T) {
	app := fiber.New()
	app.Use(Middleware(
		time.Minute,
		WithRouteProfile(fiber.MethodGet, "/stream", Profile{Name: "stream", Timeout: 0}),
	))
	app.Get("/stream", func(ctx *fiber.Ctx) error {
		if _, ok := ctx.UserContext().Deadline(); ok {
			t.Fatalf("Deadline() ok = true, want false")
		}
		return ctx.SendStatus(fiber.StatusNoContent)
	})

	response, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/stream", nil), -1)
	if err != nil {
		t.Fatalf("Test() error = %v", err)
	}
	defer response.Body.Close()
	if response.StatusCode != fiber.StatusNoContent {
		t.Fatalf("StatusCode = %d, want %d", response.StatusCode, fiber.StatusNoContent)
	}
}

// TestMiddlewareUsesPrefixProfile verifies prefix profiles can match route groups.
func TestMiddlewareUsesPrefixProfile(t *testing.T) {
	app := fiber.New()
	app.Use(Middleware(time.Second, WithPathPrefixProfile("", "/admin", Profile{Name: "admin", Timeout: time.Minute})))
	app.Get("/admin/jobs", func(ctx *fiber.Ctx) error {
		deadline, ok := ctx.UserContext().Deadline()
		if !ok {
			t.Fatalf("Deadline() ok = false, want true")
		}
		if time.Until(deadline) < 30*time.Second {
			t.Fatalf("deadline = %s, want admin profile", deadline)
		}
		return ctx.SendStatus(fiber.StatusNoContent)
	})

	response, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/admin/jobs", nil), -1)
	if err != nil {
		t.Fatalf("Test() error = %v", err)
	}
	if response.StatusCode != fiber.StatusNoContent {
		t.Fatalf("StatusCode = %d, want %d", response.StatusCode, fiber.StatusNoContent)
	}
}
