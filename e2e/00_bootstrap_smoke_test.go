// Package e2e contains in-process end-to-end ecosystem tests.
package e2e

import (
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/niflaot/gamehub-go/e2e/harness"
)

// TestBootstrapServesHealth verifies the e2e ecosystem starts the server.
func TestBootstrapServesHealth(t *testing.T) {
	ecosystem := harness.New(t)
	response := ecosystem.Test(t, harness.JSONRequest(fiber.MethodGet, "/health", ""))

	assertStatus(t, response.StatusCode, fiber.StatusNoContent, harness.ResponseBody(t, response))
}

// TestBootstrapUsesGatewayVersioning verifies service routes stay unversioned.
func TestBootstrapUsesGatewayVersioning(t *testing.T) {
	ecosystem := harness.New(t)
	response := ecosystem.Test(t, harness.JSONRequest(fiber.MethodGet, "/api"+"/v1/health", ""))

	assertStatus(t, response.StatusCode, fiber.StatusNotFound, harness.ResponseBody(t, response))
}

// assertStatus verifies an HTTP status code and includes body diagnostics.
func assertStatus(t *testing.T, got int, want int, body string) {
	t.Helper()

	if got != want {
		t.Fatalf("StatusCode = %d, want %d body = %q", got, want, body)
	}
}
