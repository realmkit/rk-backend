package cors

import (
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
)

// TestNewAllowsConfiguredOrigin verifies configured browser origins are allowed.
func TestNewAllowsConfiguredOrigin(t *testing.T) {
	app := fiber.New()
	app.Use(New(Config{Enabled: true, AllowOrigins: "http://localhost:3000"}))
	app.Get("/test", func(ctx *fiber.Ctx) error {
		return ctx.SendStatus(fiber.StatusNoContent)
	})

	req := httptest.NewRequest(fiber.MethodGet, "/test", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	res, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("Test() error = %v", err)
	}
	defer res.Body.Close()

	if res.Header.Get("Access-Control-Allow-Origin") != "http://localhost:3000" {
		t.Fatalf("Access-Control-Allow-Origin = %q, want configured origin", res.Header.Get("Access-Control-Allow-Origin"))
	}
}

// TestNewCanBeDisabled verifies disabled CORS writes no CORS headers.
func TestNewCanBeDisabled(t *testing.T) {
	app := fiber.New()
	app.Use(New(Config{Enabled: false}))
	app.Get("/test", func(ctx *fiber.Ctx) error {
		return ctx.SendStatus(fiber.StatusNoContent)
	})

	req := httptest.NewRequest(fiber.MethodGet, "/test", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	res, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("Test() error = %v", err)
	}
	defer res.Body.Close()

	if res.Header.Get("Access-Control-Allow-Origin") != "" {
		t.Fatalf("Access-Control-Allow-Origin = %q, want empty", res.Header.Get("Access-Control-Allow-Origin"))
	}
}

// TestNormalizedOriginsTrimsEmptyEntries verifies origin lists are normalized.
func TestNormalizedOriginsTrimsEmptyEntries(t *testing.T) {
	got := normalizedOrigins(" http://a.test, ,http://b.test ")
	want := "http://a.test,http://b.test"
	if got != want {
		t.Fatalf("normalizedOrigins() = %q, want %q", got, want)
	}
}
