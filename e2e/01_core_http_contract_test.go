package e2e

import (
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/realmkit/rk-backend/e2e/harness"
	"github.com/realmkit/rk-backend/pkg/api/headers"
	"github.com/realmkit/rk-backend/pkg/api/problem"
)

// TestCoreHeadersAndRateLimit verifies common response headers.
func TestCoreHeadersAndRateLimit(t *testing.T) {
	steps := harness.NewSteps(t)
	steps.Log("start in-process server")
	ecosystem := harness.New(t)

	steps.Log("request health endpoint with caller-provided correlation id")
	request := harness.JSONRequest(fiber.MethodGet, "/health", "")
	request.Header.Set(headers.CorrelationID, "trace-core")
	response := ecosystem.Test(t, request)

	assertStatus(t, response.StatusCode, fiber.StatusNoContent, harness.ResponseBody(t, response))
	assertHeaderPresent(t, response.Header.Get(headers.RequestID), headers.RequestID)
	assertHeader(t, response.Header.Get(headers.CorrelationID), "trace-core", headers.CorrelationID)
	assertHeaderPresent(t, response.Header.Get(headers.RateLimitLimit), headers.RateLimitLimit)
	assertHeaderPresent(t, response.Header.Get(headers.RateLimitRemaining), headers.RateLimitRemaining)
	assertHeaderPresent(t, response.Header.Get(headers.RateLimitReset), headers.RateLimitReset)
}

// TestCoreCORSPreflight verifies configured browser origins are allowed.
func TestCoreCORSPreflight(t *testing.T) {
	steps := harness.NewSteps(t)
	steps.Log("start in-process server with default CORS configuration")
	ecosystem := harness.New(t)

	steps.Log("send browser preflight request")
	request := httptest.NewRequest(fiber.MethodOptions, "/health", nil)
	request.Header.Set("Origin", "http://localhost:3000")
	request.Header.Set("Access-Control-Request-Method", fiber.MethodGet)
	response := ecosystem.Test(t, request)

	assertStatus(t, response.StatusCode, fiber.StatusNoContent, harness.ResponseBody(t, response))
	assertHeader(
		t,
		response.Header.Get("Access-Control-Allow-Origin"),
		"http://localhost:3000",
		"Access-Control-Allow-Origin",
	)
}

// TestCoreProblemResponse verifies server errors use the shared problem shape.
func TestCoreProblemResponse(t *testing.T) {
	steps := harness.NewSteps(t)
	steps.Log("start in-process server")
	ecosystem := harness.New(t)

	steps.Log("request missing route")
	response := ecosystem.Test(t, harness.JSONRequest(fiber.MethodGet, "/missing", ""))
	body := harness.ResponseBody(t, response)

	assertStatus(t, response.StatusCode, fiber.StatusNotFound, body)
	assertHeader(t, response.Header.Get(headers.ContentType), problem.ContentType, headers.ContentType)
	var payload map[string]any
	if err := json.Unmarshal([]byte(body), &payload); err != nil {
		t.Fatalf("problem JSON error = %v body = %q", err, body)
	}
	if payload["code"] != "not_found" || payload["request_id"] == "" {
		t.Fatalf("problem = %+v, want not_found with request_id", payload)
	}
}

// assertHeader verifies one header value.
func assertHeader(t *testing.T, got string, want string, name string) {
	t.Helper()
	if got != want {
		t.Fatalf("%s = %q, want %q", name, got, want)
	}
}

// assertHeaderPresent verifies one required header exists.
func assertHeaderPresent(t *testing.T, got string, name string) {
	t.Helper()
	if got == "" {
		t.Fatalf("%s header = empty", name)
	}
}
