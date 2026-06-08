package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/niflaot/gamehub-go/module/metadata/domain"
	"github.com/niflaot/gamehub-go/module/metadata/port"
	"github.com/niflaot/gamehub-go/pkg/orm"
	"github.com/niflaot/gamehub-go/pkg/pagination"
	"gorm.io/gorm"
)

// MetaobjectDefinitionRepository stores metaobject definitions in PostgreSQL.
type MetaobjectDefinitionRepository struct {
	store orm.Store
}

// NewMetaobjectDefinitionRepository creates a metaobject definition repository.
func NewMetaobjectDefinitionRepository(store orm.Store) MetaobjectDefinitionRepository {
	return MetaobjectDefinitionRepository{store: store}
}

// Create stores definition.
func (repository MetaobjectDefinitionRepository) Create(ctx context.Context, definition domain.MetaobjectDefinition) (domain.MetaobjectDefinition, error) {
	if _, err := repository.FindByType(ctx, definition.Type); err == nil {
		return domain.MetaobjectDefinition{}, port.ErrConflict
	} else if !errors.Is(err, port.ErrNotFound) {
		return domain.MetaobjectDefinition{}, err
	}
	model := metaobjectDefinitionModelFromDomain(definition)
	if model.Version == 0 {
		model.Version = 1
	}
	if err := repository.store.DB(ctx).Create(&model).Error; err != nil {
		return domain.MetaobjectDefinition{}, port.ErrConflict
	}
	return metaobjectDefinitionFromModel(model)
}

// Update stores mutable definition changes.
func (repository MetaobjectDefinitionRepository) Update(ctx context.Context, definition domain.MetaobjectDefinition, expectedVersion uint64) (domain.MetaobjectDefinition, error) {
	model := metaobjectDefinitionModelFromDomain(definition)
	result := repository.store.DB(ctx).Model(&MetaobjectDefinitionModel{}).Where("id = ? AND version = ?", definition.ID, expectedVersion).Updates(map[string]any{
		"name":              model.Name,
		"description":       model.Description,
		"field_definitions": model.Fields,
		"active":            model.Active,
		"version":           expectedVersion + 1,
	})
	if result.Error != nil {
		return domain.MetaobjectDefinition{}, result.Error
	}
	if result.RowsAffected == 0 {
		return domain.MetaobjectDefinition{}, port.ErrPreconditionFailed
	}
	return repository.FindByID(ctx, definition.ID)
}

// FindByID returns one definition by ID.
func (repository MetaobjectDefinitionRepository) FindByID(ctx context.Context, id uuid.UUID) (domain.MetaobjectDefinition, error) {
	var model MetaobjectDefinitionModel
	if err := repository.store.DB(ctx).First(&model, "id = ?", id).Error; err != nil {
		return domain.MetaobjectDefinition{}, mapError(err)
	}
	return metaobjectDefinitionFromModel(model)
}

// FindByType returns one definition by type.
func (repository MetaobjectDefinitionRepository) FindByType(ctx context.Context, objectType domain.MetaobjectType) (domain.MetaobjectDefinition, error) {
	var model MetaobjectDefinitionModel
	err := repository.store.DB(ctx).First(&model, "type = ?", objectType).Error
	if err != nil {
		return domain.MetaobjectDefinition{}, mapError(err)
	}
	return metaobjectDefinitionFromModel(model)
}

// List returns definitions matching filter.
func (repository MetaobjectDefinitionRepository) List(ctx context.Context, filter port.MetaobjectDefinitionFilter, page pagination.Page) (pagination.Result[domain.MetaobjectDefinition], error) {
	query := repository.store.DB(ctx).Model(&MetaobjectDefinitionModel{}).Order("type asc").Limit(page.Limit + 1)
	query = applyMetaobjectDefinitionFilter(query, filter)
	var models []MetaobjectDefinitionModel
	if err := query.Find(&models).Error; err != nil {
		return pagination.Result[domain.MetaobjectDefinition]{}, err
	}
	return metaobjectDefinitionPage(models, page.Limit)
}

// Archive soft deletes definition.
func (repository MetaobjectDefinitionRepository) Archive(ctx context.Context, id uuid.UUID, expectedVersion uint64) error {
	result := repository.store.DB(ctx).Where("id = ? AND version = ?", id, expectedVersion).Delete(&MetaobjectDefinitionModel{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return port.ErrPreconditionFailed
	}
	return nil
}

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
func (repository MetaobjectEntryRepository) Update(ctx context.Context, entry domain.MetaobjectEntry, expectedVersion uint64) (domain.MetaobjectEntry, error) {
	model := metaobjectEntryModelFromDomain(entry)
	result := repository.store.DB(ctx).Model(&MetaobjectEntryModel{}).Where("id = ? AND version = ?", entry.ID, expectedVersion).Updates(map[string]any{
		"display_name": model.DisplayName,
		"field_values": model.Fields,
		"version":      expectedVersion + 1,
	})
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
func (repository MetaobjectEntryRepository) FindByHandle(ctx context.Context, definitionID uuid.UUID, handle domain.Handle) (domain.MetaobjectEntry, error) {
	var model MetaobjectEntryModel
	err := repository.store.DB(ctx).First(&model, "definition_id = ? AND handle = ?", definitionID, handle).Error
	if err != nil {
		return domain.MetaobjectEntry{}, mapError(err)
	}
	return metaobjectEntryFromModel(model)
}

// List returns entries for definition.
func (repository MetaobjectEntryRepository) List(ctx context.Context, definitionID uuid.UUID, page pagination.Page) (pagination.Result[domain.MetaobjectEntry], error) {
	var models []MetaobjectEntryModel
	err := repository.store.DB(ctx).Where("definition_id = ?", definitionID).Order("display_name asc").Limit(page.Limit + 1).Find(&models).Error
	if err != nil {
		return pagination.Result[domain.MetaobjectEntry]{}, err
	}
	return metaobjectEntryPage(models, page.Limit)
}

// Delete soft deletes entry.
func (repository MetaobjectEntryRepository) Delete(ctx context.Context, id uuid.UUID, expectedVersion uint64) error {
	result := repository.store.DB(ctx).Where("id = ? AND version = ?", id, expectedVersion).Delete(&MetaobjectEntryModel{})
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
	err := repository.store.DB(ctx).Model(&MetaobjectEntryModel{}).Where("definition_id = ?", definitionID).Count(&count).Error
	return count, err
}

// applyMetaobjectDefinitionFilter applies definition filters.
func applyMetaobjectDefinitionFilter(query *gorm.DB, filter port.MetaobjectDefinitionFilter) *gorm.DB {
	if filter.Type != "" {
		query = query.Where("type = ?", filter.Type)
	}
	if filter.Active != nil {
		query = query.Where("active = ?", *filter.Active)
	}
	return query
}

// metaobjectDefinitionPage maps definition models into a page.
func metaobjectDefinitionPage(models []MetaobjectDefinitionModel, limit int) (pagination.Result[domain.MetaobjectDefinition], error) {
	next := ""
	if len(models) > limit {
		next = models[limit-1].ID.ID.String()
		models = models[:limit]
	}
	items := make([]domain.MetaobjectDefinition, 0, len(models))
	for _, model := range models {
		item, err := metaobjectDefinitionFromModel(model)
		if err != nil {
			return pagination.Result[domain.MetaobjectDefinition]{}, err
		}
		items = append(items, item)
	}
	return pagination.Result[domain.MetaobjectDefinition]{Items: items, NextCursor: next}, nil
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
