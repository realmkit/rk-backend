package users_e2e

import (
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/niflaot/gamehub-go/e2e/harness"
	eventdomain "github.com/niflaot/gamehub-go/pkg/events/domain"
)

// TestUsersEmitProvisionAndProfileEvents verifies user event facts.
func TestUsersEmitProvisionAndProfileEvents(t *testing.T) {
	steps := harness.NewSteps(t)
	fixture := newUsersFixture(t, false)

	steps.Log("provision user and update profile")
	userID := fixture.provisionIdentity(t, "events-user")
	current := fixture.do(t, fixture.authedJSON(t, userID, fiber.MethodGet, "/users/me", ""))
	assertUserStatus(t, current, fiber.StatusOK)
	update := fixture.do(
		t,
		configureRequest(
			fixture.authedJSON(t, userID, fiber.MethodPatch, "/users/me", `{}`),
			withUserIdempotency("events-update"),
			withUserIfMatch(userVersionFrom(t, decodeUserObject(t, current))),
		),
	)
	assertUserStatus(t, update, fiber.StatusOK)

	steps.Log("verify provision and update events")
	drafts := fixture.events.Drafts()
	assertUserEventPresent(t, drafts, eventdomain.EventUsersUserProvisioned)
	assertUserEventPresent(t, drafts, "users.user.updated")
	assertUserEventPresent(t, drafts, "users.identity.linked")
}

// TestUsersOpenAPICoversRoutes verifies user route contract coverage.
func TestUsersOpenAPICoversRoutes(t *testing.T) {
	steps := harness.NewSteps(t)
	routes := []struct {
		method string
		path   string
	}{
		{fiber.MethodGet, "/users/me"},
		{fiber.MethodPatch, "/users/me"},
		{fiber.MethodGet, "/users/me/identity/account-url"},
	}
	for _, route := range routes {
		steps.Log("verify OpenAPI operation %s %s", route.method, route.path)
		assertUserOpenAPIRoute(t, route.method, route.path)
	}
}

// assertUserEventPresent verifies one user event key was published.
func assertUserEventPresent(t *testing.T, drafts []eventdomain.Draft, key eventdomain.EventKey) {
	t.Helper()
	for _, draft := range drafts {
		if draft.Key == key {
			if draft.Producer != eventdomain.ProducerUsers {
				t.Fatalf("producer = %s, want %s", draft.Producer, eventdomain.ProducerUsers)
			}
			if len(draft.Scopes) == 0 {
				t.Fatalf("event %s has no scopes", key)
			}
			return
		}
	}
	t.Fatalf("event %s missing from %+v", key, drafts)
}
