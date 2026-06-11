package e2e

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/niflaot/gamehub-go/e2e/harness"
)

// TestCoreRequestLogging verifies request logs are structured and readable.
func TestCoreRequestLogging(t *testing.T) {
	steps := harness.NewSteps(t)
	steps.Log("start server with captured logger")
	ecosystem := harness.New(t)

	steps.Log("send health request to produce one access log")
	response := ecosystem.Test(t, harness.JSONRequest(fiber.MethodGet, "/health", ""))
	assertStatus(t, response.StatusCode, fiber.StatusNoContent, harness.ResponseBody(t, response))

	steps.Log("parse captured zap JSON log entries")
	entry := findLogEntry(t, ecosystem.LogBuffer.String(), "status", float64(fiber.StatusNoContent))
	if entry["method"] != fiber.MethodGet || entry["url"] != "/health" {
		t.Fatalf("log entry = %+v, want GET /health", entry)
	}
}

// findLogEntry returns the first JSON log entry with field equal to value.
func findLogEntry(t *testing.T, logs string, field string, value any) map[string]any {
	t.Helper()
	for _, line := range strings.Split(strings.TrimSpace(logs), "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		var entry map[string]any
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			t.Fatalf("log JSON error = %v line = %q", err, line)
		}
		if entry[field] == value {
			return entry
		}
	}
	t.Fatalf("log entry with %s=%v not found in %q", field, value, logs)
	return nil
}
