package groups_e2e

import (
	"context"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/e2e/harness"
	"github.com/realmkit/rk-backend/module/groups/domain"
)

// TestGroupsPermissionSubjects verifies public, authenticated, user, and group grants.
func TestGroupsPermissionSubjects(t *testing.T) {
	steps := harness.NewSteps(t)
	fixture := newGroupsFixture(t)
	userID := uuid.New()
	otherID := uuid.New()
	group := fixture.createGroup(t, "viewers")
	groupID := groupIDFrom(t, group)

	steps.Log("assign user to viewer group")
	membershipVersion := assignMember(t, fixture, groupID, userID, "subject-permissions")

	publicForum := uuid.New()
	authForum := uuid.New()
	userForum := uuid.New()
	groupForum := uuid.New()

	steps.Log("seed representative permission grants")
	fixture.createGrant(t, forumViewGrant(publicForum, domain.SubjectPublic, domain.PublicSubjectID()))
	fixture.createGrant(t, forumViewGrant(authForum, domain.SubjectAuthenticated, domain.AuthenticatedSubjectID()))
	fixture.createGrant(t, forumViewGrant(userForum, domain.SubjectUser, userID))
	fixture.createGrant(t, forumViewGrant(groupForum, domain.SubjectGroup, groupID))

	steps.Log("public grants allow anonymous and authenticated users")
	assertDecision(t, fixture.checkPermission(t, checkForumViewBody(uuid.Nil, publicForum)), true)
	assertDecision(t, fixture.checkPermission(t, checkForumViewBody(userID, publicForum)), true)

	steps.Log("authenticated grants deny anonymous users")
	assertDecision(t, fixture.checkPermission(t, checkForumViewBody(uuid.Nil, authForum)), false)
	assertDecision(t, fixture.checkPermission(t, checkForumViewBody(userID, authForum)), true)

	steps.Log("direct user grants only allow the matching user")
	assertDecision(t, fixture.checkPermission(t, checkForumViewBody(userID, userForum)), true)
	assertDecision(t, fixture.checkPermission(t, checkForumViewBody(otherID, userForum)), false)

	steps.Log("group grants allow active members")
	assertDecision(t, fixture.checkPermission(t, checkForumViewBody(userID, groupForum)), true)
	assertDecision(t, fixture.checkPermission(t, checkForumViewBody(otherID, groupForum)), false)

	steps.Log("removed memberships stop granting access")
	removeMember(t, fixture, groupID, userID, membershipVersion)
	assertDecision(t, fixture.checkPermission(t, checkForumViewBody(userID, groupForum)), false)
}

// TestGroupsUnknownPermissionReturnsValidationProblem verifies unknown permission checks fail safely.
func TestGroupsUnknownPermissionReturnsValidationProblem(t *testing.T) {
	steps := harness.NewSteps(t)
	fixture := newGroupsFixture(t)

	steps.Log("request an unknown permission check")
	response := fixture.do(
		t,
		harness.JSONRequest(
			fiber.MethodPost,
			"/permissions/check",
			`{"actor_user_id":"`+uuid.NewString()+`","permission":"not.real","object_type":"forum","object_id":"`+uuid.NewString()+`"}`,
		),
	)
	assertGroupsStatus(t, response, fiber.StatusUnprocessableEntity)
}

// TestGroupsCustomActionGrant verifies database-backed custom actions.
func TestGroupsCustomActionGrant(t *testing.T) {
	steps := harness.NewSteps(t)
	fixture := newGroupsFixture(t)
	userID := uuid.New()
	postID := uuid.New()

	steps.Log("seed custom posts.update action and grant")
	seedPostUpdateAction(t, fixture)
	fixture.createGrant(
		t,
		domain.PermissionGrant{
			SubjectType: domain.SubjectUser,
			SubjectID:   userID,
			Action:      domain.PermissionPostsUpdate,
			ScopeType:   domain.ObjectForumPost,
			ScopeID:     postID,
		},
	)

	steps.Log("allow when direct grant matches")
	openBody := checkPostUpdateBody(userID, postID, "open")
	assertDecision(t, fixture.checkPermission(t, openBody), true)

	steps.Log("deny another user without grant")
	decision := fixture.checkPermission(t, checkPostUpdateBody(uuid.New(), postID, "open"))
	assertDecision(t, decision, false)
}

// assignMember assigns a group membership through HTTP.
func assignMember(t *testing.T, fixture groupsFixture, groupID uuid.UUID, userID uuid.UUID, key string) uint64 {
	t.Helper()
	response := fixture.do(
		t,
		configureRequest(
			harness.JSONRequest(
				fiber.MethodPut,
				"/groups/"+groupID.String()+"/members/"+userID.String(),
				`{"status":"active","assigned_reason":"permissions"}`,
			),
			withGroupsIdempotency("assign-"+key),
		),
	)
	assertGroupsStatus(t, response, fiber.StatusOK)
	return uint64(decodeGroupsObject(t, response)["version"].(float64))
}

// removeMember removes a group membership through HTTP.
func removeMember(t *testing.T, fixture groupsFixture, groupID uuid.UUID, userID uuid.UUID, version uint64) {
	t.Helper()
	response := fixture.do(
		t,
		configureRequest(
			harness.JSONRequest(fiber.MethodDelete, "/groups/"+groupID.String()+"/members/"+userID.String(), ""),
			withGroupsIdempotency("remove-"+userID.String()),
			withGroupsIfMatch(version),
		),
	)
	assertGroupsStatus(t, response, fiber.StatusNoContent)
}

// forumViewGrant creates one forum view grant.
func forumViewGrant(
	forumID uuid.UUID,
	subjectType domain.SubjectType,
	subjectID uuid.UUID,
) domain.PermissionGrant {
	return domain.PermissionGrant{
		SubjectType: subjectType,
		SubjectID:   subjectID,
		Action:      domain.PermissionForumsView,
		ScopeType:   domain.ObjectForum,
		ScopeID:     forumID,
	}
}

// checkForumViewBody builds a forum permission check body.
func checkForumViewBody(actorID uuid.UUID, forumID uuid.UUID) string {
	return `{"actor_user_id":"` + actorID.String() + `","permission":"forums.view","object_type":"forum","object_id":"` + forumID.String() + `"}`
}

// checkPostUpdateBody builds a post update permission check body.
func checkPostUpdateBody(actorID uuid.UUID, postID uuid.UUID, status string) string {
	return `{"actor_user_id":"` + actorID.String() + `","permission":"posts.update","object_type":"forum_post","object_id":"` + postID.String() + `","context":{"thread_status":"` + status + `"}}`
}

// assertDecision verifies one permission decision.
func assertDecision(t *testing.T, payload map[string]any, want bool) {
	t.Helper()
	if payload["allowed"] != want {
		t.Fatalf("allowed = %v, want %v payload = %+v", payload["allowed"], want, payload)
	}
}

// seedPostUpdatePolicy stores a custom permission rule with conditions.
func seedPostUpdateAction(t *testing.T, fixture groupsFixture) {
	t.Helper()
	_, err := fixture.policies.UpsertAction(
		context.Background(),
		domain.PermissionAction{
			ID:           uuid.New(),
			Action:       domain.PermissionPostsUpdate,
			Area:         "posts",
			ScopeType:    domain.ObjectForumPost,
			Label:        "Update posts",
			Description:  "E2E custom post update",
			WarningLevel: domain.WarningLevelDangerous,
			Enabled:      true,
			Version:      1,
		},
	)
	if err != nil {
		t.Fatalf("UpsertAction() error = %v", err)
	}
}
