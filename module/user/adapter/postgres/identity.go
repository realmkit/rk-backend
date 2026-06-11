package postgres

import (
	"context"

	"github.com/niflaot/gamehub-go/module/user/domain"
	"github.com/niflaot/gamehub-go/module/user/port"
	"github.com/niflaot/gamehub-go/pkg/orm"
)

// IdentityLinkRepository stores provider identity links.
type IdentityLinkRepository struct {
	store orm.Store
}

// NewIdentityLinkRepository creates an identity link repository.
func NewIdentityLinkRepository(store orm.Store) IdentityLinkRepository {
	return IdentityLinkRepository{store: store}
}

// Create stores an identity link.
func (repository IdentityLinkRepository) Create(ctx context.Context, link domain.IdentityLink) (domain.IdentityLink, error) {
	model := linkModelFromDomain(link)
	if err := repository.store.DB(ctx).Create(&model).Error; err != nil {
		return domain.IdentityLink{}, port.ErrConflict
	}
	return linkFromModel(model), nil
}

// FindByIssuerSubject returns one identity link.
func (repository IdentityLinkRepository) FindByIssuerSubject(
	ctx context.Context,
	issuer string,
	subject string,
) (domain.IdentityLink, error) {
	var model IdentityLinkModel
	if err := repository.store.DB(ctx).First(&model, "issuer = ? AND subject = ?", issuer, subject).Error; err != nil {
		return domain.IdentityLink{}, mapError(err)
	}
	return linkFromModel(model), nil
}

// Touch stores last-seen and sync information.
func (repository IdentityLinkRepository) Touch(ctx context.Context, link domain.IdentityLink) error {
	result := repository.store.DB(ctx).
		Model(&IdentityLinkModel{}).
		Where("id = ?", link.ID).
		Updates(map[string]any{"last_seen_at": link.LastSeenAt, "last_synced_at": link.LastSyncedAt, "claims_hash": link.ClaimsHash})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return port.ErrNotFound
	}
	return nil
}
