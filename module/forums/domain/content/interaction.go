package content

import (
	"time"

	"github.com/google/uuid"
)

// PostLike records one active user like for one post.
type PostLike struct {
	// ID is the like identifier.
	ID uuid.UUID `json:"id"`

	// PostID is the liked post.
	PostID uuid.UUID `json:"post_id"`

	// ThreadID is the containing thread.
	ThreadID uuid.UUID `json:"thread_id"`

	// ForumID is the containing forum.
	ForumID uuid.UUID `json:"forum_id"`

	// UserID is the liking user.
	UserID uuid.UUID `json:"user_id"`

	// CreatedAt is the like creation timestamp.
	CreatedAt time.Time `json:"created_at"`
}

// Validate validates like identity fields.
func (like PostLike) Validate() error {
	var violations []Violation
	if like.ID == uuid.Nil {
		violations = AppendViolation(violations, "id", "is required")
	}
	if like.PostID == uuid.Nil {
		violations = AppendViolation(violations, "post_id", "is required")
	}
	if like.ThreadID == uuid.Nil {
		violations = AppendViolation(violations, "thread_id", "is required")
	}
	if like.ForumID == uuid.Nil {
		violations = AppendViolation(violations, "forum_id", "is required")
	}
	if like.UserID == uuid.Nil {
		violations = AppendViolation(violations, "user_id", "is required")
	}
	return NewValidationError(violations)
}

// PostLikeSummary describes the current user's like state for a post.
type PostLikeSummary struct {
	// PostID is the summarized post.
	PostID uuid.UUID `json:"post_id"`

	// LikeCount is the current like count.
	LikeCount int64 `json:"like_count"`

	// LikedByActor reports whether the current actor likes the post.
	LikedByActor bool `json:"liked_by_actor"`
}

// ThreadReadState stores how far a user has read in a thread.
type ThreadReadState struct {
	// ID is the read-state identifier.
	ID uuid.UUID `json:"id"`

	// UserID is the owning user.
	UserID uuid.UUID `json:"user_id"`

	// ForumID is the containing forum.
	ForumID uuid.UUID `json:"forum_id"`

	// ThreadID is the read thread.
	ThreadID uuid.UUID `json:"thread_id"`

	// LastReadPostSequence is the highest read post sequence.
	LastReadPostSequence int64 `json:"last_read_post_sequence"`

	// LastReadAt is when the read state was updated.
	LastReadAt time.Time `json:"last_read_at"`

	// CreatedAt is the creation timestamp.
	CreatedAt time.Time `json:"created_at"`

	// UpdatedAt is the update timestamp.
	UpdatedAt time.Time `json:"updated_at"`
}

// Validate validates read-state identity and sequence fields.
func (state ThreadReadState) Validate() error {
	var violations []Violation
	if state.ID == uuid.Nil {
		violations = AppendViolation(violations, "id", "is required")
	}
	if state.UserID == uuid.Nil {
		violations = AppendViolation(violations, "user_id", "is required")
	}
	if state.ForumID == uuid.Nil {
		violations = AppendViolation(violations, "forum_id", "is required")
	}
	if state.ThreadID == uuid.Nil {
		violations = AppendViolation(violations, "thread_id", "is required")
	}
	if state.LastReadPostSequence < 1 {
		violations = AppendViolation(violations, "last_read_post_sequence", "must be one or greater")
	}
	if state.LastReadAt.IsZero() {
		violations = AppendViolation(violations, "last_read_at", "is required")
	}
	return NewValidationError(violations)
}

// LatestPostSummary is a compact latest-post widget row.
type LatestPostSummary struct {
	// ForumID is the containing forum.
	ForumID uuid.UUID `json:"forum_id"`

	// ThreadID is the containing thread.
	ThreadID uuid.UUID `json:"thread_id"`

	// PostID is the latest post.
	PostID uuid.UUID `json:"post_id"`

	// AuthorUserID is the post author.
	AuthorUserID uuid.UUID `json:"author_user_id"`

	// Sequence is the post sequence in the thread.
	Sequence int64 `json:"sequence"`

	// ThreadTitle is the containing thread title.
	ThreadTitle string `json:"thread_title"`

	// ThreadSlug is the containing thread slug.
	ThreadSlug Slug `json:"thread_slug"`

	// Excerpt is extracted plain text for previews.
	Excerpt string `json:"excerpt"`

	// CreatedAt is the post creation time.
	CreatedAt time.Time `json:"created_at"`
}

// MostLikedPost is a compact most-liked widget row.
type MostLikedPost struct {
	// ForumID is the containing forum.
	ForumID uuid.UUID `json:"forum_id"`

	// ThreadID is the containing thread.
	ThreadID uuid.UUID `json:"thread_id"`

	// PostID is the ranked post.
	PostID uuid.UUID `json:"post_id"`

	// AuthorUserID is the post author.
	AuthorUserID uuid.UUID `json:"author_user_id"`

	// Sequence is the post sequence in the thread.
	Sequence int64 `json:"sequence"`

	// ThreadTitle is the containing thread title.
	ThreadTitle string `json:"thread_title"`

	// ThreadSlug is the containing thread slug.
	ThreadSlug Slug `json:"thread_slug"`

	// Excerpt is extracted plain text for previews.
	Excerpt string `json:"excerpt"`

	// LikeCount is the post like count.
	LikeCount int64 `json:"like_count"`

	// CreatedAt is the post creation time.
	CreatedAt time.Time `json:"created_at"`
}

// UnreadSummary describes unread thread state for visible forums.
type UnreadSummary struct {
	// UserID is the summarized user.
	UserID uuid.UUID `json:"user_id"`

	// UnreadThreadCount is the total unread visible thread count.
	UnreadThreadCount int64 `json:"unread_thread_count"`

	// Forums contains per-forum unread totals.
	Forums []ForumUnreadSummary `json:"forums"`
}

// ForumUnreadSummary describes unread state for one forum.
type ForumUnreadSummary struct {
	// ForumID is the forum identifier.
	ForumID uuid.UUID `json:"forum_id"`

	// UnreadThreadCount is the unread thread count in the forum.
	UnreadThreadCount int64 `json:"unread_thread_count"`
}
