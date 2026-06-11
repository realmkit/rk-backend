package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/module/metadata/domain"
	"github.com/realmkit/rk-backend/module/metadata/port"
	"github.com/realmkit/rk-backend/pkg/orm"
	"github.com/realmkit/rk-backend/pkg/pagination"
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
func (repository MetaobjectDefinitionRepository) Create(
	ctx context.Context,
	definition domain.MetaobjectDefinition,
) (domain.MetaobjectDefinition, error) {
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
func (repository MetaobjectDefinitionRepository) Update(
	ctx context.Context,
	definition domain.MetaobjectDefinition,
	expectedVersion uint64,
) (domain.MetaobjectDefinition, error) {
	model := metaobjectDefinitionModelFromDomain(definition)
	result := repository.store.DB(ctx).
		Model(&MetaobjectDefinitionModel{}).
		Where("id = ? AND version = ?", definition.ID, expectedVersion).
		Updates(map[string]any{
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
func (repository MetaobjectDefinitionRepository) FindByType(
	ctx context.Context,
	objectType domain.MetaobjectType,
) (domain.MetaobjectDefinition, error) {
	var model MetaobjectDefinitionModel
	err := repository.store.DB(ctx).First(&model, "type = ?", objectType).Error
	if err != nil {
		return domain.MetaobjectDefinition{}, mapError(err)
	}
	return metaobjectDefinitionFromModel(model)
}

// List returns definitions matching filter.
func (repository MetaobjectDefinitionRepository) List(
	ctx context.Context,
	filter port.MetaobjectDefinitionFilter,
	page pagination.Page,
) (pagination.Result[domain.MetaobjectDefinition], error) {
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
