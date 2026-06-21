package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/module/punishments/domain"
	"github.com/realmkit/rk-backend/module/punishments/port"
	"github.com/realmkit/rk-backend/pkg/orm"
	"github.com/realmkit/rk-backend/pkg/pagination"
	"github.com/realmkit/rk-backend/pkg/search"
	"gorm.io/gorm"
)

// DefinitionRepository stores punishment definitions.
type DefinitionRepository struct {
	store orm.Store // store stores the store value.
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
	next := ""
	if len(models) > page.Limit {
		next, err = definitionCursor(models[page.Limit-1], definitionFilterHash(filter, sort), sort)
		if err != nil {
			return pagination.Result[domain.Definition]{}, err
		}
		models = models[:page.Limit]
	}
	items := make([]domain.Definition, 0, len(models))
	for _, model := range models {
		items = append(items, definitionFromModel(model, repository.actions(ctx, model.ID.ID)))
	}
	return pagination.Result[domain.Definition]{Items: items, NextCursor: next}, nil
}

// applyDefinitionFilter applies definition list filters.
func applyDefinitionFilter(query *gorm.DB, filter port.DefinitionFilter) *gorm.DB {
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

// applyDefinitionCursor applies keyset filtering.
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

// definitionOrder returns deterministic ordering SQL.
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
	case "severity":
		return "severity"
	case "created_at":
		return "created_at"
	default:
		return "display_order"
	}
}

// definitionCursor returns an encoded definition cursor.
func definitionCursor(model DefinitionModel, filterHash string, sort search.Sort) (string, error) {
	return search.EncodeCursor(search.Cursor{
		FilterHash: filterHash,
		Sort:       sort.Key,
		Direction:  sort.Direction,
		Values:     []string{definitionModelSortValue(model, sort.Key)},
		ID:         model.ID.ID.String(),
	})
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

// actions supports package behavior.
func (repository DefinitionRepository) actions(ctx context.Context, definitionID uuid.UUID) []ActionModel {
	var actions []ActionModel
	_ = repository.store.DB(ctx).Where("definition_id = ?", definitionID).Order("display_order, id").Find(&actions).Error
	return actions
}

// DefinitionRepositoryContract verifies repository interface conformance.
var _ port.DefinitionRepository = DefinitionRepository{}
