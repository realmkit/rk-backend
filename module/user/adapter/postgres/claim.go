package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/niflaot/gamehub-go/module/user/domain"
	"github.com/niflaot/gamehub-go/module/user/port"
	"github.com/niflaot/gamehub-go/pkg/orm"
)

// ClaimCacheRepository stores provider claim cache rows.
type ClaimCacheRepository struct {
	store orm.Store
}

// NewClaimCacheRepository creates a claim cache repository.
func NewClaimCacheRepository(store orm.Store) ClaimCacheRepository {
	return ClaimCacheRepository{store: store}
}

// Upsert stores provider claim cache data.
func (repository ClaimCacheRepository) Upsert(ctx context.Context, claims domain.ClaimCache) (domain.ClaimCache, error) {
	var model ClaimCacheModel
	err := repository.store.DB(ctx).First(&model, "user_id = ? AND issuer = ? AND subject = ?", claims.UserID, claims.Issuer, claims.Subject).Error
	if err == nil {
		model.Username = claims.Username
		model.Email = claims.Email
		model.EmailVerified = claims.EmailVerified
		model.DisplayName = claims.DisplayName
		model.PictureURL = claims.PictureURL
		model.PreferredLocale = claims.PreferredLocale
		model.ClaimsHash = claims.ClaimsHash
		model.SyncedAt = claims.SyncedAt
		if err := repository.store.DB(ctx).Save(&model).Error; err != nil {
			return domain.ClaimCache{}, err
		}
		return claimFromModel(model), nil
	}
	if mapped := mapError(err); !errors.Is(mapped, port.ErrNotFound) {
		return domain.ClaimCache{}, mapped
	}
	model = claimModelFromDomain(claims)
	if err := repository.store.DB(ctx).Create(&model).Error; err != nil {
		return domain.ClaimCache{}, port.ErrConflict
	}
	return claimFromModel(model), nil
}

// FindByUserID returns provider claim cache for a user.
func (repository ClaimCacheRepository) FindByUserID(ctx context.Context, userID uuid.UUID) (domain.ClaimCache, error) {
	var model ClaimCacheModel
	if err := repository.store.DB(ctx).First(&model, "user_id = ?", userID).Error; err != nil {
		return domain.ClaimCache{}, mapError(err)
	}
	return claimFromModel(model), nil
}
