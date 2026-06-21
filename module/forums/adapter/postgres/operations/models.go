package operations

import (
	"time"

	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/module/forums/domain"
)

// threadSearchRow is a compact thread search row.
type threadSearchRow struct {
	ID           uuid.UUID // ID stores the i d value.
	ForumID      uuid.UUID // ForumID stores the forum i d value.
	AuthorUserID uuid.UUID // AuthorUserID stores the author user i d value.
	Title        string    // Title stores the title value.
	Slug         string    // Slug stores the slug value.
	CreatedAt    time.Time // CreatedAt stores the created at value.
}

// postSearchRow is a compact post search row.
type postSearchRow struct {
	PostID       uuid.UUID // PostID stores the post i d value.
	ThreadID     uuid.UUID // ThreadID stores the thread i d value.
	ForumID      uuid.UUID // ForumID stores the forum i d value.
	AuthorUserID uuid.UUID // AuthorUserID stores the author user i d value.
	Title        string    // Title stores the title value.
	Slug         string    // Slug stores the slug value.
	Excerpt      string    // Excerpt stores the excerpt value.
	CreatedAt    time.Time // CreatedAt stores the created at value.
}

// threadIDRow is a compact active thread row.
type threadIDRow struct {
	ID uuid.UUID // ID stores the i d value.
}

// forumIDRow is a compact forum stats identity row.
type forumIDRow struct {
	ForumID uuid.UUID // ForumID stores the forum i d value.
}

// threadCounterRow is a stored thread counter row.
type threadCounterRow struct {
	ID                uuid.UUID // ID stores the i d value.
	PostCount         int64     // PostCount stores the post count value.
	VisiblePostCount  int64     // VisiblePostCount stores the visible post count value.
	ReplyCount        int64     // ReplyCount stores the reply count value.
	VisibleReplyCount int64     // VisibleReplyCount stores the visible reply count value.
	LikeCount         int64     // LikeCount stores the like count value.
}

// postCounterRow is a stored post counter row.
type postCounterRow struct {
	ID        uuid.UUID // ID stores the i d value.
	LikeCount int64     // LikeCount stores the like count value.
}

// forumCounterRow is a stored forum counter row.
type forumCounterRow struct {
	ForumID            uuid.UUID // ForumID stores the forum i d value.
	ThreadCount        int64     // ThreadCount stores the thread count value.
	VisibleThreadCount int64     // VisibleThreadCount stores the visible thread count value.
	PostCount          int64     // PostCount stores the post count value.
	VisiblePostCount   int64     // VisiblePostCount stores the visible post count value.
}

// threadExpectation contains source-of-truth thread counters.
type threadExpectation struct {
	PostCount         int64 // PostCount stores the post count value.
	VisiblePostCount  int64 // VisiblePostCount stores the visible post count value.
	ReplyCount        int64 // ReplyCount stores the reply count value.
	VisibleReplyCount int64 // VisibleReplyCount stores the visible reply count value.
}

// forumExpectation contains source-of-truth forum counters.
type forumExpectation struct {
	ThreadCount        int64 // ThreadCount stores the thread count value.
	VisibleThreadCount int64 // VisibleThreadCount stores the visible thread count value.
	PostCount          int64 // PostCount stores the post count value.
	VisiblePostCount   int64 // VisiblePostCount stores the visible post count value.
}

// threadPostCounterRow is a grouped thread post count.
type threadPostCounterRow struct {
	ThreadID         uuid.UUID // ThreadID stores the thread i d value.
	PostCount        int64     // PostCount stores the post count value.
	VisiblePostCount int64     // VisiblePostCount stores the visible post count value.
}

// forumThreadCounterRow is a grouped forum thread count.
type forumThreadCounterRow struct {
	ForumID            uuid.UUID // ForumID stores the forum i d value.
	ThreadCount        int64     // ThreadCount stores the thread count value.
	VisibleThreadCount int64     // VisibleThreadCount stores the visible thread count value.
}

// forumPostCounterRow is a grouped forum post count.
type forumPostCounterRow struct {
	ForumID          uuid.UUID // ForumID stores the forum i d value.
	PostCount        int64     // PostCount stores the post count value.
	VisiblePostCount int64     // VisiblePostCount stores the visible post count value.
}

// likeCounterRow is a grouped like count.
type likeCounterRow struct {
	ID    uuid.UUID // ID stores the i d value.
	Count int64     // Count stores the count value.
}

// searchResult maps a thread row to a domain result.
func (row threadSearchRow) searchResult() domain.SearchResult {
	return domain.SearchResult{
		Type:         "thread",
		ForumID:      row.ForumID,
		ThreadID:     row.ID,
		Title:        row.Title,
		Slug:         domain.Slug(row.Slug),
		Excerpt:      row.Title,
		AuthorUserID: row.AuthorUserID,
		CreatedAt:    row.CreatedAt,
	}
}

// searchResult maps a post row to a domain result.
func (row postSearchRow) searchResult() domain.SearchResult {
	postID := row.PostID
	return domain.SearchResult{
		Type:         "post",
		ForumID:      row.ForumID,
		ThreadID:     row.ThreadID,
		PostID:       &postID,
		Title:        row.Title,
		Slug:         domain.Slug(row.Slug),
		Excerpt:      row.Excerpt,
		AuthorUserID: row.AuthorUserID,
		CreatedAt:    row.CreatedAt,
	}
}

// visiblePostStatuses returns statuses normal readers may see.
func visiblePostStatuses() []domain.PostStatus {
	return []domain.PostStatus{domain.PostStatusVisible, domain.PostStatusSystem}
}

// visibleThreadStatuses returns statuses normal readers may see.
func visibleThreadStatuses() []domain.ThreadStatus {
	return []domain.ThreadStatus{
		domain.ThreadStatusOpen,
		domain.ThreadStatusClosed,
		domain.ThreadStatusLocked,
	}
}

// Query snippets used by repair/search projections.
const (
	searchPostSelect = "p.id AS post_id, p.thread_id, p.forum_id, p.author_user_id, " +
		"p.content_text AS excerpt, p.created_at, t.title, t.slug"
	threadPostCounterSelect = "thread_id, COUNT(*) AS post_count, " +
		"SUM(CASE WHEN status IN ? THEN 1 ELSE 0 END) AS visible_post_count"
	forumThreadCounterSelect = "forum_id, COUNT(*) AS thread_count, " +
		"SUM(CASE WHEN status IN ? THEN 1 ELSE 0 END) AS visible_thread_count"
	forumPostCounterSelect = "forum_id, COUNT(*) AS post_count, " +
		"SUM(CASE WHEN status IN ? THEN 1 ELSE 0 END) AS visible_post_count"
)
