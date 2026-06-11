package forums_e2e

import (
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/niflaot/gamehub-go/e2e/harness"
)

func TestForumInteractionsWidgetsAndReadState(t *testing.T) {
	fixture := newForumsFixture(t)
	steps := harness.NewSteps(t)
	forum := fixture.readyDiscussionForum(t, "interaction_forum", fixture.member)
	thread, opener := fixture.createThread(t, forum, fixture.member, "Interaction Thread")

	steps.Do("like and unlike are idempotent and enforce authentication", func() {
		postID := forumID(t, opener, "id")
		response := fixture.do(t, forumRequest(
			fiber.MethodPut, "/posts/"+postID.String()+"/like", "",
			forumIdempotency("anonymous-like"),
		))
		assertForumStatus(t, response, fiber.StatusUnauthorized)
		response = fixture.do(t, forumRequest(
			fiber.MethodPut, "/posts/"+postID.String()+"/like", "",
			forumUser(fixture.member), forumIdempotency("like-one"),
		))
		assertForumStatus(t, response, fiber.StatusOK)
		if got := decodeForumObject(t, response)["like_count"].(float64); got != 1 {
			t.Fatalf("like count = %.0f, want 1", got)
		}
		response = fixture.do(t, forumRequest(
			fiber.MethodPut, "/posts/"+postID.String()+"/like", "",
			forumUser(fixture.member), forumIdempotency("like-two"),
		))
		assertForumStatus(t, response, fiber.StatusOK)
		if got := decodeForumObject(t, response)["like_count"].(float64); got != 1 {
			t.Fatalf("duplicate like count = %.0f, want 1", got)
		}
		response = fixture.do(t, forumRequest(
			fiber.MethodDelete, "/posts/"+postID.String()+"/like", "",
			forumUser(fixture.member), forumIdempotency("unlike-one"),
		))
		assertForumStatus(t, response, fiber.StatusOK)
		response = fixture.do(t, forumRequest(
			fiber.MethodDelete, "/posts/"+postID.String()+"/like", "",
			forumUser(fixture.member), forumIdempotency("unlike-two"),
		))
		assertForumStatus(t, response, fiber.StatusOK)
	})

	steps.Do("widgets respect visibility and include latest and most-liked posts", func() {
		postID := forumID(t, opener, "id")
		_ = fixture.do(t, forumRequest(
			fiber.MethodPut, "/posts/"+postID.String()+"/like", "",
			forumUser(fixture.member), forumIdempotency("like-for-widgets"),
		))
		response := fixture.do(t, forumRequest(
			fiber.MethodGet, "/forums/latest-posts", "", forumUser(fixture.member),
		))
		assertForumStatus(t, response, fiber.StatusOK)
		if got := len(decodeForumList(t, response)); got != 1 {
			t.Fatalf("latest posts = %d, want 1", got)
		}
		response = fixture.do(t, forumRequest(
			fiber.MethodGet, "/forums/"+forumID(t, forum, "id").String()+"/posts/most-liked",
			"", forumUser(fixture.member),
		))
		assertForumStatus(t, response, fiber.StatusOK)
		if got := len(decodeForumList(t, response)); got != 1 {
			t.Fatalf("most-liked posts = %d, want 1", got)
		}
		response = fixture.do(t, forumRequest(
			fiber.MethodGet, "/forums/latest-posts", "", forumUser(uuid.New()),
		))
		assertForumStatus(t, response, fiber.StatusOK)
		if got := len(decodeForumList(t, response)); got != 0 {
			t.Fatalf("invisible latest posts = %d, want 0", got)
		}
	})

	steps.Do("read state stores thread progress and clears forum unread counts", func() {
		response := fixture.do(t, forumRequest(
			fiber.MethodGet, "/forums/unread-summary", "", forumUser(fixture.member),
		))
		assertForumStatus(t, response, fiber.StatusOK)
		if got := decodeForumObject(t, response)["unread_thread_count"].(float64); got < 1 {
			t.Fatalf("unread before read = %.0f, want at least 1", got)
		}
		response = fixture.do(t, forumRequest(
			fiber.MethodPost, "/threads/"+forumID(t, thread, "id").String()+"/read",
			`{"last_read_post_sequence":1}`,
			forumUser(fixture.member), forumIdempotency("read-thread"),
		))
		assertForumStatus(t, response, fiber.StatusOK)
		response = fixture.do(t, forumRequest(
			fiber.MethodPost, "/forums/"+forumID(t, forum, "id").String()+"/read", "",
			forumUser(fixture.member), forumIdempotency("read-forum"),
		))
		assertForumStatus(t, response, fiber.StatusNoContent)
		response = fixture.do(t, forumRequest(
			fiber.MethodGet, "/forums/unread-summary", "", forumUser(fixture.member),
		))
		assertForumStatus(t, response, fiber.StatusOK)
		if got := decodeForumObject(t, response)["unread_thread_count"].(float64); got != 0 {
			t.Fatalf("unread after forum read = %.0f, want 0", got)
		}
	})

	steps.Do("punishment restriction seam blocks forum action keys", func() {
		fixture.restrictions.blocked[fixture.member] = map[string]bool{
			"gamehub.forums.like_posts": true,
		}
		response := fixture.do(t, forumRequest(
			fiber.MethodPut, "/posts/"+forumID(t, opener, "id").String()+"/like", "",
			forumUser(fixture.member), forumIdempotency("restricted-like"),
		))
		assertForumStatus(t, response, fiber.StatusForbidden)
	})
}

func (fixture forumsFixture) createThread(
	t *testing.T,
	forum map[string]any,
	actor uuid.UUID,
	title string,
) (map[string]any, map[string]any) {
	t.Helper()
	response := fixture.do(t, forumRequest(
		fiber.MethodPost, "/forums/"+forumID(t, forum, "id").String()+"/threads",
		threadBody(title), forumUser(actor), forumIdempotency("thread-"+slug(title)),
	))
	assertForumStatus(t, response, fiber.StatusCreated)
	payload := decodeForumObject(t, response)
	return payload["thread"].(map[string]any), payload["post"].(map[string]any)
}
