package seeding_e2e

import (
	"context"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/e2e/harness"
	"github.com/realmkit/rk-backend/pkg/postgres/seeding"
)

// TestGlobalSeedsBootstrapCommunity verifies a fresh instance can seed defaults.
func TestGlobalSeedsBootstrapCommunity(t *testing.T) {
	steps := harness.NewSteps(t)
	fixture := newSeedingFixture(t)

	steps.Log("apply global data seeds")
	status, err := fixture.runner.Up(context.Background())
	if err != nil {
		t.Fatalf("seed Up() error = %v", err)
	}
	if len(status.Applied) != 3 || len(status.Pending) != 0 {
		t.Fatalf("status = %+v, want all seeds applied", status)
	}

	steps.Log("read public seeded forum tree")
	tree := fixture.do(t, getJSON("/forums/tree"))
	assertStatus(t, tree, fiber.StatusOK)
	payload := decodeObject(t, tree)
	categories := payload["categories"].([]any)
	if len(categories) != 1 {
		t.Fatalf("categories = %d, want one seeded community category", len(categories))
	}
	category := categories[0].(map[string]any)
	forums := category["forums"].([]any)
	if len(forums) != 3 {
		t.Fatalf("forums = %d, want three seeded forums", len(forums))
	}
}

// TestSeedGrantAdminAllowsSeededPolicy verifies first-operator permissions.
func TestSeedGrantAdminAllowsSeededPolicy(t *testing.T) {
	steps := harness.NewSteps(t)
	fixture := newSeedingFixture(t)
	userID := uuid.MustParse("00000000-0000-0000-0000-000000000901")

	steps.Log("seed defaults and create local user")
	if _, err := fixture.runner.Up(context.Background()); err != nil {
		t.Fatalf("seed Up() error = %v", err)
	}
	fixture.insertUser(t, userID)

	steps.Log("grant administrator group to user")
	grant, err := fixture.runner.GrantAdmin(context.Background(), userID)
	if err != nil {
		t.Fatalf("GrantAdmin() error = %v", err)
	}
	if grant.GroupID != seeding.AdminGroupID {
		t.Fatalf("grant group = %s, want %s", grant.GroupID, seeding.AdminGroupID)
	}

	for _, permission := range []string{
		"groups.update",
		"groups.delete",
		"groups.assign_member",
		"groups.read_members",
		"groups.manage_permissions",
	} {
		steps.Log("check seeded %s policy through HTTP", permission)
		body := `{"actor_user_id":"` + userID.String() + `","permission":"` + permission + `","object_type":"group","object_id":"` + seeding.AdminGroupID.String() + `"}`
		decision := fixture.do(t, postJSON("/permissions/check", body))
		assertStatus(t, decision, fiber.StatusOK)
		payload := decodeObject(t, decision)
		if payload["allowed"] != true {
			t.Fatalf("%s allowed = %v, want true payload = %+v", permission, payload["allowed"], payload)
		}
	}
}
