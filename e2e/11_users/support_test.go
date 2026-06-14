// Package users_e2e verifies user journeys through the real server.
package users_e2e

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/e2e/harness"
	groupspostgres "github.com/realmkit/rk-backend/module/groups/adapter/postgres"
	groupsapplication "github.com/realmkit/rk-backend/module/groups/application"
	userhttp "github.com/realmkit/rk-backend/module/user/adapter/http"
	userpostgres "github.com/realmkit/rk-backend/module/user/adapter/postgres"
	userapplication "github.com/realmkit/rk-backend/module/user/application"
	userdomain "github.com/realmkit/rk-backend/module/user/domain"
	userport "github.com/realmkit/rk-backend/module/user/port"
	"github.com/realmkit/rk-backend/pkg/api/auth"
	"github.com/realmkit/rk-backend/pkg/api/headers"
	"github.com/realmkit/rk-backend/pkg/api/openapi"
	eventtesting "github.com/realmkit/rk-backend/pkg/events/testing"
	"github.com/realmkit/rk-backend/pkg/identity"
	"github.com/realmkit/rk-backend/pkg/server"
	"github.com/realmkit/rk-backend/pkg/transaction"
)

// usersFixture contains user e2e wiring.
type usersFixture struct {
	ecosystem *harness.Ecosystem
	service   userapplication.Service
	groups    groupsapplication.Service
	users     userpostgres.UserRepository
	events    *eventtesting.PublisherRecorder
}

// newUsersFixture starts a server with user routes.
func newUsersFixture(t *testing.T, includeGroups bool) usersFixture {
	t.Helper()
	database := harness.NewSQLiteDatabase(t)
	events := &eventtesting.PublisherRecorder{}
	users := userpostgres.NewUserRepository(database.Store)
	service := userapplication.NewService(userapplication.Dependencies{
		Users:        users,
		Links:        userpostgres.NewIdentityLinkRepository(database.Store),
		Claims:       userpostgres.NewClaimCacheRepository(database.Store),
		Transactions: transaction.New(database.DB),
		Provider:     "e2e_oidc",
		Events:       events,
	})
	services := userhttp.Services{Users: service}
	var groups groupsapplication.Service
	if includeGroups {
		groups = groupsapplication.NewService(
			groupspostgres.NewGroupRepository(database.Store),
			groupspostgres.NewMembershipRepository(database.Store),
			groupspostgres.NewPermissionRepository(database.Store),
		)
		services.Groups = groups
	}
	ecosystem := harness.New(
		t,
		harness.WithDevelopment(true),
		harness.WithDatabase(database),
		harness.WithServerOptions(
			server.WithAuth(auth.Config{DevelopmentBypass: true}, service),
			server.WithUsers(services),
		),
	)
	return usersFixture{
		ecosystem: ecosystem,
		service:   service,
		groups:    groups,
		users:     users,
		events:    events,
	}
}

// seedUser creates a local user row.
func (fixture usersFixture) seedUser(t *testing.T, status userdomain.Status) userdomain.User {
	t.Helper()
	now := time.Now().UTC()
	user, err := fixture.users.Create(
		context.Background(),
		userdomain.User{
			ID:          uuid.New(),
			Status:      status,
			FirstSeenAt: now,
			LastSeenAt:  &now,
			Version:     1,
		},
	)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	return user
}

// provisionIdentity creates or resolves a user through the real provisioner.
func (fixture usersFixture) provisionIdentity(t *testing.T, subject string) uuid.UUID {
	t.Helper()
	claims := map[string]any{
		"iss":                "https://auth.e2e",
		"sub":                subject,
		"preferred_username": subject,
		"email":              subject + "@example.test",
		"email_verified":     true,
		"name":               "E2E " + subject,
	}
	external := identity.ExternalIdentity{
		Issuer:          "https://auth.e2e",
		Subject:         subject,
		Username:        subject,
		Email:           subject + "@example.test",
		EmailVerified:   true,
		DisplayName:     "E2E " + subject,
		RawClaimsHash:   identity.ClaimsHash(claims),
		PreferredLocale: "en",
	}
	current, err := fixture.service.Provision(context.Background(), external, auth.Token{})
	if err != nil {
		t.Fatalf("Provision() error = %v", err)
	}
	return current.UserID
}

// authedJSON sends an authenticated JSON request.
func (fixture usersFixture) authedJSON(t *testing.T, userID uuid.UUID, method string, path string, body string) *http.Request {
	t.Helper()
	request := harness.JSONRequest(method, path, body)
	request.Header.Set(auth.DevUserIDHeader, userID.String())
	return request
}

// do sends a request through the fixture server.
func (fixture usersFixture) do(t *testing.T, request *http.Request) *http.Response {
	t.Helper()
	return fixture.ecosystem.Test(t, request)
}

// withUserIdempotency adds an idempotency key.
func withUserIdempotency(key string) func(*http.Request) {
	return func(request *http.Request) {
		request.Header.Set(headers.IdempotencyKey, key)
	}
}

// withUserIfMatch adds an If-Match header.
func withUserIfMatch(version uint64) func(*http.Request) {
	return func(request *http.Request) {
		request.Header.Set(headers.IfMatch, `"`+strconv.FormatUint(version, 10)+`"`)
	}
}

// configureRequest applies request mutations.
func configureRequest(request *http.Request, configs ...func(*http.Request)) *http.Request {
	for _, config := range configs {
		config(request)
	}
	return request
}

// decodeUserObject decodes one JSON object.
func decodeUserObject(t *testing.T, response *http.Response) map[string]any {
	t.Helper()
	var payload map[string]any
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		t.Fatalf("Decode() error = %v", err)
	}
	return payload
}

// assertUserStatus verifies response status.
func assertUserStatus(t *testing.T, response *http.Response, want int) {
	t.Helper()
	if response.StatusCode != want {
		t.Fatalf("StatusCode = %d, want %d body = %q", response.StatusCode, want, harness.ResponseBody(t, response))
	}
}

// userVersionFrom extracts nested or root version.
func userVersionFrom(t *testing.T, payload map[string]any) uint64 {
	t.Helper()
	if user, ok := payload["user"].(map[string]any); ok {
		return uint64(user["version"].(float64))
	}
	return uint64(payload["version"].(float64))
}

// assertUserOpenAPIRoute verifies an OpenAPI operation exists.
func assertUserOpenAPIRoute(t *testing.T, method string, path string) {
	t.Helper()
	ok, err := openapi.OperationExists(method, path)
	if err != nil {
		t.Fatalf("OperationExists() error = %v", err)
	}
	if !ok {
		t.Fatalf("%s %s missing OpenAPI operation", method, path)
	}
}

// ensureUserServiceShape keeps compile-time contracts visible.
var _ userport.Service = userapplication.Service{}
