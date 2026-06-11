package forums_e2e

import (
	"context"
	"net/http"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/e2e/harness"
	groupsdomain "github.com/realmkit/rk-backend/module/groups/domain"
	groupsport "github.com/realmkit/rk-backend/module/groups/port"
)

func TestForumPermissionSubjectsAndDeleteGates(t *testing.T) {
	fixture := newForumsFixture(t)
	steps := harness.NewSteps(t)

	steps.Do("authenticated grant is invisible to anonymous but visible to signed-in users", func() {
		category := fixture.createCategory(t, "auth_visibility")
		forum := fixture.createForum(t, forumWrite{
			CategoryID: forumID(t, category, "id"), Key: "auth_only",
			Slug: "auth-only", Name: "Authenticated", Kind: "discussion", Order: 1,
		})
		fixture.grantSubject(
			t, forumID(t, forum, "id"), groupsdomain.RelationViewer,
			groupsdomain.SubjectAuthenticated, groupsdomain.AuthenticatedSubjectID(), "",
		)
		response := fixture.do(t, forumRequest(fiber.MethodGet, "/forums/tree", ""))
		assertForumStatus(t, response, fiber.StatusOK)
		if got := len(decodeForumObject(t, response)["categories"].([]any)); got != 0 {
			t.Fatalf("anonymous categories = %d, want 0", got)
		}
		response = fixture.do(t, forumRequest(fiber.MethodGet, "/forums/tree", "", forumUser(fixture.member)))
		assertForumStatus(t, response, fiber.StatusOK)
		if got := len(decodeForumObject(t, response)["categories"].([]any)); got != 1 {
			t.Fatalf("authenticated categories = %d, want 1", got)
		}
	})

	steps.Do("public grant allows anonymous forum tree reads", func() {
		category := fixture.createCategory(t, "public_visibility")
		forum := fixture.createForum(t, forumWrite{
			CategoryID: forumID(t, category, "id"), Key: "public_forum",
			Slug: "public-forum", Name: "Public", Kind: "discussion", Order: 1,
		})
		fixture.grantSubject(
			t, forumID(t, forum, "id"), groupsdomain.RelationViewer,
			groupsdomain.SubjectPublic, groupsdomain.PublicSubjectID(), "",
		)
		response := fixture.do(t, forumRequest(fiber.MethodGet, "/forums/tree", ""))
		assertForumStatus(t, response, fiber.StatusOK)
		if got := len(decodeForumObject(t, response)["categories"].([]any)); got == 0 {
			t.Fatalf("public tree returned no visible category")
		}
	})

	steps.Do("thread update/delete gates non-authors and stale versions", func() {
		forum := fixture.readyDiscussionForum(t, "delete_gate_forum", fixture.member)
		thread, _ := fixture.createThread(t, forum, fixture.member, "Delete Gate")
		threadID := forumID(t, thread, "id")
		response := fixture.do(t, forumRequest(
			fiber.MethodPatch, "/threads/"+threadID.String(),
			`{"title":"Bad Actor","slug":"bad-actor"}`,
			forumUser(fixture.other), forumIdempotency("thread-update-other"),
			forumIfMatch(forumVersion(thread)),
		))
		assertForumStatus(t, response, fiber.StatusForbidden)
		response = fixture.do(t, forumRequest(
			fiber.MethodDelete, "/threads/"+threadID.String(), "",
			forumUser(fixture.member), forumIdempotency("thread-delete-stale"),
			forumIfMatch(99),
		))
		assertForumStatus(t, response, fiber.StatusPreconditionFailed)
		response = fixture.do(t, forumRequest(
			fiber.MethodDelete, "/threads/"+threadID.String(), "",
			forumUser(fixture.manager), forumIdempotency("thread-delete-manager"),
			forumIfMatch(forumVersion(thread)),
		))
		assertForumStatus(t, response, fiber.StatusNoContent)
	})
}

func TestForumOpenAPIContractCoversEveryRoute(t *testing.T) {
	routes := []struct {
		method string
		path   string
	}{
		{http.MethodGet, "/forums/tree"},
		{http.MethodPost, "/forum-categories"},
		{http.MethodGet, "/forum-categories"},
		{http.MethodGet, "/forum-categories/{category_id}"},
		{http.MethodPatch, "/forum-categories/{category_id}"},
		{http.MethodDelete, "/forum-categories/{category_id}"},
		{http.MethodPost, "/forum-categories/reorder"},
		{http.MethodPost, "/forums"},
		{http.MethodGet, "/forums"},
		{http.MethodGet, "/forums/{forum_id}"},
		{http.MethodPatch, "/forums/{forum_id}"},
		{http.MethodDelete, "/forums/{forum_id}"},
		{http.MethodPost, "/forums/{forum_id}/move"},
		{http.MethodPost, "/forums/reorder"},
		{http.MethodGet, "/forums/{forum_id}/threads"},
		{http.MethodPost, "/forums/{forum_id}/threads"},
		{http.MethodGet, "/threads/{thread_id}"},
		{http.MethodPatch, "/threads/{thread_id}"},
		{http.MethodDelete, "/threads/{thread_id}"},
		{http.MethodGet, "/threads/{thread_id}/posts"},
		{http.MethodPost, "/threads/{thread_id}/posts"},
		{http.MethodGet, "/posts/{post_id}"},
		{http.MethodPatch, "/posts/{post_id}"},
		{http.MethodDelete, "/posts/{post_id}"},
		{http.MethodPut, "/posts/{post_id}/like"},
		{http.MethodDelete, "/posts/{post_id}/like"},
		{http.MethodGet, "/posts/{post_id}/revisions"},
		{http.MethodGet, "/forums/latest-posts"},
		{http.MethodGet, "/forums/{forum_id}/latest-posts"},
		{http.MethodGet, "/forums/{forum_id}/posts/most-liked"},
		{http.MethodPost, "/threads/{thread_id}/read"},
		{http.MethodPost, "/forums/{forum_id}/read"},
		{http.MethodGet, "/forums/unread-summary"},
		{http.MethodGet, "/forums/search"},
		{http.MethodGet, "/forums/{forum_id}/search"},
		{http.MethodGet, "/forums/{forum_id}/settings"},
		{http.MethodPatch, "/forums/{forum_id}/settings"},
		{http.MethodGet, "/forums/{forum_id}/permissions"},
		{http.MethodPut, "/forums/{forum_id}/permissions"},
		{http.MethodPost, "/forums/{forum_id}/permissions/simulate"},
	}
	for _, route := range routes {
		assertForumOpenAPIRoute(t, route.method, route.path)
	}
}

func (fixture forumsFixture) grantSubject(
	t *testing.T,
	objectID uuid.UUID,
	relation groupsdomain.Relation,
	subjectType groupsdomain.SubjectType,
	subjectID uuid.UUID,
	subjectRelation groupsdomain.Relation,
) {
	t.Helper()
	_, err := fixture.groups.CreateTuple(context.Background(), groupsport.CreateTupleCommand{
		Tuple: groupsdomain.RelationTuple{
			ObjectType: groupsdomain.ObjectForum, ObjectID: objectID,
			Relation: relation, SubjectType: subjectType,
			SubjectID: subjectID, SubjectRelation: subjectRelation,
		},
	})
	if err != nil {
		t.Fatalf("CreateTuple(%s/%s) error = %v", subjectType, relation, err)
	}
	if err := fixture.service.ClearReadCache(context.Background()); err != nil {
		t.Fatalf("ClearReadCache() error = %v", err)
	}
}
