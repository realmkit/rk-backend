package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/module/metadata/domain"
	"github.com/realmkit/rk-backend/module/metadata/port"
	"github.com/realmkit/rk-backend/pkg/orm"
	"gorm.io/gorm"
)

// MetafieldValueRepository stores values in PostgreSQL.
type MetafieldValueRepository struct {
	store orm.Store
}

// NewMetafieldValueRepository creates a metafield value repository.
func NewMetafieldValueRepository(store orm.Store) MetafieldValueRepository {
	return MetafieldValueRepository{store: store}
}

// Upsert creates or updates value.
func (repository MetafieldValueRepository) Upsert(
	ctx context.Context,
	value domain.MetafieldValue,
	expectedVersion *uint64,
) (domain.MetafieldValue, bool, error) {
	var current MetafieldValueModel
	err := repository.store.DB(ctx).
		Where("definition_id = ? AND owner_type = ? AND owner_id = ?", value.DefinitionID, value.OwnerType, value.OwnerID).
		First(&current).
		Error
	if err == nil {
		return repository.updateExisting(ctx, current, value, expectedVersion)
	}
	if err != nil && err != gorm.ErrRecordNotFound {
		return domain.MetafieldValue{}, false, err
	}
	if expectedVersion != nil {
		return domain.MetafieldValue{}, false, port.ErrPreconditionFailed
	}
	model := valueModelFromDomain(value)
	model.Version = 1
	if err := repository.store.DB(ctx).Create(&model).Error; err != nil {
		return domain.MetafieldValue{}, false, port.ErrConflict
	}
	return valueFromModel(model), true, nil
}

// Find returns one owner value.
func (repository MetafieldValueRepository) Find(
	ctx context.Context,
	definitionID uuid.UUID,
	ownerType domain.OwnerType,
	ownerID uuid.UUID,
) (domain.MetafieldValue, error) {
	var model MetafieldValueModel
	err := repository.store.DB(ctx).
		First(&model, "definition_id = ? AND owner_type = ? AND owner_id = ?", definitionID, ownerType, ownerID).
		Error
	if err != nil {
		return domain.MetafieldValue{}, mapError(err)
	}
	return valueFromModel(model), nil
}

// ListForOwner returns all values for owner.
func (repository MetafieldValueRepository) ListForOwner(
	ctx context.Context,
	ownerType domain.OwnerType,
	ownerID uuid.UUID,
) ([]domain.MetafieldValue, error) {
	var models []MetafieldValueModel
	err := repository.store.DB(ctx).Find(&models, "owner_type = ? AND owner_id = ?", ownerType, ownerID).Error
	if err != nil {
		return nil, err
	}
	values := make([]domain.MetafieldValue, 0, len(models))
	for _, model := range models {
		values = append(values, valueFromModel(model))
	}
	return values, nil
}

// Delete soft deletes one owner value.
func (repository MetafieldValueRepository) Delete(
	ctx context.Context,
	definitionID uuid.UUID,
	ownerType domain.OwnerType,
	ownerID uuid.UUID,
	expectedVersion uint64,
) error {
	result := repository.store.DB(ctx).
		Where("definition_id = ? AND owner_type = ? AND owner_id = ? AND version = ?", definitionID, ownerType, ownerID, expectedVersion).
		Delete(&MetafieldValueModel{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return port.ErrPreconditionFailed
	}
	return nil
}

// CountByDefinition returns active value count for definition.
func (repository MetafieldValueRepository) CountByDefinition(ctx context.Context, definitionID uuid.UUID) (int64, error) {
	var count int64
	err := repository.store.DB(ctx).Model(&MetafieldValueModel{}).Where("definition_id = ?", definitionID).Count(&count).Error
	return count, err
}

// updateExisting updates a current value.
func (repository MetafieldValueRepository) updateExisting(
	ctx context.Context,
	current MetafieldValueModel,
	value domain.MetafieldValue,
	expectedVersion *uint64,
) (domain.MetafieldValue, bool, error) {
	if expectedVersion != nil && current.Version != *expectedVersion {
		return domain.MetafieldValue{}, false, port.ErrPreconditionFailed
	}
	nextVersion := current.Version + 1
	result := repository.store.DB(ctx).
		Model(&MetafieldValueModel{}).
		Where("id = ? AND version = ?", current.ID.ID, current.Version).
		Updates(map[string]any{
			"value_json": JSON(value.Value),
			"version":    nextVersion,
		})
	if result.Error != nil {
		return domain.MetafieldValue{}, false, result.Error
	}
	if result.RowsAffected == 0 {
		return domain.MetafieldValue{}, false, port.ErrPreconditionFailed
	}
	updated, err := repository.Find(ctx, value.DefinitionID, value.OwnerType, value.OwnerID)
	return updated, false, err
}
