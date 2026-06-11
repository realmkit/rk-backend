package groups_e2e

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/e2e/harness"
	"github.com/realmkit/rk-backend/module/groups/domain"
	"github.com/realmkit/rk-backend/module/groups/port"
)

// TestGroupsRelationTupleLifecycle verifies tuple create, duplicate, delete, and recreate behavior.
func TestGroupsRelationTupleLifecycle(t *testing.T) {
	steps := harness.NewSteps(t)
	fixture := newGroupsFixture(t)
	userID := uuid.New()
	forumID := uuid.New()
	tuple := forumViewerTuple(forumID, domain.SubjectUser, userID, "")

	steps.Log("create direct user viewer tuple")
	created := fixture.createTuple(t, tuple)
	assertDecision(t, fixture.checkPermission(t, checkForumViewBody(userID, forumID)), true)

	steps.Log("reject duplicate active tuple through service contract")
	if _, err := fixture.service.CreateTuple(context.Background(), portCreateTuple(tuple)); err == nil {
		t.Fatalf("CreateTuple() duplicate error = nil, want conflict")
	}

	steps.Log("delete tuple and verify access is removed")
	if err := fixture.service.DeleteTuple(context.Background(), portDeleteTuple(created.ID)); err != nil {
		t.Fatalf("DeleteTuple() error = %v", err)
	}
	assertDecision(t, fixture.checkPermission(t, checkForumViewBody(userID, forumID)), false)

	steps.Log("recreate tuple after deletion")
	fixture.createTuple(t, tuple)
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

// portCreateTuple creates a tuple command while keeping test bodies compact.
func portCreateTuple(tuple domain.RelationTuple) port.CreateTupleCommand {
	return port.CreateTupleCommand{Tuple: tuple}
}

// portDeleteTuple creates a delete tuple command while keeping test bodies compact.
func portDeleteTuple(id uuid.UUID) port.DeleteTupleCommand {
	return port.DeleteTupleCommand{ID: id}
}
