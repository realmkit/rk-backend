package forums_e2e

import (
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/e2e/harness"
	groupsdomain "github.com/realmkit/rk-backend/module/groups/domain"
)

func TestForumThreadsPostsAndRevisionJourney(t *testing.T) {
	fixture := newForumsFixture(t)
	steps := harness.NewSteps(t)
	forum := fixture.readyDiscussionForum(t, "content_forum", fixture.member)
	var thread, opener, reply map[string]any

	steps.Do("create thread with opener post and reject missing idempotency", func() {
		forumID := forumID(t, forum, "id")
		response := fixture.do(t, forumRequest(
			fiber.MethodPost, "/forums/"+forumID.String()+"/threads",
			threadBody("Missing Idempotency"), forumUser(fixture.member),
		))
		assertForumStatus(t, response, fiber.StatusBadRequest)
		response = fixture.do(t, forumRequest(
			fiber.MethodPost, "/forums/"+forumID.String()+"/threads",
			threadBody("Welcome Thread"), forumUser(fixture.member),
			forumIdempotency("thread-welcome"),
		))
		assertForumStatus(t, response, fiber.StatusCreated)
		payload := decodeForumObject(t, response)
		thread = payload["thread"].(map[string]any)
		opener = payload["post"].(map[string]any)
	})

	steps.Do("list, fetch, and update thread with optimistic concurrency", func() {
		currentForumID := forumID(t, forum, "id")
		response := fixture.do(t, forumRequest(
			fiber.MethodGet, "/forums/"+currentForumID.String()+"/threads?page_size=1", "",
			forumUser(fixture.member),
		))
		assertForumStatus(t, response, fiber.StatusOK)
		if got := len(decodeForumList(t, response)); got != 1 {
			t.Fatalf("threads listed = %d, want 1", got)
		}
		threadID := forumID(t, thread, "id")
		response = fixture.do(t, forumRequest(
			fiber.MethodGet, "/threads/"+threadID.String(), "", forumUser(fixture.member),
		))
		assertForumStatus(t, response, fiber.StatusOK)
		response = fixture.do(t, forumRequest(
			fiber.MethodPatch, "/threads/"+threadID.String(),
			`{"title":"Welcome Thread Edited","slug":"welcome-thread-edited"}`,
			forumUser(fixture.member), forumIdempotency("thread-edit"),
			forumIfMatch(forumVersion(thread)),
		))
		assertForumStatus(t, response, fiber.StatusOK)
		thread = decodeForumObject(t, response)
	})

	steps.Do("create reply with quote and asset references", func() {
		assetID := uuid.New()
		fixture.assets.known[assetID] = true
		body := postBodyWithReferences("quoted reply", forumID(t, opener, "id"), assetID)
		response := fixture.do(t, forumRequest(
			fiber.MethodPost, "/threads/"+forumID(t, thread, "id").String()+"/posts",
			body, forumUser(fixture.member), forumIdempotency("reply-with-references"),
		))
		assertForumStatus(t, response, fiber.StatusCreated)
		reply = decodeForumObject(t, response)
		if got := uint64(reply["sequence"].(float64)); got != 2 {
			t.Fatalf("reply sequence = %d, want 2", got)
		}
	})

	steps.Do("reject invalid attachment reference and locked thread replies", func() {
		body := postBodyWithReferences("bad asset", forumID(t, opener, "id"), uuid.New())
		response := fixture.do(t, forumRequest(
			fiber.MethodPost, "/threads/"+forumID(t, thread, "id").String()+"/posts",
			body, forumUser(fixture.member), forumIdempotency("reply-bad-asset"),
		))
		assertForumStatus(t, response, fiber.StatusNotFound)

		locked := fixture.lockedThread(t, fixture.member)
		response = fixture.do(t, forumRequest(
			fiber.MethodPost, "/threads/"+forumID(t, locked, "id").String()+"/posts",
			postBody("cannot reply"), forumUser(fixture.member),
			forumIdempotency("reply-locked"),
		))
		assertForumStatus(t, response, fiber.StatusConflict)
	})

	steps.Do("edit, list revisions, and delete posts with permission gates", func() {
		response := fixture.do(t, forumRequest(
			fiber.MethodPatch, "/posts/"+forumID(t, reply, "id").String(),
			`{"content_document_json":{"type":"doc","content":[{"type":"text","text":"edited"}]},`+
				`"edit_reason":"clarified"}`,
			forumUser(fixture.member), forumIdempotency("edit-reply"), forumIfMatch(forumVersion(reply)),
		))
		assertForumStatus(t, response, fiber.StatusOK)
		reply = decodeForumObject(t, response)

		response = fixture.do(t, forumRequest(
			fiber.MethodGet, "/posts/"+forumID(t, reply, "id").String()+"/revisions", "",
			forumUser(fixture.member),
		))
		assertForumStatus(t, response, fiber.StatusForbidden)
		response = fixture.do(t, forumRequest(
			fiber.MethodGet, "/posts/"+forumID(t, reply, "id").String()+"/revisions", "",
			forumUser(fixture.manager),
		))
		assertForumStatus(t, response, fiber.StatusOK)
		if got := len(decodeForumList(t, response)); got != 1 {
			t.Fatalf("revisions listed = %d, want 1", got)
		}

		response = fixture.do(t, forumRequest(
			fiber.MethodDelete, "/posts/"+forumID(t, reply, "id").String(), "",
			forumUser(fixture.other), forumIdempotency("delete-forbidden"),
			forumIfMatch(forumVersion(reply)),
		))
		assertForumStatus(t, response, fiber.StatusForbidden)
		response = fixture.do(t, forumRequest(
			fiber.MethodDelete, "/posts/"+forumID(t, reply, "id").String(), "",
			forumUser(fixture.manager), forumIdempotency("delete-reply"),
			forumIfMatch(forumVersion(reply)),
		))
		assertForumStatus(t, response, fiber.StatusNoContent)
	})
}

func (fixture forumsFixture) readyDiscussionForum(t *testing.T, key string, member uuid.UUID) map[string]any {
	t.Helper()
	category := fixture.createCategory(t, key+"_category")
	forum := fixture.createForum(t, forumWrite{
		CategoryID: forumID(t, category, "id"), Key: key,
		Slug: slug(key), Name: key, Kind: "discussion", Order: 1,
	})
	for _, relation := range []groupsdomain.Relation{
		groupsdomain.RelationViewer, groupsdomain.RelationCreator,
		groupsdomain.RelationReplyer, groupsdomain.RelationLiker,
	} {
		fixture.grant(t, forumID(t, forum, "id"), relation, member)
	}
	fixture.grant(t, forumID(t, forum, "id"), groupsdomain.RelationModerator, fixture.manager)
	return forum
}

func (fixture forumsFixture) lockedThread(t *testing.T, actor uuid.UUID) map[string]any {
	t.Helper()
	category := fixture.createCategory(t, "locked_category")
	forum := fixture.createForum(t, forumWrite{
		CategoryID: forumID(t, category, "id"), Key: "locked_forum",
		Slug: "locked-forum", Name: "Locked Forum", Kind: "discussion",
		DefaultStatus: "locked", Order: 1,
	})
	fixture.grant(t, forumID(t, forum, "id"), groupsdomain.RelationCreator, actor)
	fixture.grant(t, forumID(t, forum, "id"), groupsdomain.RelationReplyer, actor)
	response := fixture.do(t, forumRequest(
		fiber.MethodPost, "/forums/"+forumID(t, forum, "id").String()+"/threads",
		threadBody("Locked Thread"), forumUser(actor), forumIdempotency("locked-thread"),
	))
	assertForumStatus(t, response, fiber.StatusCreated)
	return decodeForumObject(t, response)["thread"].(map[string]any)
}

func postBodyWithReferences(text string, postID uuid.UUID, assetID uuid.UUID) string {
	return `{"content_document_json":{"type":"doc","content":[{"type":"text","text":"` + text + `"}]},` +
		`"references":[{"target_post_id":"` + postID.String() + `","reference_type":"quote",` +
		`"quote_excerpt":"Welcome"},{"target_asset_id":"` + assetID.String() + `",` +
		`"reference_type":"attachment"}]}`
}
