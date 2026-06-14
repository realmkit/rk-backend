package groups_e2e

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/e2e/harness"
	"github.com/realmkit/rk-backend/module/groups/domain"
	"github.com/realmkit/rk-backend/module/groups/port"
)

// TestGroupsPermissionGrantLifecycle verifies grant create, duplicate, delete, and recreate behavior.
func TestGroupsPermissionGrantLifecycle(t *testing.T) {
	steps := harness.NewSteps(t)
	fixture := newGroupsFixture(t)
	userID := uuid.New()
	forumID := uuid.New()
	grant := forumViewGrant(forumID, domain.SubjectUser, userID)

	steps.Log("create direct user forum view grant")
	created := fixture.createGrant(t, grant)
	assertDecision(t, fixture.checkPermission(t, checkForumViewBody(userID, forumID)), true)

	steps.Log("reject duplicate active grant through service contract")
	if _, err := fixture.service.CreatePermissionGrant(context.Background(), portCreateGrant(grant)); err == nil {
		t.Fatalf("CreatePermissionGrant() duplicate error = nil, want conflict")
	}

	steps.Log("delete grant and verify access is removed")
	if err := fixture.service.DeletePermissionGrant(context.Background(), portDeleteGrant(created.ID)); err != nil {
		t.Fatalf("DeletePermissionGrant() error = %v", err)
	}
	assertDecision(t, fixture.checkPermission(t, checkForumViewBody(userID, forumID)), false)

	steps.Log("recreate grant after deletion")
	fixture.createGrant(t, grant)
	assertDecision(t, fixture.checkPermission(t, checkForumViewBody(userID, forumID)), true)
}

// TestGroupsBuiltInPermissionsAreRecognized verifies representative module permissions.
func TestGroupsBuiltInPermissionsAreRecognized(t *testing.T) {
	steps := harness.NewSteps(t)
	fixture := newGroupsFixture(t)
	userID := uuid.New()
	routes := []struct {
		permission string
		objectType string
	}{
		{string(domain.PermissionForumsView), string(domain.ObjectForum)},
		{string(domain.PermissionPunishmentsIssue), string(domain.ObjectPunishment)},
		{string(domain.PermissionTicketsReply), string(domain.ObjectTicket)},
	}

	for _, route := range routes {
		steps.Log("check recognized permission %s", route.permission)
		body := `{"actor_user_id":"` + userID.String() +
			`","permission":"` + route.permission +
			`","object_type":"` + route.objectType +
			`","object_id":"` + uuid.NewString() + `"}`
		decision := fixture.checkPermission(t, body)
		assertDecision(t, decision, false)
		if decision["reason"] == "unknown_permission" {
			t.Fatalf("permission %s was not recognized", route.permission)
		}
	}
}

// portCreateGrant creates a grant command while keeping test bodies compact.
func portCreateGrant(grant domain.PermissionGrant) port.CreatePermissionGrantCommand {
	return port.CreatePermissionGrantCommand{Grant: grant}
}

// portDeleteGrant creates a delete grant command while keeping test bodies compact.
func portDeleteGrant(id uuid.UUID) port.DeletePermissionGrantCommand {
	return port.DeletePermissionGrantCommand{ID: id}
}
