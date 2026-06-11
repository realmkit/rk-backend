package forums_e2e

import (
	"context"
	"strconv"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/e2e/harness"
	groupsdomain "github.com/realmkit/rk-backend/module/groups/domain"
)

func TestForumStructureLifecycleAndVisibleTree(t *testing.T) {
	fixture := newForumsFixture(t)
	steps := harness.NewSteps(t)
	var category, support, help, discord map[string]any

	steps.Do("create category and root forums", func() {
		category = fixture.createCategory(t, "community")
		catID := forumID(t, category, "id")
		support = fixture.createForum(t, forumWrite{
			CategoryID: catID, Key: "support", Slug: "support",
			Name: "Support", Kind: "container", Order: 2,
		})
		help = fixture.createForum(t, forumWrite{
			CategoryID: catID, ParentID: forumID(t, support, "id"),
			Key: "help_desk", Slug: "help-desk", Name: "Help Desk",
			Kind: "discussion", Order: 1,
		})
		discord = fixture.createForum(t, forumWrite{
			CategoryID: catID, ParentID: forumID(t, support, "id"),
			Key: "discord", Slug: "discord", Name: "Discord",
			Kind: "link", ExternalURL: "https://discord.example.test", Order: 2,
		})
	})

	steps.Do("tree hides forums until view permission is granted", func() {
		response := fixture.do(t, forumRequest(fiber.MethodGet, "/forums/tree", ""))
		assertForumStatus(t, response, fiber.StatusOK)
		if got := decodeForumObject(t, response)["categories"].([]any); len(got) != 0 {
			t.Fatalf("visible categories = %d, want none", len(got))
		}
		fixture.grant(t, forumID(t, support, "id"), groupsdomain.RelationViewer, fixture.member)
		fixture.grant(t, forumID(t, help, "id"), groupsdomain.RelationViewer, fixture.member)
		fixture.grant(t, forumID(t, discord, "id"), groupsdomain.RelationViewer, fixture.member)
		response = fixture.do(t, forumRequest(fiber.MethodGet, "/forums/tree", "", forumUser(fixture.member)))
		assertForumStatus(t, response, fiber.StatusOK)
		if got := decodeForumObject(t, response)["categories"].([]any); len(got) != 1 {
			t.Fatalf("visible categories = %d, want 1", len(got))
		}
	})

	steps.Do("update, move, reorder, list, and delete structure resources", func() {
		supportVersion := forumVersion(support)
		response := fixture.do(t, forumRequest(
			fiber.MethodPatch, "/forums/"+forumID(t, support, "id").String(),
			`{"category_id":"`+forumID(t, category, "id").String()+`","kind":"container","key":"support",`+
				`"slug":"support","name":"Support Desk","display_order":2,"status":"active"}`,
			forumUser(fixture.manager), forumIdempotency("update-support"), forumIfMatch(supportVersion),
		))
		assertForumStatus(t, response, fiber.StatusOK)
		support = decodeForumObject(t, response)

		response = fixture.do(t, forumRequest(
			fiber.MethodPost, "/forums/"+forumID(t, discord, "id").String()+"/move",
			`{"category_id":"`+forumID(t, category, "id").String()+`","display_order":3}`,
			forumUser(fixture.manager), forumIdempotency("move-discord"), forumIfMatch(forumVersion(discord)),
		))
		assertForumStatus(t, response, fiber.StatusOK)
		discord = decodeForumObject(t, response)

		response = fixture.do(t, forumRequest(
			fiber.MethodPost, "/forums/reorder",
			`{"items":[{"id":"`+forumID(t, help, "id").String()+`","display_order":1},`+
				`{"id":"`+forumID(t, discord, "id").String()+`","display_order":3}]}`,
			forumUser(fixture.manager), forumIdempotency("reorder-forums"),
		))
		assertForumStatus(t, response, fiber.StatusNoContent)

		response = fixture.do(t, forumRequest(fiber.MethodGet, "/forums", ""))
		assertForumStatus(t, response, fiber.StatusOK)
		if got := len(decodeForumList(t, response)); got < 3 {
			t.Fatalf("forums listed = %d, want at least 3", got)
		}

		response = fixture.do(t, forumRequest(
			fiber.MethodDelete, "/forums/"+forumID(t, discord, "id").String(), "",
			forumUser(fixture.manager), forumIdempotency("delete-discord"), forumIfMatch(forumVersion(discord)),
		))
		assertForumStatus(t, response, fiber.StatusNoContent)
	})

	steps.Do("reject stale versions and missing precondition headers", func() {
		response := fixture.do(t, forumRequest(
			fiber.MethodPatch, "/forum-categories/"+forumID(t, category, "id").String(),
			categoryPatchBody(category, "Community Hub"),
			forumUser(fixture.manager), forumIdempotency("missing-if-match"),
		))
		assertForumStatus(t, response, fiber.StatusPreconditionRequired)
		response = fixture.do(t, forumRequest(
			fiber.MethodDelete, "/forum-categories/"+forumID(t, category, "id").String(), "",
			forumUser(fixture.manager), forumIdempotency("stale-category"), forumIfMatch(99),
		))
		assertForumStatus(t, response, fiber.StatusPreconditionFailed)
	})
}

func (fixture forumsFixture) createCategory(t *testing.T, key string) map[string]any {
	t.Helper()
	body := `{"key":"` + key + `","name":"` + key + `","display_order":1,"status":"active"}`
	response := fixture.do(t, forumRequest(
		fiber.MethodPost, "/forum-categories", body,
		forumUser(fixture.manager), forumIdempotency("category-"+key),
	))
	assertForumStatus(t, response, fiber.StatusCreated)
	return decodeForumObject(t, response)
}

type forumWrite struct {
	CategoryID    uuid.UUID
	ParentID      uuid.UUID
	Key           string
	Slug          string
	Name          string
	Kind          string
	ExternalURL   string
	DefaultStatus string
	Order         int
}

func (fixture forumsFixture) createForum(t *testing.T, write forumWrite) map[string]any {
	t.Helper()
	parent := "null"
	if write.ParentID != uuid.Nil {
		parent = `"` + write.ParentID.String() + `"`
	}
	status := write.DefaultStatus
	if status == "" {
		status = "open"
	}
	body := `{"category_id":"` + write.CategoryID.String() + `","parent_forum_id":` + parent +
		`,"kind":"` + write.Kind + `","key":"` + write.Key + `","slug":"` + write.Slug +
		`","name":"` + write.Name + `","display_order":` + strconv.Itoa(write.Order) +
		`,"external_url":"` + write.ExternalURL + `","thread_visibility_mode":"all_threads",` +
		`"max_sticky_threads":5,"default_thread_status":"` + status + `","status":"active"}`
	response := fixture.do(t, forumRequest(
		fiber.MethodPost, "/forums", body,
		forumUser(fixture.manager), forumIdempotency("forum-"+write.Key),
	))
	assertForumStatus(t, response, fiber.StatusCreated)
	forum := decodeForumObject(t, response)
	fixture.grant(t, forumID(t, forum, "id"), groupsdomain.RelationManager, fixture.manager)
	if _, err := fixture.service.GetForumSettings(
		context.Background(),
		fixture.manager,
		forumID(t, forum, "id"),
	); err != nil {
		t.Fatalf("manager grant for created forum was not effective: %v", err)
	}
	return forum
}

func categoryPatchBody(category map[string]any, name string) string {
	return `{"key":"` + category["key"].(string) + `","name":"` + name +
		`","display_order":1,"status":"active"}`
}
