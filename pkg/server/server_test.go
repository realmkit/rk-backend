package server

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	realmkitcors "github.com/realmkit/rk-backend/pkg/api/cors"
	"github.com/realmkit/rk-backend/pkg/api/headers"
	"github.com/realmkit/rk-backend/pkg/api/swagger"
	"github.com/realmkit/rk-backend/pkg/logger"
)

// TestNewServesAuthConfig verifies public auth config route wiring.
func TestNewServesAuthConfig(t *testing.T) {
	authConfig, userService, userServices := newUserServices(t)
	app := newApp(t, nil, true, WithAuth(authConfig, userService), WithUsers(userServices))
	req := httptest.NewRequest(fiber.MethodGet, "/auth/config", nil)

	res, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("Test() error = %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != fiber.StatusOK {
		t.Fatalf("StatusCode = %d, want %d", res.StatusCode, fiber.StatusOK)
	}
}

// TestNewUsesDefaultIdempotencyStore verifies isolated server construction succeeds.
func TestNewUsesDefaultIdempotencyStore(t *testing.T) {
	app := New(nil, false)
	req := httptest.NewRequest(fiber.MethodGet, "/health", nil)

	res, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("Test() error = %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != fiber.StatusNoContent {
		t.Fatalf("StatusCode = %d, want %d", res.StatusCode, fiber.StatusNoContent)
	}
}

// TestNewConfiguresLargeHeaderBuffer verifies auth cookies do not break docs.
func TestNewConfiguresLargeHeaderBuffer(t *testing.T) {
	app := newApp(t, nil, true)
	req := httptest.NewRequest(fiber.MethodGet, swagger.DocsPath, nil)
	req.Header.Set("Cookie", "realmkit="+strings.Repeat("x", 24*1024))

	res, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("Test() error = %v", err)
	}
	defer res.Body.Close()

	if app.Config().ReadBufferSize != DefaultReadBufferSize {
		t.Fatalf("ReadBufferSize = %d, want %d", app.Config().ReadBufferSize, DefaultReadBufferSize)
	}
	if res.StatusCode != fiber.StatusOK {
		t.Fatalf("StatusCode = %d, want %d", res.StatusCode, fiber.StatusOK)
	}
}

// TestNewServesHealth verifies the server exposes a health endpoint.
func TestNewServesHealth(t *testing.T) {
	app := newApp(t, nil, false)
	req := httptest.NewRequest(fiber.MethodGet, "/health", nil)

	res, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("Test() error = %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != fiber.StatusNoContent {
		t.Fatalf("StatusCode = %d, want %d", res.StatusCode, fiber.StatusNoContent)
	}
}

// TestNewRejectsGatewayVersionPrefix verifies RealmKit does not own public version prefixes.
func TestNewRejectsGatewayVersionPrefix(t *testing.T) {
	app := newApp(t, nil, false)
	req := httptest.NewRequest(fiber.MethodGet, "/api"+"/v1/health", nil)

	res, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("Test() error = %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != fiber.StatusNotFound {
		t.Fatalf("StatusCode = %d, want %d", res.StatusCode, fiber.StatusNotFound)
	}
}

// TestNewWritesRequestHeaders verifies common response headers are applied.
func TestNewWritesRequestHeaders(t *testing.T) {
	app := newApp(t, nil, false)
	req := httptest.NewRequest(fiber.MethodGet, "/health", nil)

	res, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("Test() error = %v", err)
	}
	defer res.Body.Close()

	if res.Header.Get(headers.RequestID) == "" {
		t.Fatalf("%s header = empty", headers.RequestID)
	}
	if res.Header.Get(headers.CorrelationID) == "" {
		t.Fatalf("%s header = empty", headers.CorrelationID)
	}
	if res.Header.Get(headers.RateLimitLimit) == "" {
		t.Fatalf("%s header = empty", headers.RateLimitLimit)
	}
}

// TestNewAppliesCORS verifies configured browser origins are allowed.
func TestNewAppliesCORS(t *testing.T) {
	app := newApp(t, nil, false, WithCORS(realmkitcors.Config{Enabled: true, AllowOrigins: "http://localhost:3000"}))
	req := httptest.NewRequest(fiber.MethodOptions, "/health", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	req.Header.Set("Access-Control-Request-Method", fiber.MethodGet)

	res, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("Test() error = %v", err)
	}
	defer res.Body.Close()

	if res.Header.Get("Access-Control-Allow-Origin") != "http://localhost:3000" {
		t.Fatalf("Access-Control-Allow-Origin = %q, want configured origin", res.Header.Get("Access-Control-Allow-Origin"))
	}
}

// TestNewUsesInjectedRateLimitStore verifies server options wire custom rate limit stores.
func TestNewUsesInjectedRateLimitStore(t *testing.T) {
	app := newApp(t, nil, false, WithRateLimitStore(denyingRateLimitStore{}))
	req := httptest.NewRequest(fiber.MethodGet, "/health", nil)

	res, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("Test() error = %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != fiber.StatusTooManyRequests {
		t.Fatalf("StatusCode = %d, want %d", res.StatusCode, fiber.StatusTooManyRequests)
	}
}

// TestNewUsesZapFiberMiddleware verifies Fiber access logs are emitted as JSON.
func TestNewUsesZapFiberMiddleware(t *testing.T) {
	var output bytes.Buffer
	log, err := logger.New(logger.Config{Level: "info"}, logger.WithOutput(&output))
	if err != nil {
		t.Fatalf("logger.New() error = %v", err)
	}

	app := newApp(t, log, false)
	req := httptest.NewRequest(fiber.MethodGet, "/health", nil)
	res, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("Test() error = %v", err)
	}
	defer res.Body.Close()

	if err := log.Sync(); err != nil {
		t.Fatalf("Sync() error = %v", err)
	}

	var entry map[string]any
	if err := json.Unmarshal(output.Bytes(), &entry); err != nil {
		t.Fatalf("Unmarshal() error = %v output = %q", err, output.String())
	}
	if entry["level"] != "info" {
		t.Fatalf("level = %v, want %v", entry["level"], "info")
	}
	if entry["status"] != float64(fiber.StatusNoContent) {
		t.Fatalf("status = %v, want %v", entry["status"], fiber.StatusNoContent)
	}
}

// TestNewControlsFiberStartupMessage verifies Fiber's banner is development-only.
func TestNewControlsFiberStartupMessage(t *testing.T) {
	development := newApp(t, nil, true)
	production := newApp(t, nil, false)

	if development.Config().DisableStartupMessage {
		t.Fatalf("development DisableStartupMessage = true, want false")
	}
	if !production.Config().DisableStartupMessage {
		t.Fatalf("production DisableStartupMessage = false, want true")
	}
}

// TestNewServesSwaggerOnlyInDevelopment verifies Swagger follows the development gate.
func TestNewServesSwaggerOnlyInDevelopment(t *testing.T) {
	development := newApp(t, nil, true)
	production := newApp(t, nil, false)

	devRes, err := development.Test(httptest.NewRequest(fiber.MethodGet, swagger.DocsPath, nil), -1)
	if err != nil {
		t.Fatalf("development Test() error = %v", err)
	}
	defer devRes.Body.Close()
	prodRes, err := production.Test(httptest.NewRequest(fiber.MethodGet, swagger.DocsPath, nil), -1)
	if err != nil {
		t.Fatalf("production Test() error = %v", err)
	}
	defer prodRes.Body.Close()

	if devRes.StatusCode != fiber.StatusOK {
		t.Fatalf("development StatusCode = %d, want %d", devRes.StatusCode, fiber.StatusOK)
	}
	if prodRes.StatusCode != fiber.StatusNotFound {
		t.Fatalf("production StatusCode = %d, want %d", prodRes.StatusCode, fiber.StatusNotFound)
	}
}
