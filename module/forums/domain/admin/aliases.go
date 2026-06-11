// Package admin owns forum configuration and permission simulation models.
package admin

import (
	"github.com/google/uuid"
	shared "github.com/realmkit/rk-backend/module/forums/domain/shared"
)

// PermissionSubjectType identifies an admin-configurable forum grant subject.
type PermissionSubjectType = shared.PermissionSubjectType

// Violation is one validation failure.
type Violation = shared.Violation

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

// ValidatePermissionSubjectType validates forum permission grant subjects.
func ValidatePermissionSubjectType(
	field string,
	subjectType PermissionSubjectType,
) []Violation {
	return shared.ValidatePermissionSubjectType(field, subjectType)
}

// RootForumObjectID returns the reserved forum permission target.
func RootForumObjectID() uuid.UUID {
	return shared.RootForumObjectID()
}
