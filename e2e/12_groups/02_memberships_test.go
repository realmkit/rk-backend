package groups_e2e

import (
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/niflaot/gamehub-go/e2e/harness"
)

// TestGroupsMembershipLifecycle verifies membership assign, list, user groups, and remove.
func TestGroupsMembershipLifecycle(t *testing.T) {
	steps := harness.NewSteps(t)
	fixture := newGroupsFixture(t)
	group := fixture.createGroup(t, "moderators")
	groupID := groupIDFrom(t, group)
	userID := uuid.New()

	steps.Log("reject membership assignment without idempotency")
	missingKey := fixture.do(
		t,
		harness.JSONRequest(
			fiber.MethodPut,
			"/groups/"+groupID.String()+"/members/"+userID.String(),
			`{"status":"active","assigned_reason":"missing-key"}`,
		),
	)
	assertGroupsStatus(t, missingKey, fiber.StatusBadRequest)

	steps.Log("assign active membership")
	assigned := fixture.do(
		t,
		configureRequest(
			harness.JSONRequest(
				fiber.MethodPut,
				"/groups/"+groupID.String()+"/members/"+userID.String(),
				`{"status":"active","assigned_reason":"e2e"}`,
			),
			withGroupsIdempotency("assign-moderator"),
		),
	)
	assertGroupsStatus(t, assigned, fiber.StatusOK)
	membership := decodeGroupsObject(t, assigned)

	steps.Log("list group members")
	members := fixture.do(t, harness.JSONRequest(fiber.MethodGet, "/groups/"+groupID.String()+"/members?page_size=10", ""))
	assertGroupsStatus(t, members, fiber.StatusOK)
	memberPayload := decodeGroupsObject(t, members)
	if len(memberPayload["items"].([]any)) != 1 {
		t.Fatalf("items = %v, want one membership", memberPayload["items"])
	}

	steps.Log("list user groups by explicit user route")
	userGroups := fixture.do(t, harness.JSONRequest(fiber.MethodGet, "/users/"+userID.String()+"/groups", ""))
	assertGroupsStatus(t, userGroups, fiber.StatusOK)
	assertUserGroupsCount(t, decodeGroupsObject(t, userGroups), 1)

	steps.Log("reject current-user groups without debug header")
	anonymousCurrent := fixture.do(t, harness.JSONRequest(fiber.MethodGet, "/users/me/groups", ""))
	assertGroupsStatus(t, anonymousCurrent, fiber.StatusUnauthorized)

	steps.Log("list current-user groups with debug header")
	currentGroups := fixture.do(
		t,
		configureRequest(
			harness.JSONRequest(fiber.MethodGet, "/users/me/groups", ""),
			withCurrentGroupUser(userID),
		),
	)
	assertGroupsStatus(t, currentGroups, fiber.StatusOK)
	assertUserGroupsCount(t, decodeGroupsObject(t, currentGroups), 1)

	steps.Log("reject membership removal without If-Match")
	missingVersion := fixture.do(
		t,
		configureRequest(
			harness.JSONRequest(fiber.MethodDelete, "/groups/"+groupID.String()+"/members/"+userID.String(), ""),
			withGroupsIdempotency("missing-member-version"),
		),
	)
	assertGroupsStatus(t, missingVersion, fiber.StatusPreconditionRequired)

	steps.Log("remove membership")
	removed := fixture.do(
		t,
		configureRequest(
			harness.JSONRequest(fiber.MethodDelete, "/groups/"+groupID.String()+"/members/"+userID.String(), ""),
			withGroupsIdempotency("remove-moderator"),
			withGroupsIfMatch(uint64(membership["version"].(float64))),
		),
	)
	assertGroupsStatus(t, removed, fiber.StatusNoContent)

	steps.Log("verify user groups are empty after removal")
	afterRemove := fixture.do(t, harness.JSONRequest(fiber.MethodGet, "/users/"+userID.String()+"/groups", ""))
	assertGroupsStatus(t, afterRemove, fiber.StatusOK)
	assertUserGroupsCount(t, decodeGroupsObject(t, afterRemove), 0)
}

// assertUserGroupsCount verifies user group count.
func assertUserGroupsCount(t *testing.T, payload map[string]any, want int) {
	t.Helper()
	groups := payload["groups"].([]any)
	if len(groups) != want {
		t.Fatalf("groups = %v, want %d groups", groups, want)
	}
}
