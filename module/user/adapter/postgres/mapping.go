package postgres

import (
	"github.com/realmkit/rk-backend/module/user/domain"
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
