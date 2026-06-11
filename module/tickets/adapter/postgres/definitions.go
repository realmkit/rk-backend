package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/module/tickets/domain"
	"github.com/realmkit/rk-backend/module/tickets/port"
	"github.com/realmkit/rk-backend/pkg/orm"
	"github.com/realmkit/rk-backend/pkg/pagination"
	"gorm.io/gorm"
)

// DefinitionRepository stores ticket definitions.
type DefinitionRepository struct {
	store orm.Store
}

// NewDefinitionRepository creates a definition repository.
func NewDefinitionRepository(store orm.Store) DefinitionRepository {
	return DefinitionRepository{store: store}
}

// Create stores one definition.
func (repository DefinitionRepository) Create(ctx context.Context, definition domain.Definition) (domain.Definition, error) {
	model := definitionModel(definition.Normalize())
	if err := repository.store.DB(ctx).Create(&model).Error; err != nil {
		return domain.Definition{}, port.ErrConflict
	}
	return definitionFromModel(model), nil
}

// Update stores mutable definition fields using optimistic concurrency.
func (repository DefinitionRepository) Update(
	ctx context.Context,
	definition domain.Definition,
	expectedVersion uint64,
) (domain.Definition, error) {
	model := definitionModel(definition.Normalize())
	updates := map[string]any{
		"key":                        model.Key,
		"name":                       model.Name,
		"description":                model.Description,
		"kind":                       model.Kind,
		"status":                     model.Status,
		"default_team_group_id":      model.DefaultTeamGroupID,
		"default_assignee_user_id":   model.DefaultAssigneeUserID,
		"submitter_can_close":        model.SubmitterCanClose,
		"submitter_can_reopen":       model.SubmitterCanReopen,
		"allow_anonymous_submitter":  model.AllowAnonymousSubmitter,
		"requires_target_user":       model.RequiresTargetUser,
		"requires_punishment":        model.RequiresPunishment,
		"requires_evidence":          model.RequiresEvidence,
		"max_open_per_submitter":     model.MaxOpenPerSubmitter,
		"reopen_window_seconds":      model.ReopenWindowSeconds,
		"sla_first_response_seconds": model.SLAFirstResponseSeconds,
		"sla_resolution_seconds":     model.SLAResolutionSeconds,
		"metadata_schema_key":        model.MetadataSchemaKey,
		"display_order":              model.DisplayOrder,
		"version":                    expectedVersion + 1,
	}
	result := repository.store.DB(ctx).Model(&DefinitionModel{}).
		Where("id = ? AND version = ?", definition.ID, expectedVersion).
		Updates(updates)
	if result.Error != nil {
		return domain.Definition{}, result.Error
	}
	if result.RowsAffected == 0 {
		return domain.Definition{}, port.ErrPreconditionFailed
	}
	return repository.FindByID(ctx, definition.ID)
}

// Delete soft deletes one definition.
func (repository DefinitionRepository) Delete(ctx context.Context, id uuid.UUID, expectedVersion uint64) error {
	result := repository.store.DB(ctx).
		Where("id = ? AND version = ?", id, expectedVersion).
		Delete(&DefinitionModel{})
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
	return definitionFromModel(model), nil
}

// List returns matching definitions.
func (repository DefinitionRepository) List(
	ctx context.Context,
	filter port.DefinitionFilter,
	page pagination.Page,
) (pagination.Result[domain.Definition], error) {
	query := repository.store.DB(ctx).Model(&DefinitionModel{}).
		Order("display_order asc, id asc").Limit(page.Limit + 1)
	if filter.Kind != "" {
		query = query.Where("kind = ?", filter.Kind)
	}
	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}
	var models []DefinitionModel
	if err := query.Find(&models).Error; err != nil {
		return pagination.Result[domain.Definition]{}, err
	}
	return definitionPage(models, page.Limit), nil
}

// definitionPage maps rows into a paginated result.
func definitionPage(models []DefinitionModel, limit int) pagination.Result[domain.Definition] {
	next := ""
	if len(models) > limit {
		next = models[limit-1].ID.ID.String()
		models = models[:limit]
	}
	items := make([]domain.Definition, 0, len(models))
	for _, model := range models {
		items = append(items, definitionFromModel(model))
	}
	return pagination.Result[domain.Definition]{Items: items, NextCursor: next}
}

// mapError converts GORM sentinel errors to ticket port errors.
func mapError(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return port.ErrNotFound
	}
	return err
}
