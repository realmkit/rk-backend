package port

import (
	"github.com/google/uuid"
	"github.com/niflaot/gamehub-go/module/user/domain"
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
