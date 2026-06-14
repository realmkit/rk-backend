package postgres

import (
	"time"

	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/module/user/domain"
	"github.com/realmkit/rk-backend/module/user/port"
	"github.com/realmkit/rk-backend/pkg/orm"
)

// userModelFromDomain maps domain user to persistence.
func userModelFromDomain(user domain.User) UserModel {
	return UserModel{
		ID:            orm.ID{ID: user.ID},
		Status:        string(user.Status),
		AvatarAssetID: user.AvatarAssetID,
		FirstSeenAt:   user.FirstSeenAt,
		LastSeenAt:    user.LastSeenAt,
		Version:       user.Version,
	}
}

// userFromModel maps persistence user to domain.
func userFromModel(model UserModel) domain.User {
	return domain.User{
		ID:            model.ID.ID,
		Status:        domain.Status(model.Status),
		AvatarAssetID: model.AvatarAssetID,
		FirstSeenAt:   model.FirstSeenAt,
		LastSeenAt:    model.LastSeenAt,
		Version:       model.Version,
		CreatedAt:     model.CreatedAt,
		UpdatedAt:     model.UpdatedAt,
	}
}

// userListRow contains one joined user and provider claim row.
type userListRow struct {
	ID              uuid.UUID
	Status          string
	AvatarAssetID   *uuid.UUID
	FirstSeenAt     time.Time
	LastSeenAt      *time.Time
	Version         uint64
	CreatedAt       time.Time
	UpdatedAt       time.Time
	ClaimID         *uuid.UUID
	Issuer          string
	Subject         string
	Username        string
	Email           string
	EmailVerified   bool
	DisplayName     string
	PictureURL      string
	PreferredLocale string
	ClaimsHash      string
	SyncedAt        *time.Time
}

// sortValue returns the cursor value for a user list row.
func (row userListRow) sortValue(key string) string {
	switch key {
	case "display_name":
		return row.DisplayName
	case "email":
		return row.Email
	case "last_seen_at":
		return optionalTime(row.LastSeenAt)
	default:
		return row.CreatedAt.Format(time.RFC3339Nano)
	}
}

// summary maps a user list row into a port summary.
func (row userListRow) summary() port.UserSummary {
	result := port.UserSummary{User: domain.User{
		ID:            row.ID,
		Status:        domain.Status(row.Status),
		AvatarAssetID: row.AvatarAssetID,
		FirstSeenAt:   row.FirstSeenAt,
		LastSeenAt:    row.LastSeenAt,
		Version:       row.Version,
		CreatedAt:     row.CreatedAt,
		UpdatedAt:     row.UpdatedAt,
	}}
	if row.ClaimID != nil {
		result.Claims = &domain.ClaimCache{
			ID:              *row.ClaimID,
			UserID:          row.ID,
			Issuer:          row.Issuer,
			Subject:         row.Subject,
			Username:        row.Username,
			Email:           row.Email,
			EmailVerified:   row.EmailVerified,
			DisplayName:     row.DisplayName,
			PictureURL:      row.PictureURL,
			PreferredLocale: row.PreferredLocale,
			ClaimsHash:      row.ClaimsHash,
			SyncedAt:        valueTime(row.SyncedAt),
		}
	}
	return result
}

// optionalTime formats a nullable time cursor value.
func optionalTime(value *time.Time) string {
	if value == nil {
		return time.Time{}.Format(time.RFC3339Nano)
	}
	return value.Format(time.RFC3339Nano)
}

// valueTime returns a zero time when value is nil.
func valueTime(value *time.Time) time.Time {
	if value == nil {
		return time.Time{}
	}
	return *value
}

// linkModelFromDomain maps identity link to persistence.
func linkModelFromDomain(link domain.IdentityLink) IdentityLinkModel {
	return IdentityLinkModel{
		ID:           orm.ID{ID: link.ID},
		UserID:       link.UserID,
		Provider:     link.Provider,
		Issuer:       link.Issuer,
		Subject:      link.Subject,
		SubjectHash:  link.SubjectHash,
		ClaimsHash:   link.ClaimsHash,
		LinkedAt:     link.LinkedAt,
		LastSeenAt:   link.LastSeenAt,
		LastSyncedAt: link.LastSyncedAt,
	}
}

// linkFromModel maps persistence identity link to domain.
func linkFromModel(model IdentityLinkModel) domain.IdentityLink {
	return domain.IdentityLink{
		ID:           model.ID.ID,
		UserID:       model.UserID,
		Provider:     model.Provider,
		Issuer:       model.Issuer,
		Subject:      model.Subject,
		SubjectHash:  model.SubjectHash,
		ClaimsHash:   model.ClaimsHash,
		LinkedAt:     model.LinkedAt,
		LastSeenAt:   model.LastSeenAt,
		LastSyncedAt: model.LastSyncedAt,
	}
}

// claimModelFromDomain maps claim cache to persistence.
func claimModelFromDomain(claims domain.ClaimCache) ClaimCacheModel {
	return ClaimCacheModel{
		ID:              orm.ID{ID: claims.ID},
		UserID:          claims.UserID,
		Issuer:          claims.Issuer,
		Subject:         claims.Subject,
		Username:        claims.Username,
		Email:           claims.Email,
		EmailVerified:   claims.EmailVerified,
		DisplayName:     claims.DisplayName,
		PictureURL:      claims.PictureURL,
		PreferredLocale: claims.PreferredLocale,
		ClaimsHash:      claims.ClaimsHash,
		SyncedAt:        claims.SyncedAt,
	}
}

// claimFromModel maps persistence claim cache to domain.
func claimFromModel(model ClaimCacheModel) domain.ClaimCache {
	return domain.ClaimCache{
		ID:              model.ID.ID,
		UserID:          model.UserID,
		Issuer:          model.Issuer,
		Subject:         model.Subject,
		Username:        model.Username,
		Email:           model.Email,
		EmailVerified:   model.EmailVerified,
		DisplayName:     model.DisplayName,
		PictureURL:      model.PictureURL,
		PreferredLocale: model.PreferredLocale,
		ClaimsHash:      model.ClaimsHash,
		SyncedAt:        model.SyncedAt,
	}
}
