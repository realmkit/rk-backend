package postgres

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/module/themes/domain"
	"github.com/realmkit/rk-backend/module/themes/port"
	"github.com/realmkit/rk-backend/pkg/orm"
)

// SigningKeyRepository stores trusted signing keys in PostgreSQL.
type SigningKeyRepository struct {
	store orm.Store
}

// NewSigningKeyRepository creates a signing key repository.
func NewSigningKeyRepository(store orm.Store) SigningKeyRepository {
	return SigningKeyRepository{store: store}
}

// Upsert stores or updates a trusted signing key.
func (repository SigningKeyRepository) Upsert(
	ctx context.Context,
	key domain.ThemeSigningKey,
) (domain.ThemeSigningKey, error) {
	existing, err := repository.FindByKeyID(ctx, key.KeyID)
	if err == nil {
		key.ID = existing.ID
		return repository.updateSigningKey(ctx, key)
	}
	if err != port.ErrNotFound {
		return domain.ThemeSigningKey{}, err
	}
	model := signingKeyModel(key)
	if err := repository.store.DB(ctx).Create(&model).Error; err != nil {
		return domain.ThemeSigningKey{}, port.ErrConflict
	}
	return signingKeyFromModel(model), nil
}

// FindByKeyID returns a trusted signing key.
func (repository SigningKeyRepository) FindByKeyID(ctx context.Context, keyID string) (domain.ThemeSigningKey, error) {
	var model SigningKeyModel
	if err := repository.store.DB(ctx).First(&model, "key_id = ?", keyID).Error; err != nil {
		return domain.ThemeSigningKey{}, mapError(err)
	}
	return signingKeyFromModel(model), nil
}

// List returns trusted signing keys.
func (repository SigningKeyRepository) List(ctx context.Context) ([]domain.ThemeSigningKey, error) {
	var models []SigningKeyModel
	if err := repository.store.DB(ctx).Order("trust_level ASC, key_id ASC").Find(&models).Error; err != nil {
		return nil, err
	}
	keys := make([]domain.ThemeSigningKey, 0, len(models))
	for _, model := range models {
		keys = append(keys, signingKeyFromModel(model))
	}
	return keys, nil
}

// updateSigningKey updates an existing trusted signing key.
func (repository SigningKeyRepository) updateSigningKey(
	ctx context.Context,
	key domain.ThemeSigningKey,
) (domain.ThemeSigningKey, error) {
	result := repository.store.DB(ctx).Model(&SigningKeyModel{}).
		Where("id = ?", key.ID).
		Updates(signingKeyUpdates(key))
	if result.Error != nil {
		return domain.ThemeSigningKey{}, result.Error
	}
	if result.RowsAffected == 0 {
		return domain.ThemeSigningKey{}, port.ErrNotFound
	}
	return repository.FindByKeyID(ctx, key.KeyID)
}

// PreviewTokenRepository stores preview tokens in PostgreSQL.
type PreviewTokenRepository struct {
	store orm.Store
}

// NewPreviewTokenRepository creates a preview token repository.
func NewPreviewTokenRepository(store orm.Store) PreviewTokenRepository {
	return PreviewTokenRepository{store: store}
}

// Create stores one preview token.
func (repository PreviewTokenRepository) Create(
	ctx context.Context,
	token domain.ThemePreviewToken,
) (domain.ThemePreviewToken, error) {
	model := previewTokenModel(token)
	if err := repository.store.DB(ctx).Create(&model).Error; err != nil {
		return domain.ThemePreviewToken{}, port.ErrConflict
	}
	return previewTokenFromModel(model), nil
}

// FindByTokenHash returns one active preview token.
func (repository PreviewTokenRepository) FindByTokenHash(
	ctx context.Context,
	tokenHash string,
) (domain.ThemePreviewToken, error) {
	var model PreviewTokenModel
	err := repository.store.DB(ctx).
		First(&model, "token_hash = ? AND revoked_at IS NULL", tokenHash).
		Error
	if err != nil {
		return domain.ThemePreviewToken{}, mapError(err)
	}
	return previewTokenFromModel(model), nil
}

// Revoke marks a preview token as revoked.
func (repository PreviewTokenRepository) Revoke(ctx context.Context, id uuid.UUID) error {
	result := repository.store.DB(ctx).Model(&PreviewTokenModel{}).
		Where("id = ? AND revoked_at IS NULL", id).
		Update("revoked_at", time.Now().UTC())
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return port.ErrNotFound
	}
	return nil
}

// signingKeyModel maps signing key to persistence.
func signingKeyModel(key domain.ThemeSigningKey) SigningKeyModel {
	return SigningKeyModel{
		ID:              orm.ID{ID: key.ID},
		KeyID:           key.KeyID,
		Algorithm:       string(key.Algorithm),
		PublicKey:       key.PublicKey,
		TrustLevel:      string(key.TrustLevel),
		Status:          string(key.Status),
		Source:          string(key.Source),
		NotBefore:       key.NotBefore,
		NotAfter:        key.NotAfter,
		CreatedByUserID: key.CreatedBy,
		Description:     key.Description,
		RetiredAt:       key.RetiredAt,
		RevokedAt:       key.RevokedAt,
	}
}

// signingKeyFromModel maps persistence to signing key.
func signingKeyFromModel(model SigningKeyModel) domain.ThemeSigningKey {
	return domain.ThemeSigningKey{
		ID:          model.ID.ID,
		KeyID:       model.KeyID,
		Algorithm:   domain.SignatureAlgorithm(model.Algorithm),
		PublicKey:   model.PublicKey,
		TrustLevel:  domain.SigningKeyTrustLevel(model.TrustLevel),
		Status:      domain.SigningKeyStatus(model.Status),
		Source:      domain.SigningKeySource(model.Source),
		NotBefore:   model.NotBefore,
		NotAfter:    model.NotAfter,
		CreatedBy:   model.CreatedByUserID,
		CreatedAt:   model.CreatedAt,
		UpdatedAt:   model.UpdatedAt,
		RetiredAt:   model.RetiredAt,
		RevokedAt:   model.RevokedAt,
		Description: model.Description,
	}
}

// signingKeyUpdates returns mutable signing key updates.
func signingKeyUpdates(key domain.ThemeSigningKey) map[string]any {
	return map[string]any{
		"algorithm":          string(key.Algorithm),
		"public_key":         key.PublicKey,
		"trust_level":        string(key.TrustLevel),
		"status":             string(key.Status),
		"source":             string(key.Source),
		"not_before":         key.NotBefore,
		"not_after":          key.NotAfter,
		"description":        key.Description,
		"retired_at":         key.RetiredAt,
		"revoked_at":         key.RevokedAt,
		"created_by_user_id": key.CreatedBy,
	}
}

// previewTokenModel maps preview token to persistence.
func previewTokenModel(token domain.ThemePreviewToken) PreviewTokenModel {
	return PreviewTokenModel{
		ID:              orm.ID{ID: token.ID},
		VersionID:       token.VersionID,
		TokenHash:       token.TokenHash,
		PersonaKind:     string(token.PersonaKind),
		PersonaSource:   string(token.PersonaSource),
		PersonaUserID:   token.PersonaUserID,
		ExpiresAt:       token.ExpiresAt,
		CreatedByUserID: token.CreatedBy,
		RevokedAt:       token.RevokedAt,
	}
}

// previewTokenFromModel maps persistence to preview token.
func previewTokenFromModel(model PreviewTokenModel) domain.ThemePreviewToken {
	return domain.ThemePreviewToken{
		ID:            model.ID.ID,
		VersionID:     model.VersionID,
		TokenHash:     model.TokenHash,
		PersonaKind:   domain.PreviewPersonaKind(model.PersonaKind),
		PersonaSource: domain.PreviewPersonaSource(model.PersonaSource),
		PersonaUserID: model.PersonaUserID,
		ExpiresAt:     model.ExpiresAt,
		CreatedBy:     model.CreatedByUserID,
		CreatedAt:     model.CreatedAt,
		RevokedAt:     model.RevokedAt,
	}
}

// Ensure runtime repositories implement their ports.
var (
	_ port.SigningKeyRepository   = SigningKeyRepository{}
	_ port.PreviewTokenRepository = PreviewTokenRepository{}
)
