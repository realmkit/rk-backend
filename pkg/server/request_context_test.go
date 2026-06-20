package server

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/realmkit/rk-backend/pkg/api/requestctx"
)

// TestRequestContextMiddlewareUsesAdminProfile verifies operator routes get admin timeouts.
func TestRequestContextMiddlewareUsesAdminProfile(t *testing.T) {
	app := fiber.New()
	app.Use(requestContextMiddleware(Config{RequestTimeout: time.Second, AdminRequestTimeout: time.Minute}.Defaults()))
	app.Get("/cronjobs", assertProfile(t, "admin"))

	response, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/cronjobs", nil), -1)
	if err != nil {
		t.Fatalf("Test() error = %v", err)
	}
	if response.StatusCode != fiber.StatusNoContent {
		t.Fatalf("StatusCode = %d, want %d", response.StatusCode, fiber.StatusNoContent)
	}
}

// TestRequestContextMiddlewareUsesUploadProfile verifies upload routes get longer timeouts.
func TestRequestContextMiddlewareUsesUploadProfile(t *testing.T) {
	app := fiber.New()
	app.Use(requestContextMiddleware(Config{RequestTimeout: time.Second, UploadRequestTimeout: time.Minute}.Defaults()))
	app.Post("/assets/upload-intents", assertProfile(t, "upload"))

	response, err := app.Test(httptest.NewRequest(fiber.MethodPost, "/assets/upload-intents", nil), -1)
	if err != nil {
		t.Fatalf("Test() error = %v", err)
	}
	if response.StatusCode != fiber.StatusNoContent {
		t.Fatalf("StatusCode = %d, want %d", response.StatusCode, fiber.StatusNoContent)
	}
}

// TestRequestContextMiddlewareSkipsStreams verifies websocket streams remain unbounded.
func TestRequestContextMiddlewareSkipsStreams(t *testing.T) {
	app := fiber.New()
	app.Use(requestContextMiddleware(Config{RequestTimeout: time.Second}.Defaults()))
	app.Get("/events/ws", func(ctx *fiber.Ctx) error {
		if _, ok := ctx.UserContext().Deadline(); ok {
			t.Fatalf("stream context has deadline")
		}
		return ctx.SendStatus(fiber.StatusNoContent)
	})

	response, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/events/ws", nil), -1)
	if err != nil {
		t.Fatalf("Test() error = %v", err)
	}
	if response.StatusCode != fiber.StatusNoContent {
		t.Fatalf("StatusCode = %d, want %d", response.StatusCode, fiber.StatusNoContent)
	}
}

// assertProfile returns a handler that verifies the current request profile.
func assertProfile(t *testing.T, want string) fiber.Handler {
	t.Helper()
	return func(ctx *fiber.Ctx) error {
		got, ok := requestctx.CurrentProfile(ctx)
		if !ok || got != want {
			t.Fatalf("profile = %q ok=%v, want %q", got, ok, want)
		}
		return ctx.SendStatus(fiber.StatusNoContent)
	}
}
