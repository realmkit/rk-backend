package groups_e2e

import (
	"context"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/niflaot/gamehub-go/e2e/harness"
	"github.com/niflaot/gamehub-go/module/groups/domain"
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

	steps.Log("seed representative relation tuples")
	fixture.createTuple(t, forumViewerTuple(publicForum, domain.SubjectPublic, domain.PublicSubjectID(), ""))
	fixture.createTuple(t, forumViewerTuple(authForum, domain.SubjectAuthenticated, domain.AuthenticatedSubjectID(), ""))
	fixture.createTuple(t, forumViewerTuple(userForum, domain.SubjectUser, userID, ""))
	fixture.createTuple(t, forumViewerTuple(groupForum, domain.SubjectGroup, groupID, domain.RelationMember))

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

// TestGroupsCustomPolicyConditions verifies contextual condition allow and deny paths.
func TestGroupsCustomPolicyConditions(t *testing.T) {
	steps := harness.NewSteps(t)
	fixture := newGroupsFixture(t)
	userID := uuid.New()
	postID := uuid.New()

	steps.Log("seed custom posts.update policy requiring open thread status")
	seedPostUpdatePolicy(t, fixture)
	fixture.createTuple(
		t,
		domain.RelationTuple{
			ObjectType:  domain.ObjectForumPost,
			ObjectID:    postID,
			Relation:    domain.RelationAuthor,
			SubjectType: domain.SubjectUser,
			SubjectID:   userID,
		},
	)

	steps.Log("allow when condition context matches")
	openBody := checkPostUpdateBody(userID, postID, "open")
	assertDecision(t, fixture.checkPermission(t, openBody), true)

	steps.Log("deny when condition context does not match")
	lockedBody := checkPostUpdateBody(userID, postID, "locked")
	decision := fixture.checkPermission(t, lockedBody)
	assertDecision(t, decision, false)
	if decision["reason"] != "conditions_failed" {
		t.Fatalf("reason = %v, want conditions_failed", decision["reason"])
	}
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

// forumViewerTuple creates one forum viewer tuple.
func forumViewerTuple(
	forumID uuid.UUID,
	subjectType domain.SubjectType,
	subjectID uuid.UUID,
	subjectRelation domain.Relation,
) domain.RelationTuple {
	return domain.RelationTuple{
		ObjectType:      domain.ObjectForum,
		ObjectID:        forumID,
		Relation:        domain.RelationViewer,
		SubjectType:     subjectType,
		SubjectID:       subjectID,
		SubjectRelation: subjectRelation,
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
func seedPostUpdatePolicy(t *testing.T, fixture groupsFixture) {
	t.Helper()
	_, err := fixture.policies.UpsertDefinition(
		context.Background(),
		domain.PermissionDefinition{
			ID:          uuid.New(),
			Permission:  domain.PermissionPostsUpdate,
			ObjectType:  domain.ObjectForumPost,
			Description: "E2E conditioned post update",
			Enabled:     true,
			Version:     1,
		},
	)
	if err != nil {
		t.Fatalf("UpsertDefinition() error = %v", err)
	}
	_, err = fixture.policies.UpsertRule(
		context.Background(),
		domain.PermissionRule{
			ID:         uuid.New(),
			Permission: domain.PermissionPostsUpdate,
			ObjectType: domain.ObjectForumPost,
			Relation:   domain.RelationAuthor,
			Conditions: []domain.PolicyCondition{
				{
					Type:  domain.ConditionEquals,
					Field: "thread_status",
					Value: "open",
				},
			},
			Priority: 1,
			Enabled:  true,
		},
	)
	if err != nil {
		t.Fatalf("UpsertRule() error = %v", err)
	}
}
