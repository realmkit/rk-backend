package port

import (
	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/module/user/domain"
	"github.com/realmkit/rk-backend/pkg/search"
)

// UpdateCurrentCommand updates local current-user settings.
type UpdateCurrentCommand struct {
	// UserID is the current local user identifier.
	UserID uuid.UUID

	// AvatarAssetID is the replacement avatar asset.
	AvatarAssetID *uuid.UUID

	// ExpectedVersion is the required current version.
	ExpectedVersion uint64
}

// CurrentUser contains current user response data.
type CurrentUser struct {
	// User is the local user.
	User domain.User `json:"user"`

	// Claims contains provider-owned cached claims when available.
	Claims *domain.ClaimCache `json:"provider_claims,omitempty"`
}

// UserFilter filters user list reads.
type UserFilter struct {
	// Status filters by local user status.
	Status domain.Status

	// Query filters by provider cached name, email, username, or user ID.
	Query search.TextQuery

	// Sort controls deterministic result ordering.
	Sort search.Sort
}

// UserSummary contains one user list row.
type UserSummary struct {
	// User is the local user.
	User domain.User `json:"user"`

	// Claims contains provider-owned cached display claims when available.
	Claims *domain.ClaimCache `json:"provider_claims,omitempty"`
}

// DefaultUserSort returns the default user list sort.
func DefaultUserSort() search.SortOption {
	return search.SortOption{Key: "created_at", DefaultDirection: search.DirectionDesc}
}

// AllowedUserSorts returns public user list sort keys.
func AllowedUserSorts() []search.SortOption {
	return []search.SortOption{
		DefaultUserSort(),
		{Key: "last_seen_at", DefaultDirection: search.DirectionDesc},
		{Key: "display_name", DefaultDirection: search.DirectionAsc},
		{Key: "email", DefaultDirection: search.DirectionAsc},
	}
}
