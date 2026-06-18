package authz

import (
	"testing"
	"time"

	"github.com/google/uuid"
	forumsdomain "github.com/realmkit/rk-backend/module/forums/domain"
	groupsdomain "github.com/realmkit/rk-backend/module/groups/domain"
)

// TestGrantMatchingCoversSupportedSubjectTypes verifies visibility grant matching.
func TestGrantMatchingCoversSupportedSubjectTypes(t *testing.T) {
	userID := uuid.New()
	groupID := uuid.New()
	memberships := map[uuid.UUID]bool{groupID: true}
	cases := []struct {
		name  string
		grant permissionGrantRow
		want  bool
	}{
		{
			"public",
			permissionGrantRow{SubjectType: string(groupsdomain.SubjectPublic), SubjectID: groupsdomain.PublicSubjectID()},
			true,
		},
		{
			"authenticated",
			permissionGrantRow{
				SubjectType: string(groupsdomain.SubjectAuthenticated),
				SubjectID:   groupsdomain.AuthenticatedSubjectID(),
			},
			true,
		},
		{
			"user",
			permissionGrantRow{SubjectType: string(groupsdomain.SubjectUser), SubjectID: userID},
			true,
		},
		{
			"group",
			permissionGrantRow{SubjectType: string(groupsdomain.SubjectGroup), SubjectID: groupID},
			true,
		},
		{
			"group without membership",
			permissionGrantRow{SubjectType: string(groupsdomain.SubjectGroup), SubjectID: uuid.New()},
			false,
		},
		{
			"unknown",
			permissionGrantRow{SubjectType: "robot", SubjectID: uuid.New()},
			false,
		},
	}

	for _, item := range cases {
		if got := grantMatchesActor(item.grant, userID, memberships); got != item.want {
			t.Fatalf("%s matched = %v, want %v", item.name, got, item.want)
		}
	}
}

// TestActionAndGrantMappingCoversPermissionBuckets verifies action helpers.
func TestActionAndGrantMappingCoversPermissionBuckets(t *testing.T) {
	groupID := uuid.New()
	grants := []permissionGrantRow{
		{SubjectType: string(groupsdomain.SubjectGroup), SubjectID: groupID},
		{SubjectType: string(groupsdomain.SubjectGroup), SubjectID: groupID},
		{SubjectType: string(groupsdomain.SubjectUser), SubjectID: uuid.New()},
	}
	if ids := grantGroupIDs(grants, nil); len(ids) != 1 || ids[0] != groupID {
		t.Fatalf("grantGroupIDs() = %#v, want unique group id", ids)
	}

	settings := emptyPermissionSettings(uuid.New())
	grant := forumsdomain.ForumPermissionGrant{
		SubjectType: forumsdomain.PermissionSubjectUser,
		SubjectID:   uuid.New(),
	}
	addGrantToSettings(&settings, groupsdomain.PermissionForumsView, grant)
	addGrantToSettings(&settings, groupsdomain.PermissionForumsCreateThread, grant)
	addGrantToSettings(&settings, groupsdomain.PermissionForumsReply, grant)
	addGrantToSettings(&settings, groupsdomain.PermissionForumsLikePosts, grant)
	addGrantToSettings(&settings, groupsdomain.PermissionForumsViewAllThreads, grant)
	addGrantToSettings(&settings, groupsdomain.PermissionForumsBypassThreadLimits, grant)
	addGrantToSettings(&settings, groupsdomain.PermissionForumsPinThreads, grant)
	addGrantToSettings(&settings, groupsdomain.PermissionForumsManageThreads, grant)
	addGrantToSettings(&settings, groupsdomain.PermissionForumsManagePosts, grant)
	addGrantToSettings(&settings, groupsdomain.PermissionForumsAdministrativeAccess, grant)
	addGrantToSettings(&settings, groupsdomain.PermissionForumsManageForum, grant)
	if len(allPermissionGrants(settings)) != 10 {
		t.Fatalf("allPermissionGrants() count = %d, want 10", len(allPermissionGrants(settings)))
	}

	now := time.Now().UTC()
	rows := rowsFromPermissionSettings(settings, uuid.New(), now)
	if len(rows) != 10 {
		t.Fatalf("rowsFromPermissionSettings() count = %d, want 10", len(rows))
	}
	if rows[0].CreatedByUserID == nil {
		t.Fatalf("expected created_by_user_id when actor is present")
	}
	if rowGrant := grantFromRow(permissionGrantRow{
		SubjectType: string(groupsdomain.SubjectGroup),
		SubjectID:   groupID,
	}); rowGrant.SubjectID != groupID {
		t.Fatalf("grantFromRow() = %#v", rowGrant)
	}
}

// TestSimulationActionMappingCoversForumPermissions verifies simulation action buckets.
func TestSimulationActionMappingCoversForumPermissions(t *testing.T) {
	permissions := []groupsdomain.Permission{
		groupsdomain.PermissionForumsView,
		groupsdomain.PermissionForumsManageForum,
		groupsdomain.PermissionForumsCreateThread,
		groupsdomain.PermissionForumsReply,
		groupsdomain.PermissionForumsLikePosts,
		groupsdomain.PermissionForumsPinThreads,
		groupsdomain.PermissionForumsManageThreads,
		groupsdomain.PermissionForumsManagePosts,
		groupsdomain.PermissionForumsViewAllThreads,
		groupsdomain.PermissionForumsBypassThreadLimits,
		groupsdomain.PermissionForumsAdministrativeAccess,
	}
	for _, permission := range permissions {
		actions, err := simulationActions(string(permission))
		if err != nil {
			t.Fatalf("simulationActions(%s) error = %v", permission, err)
		}
		if len(actionNames(actions)) == 0 {
			t.Fatalf("simulationActions(%s) returned no actions", permission)
		}
	}
	if _, err := simulationActions("unknown.permission"); err == nil {
		t.Fatalf("expected unsupported permission to fail")
	}
}
