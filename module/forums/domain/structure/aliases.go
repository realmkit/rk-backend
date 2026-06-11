// Package structure owns forum category, forum tree, and admin settings models.
package structure

import (
	"encoding/json"

	"github.com/google/uuid"
	shared "github.com/realmkit/rk-backend/module/forums/domain/shared"
)

// Key is a stable forum machine key.
type Key = shared.Key

// Slug is a URL-friendly forum slug.
type Slug = shared.Slug

// CategoryStatus is the category lifecycle state.
type CategoryStatus = shared.CategoryStatus

// ForumKind is the forum structural kind.
type ForumKind = shared.ForumKind

// ForumStatus is the forum lifecycle state.
type ForumStatus = shared.ForumStatus

// ThreadVisibilityMode controls user-facing thread list filtering.
type ThreadVisibilityMode = shared.ThreadVisibilityMode

// ThreadStatus is the default thread lifecycle state.
type ThreadStatus = shared.ThreadStatus

// PermissionSubjectType identifies an admin-configurable forum grant subject.
type PermissionSubjectType = shared.PermissionSubjectType

// Violation is one validation failure.
type Violation = shared.Violation

const (
	// CategoryStatusActive means the category can be displayed.
	CategoryStatusActive = shared.CategoryStatusActive
)

const (
	// ForumKindDiscussion means the forum can contain threads.
	ForumKindDiscussion = shared.ForumKindDiscussion

	// ForumKindLink means the forum points to an external URL.
	ForumKindLink = shared.ForumKindLink
)

const (
	// ForumStatusActive means the forum can be displayed and used.
	ForumStatusActive = shared.ForumStatusActive
)

const (
	// ThreadVisibilityAllThreads allows visible users to see all visible threads.
	ThreadVisibilityAllThreads = shared.ThreadVisibilityAllThreads
)

const (
	// ThreadStatusOpen is the default open thread state.
	ThreadStatusOpen = shared.ThreadStatusOpen
)

const (
	// PermissionSubjectPublic grants anonymous and authenticated users.
	PermissionSubjectPublic = shared.PermissionSubjectPublic

	// PermissionSubjectAuthenticated grants any authenticated local user.
	PermissionSubjectAuthenticated = shared.PermissionSubjectAuthenticated

	// PermissionSubjectUser grants one user.
	PermissionSubjectUser = shared.PermissionSubjectUser

	// PermissionSubjectGroup grants active members of one group.
	PermissionSubjectGroup = shared.PermissionSubjectGroup
)

const (
	// DefaultAuthorPostEditWindowSeconds is the default author self-edit window.
	DefaultAuthorPostEditWindowSeconds = shared.DefaultAuthorPostEditWindowSeconds

	// DefaultAuthorPostDeleteWindowSeconds is the default author self-delete window.
	DefaultAuthorPostDeleteWindowSeconds = shared.DefaultAuthorPostDeleteWindowSeconds
)

// AppendViolation appends one validation failure.
func AppendViolation(
	violations []Violation,
	field string,
	message string,
) []Violation {
	return shared.AppendViolation(violations, field, message)
}

// NewValidationError returns a validation error when violations are present.
func NewValidationError(violations []Violation) error {
	return shared.NewValidationError(violations)
}

// RootForumObjectID returns the reserved forum permission target.
func RootForumObjectID() uuid.UUID {
	return shared.RootForumObjectID()
}

// ValidateKey validates key.
func ValidateKey(field string, key Key) []Violation {
	return shared.ValidateKey(field, key)
}

// ValidateSlug validates slug.
func ValidateSlug(field string, slug Slug) []Violation {
	return shared.ValidateSlug(field, slug)
}

// ValidateName validates display names.
func ValidateName(field string, value string) []Violation {
	return shared.ValidateName(field, value)
}

// ValidateDescription validates descriptions.
func ValidateDescription(field string, value string) []Violation {
	return shared.ValidateDescription(field, value)
}

// ValidateDisplayOrder validates display order.
func ValidateDisplayOrder(field string, value int) []Violation {
	return shared.ValidateDisplayOrder(field, value)
}

// ValidateCategoryStatus validates category status.
func ValidateCategoryStatus(field string, status CategoryStatus) []Violation {
	return shared.ValidateCategoryStatus(field, status)
}

// ValidateForumKind validates forum kind.
func ValidateForumKind(field string, kind ForumKind) []Violation {
	return shared.ValidateForumKind(field, kind)
}

// ValidateForumStatus validates forum status.
func ValidateForumStatus(field string, status ForumStatus) []Violation {
	return shared.ValidateForumStatus(field, status)
}

// ValidateThreadVisibilityMode validates thread visibility mode.
func ValidateThreadVisibilityMode(field string, mode ThreadVisibilityMode) []Violation {
	return shared.ValidateThreadVisibilityMode(field, mode)
}

// ValidateThreadStatus validates thread status.
func ValidateThreadStatus(field string, status ThreadStatus) []Violation {
	return shared.ValidateThreadStatus(field, status)
}

// ValidatePermissionSubjectType validates forum permission grant subjects.
func ValidatePermissionSubjectType(
	field string,
	subjectType PermissionSubjectType,
) []Violation {
	return shared.ValidatePermissionSubjectType(field, subjectType)
}

// ValidateExternalURL validates optional external URL.
func ValidateExternalURL(field string, value string) []Violation {
	return shared.ValidateExternalURL(field, value)
}

// ValidateContentDocument validates stored rich-content JSON.
func ValidateContentDocument(field string, document json.RawMessage) []Violation {
	return shared.ValidateContentDocument(field, document)
}
