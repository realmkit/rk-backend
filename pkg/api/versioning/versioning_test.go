package versioning

import (
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
)

// TestV1DefinesVersionPrefix verifies v1 is centrally defined.
func TestV1DefinesVersionPrefix(t *testing.T) {
	if V1.Name != "v1" {
		t.Fatalf("Name = %q, want %q", V1.Name, "v1")
	}
	if V1.Prefix != "/api/v1" {
		t.Fatalf("Prefix = %q, want %q", V1.Prefix, "/api/v1")
	}
}

// TestGroupScopesRoutes verifies routes are registered under the version prefix.
func TestGroupScopesRoutes(t *testing.T) {
	app := fiber.New()
	V1.Group(app).Get("/test", func(ctx *fiber.Ctx) error {
		return ctx.SendStatus(fiber.StatusNoContent)
	})

	res, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/api/v1/test", nil), -1)
	if err != nil {
		t.Fatalf("Test() error = %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != fiber.StatusNoContent {
		t.Fatalf("StatusCode = %d, want %d", res.StatusCode, fiber.StatusNoContent)
	}
}
