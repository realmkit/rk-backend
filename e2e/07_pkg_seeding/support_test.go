// Package seeding_e2e verifies global data seeding through real services.
package seeding_e2e

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/e2e/harness"
	forumshttp "github.com/realmkit/rk-backend/module/forums/adapter/http"
	forumspostgres "github.com/realmkit/rk-backend/module/forums/adapter/postgres"
	forumsapp "github.com/realmkit/rk-backend/module/forums/application"
	groupshttp "github.com/realmkit/rk-backend/module/groups/adapter/http"
	groupspostgres "github.com/realmkit/rk-backend/module/groups/adapter/postgres"
	groupsapp "github.com/realmkit/rk-backend/module/groups/application"
	"github.com/realmkit/rk-backend/pkg/api/auth"
	"github.com/realmkit/rk-backend/pkg/postgres/seeding"
	"github.com/realmkit/rk-backend/pkg/server"
	"github.com/realmkit/rk-backend/pkg/transaction"
)

// seedingFixture contains services used by seeding e2e tests.
type seedingFixture struct {
	ecosystem *harness.Ecosystem
	runner    seeding.Runner
	database  *harness.Database
}

// newSeedingFixture starts a server backed by a seeded-capable database.
func newSeedingFixture(t *testing.T) seedingFixture {
	t.Helper()
	database := harness.NewSQLiteDatabase(t)
	groups := groupsapp.NewService(
		groupspostgres.NewGroupRepository(database.Store),
		groupspostgres.NewMembershipRepository(database.Store),
		groupspostgres.NewTupleRepository(database.Store),
		groupspostgres.NewPermissionRepository(database.Store),
	)
	forums := forumsapp.NewService(forumsapp.Dependencies{
		Categories:   forumspostgres.NewCategoryRepository(database.Store),
		Forums:       forumspostgres.NewForumRepository(database.Store),
		Threads:      forumspostgres.NewThreadRepository(database.Store),
		Posts:        forumspostgres.NewPostRepository(database.Store),
		Interactions: forumspostgres.NewInteractionRepository(database.Store),
		Operations:   forumspostgres.NewOperationsRepository(database.Store),
		Authorizer:   forumspostgres.NewVisibilityAuthorizer(database.Store),
		Transactions: transaction.New(database.DB),
	})
	ecosystem := harness.New(
		t,
		harness.WithDevelopment(true),
		harness.WithDatabase(database),
		harness.WithServerOptions(
			server.WithAuth(auth.Config{DevelopmentBypass: true}, harness.DevProvisioner{}),
			server.WithGroups(groupshttp.Services{
				Groups:      groups,
				Memberships: groups,
				Tuples:      groups,
				Checker:     groups,
			}),
			server.WithForums(forumshttp.Services{
				Structure:   forums,
				Content:     forums,
				Interaction: forums,
				Operations:  forums,
				Admin:       forums,
			}),
		),
	)
	return seedingFixture{
		ecosystem: ecosystem,
		runner:    seeding.NewRunner(database.DB, seeding.DefaultSource()),
		database:  database,
	}
}

// do sends a request through the fixture server.
func (fixture seedingFixture) do(t *testing.T, request *http.Request) *http.Response {
	t.Helper()
	return fixture.ecosystem.Test(t, request)
}

// insertUser creates a minimal local user.
func (fixture seedingFixture) insertUser(t *testing.T, userID uuid.UUID) {
	t.Helper()
	insert := `
INSERT INTO users(id, status, avatar_asset_id, first_seen_at, last_seen_at, version, created_at, updated_at, deleted_at)
VALUES(?, 'active', NULL, CURRENT_TIMESTAMP, NULL, 1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, NULL)`
	if err := fixture.database.DB.WithContext(context.Background()).Exec(insert, userID).Error; err != nil {
		t.Fatalf("insert user error = %v", err)
	}
}

// decodeObject decodes one JSON object.
func decodeObject(t *testing.T, response *http.Response) map[string]any {
	t.Helper()
	var payload map[string]any
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		t.Fatalf("Decode() error = %v", err)
	}
	return payload
}

// assertStatus verifies response status.
func assertStatus(t *testing.T, response *http.Response, want int) {
	t.Helper()
	if response.StatusCode != want {
		t.Fatalf("StatusCode = %d, want %d body = %q", response.StatusCode, want, harness.ResponseBody(t, response))
	}
}

// getJSON sends a JSON GET request.
func getJSON(path string) *http.Request {
	return harness.JSONRequest(fiber.MethodGet, path, "")
}

// postJSON sends a JSON POST request.
func postJSON(path string, body string) *http.Request {
	return harness.JSONRequest(fiber.MethodPost, path, body)
}
