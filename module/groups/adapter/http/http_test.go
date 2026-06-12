package http

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/module/groups/domain"
	"github.com/realmkit/rk-backend/module/groups/port"
	"github.com/realmkit/rk-backend/pkg/api/headers"
	"github.com/realmkit/rk-backend/pkg/api/problem"
	"github.com/realmkit/rk-backend/pkg/pagination"
)

// TestCreateGroupRequiresIdempotency verifies mutating routes require idempotency.
func TestCreateGroupRequiresIdempotency(t *testing.T) {
	app := testApp(&httpService{})
	req := testRequest(http.MethodPost, "/groups", `{"key":"admin"}`)

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Test() error = %v", err)
	}
	if resp.StatusCode != fiber.StatusBadRequest {
		t.Fatalf("StatusCode = %d, want %d", resp.StatusCode, fiber.StatusBadRequest)
	}
}

// TestGroupRoutesExerciseLifecycle verifies group route success paths.
func TestGroupRoutesExerciseLifecycle(t *testing.T) {
	group := testHTTPGroup()
	app := testApp(&httpService{group: group})
	cases := []struct {
		method string
		path   string
		body   string
		status int
		header map[string]string
	}{
		{
			method: http.MethodPost,
			path:   "/groups",
			body:   `{"key":"admin","name":"Admin","color":"#ff0000","weight":100,"status":"active"}`,
			status: fiber.StatusCreated,
			header: map[string]string{headers.IdempotencyKey: "create"},
		},
		{method: http.MethodGet, path: "/groups", status: fiber.StatusOK},
		{method: http.MethodGet, path: "/groups/" + group.ID.String(), status: fiber.StatusOK},
		{
			method: http.MethodPatch,
			path:   "/groups/" + group.ID.String(),
			body:   `{"name":"Admin","color":"#ff0000","weight":100,"status":"active"}`,
			status: fiber.StatusOK,
			header: map[string]string{headers.IdempotencyKey: "update", headers.IfMatch: `"1"`},
		},
		{
			method: http.MethodDelete,
			path:   "/groups/" + group.ID.String(),
			status: fiber.StatusNoContent,
			header: map[string]string{headers.IdempotencyKey: "delete", headers.IfMatch: `"1"`},
		},
	}
	for _, tt := range cases {
		req := testRequest(tt.method, tt.path, tt.body)
		for key, value := range tt.header {
			req.Header.Set(key, value)
		}
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("%s %s Test() error = %v", tt.method, tt.path, err)
		}
		if resp.StatusCode != tt.status {
			t.Fatalf("%s %s status = %d, want %d", tt.method, tt.path, resp.StatusCode, tt.status)
		}
	}
}

// TestMembershipAndPermissionRoutes verifies membership and permission route success paths.
func TestMembershipAndPermissionRoutes(t *testing.T) {
	group := testHTTPGroup()
	userID := uuid.New()
	service := &httpService{
		group:      group,
		membership: domain.Membership{ID: uuid.New(), GroupID: group.ID, UserID: userID, Status: domain.MembershipStatusActive, Version: 1},
		decision:   port.Decision{Allowed: true, Reason: "matched_relation"},
	}
	app := testApp(service)
	cases := []struct {
		method string
		path   string
		body   string
		status int
		header map[string]string
	}{
		{method: http.MethodGet, path: "/groups/" + group.ID.String() + "/members", status: fiber.StatusOK},
		{
			method: http.MethodPut,
			path:   "/groups/" + group.ID.String() + "/members/" + userID.String(),
			body:   `{"status":"active"}`,
			status: fiber.StatusOK,
			header: map[string]string{headers.IdempotencyKey: "assign"},
		},
		{
			method: http.MethodDelete,
			path:   "/groups/" + group.ID.String() + "/members/" + userID.String(),
			status: fiber.StatusNoContent,
			header: map[string]string{headers.IdempotencyKey: "remove", headers.IfMatch: `"1"`},
		},
		{method: http.MethodGet, path: "/users/" + userID.String() + "/groups", status: fiber.StatusOK},
		{
			method: http.MethodGet,
			path:   "/users/me/groups",
			status: fiber.StatusOK,
			header: map[string]string{currentUserIDHeader: userID.String()},
		},
		{
			method: http.MethodPost,
			path:   "/permissions/check",
			body:   `{"actor_user_id":"` + userID.String() + `","permission":"groups.read","object_type":"group","object_id":"` + group.ID.String() + `"}`,
			status: fiber.StatusOK,
		},
	}
	for _, tt := range cases {
		req := testRequest(tt.method, tt.path, tt.body)
		for key, value := range tt.header {
			req.Header.Set(key, value)
		}
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("%s %s Test() error = %v", tt.method, tt.path, err)
		}
		if resp.StatusCode != tt.status {
			body, _ := io.ReadAll(resp.Body)
			t.Fatalf("%s %s status = %d body=%s, want %d", tt.method, tt.path, resp.StatusCode, body, tt.status)
		}
	}
}

// TestGetGroupMapsNotFound verifies problem mapping.
func TestGetGroupMapsNotFound(t *testing.T) {
	app := testApp(&httpService{err: port.ErrNotFound})
	resp, err := app.Test(testRequest(http.MethodGet, "/groups/"+uuid.NewString(), ""))
	if err != nil {
		t.Fatalf("Test() error = %v", err)
	}
	if resp.StatusCode != fiber.StatusNotFound {
		t.Fatalf("StatusCode = %d, want %d", resp.StatusCode, fiber.StatusNotFound)
	}
}

// testApp creates a groups HTTP app.
func testApp(service *httpService) *fiber.App {
	app := fiber.New(fiber.Config{ErrorHandler: problem.Handler})
	useTestPrincipal(app)
	Register(app, Services{Groups: service, Memberships: service, Tuples: service, Checker: service})
	return app
}

// testRequest creates a JSON request.
func testRequest(method string, target string, body string) *http.Request {
	var reader io.Reader
	if body != "" {
		reader = strings.NewReader(body)
	}
	req, _ := http.NewRequest(method, target, reader)
	req.Header.Set(headers.Accept, "application/json")
	if body != "" {
		req.Header.Set(headers.ContentType, "application/json")
	}
	return req
}

// testHTTPGroup returns an HTTP test group.
func testHTTPGroup() domain.Group {
	return domain.Group{
		ID:      uuid.MustParse("11111111-1111-1111-1111-111111111111"),
		Key:     "admin",
		Name:    "Admin",
		Color:   "#ff0000",
		Weight:  100,
		Status:  domain.GroupStatusActive,
		Version: 1,
	}
}

// httpService is a fake groups service.
type httpService struct {
	group      domain.Group
	membership domain.Membership
	decision   port.Decision
	err        error
}

// Create creates a group.
func (service *httpService) Create(context.Context, port.CreateGroupCommand) (domain.Group, error) {
	return service.group, service.err
}

// Update updates a group.
func (service *httpService) Update(context.Context, port.UpdateGroupCommand) (domain.Group, error) {
	return service.group, service.err
}

// Get returns one group.
func (service *httpService) Get(context.Context, uuid.UUID) (domain.Group, error) {
	return service.group, service.err
}

// List lists groups.
func (service *httpService) List(context.Context, port.GroupFilter, pagination.Page) (pagination.Result[domain.Group], error) {
	return pagination.Result[domain.Group]{Items: []domain.Group{service.group}}, service.err
}

// Delete deletes a group.
func (service *httpService) Delete(context.Context, port.DeleteGroupCommand) error {
	return service.err
}

// Assign assigns a user to a group.
func (service *httpService) Assign(context.Context, port.AssignMembershipCommand) (domain.Membership, error) {
	return service.membership, service.err
}

// Remove removes a membership.
func (service *httpService) Remove(context.Context, port.RemoveMembershipCommand) error {
	return service.err
}

// ListGroupMembers lists memberships for a group.
func (service *httpService) ListGroupMembers(context.Context, uuid.UUID, pagination.Page) (pagination.Result[domain.Membership], error) {
	return pagination.Result[domain.Membership]{Items: []domain.Membership{service.membership}}, service.err
}

// ListUserGroups returns active groups for user.
func (service *httpService) ListUserGroups(context.Context, uuid.UUID) (port.UserGroups, error) {
	return port.UserGroups{Groups: []domain.Group{service.group}, DisplayGroup: &service.group}, service.err
}

// CreateTuple creates a tuple.
func (service *httpService) CreateTuple(context.Context, port.CreateTupleCommand) (domain.RelationTuple, error) {
	return domain.RelationTuple{}, service.err
}

// DeleteTuple deletes a tuple.
func (service *httpService) DeleteTuple(context.Context, port.DeleteTupleCommand) error {
	return service.err
}

// Check returns an authorization decision.
func (service *httpService) Check(context.Context, port.CheckRequest) (port.Decision, error) {
	return service.decision, service.err
}
