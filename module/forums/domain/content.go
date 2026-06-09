// Package domain contains forum entities, value objects, and validation rules.
package domain

import (
	"encoding/json"
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
	if thread.ReplyCount < 0 || thread.VisibleReplyCount < 0 || thread.PostCount < 0 || thread.VisiblePostCount < 0 || thread.LikeCount < 0 || thread.ViewCount < 0 {
		violations = AppendViolation(violations, "counts", "must be zero or greater")
	}
	return NewValidationError(violations)
}

// Visible reports whether normal readers can see the thread.
func (thread Thread) Visible() bool {
	return thread.Status == ThreadStatusOpen || thread.Status == ThreadStatusClosed || thread.Status == ThreadStatusLocked
}

// Replyable reports whether users can add replies.
func (thread Thread) Replyable() bool {
	return thread.Status == ThreadStatusOpen
}

// Post is one message in a thread timeline.
type Post struct {
	// ID is the post identifier.
	ID uuid.UUID `json:"id"`

	// ThreadID is the containing thread.
	ThreadID uuid.UUID `json:"thread_id"`

	// ForumID is the containing forum.
	ForumID uuid.UUID `json:"forum_id"`

	// AuthorUserID is the post author.
	AuthorUserID uuid.UUID `json:"author_user_id"`

	// Sequence is the stable sequence inside the thread.
	Sequence int64 `json:"sequence"`

	// Status is the post lifecycle state.
	Status PostStatus `json:"status"`

	// ContentFormat identifies the document schema.
	ContentFormat ContentFormat `json:"content_format"`

	// ContentDocumentJSON is the canonical rich-content document.
	ContentDocumentJSON json.RawMessage `json:"content_document_json"`

	// ContentText is extracted plain text.
	ContentText string `json:"content_text"`

	// ContentChecksum identifies duplicate content.
	ContentChecksum string `json:"content_checksum"`

	// EditedAt is set after edits.
	EditedAt *time.Time `json:"edited_at,omitempty"`

	// EditedByUserID is the last editor.
	EditedByUserID *uuid.UUID `json:"edited_by_user_id,omitempty"`

	// EditCount is the number of edits.
	EditCount int64 `json:"edit_count"`

	// LikeCount is the post like count.
	LikeCount int64 `json:"like_count"`

	// ReplyReferenceCount is the number of reply/quote references.
	ReplyReferenceCount int64 `json:"reply_reference_count"`

	// Version is the optimistic concurrency version.
	Version uint64 `json:"version"`

	// CreatedAt is the creation timestamp.
	CreatedAt time.Time `json:"created_at"`

	// UpdatedAt is the update timestamp.
	UpdatedAt time.Time `json:"updated_at"`
}

// Normalize returns a normalized post copy.
func (post Post) Normalize() Post {
	post.ContentText = strings.TrimSpace(post.ContentText)
	post.ContentChecksum = strings.TrimSpace(post.ContentChecksum)
	if post.Status == "" {
		post.Status = PostStatusVisible
	}
	if post.ContentFormat == "" {
		post.ContentFormat = ContentFormatProseMirror
	}
	if post.Version == 0 {
		post.Version = 1
	}
	return post
}

// Validate validates post fields.
func (post Post) Validate() error {
	var violations []Violation
	if post.ID == uuid.Nil {
		violations = AppendViolation(violations, "id", "is required")
	}
	if post.ThreadID == uuid.Nil {
		violations = AppendViolation(violations, "thread_id", "is required")
	}
	if post.ForumID == uuid.Nil {
		violations = AppendViolation(violations, "forum_id", "is required")
	}
	if post.AuthorUserID == uuid.Nil {
		violations = AppendViolation(violations, "author_user_id", "is required")
	}
	if post.Sequence < 1 {
		violations = AppendViolation(violations, "sequence", "must be one or greater")
	}
	violations = append(violations, ValidatePostStatus("status", post.Status)...)
	violations = append(violations, ValidateContentFormat("content_format", post.ContentFormat)...)
	violations = append(violations, ValidateContentDocument("content_document_json", post.ContentDocumentJSON)...)
	violations = append(violations, ValidateContentText("content_text", post.ContentText)...)
	if len(post.ContentChecksum) > 128 {
		violations = AppendViolation(violations, "content_checksum", "must be at most 128 characters")
	}
	if post.EditCount < 0 || post.LikeCount < 0 || post.ReplyReferenceCount < 0 {
		violations = AppendViolation(violations, "counts", "must be zero or greater")
	}
	return NewValidationError(violations)
}

// Visible reports whether normal readers can see the post.
func (post Post) Visible() bool {
	return post.Status == PostStatusVisible || post.Status == PostStatusSystem
}

// PostRevision stores content before an edit.
type PostRevision struct {
	// ID is the revision identifier.
	ID uuid.UUID `json:"id"`

	// PostID is the edited post.
	PostID uuid.UUID `json:"post_id"`

	// EditedByUserID is the editor.
	EditedByUserID uuid.UUID `json:"edited_by_user_id"`

	// PreviousContentDocumentJSON is the prior document.
	PreviousContentDocumentJSON json.RawMessage `json:"previous_content_document_json"`

	// PreviousContentText is the prior extracted text.
	PreviousContentText string `json:"previous_content_text"`

	// EditReason explains the edit.
	EditReason string `json:"edit_reason"`

	// CreatedAt is the revision timestamp.
	CreatedAt time.Time `json:"created_at"`
}

// PostReference stores structured relationships extracted from a post.
type PostReference struct {
	// ID is the reference identifier.
	ID uuid.UUID `json:"id"`

	// SourcePostID is the post containing the reference.
	SourcePostID uuid.UUID `json:"source_post_id"`

	// TargetPostID is the referenced post when applicable.
	TargetPostID *uuid.UUID `json:"target_post_id,omitempty"`

	// TargetUserID is the referenced user when applicable.
	TargetUserID *uuid.UUID `json:"target_user_id,omitempty"`

	// TargetAssetID is the referenced asset when applicable.
	TargetAssetID *uuid.UUID `json:"target_asset_id,omitempty"`

	// ReferenceType identifies the relationship kind.
	ReferenceType ReferenceType `json:"reference_type"`

	// QuoteExcerpt preserves stable quote rendering.
	QuoteExcerpt string `json:"quote_excerpt"`

	// LinkURL stores extracted links.
	LinkURL string `json:"link_url"`

	// CreatedAt is the creation timestamp.
	CreatedAt time.Time `json:"created_at"`
}

// Validate validates post reference fields.
func (reference PostReference) Validate() error {
	var violations []Violation
	if reference.SourcePostID == uuid.Nil {
		violations = AppendViolation(violations, "source_post_id", "is required")
	}
	violations = append(violations, ValidateReferenceType("reference_type", reference.ReferenceType)...)
	if len(strings.TrimSpace(reference.QuoteExcerpt)) > 500 {
		violations = AppendViolation(violations, "quote_excerpt", "must be at most 500 characters")
	}
	if reference.ReferenceType == ReferenceAttachment && (reference.TargetAssetID == nil || *reference.TargetAssetID == uuid.Nil) {
		violations = AppendViolation(violations, "target_asset_id", "is required for attachments")
	}
	if reference.ReferenceType == ReferenceMention && (reference.TargetUserID == nil || *reference.TargetUserID == uuid.Nil) {
		violations = AppendViolation(violations, "target_user_id", "is required for mentions")
	}
	if (reference.ReferenceType == ReferenceReplyTo || reference.ReferenceType == ReferenceQuote) && (reference.TargetPostID == nil || *reference.TargetPostID == uuid.Nil) {
		violations = AppendViolation(violations, "target_post_id", "is required for post references")
	}
	if reference.ReferenceType == ReferenceLink {
		violations = append(violations, ValidateExternalURL("link_url", reference.LinkURL)...)
	}
	return NewValidationError(violations)
}
