package interaction

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/module/forums/domain"
)

// TestLikeAndReadModelsMapDomainFields verifies persistence mapping helpers.
func TestLikeAndReadModelsMapDomainFields(t *testing.T) {
	now := time.Now().UTC()
	like := domain.PostLike{ID: uuid.New(), PostID: uuid.New(), ThreadID: uuid.New(), ForumID: uuid.New(), UserID: uuid.New(), CreatedAt: now}
	likeModel := likeModelFromDomain(like)
	if likeModel.ID != like.ID || likeModel.PostID != like.PostID || !likeModel.CreatedAt.Equal(now) {
		t.Fatalf("likeModelFromDomain() = %#v", likeModel)
	}
	state := domain.ThreadReadState{ID: uuid.New(), UserID: uuid.New(), ForumID: uuid.New(), ThreadID: uuid.New(), LastReadPostSequence: 9, LastReadAt: now}
	readModel := readStateModelFromDomain(state)
	if readModel.ID != state.ID || readModel.LastReadPostSequence != 9 || !readModel.LastReadAt.Equal(now) {
		t.Fatalf("readStateModelFromDomain() = %#v", readModel)
	}
}

// TestWidgetPagesTrimResultsAndSetCursor verifies widget pagination helpers.
func TestWidgetPagesTrimResultsAndSetCursor(t *testing.T) {
	first := latestPostRow{PostID: uuid.New(), ThreadSlug: "first"}
	second := latestPostRow{PostID: uuid.New(), ThreadSlug: "second"}
	page := latestPostPage([]latestPostRow{first, second}, 1)
	if len(page.Items) != 1 || page.NextCursor != first.PostID.String() {
		t.Fatalf("latestPostPage() = %#v", page)
	}
	liked := mostLikedPostPage([]mostLikedPostRow{{latestPostRow: first, LikeCount: 4}}, 10)
	if len(liked.Items) != 1 || liked.Items[0].LikeCount != 4 {
		t.Fatalf("mostLikedPostPage() = %#v", liked)
	}
}

// TestVisibleStatusCatalogs verifies widget-visible status sets.
func TestVisibleStatusCatalogs(t *testing.T) {
	if len(visiblePostStatuses()) != 2 || len(visibleThreadStatuses()) != 3 {
		t.Fatalf("visible status catalogs are incomplete")
	}
}
