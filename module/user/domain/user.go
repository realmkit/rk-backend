package domain

import (
	"time"

	"github.com/google/uuid"
)

// User is the local RealmKit account anchor.
type User struct {
	// ID is the local user identifier.
	ID uuid.UUID `json:"id"`

	// Status is the local lifecycle state.
	Status Status `json:"status"`

	// AvatarAssetID is the optional RealmKit avatar asset.
	AvatarAssetID *uuid.UUID `json:"avatar_asset_id,omitempty"`

	// FirstSeenAt is the first authentication timestamp.
	FirstSeenAt time.Time `json:"first_seen_at"`

	// LastSeenAt is the last authentication timestamp.
	LastSeenAt *time.Time `json:"last_seen_at,omitempty"`

	// Version is the optimistic concurrency version.
	Version uint64 `json:"version"`

	// CreatedAt is the creation timestamp.
	CreatedAt time.Time `json:"created_at"`

	// UpdatedAt is the update timestamp.
	UpdatedAt time.Time `json:"updated_at"`
}

// Validate validates user fields.
func (user User) Validate() error {
	var violations []Violation
	violations = append(violations, ValidateStatus("status", user.Status)...)
	if user.FirstSeenAt.IsZero() {
		violations = AppendViolation(violations, "first_seen_at", "is required")
	}
	return NewValidationError(violations)
}

// CanAuthenticate reports whether the local user may authenticate.
func (user User) CanAuthenticate() bool {
	return user.Status == StatusActive || user.Status == StatusPendingProfile
}
