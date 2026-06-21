package structure

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

// ForumSettings contains admin-editable forum behavior settings.
type ForumSettings struct {
	// ForumID is the configured forum.
	ForumID uuid.UUID `json:"forum_id"`

	// Kind is the forum structural kind.
	Kind ForumKind `json:"kind"`

	// ExternalURL is required for link forums.
	ExternalURL string `json:"external_url"`

	// ThreadVisibilityMode shapes thread-list SQL for normal readers.
	ThreadVisibilityMode ThreadVisibilityMode `json:"thread_visibility_mode"`

	// MaxStickyThreads limits sticky threads.
	MaxStickyThreads int `json:"max_sticky_threads"`

	// DefaultThreadStatus is applied to new threads.
	DefaultThreadStatus ThreadStatus `json:"default_thread_status"`

	// AuthorPostEditWindowSeconds is the self-edit window, or -1 when disabled.
	AuthorPostEditWindowSeconds int `json:"author_post_edit_window_seconds"`

	// AuthorPostDeleteWindowSeconds is the self-delete window, or -1 when disabled.
	AuthorPostDeleteWindowSeconds int `json:"author_post_delete_window_seconds"`

	// Version is the forum optimistic concurrency version.
	Version uint64 `json:"version"`

	// UpdatedAt is the forum settings update timestamp.
	UpdatedAt time.Time `json:"updated_at"`
}

// Normalize returns a normalized settings copy.
func (settings ForumSettings) Normalize() ForumSettings {
	settings.ExternalURL = strings.TrimSpace(settings.ExternalURL)
	if settings.Kind == "" {
		settings.Kind = ForumKindDiscussion
	}
	if settings.ThreadVisibilityMode == "" {
		settings.ThreadVisibilityMode = ThreadVisibilityAllThreads
	}
	if settings.DefaultThreadStatus == "" {
		settings.DefaultThreadStatus = ThreadStatusOpen
	}
	if settings.AuthorPostEditWindowSeconds == 0 {
		settings.AuthorPostEditWindowSeconds = DefaultAuthorPostEditWindowSeconds
	}
	if settings.AuthorPostDeleteWindowSeconds == 0 {
		settings.AuthorPostDeleteWindowSeconds = DefaultAuthorPostDeleteWindowSeconds
	}
	return settings
}

// Validate validates forum settings.
func (settings ForumSettings) Validate() error {
	var violations []Violation
	if settings.ForumID == uuid.Nil {
		violations = AppendViolation(violations, "forum_id", "is required")
	}
	violations = append(violations, ValidateForumKind("kind", settings.Kind)...)
	violations = append(
		violations,
		ValidateThreadVisibilityMode("thread_visibility_mode", settings.ThreadVisibilityMode)...,
	)
	violations = append(
		violations,
		ValidateThreadStatus("default_thread_status", settings.DefaultThreadStatus)...,
	)
	violations = validateForumBehavior(settings, violations)
	return NewValidationError(violations)
}

// behaviorSettings defines package data.
type behaviorSettings interface {
	behaviorKind() ForumKind
	behaviorExternalURL() string
	behaviorMaxStickyThreads() int
	behaviorEditWindowSeconds() int
	behaviorDeleteWindowSeconds() int
}

// behaviorKind supports package behavior.
func (forum Forum) behaviorKind() ForumKind {
	return forum.Kind
}

// behaviorExternalURL supports package behavior.
func (forum Forum) behaviorExternalURL() string {
	return forum.ExternalURL
}

// behaviorMaxStickyThreads supports package behavior.
func (forum Forum) behaviorMaxStickyThreads() int {
	return forum.MaxStickyThreads
}

// behaviorEditWindowSeconds supports package behavior.
func (forum Forum) behaviorEditWindowSeconds() int {
	return forum.AuthorPostEditWindowSeconds
}

// behaviorDeleteWindowSeconds supports package behavior.
func (forum Forum) behaviorDeleteWindowSeconds() int {
	return forum.AuthorPostDeleteWindowSeconds
}

// behaviorKind supports package behavior.
func (settings ForumSettings) behaviorKind() ForumKind {
	return settings.Kind
}

// behaviorExternalURL supports package behavior.
func (settings ForumSettings) behaviorExternalURL() string {
	return settings.ExternalURL
}

// behaviorMaxStickyThreads supports package behavior.
func (settings ForumSettings) behaviorMaxStickyThreads() int {
	return settings.MaxStickyThreads
}

// behaviorEditWindowSeconds supports package behavior.
func (settings ForumSettings) behaviorEditWindowSeconds() int {
	return settings.AuthorPostEditWindowSeconds
}

// behaviorDeleteWindowSeconds supports package behavior.
func (settings ForumSettings) behaviorDeleteWindowSeconds() int {
	return settings.AuthorPostDeleteWindowSeconds
}

// validateForumBehavior supports package behavior.
func validateForumBehavior(
	settings behaviorSettings,
	violations []Violation,
) []Violation {
	if settings.behaviorMaxStickyThreads() < 0 {
		violations = AppendViolation(violations, "max_sticky_threads", "must be zero or greater")
	}
	if settings.behaviorEditWindowSeconds() < -1 {
		violations = AppendViolation(violations, "author_post_edit_window_seconds", "must be -1 or greater")
	}
	if settings.behaviorDeleteWindowSeconds() < -1 {
		violations = AppendViolation(violations, "author_post_delete_window_seconds", "must be -1 or greater")
	}
	if settings.behaviorKind() == ForumKindLink {
		violations = append(
			violations,
			ValidateExternalURL("external_url", settings.behaviorExternalURL())...,
		)
	}
	if settings.behaviorKind() != ForumKindLink && settings.behaviorExternalURL() != "" {
		violations = AppendViolation(violations, "external_url", "is only supported by link forums")
	}
	return violations
}
