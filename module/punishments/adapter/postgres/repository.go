package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/module/punishments/domain"
	"github.com/realmkit/rk-backend/module/punishments/port"
	"github.com/realmkit/rk-backend/pkg/orm"
	"github.com/realmkit/rk-backend/pkg/pagination"
	"gorm.io/gorm"
)

// DefinitionRepository stores punishment definitions.
type DefinitionRepository struct {
	store orm.Store
}

// NewDefinitionRepository creates a punishment definition repository.
func NewDefinitionRepository(store orm.Store) DefinitionRepository {
	return DefinitionRepository{store: store}
}

// Create stores a definition and action templates.
func (repository DefinitionRepository) Create(ctx context.Context, definition domain.Definition) (domain.Definition, error) {
	model := definitionModel(definition)
	if err := repository.store.DB(ctx).Create(&model).Error; err != nil {
		return domain.Definition{}, port.ErrConflict
	}
	for index, action := range definition.Actions {
		action.DefinitionID = model.ID.ID
		action.DisplayOrder = index
		model := actionModel(action)
		if err := repository.store.DB(ctx).Create(&model).Error; err != nil {
			return domain.Definition{}, err
		}
	}
	return repository.FindByID(ctx, model.ID.ID)
}

// Update updates a definition.
func (repository DefinitionRepository) Update(
	ctx context.Context,
	definition domain.Definition,
	expectedVersion uint64,
) (domain.Definition, error) {
	result := repository.store.DB(ctx).Model(&DefinitionModel{}).
		Where("id = ? AND version = ?", definition.ID, expectedVersion).
		Updates(map[string]any{
			"name": definition.Name, "description": definition.Description,
			"color": string(definition.Color), "severity": definition.Severity,
			"status": string(definition.Status), "allow_permanent": definition.AllowPermanent,
			"requires_reason": definition.RequiresReason, "requires_target_ip": definition.RequiresTargetIP,
			"default_duration_seconds": definition.DefaultDurationSeconds,
			"min_duration_seconds":     definition.MinDurationSeconds,
			"max_duration_seconds":     definition.MaxDurationSeconds,
			"display_order":            definition.DisplayOrder,
			"version":                  expectedVersion + 1,
		})
	if result.Error != nil {
		return domain.Definition{}, result.Error
	}
	if result.RowsAffected == 0 {
		return domain.Definition{}, port.ErrPreconditionFailed
	}
	return repository.FindByID(ctx, definition.ID)
}

// Delete soft deletes a definition.
func (repository DefinitionRepository) Delete(ctx context.Context, id uuid.UUID, expectedVersion uint64) error {
	result := repository.store.DB(ctx).Where("id = ? AND version = ?", id, expectedVersion).Delete(&DefinitionModel{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return port.ErrPreconditionFailed
	}
	return nil
}

// FindByID returns one definition.
func (repository DefinitionRepository) FindByID(ctx context.Context, id uuid.UUID) (domain.Definition, error) {
	var model DefinitionModel
	if err := repository.store.DB(ctx).First(&model, "id = ?", id).Error; err != nil {
		return domain.Definition{}, mapError(err)
	}
	return definitionFromModel(model, repository.actions(ctx, id)), nil
}

// List returns definitions.
func (repository DefinitionRepository) List(
	ctx context.Context,
	filter port.DefinitionFilter,
	page pagination.Page,
) (pagination.Result[domain.Definition], error) {
	query := repository.store.DB(ctx).Model(&DefinitionModel{}).Order("display_order, severity desc, name, id").Limit(page.Limit + 1)
	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}
	var models []DefinitionModel
	if err := query.Find(&models).Error; err != nil {
		return pagination.Result[domain.Definition]{}, err
	}
	next := ""
	if len(models) > page.Limit {
		next = models[page.Limit-1].ID.ID.String()
		models = models[:page.Limit]
	}
	items := make([]domain.Definition, 0, len(models))
	for _, model := range models {
		items = append(items, definitionFromModel(model, repository.actions(ctx, model.ID.ID)))
	}
	return pagination.Result[domain.Definition]{Items: items, NextCursor: next}, nil
}

// ReorderActions updates action display order.
func (repository DefinitionRepository) ReorderActions(ctx context.Context, definitionID uuid.UUID, actionIDs []uuid.UUID) error {
	for index, id := range actionIDs {
		result := repository.store.DB(ctx).Model(&ActionModel{}).
			Where("id = ? AND definition_id = ?", id, definitionID).
			Update("display_order", index)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return port.ErrNotFound
		}
	}
	return nil
}

func (repository DefinitionRepository) actions(ctx context.Context, definitionID uuid.UUID) []ActionModel {
	var actions []ActionModel
	_ = repository.store.DB(ctx).Where("definition_id = ?", definitionID).Order("display_order, id").Find(&actions).Error
	return actions
}

// mapError translates GORM and ORM errors to punishment ports.
func mapError(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return port.ErrNotFound
	}
	if errors.Is(orm.TranslateError(err), orm.ErrConflict) {
		return port.ErrConflict
	}
	return err
}

var _ port.DefinitionRepository = DefinitionRepository{}
