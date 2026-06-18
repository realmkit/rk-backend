package forums_e2e

import (
	"context"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/e2e/harness"
	forumsdomain "github.com/realmkit/rk-backend/module/forums/domain"
	groupsdomain "github.com/realmkit/rk-backend/module/groups/domain"
	groupsport "github.com/realmkit/rk-backend/module/groups/port"
)

func TestForumAdminSearchAndOperations(t *testing.T) {
	fixture := newForumsFixture(t)
	steps := harness.NewSteps(t)
	forum := fixture.readyDiscussionForum(t, "admin_forum", fixture.member)
	thread, opener := fixture.createThread(t, forum, fixture.member, "Searchable Alpha")

	steps.Do("update settings and reject invalid settings", func() {
		forumID := forumID(t, forum, "id")
		response := fixture.do(t, forumRequest(
			fiber.MethodGet, "/forums/"+forumID.String()+"/settings", "",
			forumUser(fixture.manager),
		))
		assertForumStatus(t, response, fiber.StatusOK)
		settings := decodeForumObject(t, response)
		response = fixture.do(t, forumRequest(
			fiber.MethodPatch, "/forums/"+forumID.String()+"/settings",
			settingsBody(forumID, "own_threads", "open"),
			forumUser(fixture.manager), forumIdempotency("settings-ok"),
			forumIfMatch(forumVersion(settings)),
		))
		assertForumStatus(t, response, fiber.StatusOK)
		response = fixture.do(t, forumRequest(
			fiber.MethodPatch, "/forums/"+forumID.String()+"/settings",
			settingsBody(forumID, "all_threads", "unknown"),
			forumUser(fixture.manager), forumIdempotency("settings-invalid"),
			forumIfMatch(2),
		))
		assertForumStatus(t, response, fiber.StatusUnprocessableEntity)
	})

	steps.Do("replace permission grants and simulate group-based access", func() {
		groupID := fixture.createGroupWithMember(t, fixture.other)
		forumID := forumID(t, forum, "id")
		body := permissionBody(forumID, groupID)
		response := fixture.do(t, forumRequest(
			fiber.MethodPut, "/forums/"+forumID.String()+"/permissions", body,
			forumUser(fixture.manager), forumIdempotency("permission-settings"),
		))
		assertForumStatus(t, response, fiber.StatusNoContent)
		response = fixture.do(t, forumRequest(
			fiber.MethodPost, "/forums/"+forumID.String()+"/permissions/simulate",
			`{"actor_user_id":"`+fixture.other.String()+`","permission":"forums.create_thread"}`,
			forumUser(fixture.manager),
		))
		assertForumStatus(t, response, fiber.StatusOK)
		if allowed := decodeForumObject(t, response)["allowed"].(bool); !allowed {
			t.Fatalf("permission simulation denied group member")
		}
		response = fixture.do(t, forumRequest(
			fiber.MethodGet, "/forums/"+forumID.String()+"/permissions", "",
			forumUser(fixture.manager),
		))
		assertForumStatus(t, response, fiber.StatusOK)
	})

	steps.Do("search visible content and reject invalid queries", func() {
		path := "/forums/search?query=Alpha"
		response := fixture.do(t, forumRequest(fiber.MethodGet, path, "", forumUser(fixture.member)))
		assertForumStatus(t, response, fiber.StatusOK)
		if got := len(decodeForumList(t, response)); got == 0 {
			t.Fatalf("global search returned no results")
		}
		path = "/forums/" + forumID(t, forum, "id").String() + "/search?query=Alpha"
		response = fixture.do(t, forumRequest(fiber.MethodGet, path, "", forumUser(fixture.member)))
		assertForumStatus(t, response, fiber.StatusOK)
		response = fixture.do(t, forumRequest(fiber.MethodGet, "/forums/search?query=a", ""))
		assertForumStatus(t, response, fiber.StatusUnprocessableEntity)
	})

	steps.Do("verify, rebuild, flush views, and clear read caches", func() {
		ctx := context.Background()
		db := fixture.ecosystem.Database.Store.DB(ctx)
		currentForumID := forumID(t, forum, "id")
		postID := forumID(t, opener, "id")
		threadID := forumID(t, thread, "id")
		db.Exec("UPDATE forum_stats SET post_count = 99 WHERE forum_id = ?", currentForumID)
		report, err := fixture.service.VerifyStats(ctx)
		if err != nil || len(report.Mismatches) == 0 {
			t.Fatalf("VerifyStats() report=%+v err=%v, want drift", report, err)
		}
		if _, err := fixture.service.RebuildStats(ctx); err != nil {
			t.Fatalf("RebuildStats() error = %v", err)
		}
		db.Exec("UPDATE forum_posts SET like_count = 42 WHERE id = ?", postID)
		report, err = fixture.service.VerifyLikes(ctx)
		if err != nil || len(report.Mismatches) == 0 {
			t.Fatalf("VerifyLikes() report=%+v err=%v, want drift", report, err)
		}
		if _, err := fixture.service.RebuildLikes(ctx); err != nil {
			t.Fatalf("RebuildLikes() error = %v", err)
		}
		_ = fixture.do(t, forumRequest(
			fiber.MethodGet, "/threads/"+threadID.String(), "", forumUser(fixture.member),
		))
		flushed, err := fixture.service.FlushThreadViews(ctx)
		if err != nil || flushed == 0 {
			t.Fatalf("FlushThreadViews() = %d, %v; want flushed views", flushed, err)
		}
		if err := fixture.service.ClearReadCache(ctx); err != nil {
			t.Fatalf("ClearReadCache() error = %v", err)
		}
	})
}

func settingsBody(id uuid.UUID, mode string, status string) string {
	return `{"forum_id":"` + id.String() + `","kind":"discussion",` +
		`"thread_visibility_mode":"` + mode + `","max_sticky_threads":3,` +
		`"default_thread_status":"` + status + `",` +
		`"author_post_edit_window_seconds":600,"author_post_delete_window_seconds":300}`
}

func permissionBody(forumID uuid.UUID, groupID uuid.UUID) string {
	publicID := forumsdomain.PublicPermissionSubjectID().String()
	groupGrant := `{"subject_type":"group","subject_id":"` + groupID.String() + `"}`
	publicGrant := `{"subject_type":"public","subject_id":"` + publicID + `"}`
	return `{"forum_id":"` + forumID.String() + `","viewers":[` + publicGrant + `],` +
		`"creators":[` + groupGrant + `],"replyers":[],"likers":[],` +
		`"thread_pinners":[],"thread_managers":[],"post_managers":[],` +
		`"limit_bypassers":[],"all_thread_viewers":[],"administrators":[]}`
}

func (fixture forumsFixture) createGroupWithMember(t *testing.T, userID uuid.UUID) uuid.UUID {
	t.Helper()
	group, err := fixture.groups.Create(context.Background(), groupsport.CreateGroupCommand{
		Group: groupsdomain.Group{
			ID: uuid.New(), Key: groupsdomain.Key("forum_e2e_group_" + uuid.NewString()[:8]),
			Name: "Forum E2E Group", Color: "#3366ff", Weight: 10,
			Status: groupsdomain.GroupStatusActive,
		},
	})
	if err != nil {
		t.Fatalf("CreateGroup() error = %v", err)
	}
	_, err = fixture.groups.Assign(context.Background(), groupsport.AssignMembershipCommand{
		Membership: groupsdomain.Membership{
			ID: uuid.New(), GroupID: group.ID, UserID: userID,
			Status: groupsdomain.MembershipStatusActive,
		},
	})
	if err != nil {
		t.Fatalf("AssignMembership() error = %v", err)
	}
	return group.ID
}
