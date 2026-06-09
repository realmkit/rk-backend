package domain

import (
	"time"

	"github.com/google/uuid"
)

// Membership assigns a user to a group.
type Membership struct {
	// ID is the membership identifier.
	ID uuid.UUID `json:"id"`

	// GroupID is the group identifier.
	GroupID uuid.UUID `json:"group_id"`

	// UserID is the user identifier.
	UserID uuid.UUID `json:"user_id"`

	// Status is the membership lifecycle state.
	Status MembershipStatus `json:"status"`

	// AssignedByUserID is the assigner when known.
	AssignedByUserID *uuid.UUID `json:"assigned_by_user_id,omitempty"`

	// AssignedReason explains why the membership exists.
	AssignedReason string `json:"assigned_reason"`

	// StartsAt is when the membership starts.
	StartsAt *time.Time `json:"starts_at,omitempty"`

	// ExpiresAt is when the membership expires.
	ExpiresAt *time.Time `json:"expires_at,omitempty"`

	// Version is the optimistic concurrency version.
	Version uint64 `json:"version"`

	// CreatedAt is the creation timestamp.
	CreatedAt time.Time `json:"created_at"`

	// UpdatedAt is the update timestamp.
	UpdatedAt time.Time `json:"updated_at"`
}

// Validate validates membership fields.
func (membership Membership) Validate() error {
	var violations []Violation
	if membership.GroupID == uuid.Nil {
		violations = AppendViolation(violations, "group_id", "is required")
	}
	if membership.UserID == uuid.Nil {
		violations = AppendViolation(violations, "user_id", "is required")
	}
	violations = append(violations, ValidateMembershipStatus("status", membership.Status)...)
	if len(membership.AssignedReason) > 500 {
		violations = AppendViolation(violations, "assigned_reason", "must be at most 500 characters")
	}
	if membership.StartsAt != nil && membership.ExpiresAt != nil && membership.StartsAt.After(*membership.ExpiresAt) {
		violations = AppendViolation(violations, "expires_at", "must be after starts_at")
	}
	return NewValidationError(violations)
}

// ActiveAt reports whether membership grants permissions at instant.
func (membership Membership) ActiveAt(instant time.Time) bool {
	if membership.Status != MembershipStatusActive {
		return false
	}
	if membership.StartsAt != nil && instant.Before(*membership.StartsAt) {
		return false
	}
	if membership.ExpiresAt != nil && !instant.Before(*membership.ExpiresAt) {
		return false
	}
	return true
}
