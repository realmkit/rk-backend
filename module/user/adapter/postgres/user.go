package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/niflaot/gamehub-go/module/user/domain"
	"github.com/niflaot/gamehub-go/module/user/port"
	"github.com/niflaot/gamehub-go/pkg/orm"
	"gorm.io/gorm"
)

// UserRepository stores local users.
type UserRepository struct {
	store orm.Store
}

// NewUserRepository creates a user repository.
func NewUserRepository(store orm.Store) UserRepository {
	return UserRepository{store: store}
}

// Create stores a user.
func (repository UserRepository) Create(ctx context.Context, user domain.User) (domain.User, error) {
	model := userModelFromDomain(user)
	if err := repository.store.DB(ctx).Create(&model).Error; err != nil {
		return domain.User{}, port.ErrConflict
	}
	return userFromModel(model), nil
}

// Update stores mutable user fields.
func (repository UserRepository) Update(ctx context.Context, user domain.User, expectedVersion uint64) (domain.User, error) {
	result := repository.store.DB(ctx).
		Model(&UserModel{}).
		Where("id = ? AND version = ?", user.ID, expectedVersion).
		Updates(map[string]any{"avatar_asset_id": user.AvatarAssetID, "version": expectedVersion + 1})
	if result.Error != nil {
		return domain.User{}, result.Error
	}
	if result.RowsAffected == 0 {
		return domain.User{}, port.ErrPreconditionFailed
	}
	return repository.FindByID(ctx, user.ID)
}

// FindByID returns one user.
func (repository UserRepository) FindByID(ctx context.Context, id uuid.UUID) (domain.User, error) {
	var model UserModel
	if err := repository.store.DB(ctx).First(&model, "id = ?", id).Error; err != nil {
		return domain.User{}, mapError(err)
	}
	return userFromModel(model), nil
}

// TouchLastSeen stores the last-seen timestamp.
func (repository UserRepository) TouchLastSeen(ctx context.Context, id uuid.UUID) error {
	result := repository.store.DB(ctx).Model(&UserModel{}).Where("id = ?", id).Update("last_seen_at", time.Now().UTC())
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return port.ErrNotFound
	}
	return nil
}

// mapError maps persistence errors to port errors.
func mapError(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return port.ErrNotFound
	}
	return err
}
