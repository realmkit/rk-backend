package users_e2e

import (
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/niflaot/gamehub-go/e2e/harness"
	"github.com/niflaot/gamehub-go/module/user/domain"
)

// TestUsersDisabledUserCannotAuthenticate verifies disabled account enforcement.
func TestUsersDisabledUserCannotAuthenticate(t *testing.T) {
	steps := harness.NewSteps(t)
	fixture := newUsersFixture(t, false)
	user := fixture.seedUser(t, domain.StatusDisabled)

	steps.Log("request current user as disabled development user")
	response := fixture.do(t, fixture.authedJSON(t, user.ID, fiber.MethodGet, "/users/me", ""))
	assertUserStatus(t, response, fiber.StatusForbidden)
}

// TestUsersAccountURLUnavailableReturnsProblem verifies provider URL fallback.
func TestUsersAccountURLUnavailableReturnsProblem(t *testing.T) {
	steps := harness.NewSteps(t)
	fixture := newUsersFixture(t, false)
	userID := fixture.provisionIdentity(t, "account-url")

	steps.Log("request account URL when no provider account URL is configured")
	response := fixture.do(t, fixture.authedJSON(t, userID, fiber.MethodGet, "/users/me/identity/account-url", ""))
	assertUserStatus(t, response, fiber.StatusNotFound)
}
