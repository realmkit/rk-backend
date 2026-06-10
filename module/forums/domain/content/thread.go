package content

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

// Thread is a forum conversation timeline.
type Thread struct {
	// ID is the thread identifier.
	ID uuid.UUID `json:"id"`

	// ForumID is the containing discussion forum.
	ForumID uuid.UUID `json:"forum_id"`

	// AuthorUserID is the user that opened the thread.
	AuthorUserID uuid.UUID `json:"author_user_id"`

	// OpenerPostID is the first post in the thread.
	OpenerPostID uuid.UUID `json:"opener_post_id"`

	// LatestPostID is the latest visible activity post.
	LatestPostID uuid.UUID `json:"latest_post_id"`

	// LatestPostAuthorUserID is the latest activity author.
	LatestPostAuthorUserID uuid.UUID `json:"latest_post_author_user_id"`

	// LatestPostAt is the latest activity timestamp.
	LatestPostAt time.Time `json:"latest_post_at"`

	// Title is the thread title.
	Title string `json:"title"`

	// Slug is the URL slug.
	Slug Slug `json:"slug"`

	// Status is the thread lifecycle state.
	Status ThreadStatus `json:"status"`

	// StickyState is the pinned display state.
	StickyState StickyState `json:"sticky_state"`

	// StickyOrder controls ordering among sticky threads.
	StickyOrder int `json:"sticky_order"`

	// StickyUntil optionally expires sticky state.
	StickyUntil *time.Time `json:"sticky_until,omitempty"`

	// LockedReason explains locked state.
	LockedReason string `json:"locked_reason"`

	// ReplyCount excludes the opener post.
	ReplyCount int64 `json:"reply_count"`

	// VisibleReplyCount excludes hidden replies.
	VisibleReplyCount int64 `json:"visible_reply_count"`

	// PostCount includes the opener post.
	PostCount int64 `json:"post_count"`

	// VisiblePostCount includes visible opener and replies.
	VisiblePostCount int64 `json:"visible_post_count"`

	// LikeCount is the aggregate post like count.
	LikeCount int64 `json:"like_count"`

	// ViewCount is the persisted view count.
	ViewCount int64 `json:"view_count"`

	// Version is the optimistic concurrency version.
	Version uint64 `json:"version"`

	// CreatedAt is the creation timestamp.
	CreatedAt time.Time `json:"created_at"`

	// UpdatedAt is the update timestamp.
	UpdatedAt time.Time `json:"updated_at"`
}

// Normalize returns a normalized thread copy.
func (thread Thread) Normalize() Thread {
	thread.Title = strings.TrimSpace(thread.Title)
	thread.Slug = Slug(strings.TrimSpace(string(thread.Slug)))
	thread.LockedReason = strings.TrimSpace(thread.LockedReason)
	if thread.Status == "" {
		thread.Status = ThreadStatusOpen
	}
	if thread.StickyState == "" {
		thread.StickyState = StickyStateNormal
	}
	if thread.Version == 0 {
		thread.Version = 1
	}
	return thread
}

// Validate validates thread fields.
func (thread Thread) Validate() error {
	var violations []Violation
	if thread.ID == uuid.Nil {
		violations = AppendViolation(violations, "id", "is required")
	}
	if thread.ForumID == uuid.Nil {
		violations = AppendViolation(violations, "forum_id", "is required")
	}
	if thread.AuthorUserID == uuid.Nil {
		violations = AppendViolation(violations, "author_user_id", "is required")
	}
	violations = append(violations, ValidateTitle("title", thread.Title)...)
	violations = append(violations, ValidateSlug("slug", thread.Slug)...)
	violations = append(violations, ValidateThreadStatus("status", thread.Status)...)
	violations = append(violations, ValidateStickyState("sticky_state", thread.StickyState)...)
	if thread.StickyOrder < 0 {
		violations = AppendViolation(violations, "sticky_order", "must be zero or greater")
	}
	if len(thread.LockedReason) > 500 {
		violations = AppendViolation(violations, "locked_reason", "must be at most 500 characters")
	}
	if thread.hasNegativeCounts() {
		violations = AppendViolation(violations, "counts", "must be zero or greater")
	}
	return NewValidationError(violations)
}

// Visible reports whether normal readers can see the thread.
func (thread Thread) Visible() bool {
	return thread.Status == ThreadStatusOpen ||
		thread.Status == ThreadStatusClosed ||
		thread.Status == ThreadStatusLocked
}

// Replyable reports whether users can add replies.
func (thread Thread) Replyable() bool {
	return thread.Status == ThreadStatusOpen
}

func (thread Thread) hasNegativeCounts() bool {
	return thread.ReplyCount < 0 ||
		thread.VisibleReplyCount < 0 ||
		thread.PostCount < 0 ||
		thread.VisiblePostCount < 0 ||
		thread.LikeCount < 0 ||
		thread.ViewCount < 0
}
