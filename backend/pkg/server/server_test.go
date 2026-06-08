package server

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/niflaot/gamehub/backend/pkg/logger"
)

// TestConfigAddress verifies server addresses are formatted for Listen.
func TestConfigAddress(t *testing.T) {
	cfg := Config{Host: "127.0.0.1", Port: 9090}

	if cfg.Address() != "127.0.0.1:9090" {
		t.Fatalf("Address() = %q, want %q", cfg.Address(), "127.0.0.1:9090")
	}
}

// TestNewServesHealth verifies the server exposes a health endpoint.
func TestNewServesHealth(t *testing.T) {
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

// TestNewUsesZapFiberMiddleware verifies Fiber access logs are emitted as JSON.
func TestNewUsesZapFiberMiddleware(t *testing.T) {
	var output bytes.Buffer
	log, err := logger.New(logger.Config{Level: "info"}, logger.WithOutput(&output))
	if err != nil {
		t.Fatalf("logger.New() error = %v", err)
	}

	app := New(log, false)
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
	development := New(nil, true)
	production := New(nil, false)

	if development.Config().DisableStartupMessage {
		t.Fatalf("development DisableStartupMessage = true, want false")
	}
	if !production.Config().DisableStartupMessage {
		t.Fatalf("production DisableStartupMessage = false, want true")
	}
}
