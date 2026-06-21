package structure

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

// Forum is a discussion board, container, or utility link.
type Forum struct {
	// ID is the forum identifier.
	ID uuid.UUID `json:"id"`

	// CategoryID is the parent category.
	CategoryID uuid.UUID `json:"category_id"`

	// ParentForumID is the optional parent forum.
	ParentForumID *uuid.UUID `json:"parent_forum_id,omitempty"`

	// Kind is the forum structural kind.
	Kind ForumKind `json:"kind"`

	// Key is the stable forum key.
	Key Key `json:"key"`

	// Slug is the URL slug.
	Slug Slug `json:"slug"`

	// Name is the display name.
	Name string `json:"name"`

	// Description explains the forum.
	Description string `json:"description"`

	// DisplayOrder controls forum ordering among siblings.
	DisplayOrder int `json:"display_order"`

	// Path is the materialized tree path.
	Path string `json:"path"`

	// Depth is the tree depth.
	Depth int `json:"depth"`

	// ExternalURL is the target URL for link forums.
	ExternalURL string `json:"external_url"`

	// IconAssetID is the optional icon asset.
	IconAssetID *uuid.UUID `json:"icon_asset_id,omitempty"`

	// ThreadVisibilityMode controls thread list filtering.
	ThreadVisibilityMode ThreadVisibilityMode `json:"thread_visibility_mode"`

	// MaxStickyThreads limits sticky threads in this forum.
	MaxStickyThreads int `json:"max_sticky_threads"`

	// DefaultThreadStatus is the initial thread status.
	DefaultThreadStatus ThreadStatus `json:"default_thread_status"`

	// AuthorPostEditWindowSeconds is the author self-edit window in seconds.
	AuthorPostEditWindowSeconds int `json:"author_post_edit_window_seconds"`

	// AuthorPostDeleteWindowSeconds is the author self-delete window in seconds.
	AuthorPostDeleteWindowSeconds int `json:"author_post_delete_window_seconds"`

	// Status is the forum lifecycle state.
	Status ForumStatus `json:"status"`

	// Version is the optimistic concurrency version.
	Version uint64 `json:"version"`

	// CreatedAt is the creation timestamp.
	CreatedAt time.Time `json:"created_at"`

	// UpdatedAt is the update timestamp.
	UpdatedAt time.Time `json:"updated_at"`
}

// Normalize returns a normalized forum copy.
func (forum Forum) Normalize() Forum {
	forum.Key = Key(strings.TrimSpace(string(forum.Key)))
	forum.Slug = Slug(strings.TrimSpace(string(forum.Slug)))
	forum.Name = strings.TrimSpace(forum.Name)
	forum.Description = strings.TrimSpace(forum.Description)
	forum.ExternalURL = strings.TrimSpace(forum.ExternalURL)
	forum = forum.normalizeDefaults()
	return forum
}

// Validate validates forum fields.
func (forum Forum) Validate() error {
	var violations []Violation
	violations = append(violations, forum.validateIdentity()...)
	violations = append(violations, forum.validateSettings()...)
	violations = append(violations, forum.validateStructure()...)
	return NewValidationError(violations)
}

// Discussion reports whether the forum can contain threads.
func (forum Forum) Discussion() bool {
	return forum.Kind == ForumKindDiscussion
}

// Settings returns the admin settings view for the forum.
func (forum Forum) Settings() ForumSettings {
	return ForumSettings{
		ForumID:                       forum.ID,
		Kind:                          forum.Kind,
		ExternalURL:                   forum.ExternalURL,
		ThreadVisibilityMode:          forum.ThreadVisibilityMode,
		MaxStickyThreads:              forum.MaxStickyThreads,
		DefaultThreadStatus:           forum.DefaultThreadStatus,
		AuthorPostEditWindowSeconds:   forum.AuthorPostEditWindowSeconds,
		AuthorPostDeleteWindowSeconds: forum.AuthorPostDeleteWindowSeconds,
		Version:                       forum.Version,
		UpdatedAt:                     forum.UpdatedAt,
	}
}

// normalizeDefaults supports package behavior.
func (forum Forum) normalizeDefaults() Forum {
	if forum.Kind == "" {
		forum.Kind = ForumKindDiscussion
	}
	if forum.ThreadVisibilityMode == "" {
		forum.ThreadVisibilityMode = ThreadVisibilityAllThreads
	}
	if forum.DefaultThreadStatus == "" {
		forum.DefaultThreadStatus = ThreadStatusOpen
	}
	if forum.AuthorPostEditWindowSeconds == 0 {
		forum.AuthorPostEditWindowSeconds = DefaultAuthorPostEditWindowSeconds
	}
	if forum.AuthorPostDeleteWindowSeconds == 0 {
		forum.AuthorPostDeleteWindowSeconds = DefaultAuthorPostDeleteWindowSeconds
	}
	if forum.Status == "" {
		forum.Status = ForumStatusActive
	}
	if forum.Version == 0 {
		forum.Version = 1
	}
	return forum
}

// validateIdentity supports package behavior.
func (forum Forum) validateIdentity() []Violation {
	var violations []Violation
	violations = append(violations, ValidateKey("key", forum.Key)...)
	violations = append(violations, ValidateSlug("slug", forum.Slug)...)
	violations = append(violations, ValidateName("name", forum.Name)...)
	violations = append(violations, ValidateDescription("description", forum.Description)...)
	if forum.CategoryID == uuid.Nil {
		violations = AppendViolation(violations, "category_id", "is required")
	}
	return violations
}

// validateSettings supports package behavior.
func (forum Forum) validateSettings() []Violation {
	var violations []Violation
	violations = append(violations, ValidateDisplayOrder("display_order", forum.DisplayOrder)...)
	violations = append(violations, ValidateForumKind("kind", forum.Kind)...)
	violations = append(violations, ValidateForumStatus("status", forum.Status)...)
	violations = append(
		violations,
		ValidateThreadVisibilityMode("thread_visibility_mode", forum.ThreadVisibilityMode)...,
	)
	violations = append(
		violations,
		ValidateThreadStatus("default_thread_status", forum.DefaultThreadStatus)...,
	)
	return validateForumBehavior(forum, violations)
}

// validateStructure supports package behavior.
func (forum Forum) validateStructure() []Violation {
	var violations []Violation
	if forum.Depth < 0 || forum.Depth > 5 {
		violations = AppendViolation(violations, "depth", "must be between 0 and 5")
	}
	if !validPath(forum.Path, forum.ID) {
		violations = AppendViolation(violations, "path", "must be a materialized path ending with the forum id")
	}
	return violations
}
