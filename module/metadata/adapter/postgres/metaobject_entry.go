package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/niflaot/gamehub-go/module/metadata/domain"
	"github.com/niflaot/gamehub-go/module/metadata/port"
	"github.com/niflaot/gamehub-go/pkg/orm"
	"github.com/niflaot/gamehub-go/pkg/pagination"
)

// MetaobjectEntryRepository stores metaobject entries in PostgreSQL.
type MetaobjectEntryRepository struct {
	store orm.Store
}

// NewMetaobjectEntryRepository creates a metaobject entry repository.
func NewMetaobjectEntryRepository(store orm.Store) MetaobjectEntryRepository {
	return MetaobjectEntryRepository{store: store}
}

// Create stores entry.
func (repository MetaobjectEntryRepository) Create(ctx context.Context, entry domain.MetaobjectEntry) (domain.MetaobjectEntry, error) {
	if _, err := repository.FindByHandle(ctx, entry.DefinitionID, entry.Handle); err == nil {
		return domain.MetaobjectEntry{}, port.ErrConflict
	} else if !errors.Is(err, port.ErrNotFound) {
		return domain.MetaobjectEntry{}, err
	}
	model := metaobjectEntryModelFromDomain(entry)
	if model.Version == 0 {
		model.Version = 1
	}
	if err := repository.store.DB(ctx).Create(&model).Error; err != nil {
		return domain.MetaobjectEntry{}, port.ErrConflict
	}
	return metaobjectEntryFromModel(model)
}

// Update stores mutable entry changes.
func (repository MetaobjectEntryRepository) Update(
	ctx context.Context,
	entry domain.MetaobjectEntry,
	expectedVersion uint64,
) (domain.MetaobjectEntry, error) {
	model := metaobjectEntryModelFromDomain(entry)
	result := repository.store.DB(ctx).
		Model(&MetaobjectEntryModel{}).
		Where("id = ? AND version = ?", entry.ID, expectedVersion).
		Updates(metaobjectEntryUpdates(model, expectedVersion))
	if result.Error != nil {
		return domain.MetaobjectEntry{}, result.Error
	}
	if result.RowsAffected == 0 {
		return domain.MetaobjectEntry{}, port.ErrPreconditionFailed
	}
	return repository.FindByID(ctx, entry.ID)
}

// FindByID returns one entry by ID.
func (repository MetaobjectEntryRepository) FindByID(ctx context.Context, id uuid.UUID) (domain.MetaobjectEntry, error) {
	var model MetaobjectEntryModel
	if err := repository.store.DB(ctx).First(&model, "id = ?", id).Error; err != nil {
		return domain.MetaobjectEntry{}, mapError(err)
	}
	return metaobjectEntryFromModel(model)
}

// FindByHandle returns one entry by definition and handle.
func (repository MetaobjectEntryRepository) FindByHandle(
	ctx context.Context,
	definitionID uuid.UUID,
	handle domain.Handle,
) (domain.MetaobjectEntry, error) {
	var model MetaobjectEntryModel
	err := repository.store.DB(ctx).
		First(&model, "definition_id = ? AND handle = ?", definitionID, handle).
		Error
	if err != nil {
		return domain.MetaobjectEntry{}, mapError(err)
	}
	return metaobjectEntryFromModel(model)
}

// List returns entries for definition.
func (repository MetaobjectEntryRepository) List(
	ctx context.Context,
	definitionID uuid.UUID,
	page pagination.Page,
) (pagination.Result[domain.MetaobjectEntry], error) {
	var models []MetaobjectEntryModel
	err := repository.store.DB(ctx).
		Where("definition_id = ?", definitionID).
		Order("display_name asc").
		Limit(page.Limit + 1).
		Find(&models).
		Error
	if err != nil {
		return pagination.Result[domain.MetaobjectEntry]{}, err
	}
	return metaobjectEntryPage(models, page.Limit)
}

// Delete soft deletes entry.
func (repository MetaobjectEntryRepository) Delete(ctx context.Context, id uuid.UUID, expectedVersion uint64) error {
	result := repository.store.DB(ctx).
		Where("id = ? AND version = ?", id, expectedVersion).
		Delete(&MetaobjectEntryModel{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return port.ErrPreconditionFailed
	}
	return nil
}

// CountByDefinition returns active entry count for definition.
func (repository MetaobjectEntryRepository) CountByDefinition(ctx context.Context, definitionID uuid.UUID) (int64, error) {
	var count int64
	err := repository.store.DB(ctx).
		Model(&MetaobjectEntryModel{}).
		Where("definition_id = ?", definitionID).
		Count(&count).
		Error
	return count, err
}

// metaobjectEntryUpdates returns update fields for an entry.
func metaobjectEntryUpdates(model MetaobjectEntryModel, expectedVersion uint64) map[string]any {
	return map[string]any{
		"display_name": model.DisplayName,
		"field_values": model.Fields,
		"version":      expectedVersion + 1,
	}
}

// metaobjectEntryPage maps entry models into a page.
func metaobjectEntryPage(models []MetaobjectEntryModel, limit int) (pagination.Result[domain.MetaobjectEntry], error) {
	next := ""
	if len(models) > limit {
		next = models[limit-1].ID.ID.String()
		models = models[:limit]
	}
	items := make([]domain.MetaobjectEntry, 0, len(models))
	for _, model := range models {
		item, err := metaobjectEntryFromModel(model)
		if err != nil {
			return pagination.Result[domain.MetaobjectEntry]{}, err
		}
		items = append(items, item)
	}
	return pagination.Result[domain.MetaobjectEntry]{Items: items, NextCursor: next}, nil
}
