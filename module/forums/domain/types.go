package domain

import (
	"net/url"
	"regexp"
	"slices"
	"strings"

	"github.com/google/uuid"
)

// Key is a stable forum machine key.
type Key string

// Slug is a URL-friendly forum slug.
type Slug string

// CategoryStatus is the category lifecycle state.
type CategoryStatus string

// ForumKind is the forum structural kind.
type ForumKind string

// ForumStatus is the forum lifecycle state.
type ForumStatus string

// ThreadVisibilityMode controls user-facing thread list filtering.
type ThreadVisibilityMode string

// ThreadStatus is the default thread lifecycle state.
type ThreadStatus string

const (
	// CategoryStatusActive means the category can be displayed.
	CategoryStatusActive CategoryStatus = "active"

	// CategoryStatusHidden means the category is hidden.
	CategoryStatusHidden CategoryStatus = "hidden"

	// CategoryStatusArchived means the category is archived.
	CategoryStatusArchived CategoryStatus = "archived"
)

const (
	// ForumKindDiscussion means the forum can contain threads.
	ForumKindDiscussion ForumKind = "discussion"

	// ForumKindLink means the forum points to an external URL.
	ForumKindLink ForumKind = "link"

	// ForumKindContainer means the forum groups child forums.
	ForumKindContainer ForumKind = "container"
)

const (
	// ForumStatusActive means the forum can be displayed and used.
	ForumStatusActive ForumStatus = "active"

	// ForumStatusHidden means the forum is hidden.
	ForumStatusHidden ForumStatus = "hidden"

	// ForumStatusArchived means the forum is archived.
	ForumStatusArchived ForumStatus = "archived"
)

const (
	// ThreadVisibilityAllThreads allows visible users to see all visible threads.
	ThreadVisibilityAllThreads ThreadVisibilityMode = "all_threads"

	// ThreadVisibilityOwnThreads allows visible users to see only their authored threads.
	ThreadVisibilityOwnThreads ThreadVisibilityMode = "own_threads"

	// ThreadVisibilityOwnOrStickyThreads allows visible users to see authored or sticky threads.
	ThreadVisibilityOwnOrStickyThreads ThreadVisibilityMode = "own_or_sticky_threads"
)

const (
	// ThreadStatusOpen is the default open thread state.
	ThreadStatusOpen ThreadStatus = "open"

	// ThreadStatusClosed is the default closed thread state.
	ThreadStatusClosed ThreadStatus = "closed"

	// ThreadStatusLocked is the default locked thread state.
	ThreadStatusLocked ThreadStatus = "locked"
)

// rootForumObjectID is the reserved permission object for category and root forum administration.
const rootForumObjectID = "00000000-0000-0000-0000-000000000101"

// RootForumObjectID returns the reserved forum permission target for structure administration.
func RootForumObjectID() uuid.UUID {
	return uuid.MustParse(rootForumObjectID)
}

// keyPattern matches stable lower snake identifiers.
var keyPattern = regexp.MustCompile(`^[a-z][a-z0-9_]{1,62}[a-z0-9]$`)

// slugPattern matches URL slugs.
var slugPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{1,118}[a-z0-9]$`)

// ValidateKey validates key.
func ValidateKey(field string, key Key) []Violation {
	if !keyPattern.MatchString(strings.TrimSpace(string(key))) {
		return []Violation{{Field: field, Message: "must be lower snake case and between 3 and 64 characters"}}
	}
	return nil
}

// ValidateSlug validates slug.
func ValidateSlug(field string, slug Slug) []Violation {
	if !slugPattern.MatchString(strings.TrimSpace(string(slug))) {
		return []Violation{{Field: field, Message: "must be lower kebab case and between 3 and 120 characters"}}
	}
	return nil
}

// ValidateName validates display names.
func ValidateName(field string, value string) []Violation {
	length := len(strings.TrimSpace(value))
	if length < 1 || length > 120 {
		return []Violation{{Field: field, Message: "must be between 1 and 120 characters"}}
	}
	return nil
}

// ValidateDescription validates descriptions.
func ValidateDescription(field string, value string) []Violation {
	if len(strings.TrimSpace(value)) > 1000 {
		return []Violation{{Field: field, Message: "must be at most 1000 characters"}}
	}
	return nil
}

// ValidateDisplayOrder validates display order.
func ValidateDisplayOrder(field string, value int) []Violation {
	if value < 0 {
		return []Violation{{Field: field, Message: "must be zero or greater"}}
	}
	return nil
}

// ValidateCategoryStatus validates category status.
func ValidateCategoryStatus(field string, status CategoryStatus) []Violation {
	if slices.Contains([]CategoryStatus{CategoryStatusActive, CategoryStatusHidden, CategoryStatusArchived}, status) {
		return nil
	}
	return []Violation{{Field: field, Message: "is not supported"}}
}

// ValidateForumKind validates forum kind.
func ValidateForumKind(field string, kind ForumKind) []Violation {
	if slices.Contains([]ForumKind{ForumKindDiscussion, ForumKindLink, ForumKindContainer}, kind) {
		return nil
	}
	return []Violation{{Field: field, Message: "is not supported"}}
}

// ValidateForumStatus validates forum status.
func ValidateForumStatus(field string, status ForumStatus) []Violation {
	if slices.Contains([]ForumStatus{ForumStatusActive, ForumStatusHidden, ForumStatusArchived}, status) {
		return nil
	}
	return []Violation{{Field: field, Message: "is not supported"}}
}

// ValidateThreadVisibilityMode validates thread visibility mode.
func ValidateThreadVisibilityMode(field string, mode ThreadVisibilityMode) []Violation {
	if slices.Contains([]ThreadVisibilityMode{ThreadVisibilityAllThreads, ThreadVisibilityOwnThreads, ThreadVisibilityOwnOrStickyThreads}, mode) {
		return nil
	}
	return []Violation{{Field: field, Message: "is not supported"}}
}

// ValidateThreadStatus validates thread status.
func ValidateThreadStatus(field string, status ThreadStatus) []Violation {
	if slices.Contains([]ThreadStatus{ThreadStatusOpen, ThreadStatusClosed, ThreadStatusLocked}, status) {
		return nil
	}
	return []Violation{{Field: field, Message: "is not supported"}}
}

// ValidateExternalURL validates optional external URL.
func ValidateExternalURL(field string, value string) []Violation {
	if strings.TrimSpace(value) == "" {
		return []Violation{{Field: field, Message: "is required"}}
	}
	parsed, err := url.ParseRequestURI(strings.TrimSpace(value))
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return []Violation{{Field: field, Message: "must be an absolute URL"}}
	}
	return nil
}
