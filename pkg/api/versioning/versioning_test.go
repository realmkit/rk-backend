package versioning

import (
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
)

// TestServiceDefinesUnversionedPrefix verifies GameHub does not own public version prefixes.
func TestServiceDefinesUnversionedPrefix(t *testing.T) {
	if Service.Name != "gateway-owned" {
		t.Fatalf("Name = %q, want %q", Service.Name, "gateway-owned")
	}
	if Service.Prefix != "" {
		t.Fatalf("Prefix = %q, want empty", Service.Prefix)
	}
}

// TestGroupScopesRoutes verifies routes are registered without a service prefix.
func TestGroupScopesRoutes(t *testing.T) {
	app := fiber.New()
	Service.Group(app).Get("/test", func(ctx *fiber.Ctx) error {
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
}
