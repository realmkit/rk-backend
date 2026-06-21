package operations

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/pkg/search"
)

// TestSearchRowsMapDomainResults verifies search projection mapping.
func TestSearchRowsMapDomainResults(t *testing.T) {
	now := time.Now().UTC()
	thread := threadSearchRow{ID: uuid.New(), ForumID: uuid.New(), AuthorUserID: uuid.New(), Title: "Thread", Slug: "thread", CreatedAt: now}
	threadResult := thread.searchResult()
	if threadResult.Type != "thread" || threadResult.ThreadID != thread.ID || threadResult.Excerpt != "Thread" {
		t.Fatalf("thread searchResult() = %#v", threadResult)
	}
	post := postSearchRow{PostID: uuid.New(), ThreadID: uuid.New(), ForumID: uuid.New(), Title: "Thread", Slug: "thread", Excerpt: "Post", CreatedAt: now}
	postResult := post.searchResult()
	if postResult.Type != "post" || postResult.PostID == nil || *postResult.PostID != post.PostID {
		t.Fatalf("post searchResult() = %#v", postResult)
	}
}

// TestSearchPageSortsAndCursors verifies merged search pagination.
func TestSearchPageSortsAndCursors(t *testing.T) {
	now := time.Now().UTC()
	older := threadSearchRow{ID: uuid.New(), CreatedAt: now.Add(-time.Minute)}
	newer := postSearchRow{PostID: uuid.New(), ThreadID: uuid.New(), CreatedAt: now}
	page, err := searchPage([]threadSearchRow{older}, []postSearchRow{newer}, 1, "filter", search.Sort{Key: "created_at", Direction: search.DirectionDesc})
	if err != nil {
		t.Fatalf("searchPage() error = %v", err)
	}
	if len(page.Items) != 1 || page.Items[0].Type != "post" || page.NextCursor == "" {
		t.Fatalf("searchPage() = %#v", page)
	}
}

// TestVisibleStatusCatalogs verifies search-visible status sets.
func TestVisibleStatusCatalogs(t *testing.T) {
	if len(visiblePostStatuses()) != 2 || len(visibleThreadStatuses()) != 3 {
		t.Fatalf("visible status catalogs are incomplete")
	}
}
