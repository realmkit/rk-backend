package content

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/google/uuid"
)

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
	violations = append(
		violations,
		ValidateReferenceType("reference_type", reference.ReferenceType)...,
	)
	if len(strings.TrimSpace(reference.QuoteExcerpt)) > 500 {
		violations = AppendViolation(violations, "quote_excerpt", "must be at most 500 characters")
	}
	violations = reference.validateTargets(violations)
	return NewValidationError(violations)
}

func (reference PostReference) validateTargets(violations []Violation) []Violation {
	if reference.ReferenceType == ReferenceAttachment && missingUUID(reference.TargetAssetID) {
		violations = AppendViolation(violations, "target_asset_id", "is required for attachments")
	}
	if reference.ReferenceType == ReferenceMention && missingUUID(reference.TargetUserID) {
		violations = AppendViolation(violations, "target_user_id", "is required for mentions")
	}
	if reference.requiresTargetPost() && missingUUID(reference.TargetPostID) {
		violations = AppendViolation(violations, "target_post_id", "is required for post references")
	}
	if reference.ReferenceType == ReferenceLink {
		violations = append(violations, ValidateExternalURL("link_url", reference.LinkURL)...)
	}
	return violations
}

func (reference PostReference) requiresTargetPost() bool {
	return reference.ReferenceType == ReferenceReplyTo ||
		reference.ReferenceType == ReferenceQuote
}

func missingUUID(id *uuid.UUID) bool {
	return id == nil || *id == uuid.Nil
}
