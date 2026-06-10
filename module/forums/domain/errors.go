package domain

import (
	"encoding/json"

	shared "github.com/niflaot/gamehub-go/module/forums/domain/shared"
)

// ErrInvalid reports invalid forum domain data.
var ErrInvalid = shared.ErrInvalid

// Violation describes one validation failure.
type Violation = shared.Violation

// ValidationError contains domain validation failures.
type ValidationError = shared.ValidationError

// NewValidationError returns nil when violations is empty.
func NewValidationError(violations []Violation) error {
	return shared.NewValidationError(violations)
}

// AppendViolation appends one validation failure.
func AppendViolation(
	violations []Violation,
	field string,
	message string,
) []Violation {
	return shared.AppendViolation(violations, field, message)
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

// ValidateStickyState validates sticky state.
func ValidateStickyState(field string, state StickyState) []Violation {
	return shared.ValidateStickyState(field, state)
}

// ValidatePostStatus validates post status.
func ValidatePostStatus(field string, status PostStatus) []Violation {
	return shared.ValidatePostStatus(field, status)
}

// ValidateContentFormat validates content format.
func ValidateContentFormat(field string, format ContentFormat) []Violation {
	return shared.ValidateContentFormat(field, format)
}

// ValidateReferenceType validates reference type.
func ValidateReferenceType(field string, referenceType ReferenceType) []Violation {
	return shared.ValidateReferenceType(field, referenceType)
}

// ValidatePermissionSubjectType validates forum permission grant subjects.
func ValidatePermissionSubjectType(field string, subjectType PermissionSubjectType) []Violation {
	return shared.ValidatePermissionSubjectType(field, subjectType)
}

// ValidateExternalURL validates optional external URL.
func ValidateExternalURL(field string, value string) []Violation {
	return shared.ValidateExternalURL(field, value)
}

// ValidateTitle validates thread titles.
func ValidateTitle(field string, value string) []Violation {
	return shared.ValidateTitle(field, value)
}

// ValidateContentDocument validates stored rich-content JSON.
func ValidateContentDocument(field string, document json.RawMessage) []Violation {
	return shared.ValidateContentDocument(field, document)
}

// ValidateContentText validates extracted text.
func ValidateContentText(field string, value string) []Violation {
	return shared.ValidateContentText(field, value)
}
