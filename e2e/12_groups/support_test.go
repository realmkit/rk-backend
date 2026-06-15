// Package groups_e2e verifies group and permission journeys through the real server.
package groups_e2e

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/e2e/harness"
	groupshttp "github.com/realmkit/rk-backend/module/groups/adapter/http"
	groupspostgres "github.com/realmkit/rk-backend/module/groups/adapter/postgres"
	groupsapplication "github.com/realmkit/rk-backend/module/groups/application"
	groupsdomain "github.com/realmkit/rk-backend/module/groups/domain"
	groupsport "github.com/realmkit/rk-backend/module/groups/port"
	"github.com/realmkit/rk-backend/pkg/api/auth"
	"github.com/realmkit/rk-backend/pkg/api/headers"
	"github.com/realmkit/rk-backend/pkg/api/openapi"
	eventtesting "github.com/realmkit/rk-backend/pkg/events/testing"
	"github.com/realmkit/rk-backend/pkg/server"
)

// groupsFixture contains groups e2e wiring.
type groupsFixture struct {
	ecosystem *harness.Ecosystem
	service   groupsapplication.Service
	policies  groupspostgres.PermissionRepository
	events    *eventtesting.PublisherRecorder
}

// newGroupsFixture starts a server with groups routes.
func newGroupsFixture(t *testing.T) groupsFixture {
	t.Helper()
	database := harness.NewSQLiteDatabase(t)
	events := &eventtesting.PublisherRecorder{}
	policies := groupspostgres.NewPermissionRepository(database.Store)
	service := groupsapplication.NewService(
		groupspostgres.NewGroupRepository(database.Store),
		groupspostgres.NewMembershipRepository(database.Store),
		policies,
	).WithEvents(events)
	ecosystem := harness.New(
		t,
		harness.WithDevelopment(true),
		harness.WithDatabase(database),
		harness.WithServerOptions(
			server.WithAuth(auth.Config{DevelopmentBypass: true}, harness.DevProvisioner{}),
			server.WithGroups(
				groupshttp.Services{
					Groups:      service,
					Memberships: service,
					Grants:      service,
					Checker:     service,
				},
			),
		),
	)
	return groupsFixture{
		ecosystem: ecosystem,
		service:   service,
		policies:  policies,
		events:    events,
	}
}

// do sends a request through the fixture server.
func (fixture groupsFixture) do(t *testing.T, request *http.Request) *http.Response {
	t.Helper()
	return fixture.ecosystem.Test(t, request)
}

// createGroup creates one group through HTTP.
func (fixture groupsFixture) createGroup(t *testing.T, key string) map[string]any {
	t.Helper()
	body := `{"key":"` + key + `","name":"` + key + ` team","description":"E2E group","color":"#3366ff","weight":50,"status":"active"}`
	response := fixture.do(
		t,
		configureRequest(
			harness.JSONRequest(fiber.MethodPost, "/groups", body),
			withGroupsIdempotency("create-"+key),
		),
	)
	assertGroupsStatus(t, response, fiber.StatusCreated)
	return decodeGroupsObject(t, response)
}

// createGrant creates a permission grant through the application seam.
func (fixture groupsFixture) createGrant(
	t *testing.T,
	groupID uuid.UUID,
	grant groupsdomain.PermissionGrant,
) groupsdomain.PermissionGrant {
	t.Helper()
	created, err := fixture.service.CreatePermissionGrant(
		context.Background(),
		groupsport.CreatePermissionGrantCommand{GroupID: groupID, Grant: grant},
	)
	if err != nil {
		t.Fatalf("CreatePermissionGrant() error = %v", err)
	}
	return created
}

// checkPermission sends a permission check request.
func (fixture groupsFixture) checkPermission(t *testing.T, body string) map[string]any {
	t.Helper()
	response := fixture.do(t, harness.JSONRequest(fiber.MethodPost, "/permissions/check", body))
	assertGroupsStatus(t, response, fiber.StatusOK)
	return decodeGroupsObject(t, response)
}

// withGroupsIdempotency adds an idempotency key.
func withGroupsIdempotency(key string) func(*http.Request) {
	return func(request *http.Request) {
		request.Header.Set(headers.IdempotencyKey, key)
	}
}

// withGroupsIfMatch adds an If-Match header.
func withGroupsIfMatch(version uint64) func(*http.Request) {
	return func(request *http.Request) {
		request.Header.Set(headers.IfMatch, `"`+strconv.FormatUint(version, 10)+`"`)
	}
}

// withCurrentGroupUser adds the temporary current-user header.
func withCurrentGroupUser(userID uuid.UUID) func(*http.Request) {
	return func(request *http.Request) {
		request.Header.Set(auth.DevUserIDHeader, userID.String())
	}
}

// configureRequest applies request mutations.
func configureRequest(request *http.Request, configs ...func(*http.Request)) *http.Request {
	for _, config := range configs {
		config(request)
	}
	return request
}

// decodeGroupsObject decodes one JSON object.
func decodeGroupsObject(t *testing.T, response *http.Response) map[string]any {
	t.Helper()
	var payload map[string]any
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		t.Fatalf("Decode() error = %v", err)
	}
	return payload
}

// assertGroupsStatus verifies response status.
func assertGroupsStatus(t *testing.T, response *http.Response, want int) {
	t.Helper()
	if response.StatusCode != want {
		t.Fatalf("StatusCode = %d, want %d body = %q", response.StatusCode, want, harness.ResponseBody(t, response))
	}
}

// groupIDFrom extracts group ID.
func groupIDFrom(t *testing.T, payload map[string]any) uuid.UUID {
	t.Helper()
	id, err := uuid.Parse(payload["id"].(string))
	if err != nil {
		t.Fatalf("Parse(id) error = %v", err)
	}
	return id
}

// groupVersionFrom extracts group version.
func groupVersionFrom(t *testing.T, payload map[string]any) uint64 {
	t.Helper()
	return uint64(payload["version"].(float64))
}

// assertGroupsOpenAPIRoute verifies an OpenAPI operation exists.
func assertGroupsOpenAPIRoute(t *testing.T, method string, path string) {
	t.Helper()
	ok, err := openapi.OperationExists(method, path)
	if err != nil {
		t.Fatalf("OperationExists() error = %v", err)
	}
	if !ok {
		t.Fatalf("%s %s missing OpenAPI operation", method, path)
	}
}
