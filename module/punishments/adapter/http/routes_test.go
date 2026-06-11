package http

import (
	"bytes"
	"net/http"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/module/punishments/domain"
	"github.com/realmkit/rk-backend/module/punishments/port"
	"github.com/realmkit/rk-backend/pkg/api/headers"
)

// TestDefinitionRoutesCoverSuccessPaths verifies definition admin route mappings.
func TestDefinitionRoutesCoverSuccessPaths(t *testing.T) {
	app := newTestApp(httpService{})
	id := uuid.NewString()
	routes := []struct {
		method string
		path   string
		body   string
		status int
	}{
		{http.MethodPost, "/punishment-definitions", validDefinitionBody(), fiber.StatusCreated},
		{http.MethodGet, "/punishment-definitions?status=active&page_size=10", "", fiber.StatusOK},
		{http.MethodGet, "/punishment-definitions/" + id, "", fiber.StatusOK},
		{http.MethodPatch, "/punishment-definitions/" + id, validDefinitionBody(), fiber.StatusOK},
		{http.MethodDelete, "/punishment-definitions/" + id, "", fiber.StatusNoContent},
		{
			http.MethodPost,
			"/punishment-definitions/" + id + "/actions/reorder",
			`{"ids":["` + uuid.NewString() + `"]}`,
			fiber.StatusNoContent,
		},
	}

	for _, route := range routes {
		req := authenticatedRequest(t, route.method, route.path, route.body)
		req.Header.Set(headers.IdempotencyKey, "definition-route")
		req.Header.Set(headers.IfMatch, `"1"`)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("%s %s error = %v", route.method, route.path, err)
		}
		if resp.StatusCode != route.status {
			t.Fatalf("%s %s status = %d, want %d", route.method, route.path, resp.StatusCode, route.status)
		}
	}
}

// TestPunishmentRoutesCoverSuccessPaths verifies punishment case route mappings.
func TestPunishmentRoutesCoverSuccessPaths(t *testing.T) {
	app := newTestApp(httpService{})
	punishmentID := uuid.NewString()
	userID := uuid.NewString()
	routes := []struct {
		method string
		path   string
		body   string
		status int
	}{
		{http.MethodGet, "/punishments?target_user_id=" + userID + "&status=active", "", fiber.StatusOK},
		{http.MethodGet, "/punishments/" + punishmentID, "", fiber.StatusOK},
		{http.MethodPatch, "/punishments/" + punishmentID, `{"reason":"updated"}`, fiber.StatusOK},
		{http.MethodPost, "/punishments/" + punishmentID + "/revoke", `{"reason":"appeal"}`, fiber.StatusNoContent},
		{http.MethodGet, "/users/" + userID + "/punishments", "", fiber.StatusOK},
		{http.MethodGet, "/users/" + userID + "/punishments/active", "", fiber.StatusOK},
		{http.MethodGet, "/users/" + userID + "/punishments/restrictions", "", fiber.StatusOK},
	}

	for _, route := range routes {
		req := authenticatedRequest(t, route.method, route.path, route.body)
		req.Header.Set(headers.IdempotencyKey, "punishment-route")
		req.Header.Set(headers.IfMatch, `"1"`)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("%s %s error = %v", route.method, route.path, err)
		}
		if resp.StatusCode != route.status {
			t.Fatalf("%s %s status = %d, want %d", route.method, route.path, resp.StatusCode, route.status)
		}
	}
}

// TestPunishmentHTTPProblemMappings verifies adapter-level validation and application errors.
func TestPunishmentHTTPProblemMappings(t *testing.T) {
	validationErr := domain.NewValidationError([]domain.Violation{
		{Field: "reason", Message: "is required"},
	})
	cases := []struct {
		name    string
		service httpService
		method  string
		path    string
		body    string
		status  int
	}{
		{"invalid user", httpService{}, http.MethodGet, "/punishments", "", fiber.StatusBadRequest},
		{"invalid path", httpService{}, http.MethodGet, "/punishments/not-a-uuid", "", fiber.StatusBadRequest},
		{"invalid pagination", httpService{}, http.MethodGet, "/punishments?page_size=-1", "", fiber.StatusBadRequest},
		{"invalid if match", httpService{}, http.MethodPatch, "/punishments/" + uuid.NewString(), `{}`, fiber.StatusBadRequest},
		{"not found", httpService{err: port.ErrNotFound}, http.MethodGet, "/punishments/" + uuid.NewString(), "", fiber.StatusNotFound},
		{
			"precondition",
			httpService{err: port.ErrPreconditionFailed},
			http.MethodPatch,
			"/punishments/" + uuid.NewString(),
			`{"reason":"x"}`,
			fiber.StatusPreconditionFailed,
		},
		{
			"conflict",
			httpService{err: port.ErrConflict},
			http.MethodPost,
			"/punishments",
			validIssueBody(),
			fiber.StatusConflict,
		},
		{
			"forbidden",
			httpService{err: port.ErrForbidden},
			http.MethodPost,
			"/punishments/restrictions/check",
			`{"user_id":"` + uuid.NewString() + `","action_key":"` + domain.ActionForumsReply + `"}`,
			fiber.StatusForbidden,
		},
		{
			"validation",
			httpService{err: validationErr},
			http.MethodPost,
			"/punishment-definitions",
			validDefinitionBody(),
			fiber.StatusUnprocessableEntity,
		},
	}

	for _, item := range cases {
		app := newTestApp(item.service)
		req := authenticatedRequest(t, item.method, item.path, item.body)
		req.Header.Set(headers.IdempotencyKey, "problem-route")
		req.Header.Set(headers.IfMatch, `"1"`)
		if item.name == "invalid if match" {
			req.Header.Set(headers.IfMatch, "not-a-version")
		}
		if item.name == "invalid user" {
			req.Header.Set(currentUserIDHeader, "not-a-uuid")
		}
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("%s error = %v", item.name, err)
		}
		if resp.StatusCode != item.status {
			t.Fatalf("%s status = %d, want %d", item.name, resp.StatusCode, item.status)
		}
	}
}

// authenticatedRequest builds a request with a current-user header.
func authenticatedRequest(t *testing.T, method string, path string, body string) *http.Request {
	t.Helper()
	req, err := http.NewRequest(method, path, bytes.NewBufferString(body))
	if err != nil {
		t.Fatalf("NewRequest() error = %v", err)
	}
	req.Header.Set(headers.ContentType, "application/json")
	req.Header.Set(currentUserIDHeader, uuid.NewString())
	return req
}

// validDefinitionBody returns a valid definition request payload.
func validDefinitionBody() string {
	return `{
		"key":"chat_ban",
		"name":"Chat Ban",
		"color":"#ff5555",
		"status":"active",
		"allow_permanent":true,
		"requires_reason":true,
		"actions":[{
			"id":"` + uuid.NewString() + `",
			"target_system":"realmkit",
			"action_key":"` + domain.ActionForumsReply + `",
			"effect":"restrict",
			"configuration_json":{},
			"status":"active"
		}]
	}`
}

// validIssueBody returns a valid punishment issue request payload.
func validIssueBody() string {
	return `{
		"definition_id":"` + uuid.NewString() + `",
		"target_user_id":"` + uuid.NewString() + `",
		"issuer_type":"system",
		"issuer_key":"test",
		"reason":"spam"
	}`
}
