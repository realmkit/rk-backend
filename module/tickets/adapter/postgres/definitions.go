package postgres

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/module/tickets/domain"
	"github.com/realmkit/rk-backend/module/tickets/port"
	"github.com/realmkit/rk-backend/pkg/orm"
	"github.com/realmkit/rk-backend/pkg/pagination"
	"github.com/realmkit/rk-backend/pkg/search"
	"gorm.io/gorm"
)

// DefinitionRepository stores ticket definitions.
type DefinitionRepository struct {
	store orm.Store // store stores the store value.
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
	sort := filter.Sort
	if sort.Key == "" {
		sort, _ = search.NewSort("", "", port.DefaultDefinitionSort(), port.AllowedDefinitionSorts())
	}
	filterHash := definitionFilterHash(filter, sort)
	cursor, hasCursor, err := search.RequireCursor(page.Cursor, filterHash, sort)
	if err != nil {
		return pagination.Result[domain.Definition]{}, err
	}
	query := applyDefinitionFilter(repository.store.DB(ctx).Model(&DefinitionModel{}), filter)
	query, err = applyDefinitionCursor(query, cursor, hasCursor, sort)
	if err != nil {
		return pagination.Result[domain.Definition]{}, err
	}
	query = query.Order(definitionOrder(sort)).Limit(page.Limit + 1)
	var models []DefinitionModel
	if err := query.Find(&models).Error; err != nil {
		return pagination.Result[domain.Definition]{}, err
	}
	return definitionPage(models, page.Limit, filterHash, sort)
}

// definitionPage maps rows into a paginated result.
func definitionPage(models []DefinitionModel, limit int, filterHash string, sort search.Sort) (pagination.Result[domain.Definition], error) {
	next := ""
	if len(models) > limit {
		cursor, err := definitionCursor(models[limit-1], filterHash, sort)
		if err != nil {
			return pagination.Result[domain.Definition]{}, err
		}
		next = cursor
		models = models[:limit]
	}
	items := make([]domain.Definition, 0, len(models))
	for _, model := range models {
		items = append(items, definitionFromModel(model))
	}
	return pagination.Result[domain.Definition]{Items: items, NextCursor: next}, nil
}

// applyDefinitionFilter applies ticket definition list filters.
func applyDefinitionFilter(query *gorm.DB, filter port.DefinitionFilter) *gorm.DB {
	if filter.Kind != "" {
		query = query.Where("kind = ?", filter.Kind)
	}
	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}
	if !filter.Query.Empty() {
		if query.Dialector.Name() == "postgres" {
			query = query.Where("to_tsvector('simple', coalesce(key, '') || ' ' || coalesce(name, '') || ' ' || coalesce(description, '')) @@ plainto_tsquery('simple', ?)", filter.Query.String())
		} else {
			like := filter.Query.LowerLike()
			query = query.Where("LOWER(key) LIKE ? OR LOWER(name) LIKE ? OR LOWER(description) LIKE ?", like, like, like)
		}
	}
	return query
}

// applyDefinitionCursor applies keyset cursor filtering.
func applyDefinitionCursor(query *gorm.DB, cursor search.Cursor, ok bool, sort search.Sort) (*gorm.DB, error) {
	if !ok || len(cursor.Values) == 0 {
		return query, nil
	}
	id, err := uuid.Parse(cursor.ID)
	if err != nil {
		return nil, search.ErrInvalidCursor
	}
	column := definitionSortColumn(sort.Key)
	value := definitionCursorValue(cursor.Values[0], sort.Key)
	if sort.Desc() {
		return query.Where(column+" < ? OR ("+column+" = ? AND id > ?)", value, value, id), nil
	}
	return query.Where(column+" > ? OR ("+column+" = ? AND id > ?)", value, value, id), nil
}

// definitionOrder returns deterministic ticket definition ordering.
func definitionOrder(sort search.Sort) string {
	direction := "ASC"
	if sort.Desc() {
		direction = "DESC"
	}
	return definitionSortColumn(sort.Key) + " " + direction + ", id ASC"
}

// definitionSortColumn maps public sort keys to columns.
func definitionSortColumn(key string) string {
	switch key {
	case "name":
		return "name"
	case "created_at":
		return "created_at"
	default:
		return "display_order"
	}
}

// definitionCursor returns an encoded ticket definition cursor.
func definitionCursor(model DefinitionModel, filterHash string, sort search.Sort) (string, error) {
	return search.EncodeCursor(search.Cursor{
		FilterHash: filterHash,
		Sort:       sort.Key,
		Direction:  sort.Direction,
		Values:     []string{definitionModelSortValue(model, sort.Key)},
		ID:         model.ID.ID.String(),
	})
}

// definitionModelSortValue returns the definition cursor value.
func definitionModelSortValue(model DefinitionModel, key string) string {
	switch key {
	case "name":
		return model.Name
	case "created_at":
		return model.CreatedAt.Format(time.RFC3339Nano)
	default:
		return strconv.Itoa(model.DisplayOrder)
	}
}

// definitionCursorValue converts a cursor value to the matching SQL type.
func definitionCursorValue(value string, key string) any {
	if key == "created_at" {
		parsed, _ := time.Parse(time.RFC3339Nano, value)
		return parsed
	}
	if key == "display_order" || key == "" {
		parsed, _ := strconv.Atoi(value)
		return parsed
	}
	return value
}

// mapError converts GORM sentinel errors to ticket port errors.
func mapError(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return port.ErrNotFound
	}
	return err
}
