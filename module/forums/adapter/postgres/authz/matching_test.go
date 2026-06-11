package authz

import (
	"testing"
	"time"

	"github.com/google/uuid"
	forumsdomain "github.com/realmkit/rk-backend/module/forums/domain"
	groupsdomain "github.com/realmkit/rk-backend/module/groups/domain"
)

// TestTupleMatchingCoversSupportedSubjectTypes verifies visibility grant matching.
func TestTupleMatchingCoversSupportedSubjectTypes(t *testing.T) {
	userID := uuid.New()
	groupID := uuid.New()
	memberships := map[uuid.UUID]bool{groupID: true}
	cases := []struct {
		name  string
		tuple relationTupleRow
		want  bool
	}{
		{
			"public",
			relationTupleRow{SubjectType: string(groupsdomain.SubjectPublic), SubjectID: groupsdomain.PublicSubjectID()},
			true,
		},
		{
			"authenticated",
			relationTupleRow{
				SubjectType: string(groupsdomain.SubjectAuthenticated),
				SubjectID:   groupsdomain.AuthenticatedSubjectID(),
			},
			true,
		},
		{
			"user",
			relationTupleRow{SubjectType: string(groupsdomain.SubjectUser), SubjectID: userID},
			true,
		},
		{
			"group",
			relationTupleRow{
				SubjectType:     string(groupsdomain.SubjectGroup),
				SubjectID:       groupID,
				SubjectRelation: string(groupsdomain.RelationMember),
			},
			true,
		},
		{
			"group wrong relation",
			relationTupleRow{
				SubjectType:     string(groupsdomain.SubjectGroup),
				SubjectID:       groupID,
				SubjectRelation: string(groupsdomain.RelationOwner),
			},
			false,
		},
		{
			"unknown",
			relationTupleRow{SubjectType: "robot", SubjectID: uuid.New()},
			false,
		},
	}

	for _, item := range cases {
		if got := tupleMatchesActor(item.tuple, userID, memberships); got != item.want {
			t.Fatalf("%s matched = %v, want %v", item.name, got, item.want)
		}
	}
}

// TestRelationAndGrantMappingCoversPermissionBuckets verifies relation helpers.
func TestRelationAndGrantMappingCoversPermissionBuckets(t *testing.T) {
	groupID := uuid.New()
	tuples := []relationTupleRow{
		{SubjectType: string(groupsdomain.SubjectGroup), SubjectID: groupID},
		{SubjectType: string(groupsdomain.SubjectGroup), SubjectID: groupID},
		{SubjectType: string(groupsdomain.SubjectUser), SubjectID: uuid.New()},
	}
	if ids := groupSubjectIDs(tuples); len(ids) != 1 || ids[0] != groupID {
		t.Fatalf("groupSubjectIDs() = %#v, want unique group id", ids)
	}

	settings := emptyPermissionSettings(uuid.New())
	grant := forumsdomain.ForumPermissionGrant{
		SubjectType: forumsdomain.PermissionSubjectUser,
		SubjectID:   uuid.New(),
	}
	addGrantToSettings(&settings, groupsdomain.RelationViewer, grant)
	addGrantToSettings(&settings, groupsdomain.RelationCreator, grant)
	addGrantToSettings(&settings, groupsdomain.RelationReplyer, grant)
	addGrantToSettings(&settings, groupsdomain.RelationLiker, grant)
	addGrantToSettings(&settings, groupsdomain.RelationModerator, grant)
	addGrantToSettings(&settings, groupsdomain.RelationManager, grant)
	if len(allPermissionGrants(settings)) != 6 {
		t.Fatalf("allPermissionGrants() count = %d, want 6", len(allPermissionGrants(settings)))
	}

	now := time.Now().UTC()
	rows := tuplesFromPermissionSettings(settings, uuid.New(), now)
	if len(rows) != 6 {
		t.Fatalf("tuplesFromPermissionSettings() count = %d, want 6", len(rows))
	}
	if rows[0].CreatedByUserID == nil {
		t.Fatalf("expected created_by_user_id when actor is present")
	}
	if rowGrant := grantFromTuple(relationTupleRow{
		SubjectType:     string(groupsdomain.SubjectGroup),
		SubjectID:       groupID,
		SubjectRelation: string(groupsdomain.RelationMember),
	}); rowGrant.SubjectID != groupID {
		t.Fatalf("grantFromTuple() = %#v", rowGrant)
	}
}

// TestSimulationRelationMappingCoversForumPermissions verifies simulation relation buckets.
func TestSimulationRelationMappingCoversForumPermissions(t *testing.T) {
	permissions := []groupsdomain.Permission{
		groupsdomain.PermissionForumsView,
		groupsdomain.PermissionForumsManageForum,
		groupsdomain.PermissionForumsCreateThread,
		groupsdomain.PermissionForumsReply,
		groupsdomain.PermissionForumsLikePosts,
		groupsdomain.PermissionForumsManageThreads,
	}
	for _, permission := range permissions {
		relations, err := simulationRelations(string(permission))
		if err != nil {
			t.Fatalf("simulationRelations(%s) error = %v", permission, err)
		}
		if len(relationNames(relations)) == 0 {
			t.Fatalf("simulationRelations(%s) returned no relations", permission)
		}
	}
	if _, err := simulationRelations("unknown.permission"); err == nil {
		t.Fatalf("expected unsupported permission to fail")
	}
}
