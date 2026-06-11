package users_e2e

import (
	"context"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/niflaot/gamehub-go/e2e/harness"
	groupsdomain "github.com/niflaot/gamehub-go/module/groups/domain"
	groupsport "github.com/niflaot/gamehub-go/module/groups/port"
	"github.com/niflaot/gamehub-go/module/user/domain"
)

// TestUsersMeReturnsProvisionedCurrentUser verifies current-user reads.
func TestUsersMeReturnsProvisionedCurrentUser(t *testing.T) {
	steps := harness.NewSteps(t)
	fixture := newUsersFixture(t, false)

	steps.Log("provision local user through the real user service")
	userID := fixture.provisionIdentity(t, "current-user")

	steps.Log("fetch current user through development auth middleware")
	response := fixture.do(t, fixture.authedJSON(t, userID, fiber.MethodGet, "/users/me", ""))
	assertUserStatus(t, response, fiber.StatusOK)
	payload := decodeUserObject(t, response)
	user := payload["user"].(map[string]any)
	if user["id"] != userID.String() || user["status"] != string(domain.StatusActive) {
		t.Fatalf("user = %+v, want provisioned active user", user)
	}
	if payload["provider_claims"] == nil {
		t.Fatalf("provider_claims missing from current user response")
	}
}

// TestUsersProvisioningIsStable verifies repeated identity provisioning is deterministic.
func TestUsersProvisioningIsStable(t *testing.T) {
	steps := harness.NewSteps(t)
	fixture := newUsersFixture(t, false)

	steps.Log("provision the same external identity twice")
	firstID := fixture.provisionIdentity(t, "stable-user")
	secondID := fixture.provisionIdentity(t, "stable-user")
	if firstID != secondID {
		t.Fatalf("second user id = %s, want %s", secondID, firstID)
	}
}

// TestUsersMeIncludesGroupSummaryWhenWired verifies safe group summary integration.
func TestUsersMeIncludesGroupSummaryWhenWired(t *testing.T) {
	steps := harness.NewSteps(t)
	fixture := newUsersFixture(t, true)
	userID := fixture.provisionIdentity(t, "group-summary")

	steps.Log("create a display group and membership through the groups service")
	group, err := fixture.groups.Create(
		context.Background(),
		groupsport.CreateGroupCommand{
			Group: groupsdomain.Group{
				ID:      uuid.New(),
				Key:     "vip",
				Name:    "VIP",
				Color:   "#3366ff",
				Weight:  100,
				Status:  groupsdomain.GroupStatusActive,
				Version: 1,
			},
		},
	)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	_, err = fixture.groups.Assign(
		context.Background(),
		groupsport.AssignMembershipCommand{
			Membership: groupsdomain.Membership{
				ID:      uuid.New(),
				GroupID: group.ID,
				UserID:  userID,
				Status:  groupsdomain.MembershipStatusActive,
				Version: 1,
			},
		},
	)
	if err != nil {
		t.Fatalf("Assign() error = %v", err)
	}

	steps.Log("fetch current user with group summary")
	response := fixture.do(t, fixture.authedJSON(t, userID, fiber.MethodGet, "/users/me", ""))
	assertUserStatus(t, response, fiber.StatusOK)
	payload := decodeUserObject(t, response)
	groups := payload["groups"].(map[string]any)
	if len(groups["groups"].([]any)) != 1 || groups["display_group"] == nil {
		t.Fatalf("groups = %+v, want one group and display group", groups)
	}
}

// TestUsersMeRequiresAuthentication verifies anonymous access is denied.
func TestUsersMeRequiresAuthentication(t *testing.T) {
	steps := harness.NewSteps(t)
	fixture := newUsersFixture(t, false)

	steps.Log("request current user without auth")
	response := fixture.do(t, harness.JSONRequest(fiber.MethodGet, "/users/me", ""))
	assertUserStatus(t, response, fiber.StatusUnauthorized)
}
