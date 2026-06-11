package http

import (
	"bytes"
	"context"
	"net/http"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/module/punishments/domain"
	"github.com/realmkit/rk-backend/module/punishments/port"
	"github.com/realmkit/rk-backend/pkg/api/headers"
	"github.com/realmkit/rk-backend/pkg/api/problem"
	"github.com/realmkit/rk-backend/pkg/pagination"
)

// TestIssuePunishmentRequiresIdempotency verifies retryable issue commands.
func TestIssuePunishmentRequiresIdempotency(t *testing.T) {
	app := newTestApp(httpService{})
	req, _ := http.NewRequest(http.MethodPost, "/punishments", bytes.NewBufferString(`{}`))
	req.Header.Set(headers.ContentType, "application/json")
	req.Header.Set(currentUserIDHeader, uuid.NewString())

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Test() error = %v", err)
	}
	if resp.StatusCode != fiber.StatusBadRequest {
		t.Fatalf("status = %d, want %d", resp.StatusCode, fiber.StatusBadRequest)
	}
}

// TestIssuePunishmentReturnsCreated verifies successful issue response metadata.
func TestIssuePunishmentReturnsCreated(t *testing.T) {
	app := newTestApp(httpService{})
	body := `{"definition_id":"` + uuid.NewString() + `","target_user_id":"` +
		uuid.NewString() + `","issuer_type":"system","issuer_key":"test","reason":"spam"}`
	req, _ := http.NewRequest(http.MethodPost, "/punishments", bytes.NewBufferString(body))
	req.Header.Set(headers.ContentType, "application/json")
	req.Header.Set(headers.IdempotencyKey, "issue-1")
	req.Header.Set(currentUserIDHeader, uuid.NewString())

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Test() error = %v", err)
	}
	if resp.StatusCode != fiber.StatusCreated {
		t.Fatalf("status = %d, want %d", resp.StatusCode, fiber.StatusCreated)
	}
	if resp.Header.Get(headers.ETag) != `"1"` {
		t.Fatalf("ETag = %q, want %q", resp.Header.Get(headers.ETag), `"1"`)
	}
}

// TestUpdatePunishmentRequiresIfMatch verifies optimistic concurrency headers.
func TestUpdatePunishmentRequiresIfMatch(t *testing.T) {
	app := newTestApp(httpService{})
	req, _ := http.NewRequest(
		http.MethodPatch,
		"/punishments/"+uuid.NewString(),
		bytes.NewBufferString(`{"reason":"updated"}`),
	)
	req.Header.Set(headers.ContentType, "application/json")
	req.Header.Set(currentUserIDHeader, uuid.NewString())

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Test() error = %v", err)
	}
	if resp.StatusCode != fiber.StatusPreconditionRequired {
		t.Fatalf("status = %d, want %d", resp.StatusCode, fiber.StatusPreconditionRequired)
	}
}

// TestCheckRestrictionReturnsDenied verifies restriction checks are exposed.
func TestCheckRestrictionReturnsDenied(t *testing.T) {
	app := newTestApp(httpService{denied: true})
	body := `{"user_id":"` + uuid.NewString() + `","action_key":"` +
		domain.ActionForumsReply + `"}`
	req, _ := http.NewRequest(
		http.MethodPost,
		"/punishments/restrictions/check",
		bytes.NewBufferString(body),
	)
	req.Header.Set(headers.ContentType, "application/json")
	req.Header.Set(currentUserIDHeader, uuid.NewString())

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Test() error = %v", err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, fiber.StatusOK)
	}
}

// TestPrivatePunishmentRoutesRequireUser verifies moderation state is not anonymously readable.
func TestPrivatePunishmentRoutesRequireUser(t *testing.T) {
	app := newTestApp(httpService{})
	id := uuid.NewString()
	userID := uuid.NewString()
	for _, req := range []*http.Request{
		newHTTPTestRequest(t, http.MethodPost, "/punishment-definitions", `{}`),
		newHTTPTestRequest(t, http.MethodGet, "/punishment-definitions", ``),
		newHTTPTestRequest(t, http.MethodGet, "/punishment-definitions/"+id, ``),
		newHTTPTestRequest(t, http.MethodPatch, "/punishment-definitions/"+id, `{}`),
		newHTTPTestRequest(t, http.MethodDelete, "/punishment-definitions/"+id, ``),
		newHTTPTestRequest(t, http.MethodPost, "/punishment-definitions/"+id+"/actions/reorder", `{}`),
		newHTTPTestRequest(t, http.MethodPost, "/punishments", `{}`),
		newHTTPTestRequest(t, http.MethodGet, "/punishments", ``),
		newHTTPTestRequest(t, http.MethodGet, "/punishments/"+id, ``),
		newHTTPTestRequest(t, http.MethodPatch, "/punishments/"+id, `{}`),
		newHTTPTestRequest(t, http.MethodPost, "/punishments/"+id+"/revoke", `{}`),
		newHTTPTestRequest(t, http.MethodGet, "/users/"+userID+"/punishments", ``),
		newHTTPTestRequest(t, http.MethodGet, "/users/"+userID+"/punishments/active", ``),
		newHTTPTestRequest(t, http.MethodPost, "/punishments/restrictions/check", `{}`),
		newHTTPTestRequest(t, http.MethodGet, "/users/"+userID+"/punishments/restrictions", ``),
	} {
		req.Header.Set(headers.ContentType, "application/json")
		req.Header.Set(headers.IdempotencyKey, "punishment-security")
		req.Header.Set(headers.IfMatch, `"1"`)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("%s %s error = %v", req.Method, req.URL.Path, err)
		}
		if resp.StatusCode != fiber.StatusUnauthorized {
			t.Fatalf("%s %s status = %d, want 401", req.Method, req.URL.Path, resp.StatusCode)
		}
	}
}

func newHTTPTestRequest(t *testing.T, method string, path string, body string) *http.Request {
	t.Helper()
	req, err := http.NewRequest(method, path, bytes.NewBufferString(body))
	if err != nil {
		t.Fatalf("NewRequest() error = %v", err)
	}
	return req
}

func newTestApp(service httpService) *fiber.App {
	app := fiber.New(fiber.Config{ErrorHandler: problem.Handler})
	Register(app, Services{Punishments: service})
	return app
}

type httpService struct {
	denied bool
	err    error
}

func (service httpService) CreateDefinition(context.Context, domain.Definition) (domain.Definition, error) {
	if service.err != nil {
		return domain.Definition{}, service.err
	}
	return domain.Definition{ID: uuid.New(), Version: 1}, nil
}

func (service httpService) UpdateDefinition(context.Context, domain.Definition, uint64) (domain.Definition, error) {
	if service.err != nil {
		return domain.Definition{}, service.err
	}
	return domain.Definition{ID: uuid.New(), Version: 2}, nil
}

func (service httpService) DeleteDefinition(context.Context, uuid.UUID, uint64) error {
	return service.err
}

func (service httpService) GetDefinition(context.Context, uuid.UUID) (domain.Definition, error) {
	if service.err != nil {
		return domain.Definition{}, service.err
	}
	return domain.Definition{ID: uuid.New(), Version: 1}, nil
}

func (service httpService) ListDefinitions(
	context.Context,
	port.DefinitionFilter,
	pagination.Page,
) (pagination.Result[domain.Definition], error) {
	if service.err != nil {
		return pagination.Result[domain.Definition]{}, service.err
	}
	return pagination.Result[domain.Definition]{}, nil
}

func (service httpService) ReorderDefinitionActions(context.Context, uuid.UUID, []uuid.UUID) error {
	return service.err
}

func (service httpService) IssuePunishment(context.Context, port.IssueCommand) (domain.Punishment, error) {
	if service.err != nil {
		return domain.Punishment{}, service.err
	}
	return domain.Punishment{ID: uuid.New(), Version: 1}, nil
}

func (service httpService) UpdatePunishment(context.Context, port.UpdateCommand) (domain.Punishment, error) {
	if service.err != nil {
		return domain.Punishment{}, service.err
	}
	return domain.Punishment{ID: uuid.New(), Version: 2}, nil
}

func (service httpService) RevokePunishment(context.Context, port.RevokeCommand) error {
	return service.err
}

func (service httpService) GetPunishment(context.Context, uuid.UUID) (domain.Punishment, error) {
	if service.err != nil {
		return domain.Punishment{}, service.err
	}
	return domain.Punishment{ID: uuid.New(), Version: 1}, nil
}

func (service httpService) ListPunishments(
	context.Context,
	port.PunishmentFilter,
	pagination.Page,
) (pagination.Result[domain.Punishment], error) {
	if service.err != nil {
		return pagination.Result[domain.Punishment]{}, service.err
	}
	return pagination.Result[domain.Punishment]{}, nil
}

func (service httpService) CheckRestriction(context.Context, port.CheckCommand) (domain.CheckResult, error) {
	if service.err != nil {
		return domain.CheckResult{}, service.err
	}
	return domain.CheckResult{Allowed: !service.denied}, nil
}

func (service httpService) ListActiveRestrictions(context.Context, uuid.UUID) ([]domain.ActiveRestriction, error) {
	if service.err != nil {
		return nil, service.err
	}
	return nil, nil
}

func (service httpService) ExpirePunishments(context.Context) (int64, error) {
	return 0, nil
}

func (service httpService) VerifyRestrictions(context.Context) (domain.DriftReport, error) {
	if service.err != nil {
		return domain.DriftReport{}, service.err
	}
	return domain.DriftReport{}, nil
}

func (service httpService) RebuildRestrictions(context.Context) (domain.DriftReport, error) {
	if service.err != nil {
		return domain.DriftReport{}, service.err
	}
	return domain.DriftReport{Repaired: true}, nil
}

func (service httpService) ClearRestrictionCache(context.Context) error { return nil }
