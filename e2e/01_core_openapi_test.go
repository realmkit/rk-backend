package e2e

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/niflaot/gamehub-go/e2e/harness"
	"github.com/niflaot/gamehub-go/pkg/api/openapi"
	"github.com/niflaot/gamehub-go/pkg/api/swagger"
)

// TestCoreOpenAPIServedInDevelopment verifies docs are development-only.
func TestCoreOpenAPIServedInDevelopment(t *testing.T) {
	steps := harness.NewSteps(t)
	steps.Log("start development server")
	development := harness.New(t, harness.WithDevelopment(true))

	steps.Log("fetch OpenAPI document")
	response := development.Test(t, harness.JSONRequest(fiber.MethodGet, swagger.OpenAPIPath, ""))
	body := harness.ResponseBody(t, response)
	assertStatus(t, response.StatusCode, fiber.StatusOK, body)

	var document map[string]any
	if err := json.Unmarshal([]byte(body), &document); err != nil {
		t.Fatalf("OpenAPI JSON error = %v", err)
	}
	if document["openapi"] == "" || document["paths"] == nil {
		t.Fatalf("OpenAPI document missing required fields: %+v", document)
	}

	steps.Log("start production-mode server")
	production := harness.New(t)
	response = production.Test(t, harness.JSONRequest(fiber.MethodGet, swagger.OpenAPIPath, ""))
	assertStatus(t, response.StatusCode, fiber.StatusNotFound, harness.ResponseBody(t, response))
}

// TestCoreSwaggerUIServedInDevelopment verifies Swagger UI is available.
func TestCoreSwaggerUIServedInDevelopment(t *testing.T) {
	steps := harness.NewSteps(t)
	steps.Log("start development server")
	ecosystem := harness.New(t, harness.WithDevelopment(true))

	steps.Log("fetch Swagger UI")
	response := ecosystem.Test(t, harness.JSONRequest(fiber.MethodGet, swagger.DocsPath, ""))
	body := harness.ResponseBody(t, response)

	assertStatus(t, response.StatusCode, fiber.StatusOK, body)
	if !strings.Contains(body, "SwaggerUIBundle") {
		t.Fatalf("Swagger UI body missing bundle bootstrap")
	}
}

// TestCoreOpenAPICoversInfrastructureRoutes verifies e2e-used routes are documented.
func TestCoreOpenAPICoversInfrastructureRoutes(t *testing.T) {
	steps := harness.NewSteps(t)
	for _, route := range infrastructureRoutes() {
		steps.Log("verify OpenAPI operation %s %s", route.method, route.path)
		exists, err := openapi.OperationExists(route.method, route.path)
		if err != nil {
			t.Fatalf("OperationExists() error = %v", err)
		}
		if !exists {
			t.Fatalf("OpenAPI operation missing for %s %s", route.method, route.path)
		}
	}
}

// routeCheck identifies one documented route.
type routeCheck struct {
	method string
	path   string
}

// infrastructureRoutes returns package infrastructure routes exercised by e2e.
func infrastructureRoutes() []routeCheck {
	return []routeCheck{
		{method: fiber.MethodGet, path: "/health"},
		{method: fiber.MethodGet, path: "/auth/config"},
		{method: fiber.MethodGet, path: "/events"},
		{method: fiber.MethodGet, path: "/cronjobs"},
		{method: fiber.MethodPost, path: "/cronjobs/{job_key}/run"},
		{method: fiber.MethodPost, path: "/assets/upload-intents"},
		{method: fiber.MethodGet, path: "/forums/tree"},
	}
}
