// Package pkgapi_e2e verifies shared API middleware through real server requests.
package pkgapi_e2e

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/e2e/harness"
	"github.com/realmkit/rk-backend/pkg/api/auth"
	"github.com/realmkit/rk-backend/pkg/api/headers"
	"github.com/realmkit/rk-backend/pkg/api/principal"
	"github.com/realmkit/rk-backend/pkg/api/problem"
	"github.com/realmkit/rk-backend/pkg/api/ratelimit"
	"github.com/realmkit/rk-backend/pkg/identity"
	"github.com/realmkit/rk-backend/pkg/server"
)

// TestAPIIdempotencyReplayAndConflict verifies POST replay behavior.
func TestAPIIdempotencyReplayAndConflict(t *testing.T) {
	steps := harness.NewSteps(t)
	steps.Log("start server with Redis-backed idempotency")
	ecosystem := harness.New(t)
	registerIdempotentRoute(ecosystem)

	steps.Log("send first idempotent request")
	first := postWithKey("same-key", `{"name":"one"}`)
	firstResponse := ecosystem.Test(t, first)
	firstBody := harness.ResponseBody(t, firstResponse)
	assertStatus(t, firstResponse.StatusCode, fiber.StatusCreated, firstBody)

	steps.Log("replay identical request with same idempotency key")
	replay := postWithKey("same-key", `{"name":"one"}`)
	replayResponse := ecosystem.Test(t, replay)
	replayBody := harness.ResponseBody(t, replayResponse)
	assertStatus(t, replayResponse.StatusCode, fiber.StatusCreated, replayBody)
	if replayBody != firstBody {
		t.Fatalf("replay body = %q, want %q", replayBody, firstBody)
	}

	steps.Log("reuse same idempotency key with different request body")
	conflict := postWithKey("same-key", `{"name":"two"}`)
	conflictResponse := ecosystem.Test(t, conflict)
	conflictBody := harness.ResponseBody(t, conflictResponse)
	assertStatus(t, conflictResponse.StatusCode, fiber.StatusConflict, conflictBody)
}

// TestAPIRateLimitProblem verifies rate limit denials use problem responses.
func TestAPIRateLimitProblem(t *testing.T) {
	steps := harness.NewSteps(t)
	steps.Log("start server with denying rate limit store")
	ecosystem := harness.New(t, harness.WithServerOptions(server.WithRateLimitStore(denyingStore{})))

	steps.Log("send request that exceeds configured rate limit")
	response := ecosystem.Test(t, harness.JSONRequest(fiber.MethodGet, "/health", ""))
	body := harness.ResponseBody(t, response)

	assertStatus(t, response.StatusCode, fiber.StatusTooManyRequests, body)
	assertHeaderPresent(t, response.Header.Get(headers.RetryAfter), headers.RetryAfter)
	assertProblemCode(t, body, "rate_limited")
}

// TestAPIJSONMiddleware verifies shared JSON content negotiation.
func TestAPIJSONMiddleware(t *testing.T) {
	steps := harness.NewSteps(t)
	steps.Log("start server and attach JSON-enforced test route")
	ecosystem := harness.New(t)
	ecosystem.App.Group("", headers.RequireJSON()).Post("/e2e/json", func(ctx *fiber.Ctx) error {
		return ctx.JSON(map[string]string{"ok": "true"})
	})

	steps.Log("send request with unsupported content type")
	request := harness.JSONRequest(fiber.MethodPost, "/e2e/json", `{"ok":true}`)
	request.Header.Set(headers.ContentType, "text/plain")
	response := ecosystem.Test(t, request)
	body := harness.ResponseBody(t, response)

	assertStatus(t, response.StatusCode, fiber.StatusUnsupportedMediaType, body)
	assertProblemCode(t, body, "unsupported_media_type")
}

// TestAPIAuthDevelopmentPrincipal verifies auth middleware sets principals.
func TestAPIAuthDevelopmentPrincipal(t *testing.T) {
	steps := harness.NewSteps(t)
	userID := uuid.New()
	provisioner := &developmentProvisioner{
		principal: principal.Principal{
			UserID:            userID,
			SubjectHash:       "dev:" + userID.String(),
			DevelopmentBypass: true,
		},
	}

	steps.Log("start server and attach protected development route")
	ecosystem := harness.New(t, harness.WithDevelopment(true))
	ecosystem.App.Get(
		"/e2e/protected",
		auth.Middleware(
			auth.Config{DevelopmentBypass: true},
			nil,
			provisioner,
			auth.MiddlewareConfig{Development: true, Log: ecosystem.Log},
		),
		func(ctx *fiber.Ctx) error {
			current, err := principal.Require(ctx)
			if err != nil {
				return err
			}
			return ctx.JSON(map[string]string{"user_id": current.UserID.String()})
		},
	)

	steps.Log("verify anonymous request is rejected")
	anonymous := ecosystem.Test(t, harness.JSONRequest(fiber.MethodGet, "/e2e/protected", ""))
	anonymousBody := harness.ResponseBody(t, anonymous)
	assertStatus(t, anonymous.StatusCode, fiber.StatusUnauthorized, anonymousBody)
	assertProblemCode(t, anonymousBody, "unauthenticated")

	steps.Log("verify development principal reaches handler")
	request := harness.JSONRequest(fiber.MethodGet, "/e2e/protected", "")
	request.Header.Set(auth.DevUserIDHeader, userID.String())
	response := ecosystem.Test(t, request)
	body := harness.ResponseBody(t, response)
	assertStatus(t, response.StatusCode, fiber.StatusOK, body)
	if provisioner.developmentCalls != 1 {
		t.Fatalf("developmentCalls = %d, want 1", provisioner.developmentCalls)
	}
	if !contains(body, userID.String()) {
		t.Fatalf("body = %q, want user id", body)
	}
}

// registerIdempotentRoute adds a POST route behind global middleware.
func registerIdempotentRoute(ecosystem *harness.Ecosystem) {
	counter := 0
	ecosystem.App.Post("/e2e/idempotent", func(ctx *fiber.Ctx) error {
		counter++
		return ctx.Status(fiber.StatusCreated).JSON(map[string]any{
			"counter": counter,
			"body":    string(ctx.Body()),
		})
	})
}

// postWithKey creates an idempotent POST request.
func postWithKey(key string, body string) *http.Request {
	request := harness.JSONRequest(fiber.MethodPost, "/e2e/idempotent", body)
	request.Header.Set(headers.IdempotencyKey, key)
	return request
}

// denyingStore rejects every rate-limit decision.
type denyingStore struct{}

// Allow returns a denied decision.
func (denyingStore) Allow(context.Context, string, ratelimit.Policy) (ratelimit.Decision, error) {
	return ratelimit.Decision{
		Allowed:   false,
		Limit:     1,
		Remaining: 0,
		ResetAt:   time.Now().Add(time.Minute),
	}, nil
}

// developmentProvisioner returns development principals for auth e2e tests.
type developmentProvisioner struct {
	principal        principal.Principal
	developmentCalls int
}

// Provision resolves bearer identities.
func (provisioner *developmentProvisioner) Provision(
	context.Context,
	identity.ExternalIdentity,
	auth.Token,
) (principal.Principal, error) {
	return provisioner.principal, nil
}

// DevelopmentPrincipal returns the configured development principal.
func (provisioner *developmentProvisioner) DevelopmentPrincipal(
	context.Context,
	uuid.UUID,
) (principal.Principal, error) {
	provisioner.developmentCalls++
	return provisioner.principal, nil
}

// assertStatus verifies a response status.
func assertStatus(t *testing.T, got int, want int, body string) {
	t.Helper()
	if got != want {
		t.Fatalf("StatusCode = %d, want %d body = %q", got, want, body)
	}
}

// assertHeader verifies a response header.
func assertHeader(t *testing.T, got string, want string, name string) {
	t.Helper()
	if got != want {
		t.Fatalf("%s = %q, want %q", name, got, want)
	}
}

// assertHeaderPresent verifies a response header exists.
func assertHeaderPresent(t *testing.T, got string, name string) {
	t.Helper()
	if got == "" {
		t.Fatalf("%s header = empty", name)
	}
}

// assertProblemCode verifies a problem response code.
func assertProblemCode(t *testing.T, body string, code string) {
	t.Helper()
	var payload problem.Problem
	if err := json.Unmarshal([]byte(body), &payload); err != nil {
		t.Fatalf("problem JSON error = %v body = %q", err, body)
	}
	if payload.Code != code || payload.RequestID == "" {
		t.Fatalf("problem = %+v, want code %q with request id", payload, code)
	}
}

// contains reports whether text contains part.
func contains(text string, part string) bool {
	return strings.Contains(text, part)
}
