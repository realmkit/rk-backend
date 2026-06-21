package interaction

import (
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/module/forums/domain"
	"github.com/realmkit/rk-backend/module/forums/port"
	"github.com/realmkit/rk-backend/pkg/pagination"
)

// TestNewPostLikeCopiesPostScope verifies likes retain post, thread, and forum scope.
func TestNewPostLikeCopiesPostScope(t *testing.T) {
	post := domain.Post{ID: uuid.New(), ThreadID: uuid.New(), ForumID: uuid.New()}
	actorID := uuid.New()
	like := newPostLike(post, actorID)
	if like.PostID != post.ID || like.ThreadID != post.ThreadID || like.ForumID != post.ForumID || like.UserID != actorID {
		t.Fatalf("newPostLike() = %#v", like)
	}
}

// TestThreadReadStateDefaultsToVisiblePostCount verifies read sequence fallback.
func TestThreadReadStateDefaultsToVisiblePostCount(t *testing.T) {
	thread := domain.Thread{ID: uuid.New(), ForumID: uuid.New(), VisiblePostCount: 7}
	state := threadReadState(port.MarkThreadReadCommand{ActorUserID: uuid.New()}, thread)
	if state.LastReadPostSequence != 7 || state.ThreadID != thread.ID || state.ForumID != thread.ForumID {
		t.Fatalf("threadReadState() = %#v", state)
	}
}

// TestWidgetCacheKeysIncludeScope verifies cache key partitioning.
func TestWidgetCacheKeysIncludeScope(t *testing.T) {
	actorID := uuid.New()
	forumID := uuid.New()
	page := pagination.Page{Limit: 25, Cursor: "cursor"}
	latest := latestPostsCacheKey(actorID, forumID, page)
	if !strings.Contains(latest, "forum:"+forumID.String()) || !strings.Contains(latest, "user:"+actorID.String()) {
		t.Fatalf("latestPostsCacheKey() = %q", latest)
	}
	if got := actorScope(uuid.Nil); got != "anonymous" {
		t.Fatalf("actorScope(nil) = %q, want anonymous", got)
	}
}
