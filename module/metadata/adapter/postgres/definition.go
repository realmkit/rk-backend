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

// MetafieldDefinitionRepository stores definitions in PostgreSQL.
type MetafieldDefinitionRepository struct {
	store orm.Store
}

// NewMetafieldDefinitionRepository creates a metafield definition repository.
func NewMetafieldDefinitionRepository(store orm.Store) MetafieldDefinitionRepository {
	return MetafieldDefinitionRepository{store: store}
}

// Create stores definition.
func (repository MetafieldDefinitionRepository) Create(
	ctx context.Context,
	definition domain.MetafieldDefinition,
) (domain.MetafieldDefinition, error) {
	if _, err := repository.FindByKey(ctx, definition.OwnerType, definition.Key); err == nil {
		return domain.MetafieldDefinition{}, port.ErrConflict
	} else if !errors.Is(err, port.ErrNotFound) {
		return domain.MetafieldDefinition{}, err
	}
	model := definitionModelFromDomain(definition)
	if model.Version == 0 {
		model.Version = 1
	}
	if err := repository.store.DB(ctx).Create(&model).Error; err != nil {
		return domain.MetafieldDefinition{}, mapCreateError(err)
	}
	return definitionFromModel(model)
}

// Update stores mutable definition changes.
func (repository MetafieldDefinitionRepository) Update(
	ctx context.Context,
	definition domain.MetafieldDefinition,
	expectedVersion uint64,
) (domain.MetafieldDefinition, error) {
	model := definitionModelFromDomain(definition)
	result := repository.store.DB(ctx).
		Model(&MetafieldDefinitionModel{}).
		Where("id = ? AND version = ?", definition.ID, expectedVersion).
		Updates(map[string]any{
			"name":        model.Name,
			"description": model.Description,
			"is_required": model.Required,
			"rules":       model.Rules,
			"sort_order":  model.SortOrder,
			"active":      model.Active,
			"version":     expectedVersion + 1,
		})
	if result.Error != nil {
		return domain.MetafieldDefinition{}, result.Error
	}
	if result.RowsAffected == 0 {
		return domain.MetafieldDefinition{}, port.ErrPreconditionFailed
	}
	return repository.FindByID(ctx, definition.ID)
}

// FindByID returns one definition by ID.
func (repository MetafieldDefinitionRepository) FindByID(ctx context.Context, id uuid.UUID) (domain.MetafieldDefinition, error) {
	var model MetafieldDefinitionModel
	if err := repository.store.DB(ctx).First(&model, "id = ?", id).Error; err != nil {
		return domain.MetafieldDefinition{}, mapError(err)
	}
	return definitionFromModel(model)
}

// FindByKey returns one definition by owner type and key.
func (repository MetafieldDefinitionRepository) FindByKey(
	ctx context.Context,
	ownerType domain.OwnerType,
	key domain.Key,
) (domain.MetafieldDefinition, error) {
	var model MetafieldDefinitionModel
	err := repository.store.DB(ctx).First(&model, "owner_type = ? AND key = ?", ownerType, key).Error
	if err != nil {
		return domain.MetafieldDefinition{}, mapError(err)
	}
	return definitionFromModel(model)
}

// List returns definitions matching filter.
func (repository MetafieldDefinitionRepository) List(
	ctx context.Context,
	filter port.DefinitionFilter,
	page pagination.Page,
) (pagination.Result[domain.MetafieldDefinition], error) {
	query := repository.store.DB(ctx).Model(&MetafieldDefinitionModel{}).Order("sort_order asc, created_at asc").Limit(page.Limit + 1)
	query = applyDefinitionFilter(query, filter)
	var models []MetafieldDefinitionModel
	if err := query.Find(&models).Error; err != nil {
		return pagination.Result[domain.MetafieldDefinition]{}, err
	}
	return definitionPage(models, page.Limit)
}

// Archive soft deletes definition.
func (repository MetafieldDefinitionRepository) Archive(ctx context.Context, id uuid.UUID, expectedVersion uint64) error {
	result := repository.store.DB(ctx).Where("id = ? AND version = ?", id, expectedVersion).Delete(&MetafieldDefinitionModel{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return port.ErrPreconditionFailed
	}
	return nil
}

// applyDefinitionFilter applies definition filters.
func applyDefinitionFilter(query *gorm.DB, filter port.DefinitionFilter) *gorm.DB {
	if filter.OwnerType != "" {
		query = query.Where("owner_type = ?", filter.OwnerType)
	}
	if filter.Active != nil {
		query = query.Where("active = ?", *filter.Active)
	}
	return query
}

// definitionPage maps definition models into a page.
func definitionPage(models []MetafieldDefinitionModel, limit int) (pagination.Result[domain.MetafieldDefinition], error) {
	next := ""
	if len(models) > limit {
		next = models[limit-1].ID.ID.String()
		models = models[:limit]
	}
	items := make([]domain.MetafieldDefinition, 0, len(models))
	for _, model := range models {
		item, err := definitionFromModel(model)
		if err != nil {
			return pagination.Result[domain.MetafieldDefinition]{}, err
		}
		items = append(items, item)
	}
	return pagination.Result[domain.MetafieldDefinition]{Items: items, NextCursor: next}, nil
}

// mapError maps GORM errors into metadata errors.
func mapError(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return port.ErrNotFound
	}
	return err
}

// mapCreateError maps insert errors into metadata errors.
func mapCreateError(err error) error {
	translated := orm.TranslateError(err)
	if errors.Is(translated, orm.ErrConflict) {
		return port.ErrConflict
	}
	return translated
}
