package users_e2e

import (
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/niflaot/gamehub-go/e2e/harness"
)

// TestUsersProfileUpdateRequiresIdempotencyAndIfMatch verifies update headers.
func TestUsersProfileUpdateRequiresIdempotencyAndIfMatch(t *testing.T) {
	steps := harness.NewSteps(t)
	fixture := newUsersFixture(t, false)
	userID := fixture.provisionIdentity(t, "profile-update")

	steps.Log("fetch current user version")
	current := fixture.do(t, fixture.authedJSON(t, userID, fiber.MethodGet, "/users/me", ""))
	assertUserStatus(t, current, fiber.StatusOK)
	version := userVersionFrom(t, decodeUserObject(t, current))

	steps.Log("reject update without idempotency key")
	missingKey := fixture.do(t, fixture.authedJSON(t, userID, fiber.MethodPatch, "/users/me", `{}`))
	assertUserStatus(t, missingKey, fiber.StatusBadRequest)

	steps.Log("reject update without If-Match")
	missingVersion := fixture.do(
		t,
		configureRequest(
			fixture.authedJSON(t, userID, fiber.MethodPatch, "/users/me", `{}`),
			withUserIdempotency("missing-version"),
		),
	)
	assertUserStatus(t, missingVersion, fiber.StatusPreconditionRequired)

	steps.Log("update avatar asset id")
	avatarID := uuid.New()
	update := fixture.do(
		t,
		configureRequest(
			fixture.authedJSON(t, userID, fiber.MethodPatch, "/users/me", `{"avatar_asset_id":"`+avatarID.String()+`"}`),
			withUserIdempotency("update-avatar"),
			withUserIfMatch(version),
		),
	)
	assertUserStatus(t, update, fiber.StatusOK)
	payload := decodeUserObject(t, update)
	if payload["avatar_asset_id"] != avatarID.String() {
		t.Fatalf("avatar_asset_id = %v, want %s", payload["avatar_asset_id"], avatarID)
	}

	steps.Log("reject stale version update")
	stale := fixture.do(
		t,
		configureRequest(
			fixture.authedJSON(t, userID, fiber.MethodPatch, "/users/me", `{}`),
			withUserIdempotency("stale-update"),
			withUserIfMatch(version),
		),
	)
	assertUserStatus(t, stale, fiber.StatusPreconditionFailed)
}
