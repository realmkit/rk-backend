package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/pkg/events/domain"
	"github.com/realmkit/rk-backend/pkg/events/port"
	"github.com/realmkit/rk-backend/pkg/orm"
	"github.com/realmkit/rk-backend/pkg/pagination"
	"gorm.io/gorm"
)

// Repository stores events in PostgreSQL.
type Repository struct {
	store orm.Store // store stores the store value.
}

// NewRepository creates an event repository.
func NewRepository(store orm.Store) Repository {
	return Repository{store: store}
}

// Publish stores one event and its scopes.
func (repository Repository) Publish(ctx context.Context, draft domain.Draft, now time.Time) (domain.Event, error) {
	model, scopes, err := modelFromDraft(draft, now)
	if err != nil {
		return domain.Event{}, err
	}
	db := repository.store.DB(ctx)
	if err := db.Create(&model).Error; err != nil {
		return domain.Event{}, translate(err)
	}
	if len(scopes) > 0 {
		if err := db.Create(&scopes).Error; err != nil {
			return domain.Event{}, translate(err)
		}
	}
	return eventFromModel(model, scopes), nil
}

// Get returns one event by id.
func (repository Repository) Get(ctx context.Context, id uuid.UUID) (domain.Event, error) {
	var model EventModel
	if err := repository.store.DB(ctx).Where("id = ?", id).First(&model).Error; err != nil {
		return domain.Event{}, translate(err)
	}
	scopes, err := repository.scopes(ctx, []uuid.UUID{id})
	if err != nil {
		return domain.Event{}, err
	}
	return eventFromModel(model, scopes[id]), nil
}

// List returns matching events.
func (repository Repository) List(
	ctx context.Context,
	filter port.ListFilter,
	page pagination.Page,
) (pagination.Result[domain.Event], error) {
	var models []EventModel
	query := repository.filter(repository.store.DB(ctx), filter)
	if err := query.Order("occurred_at DESC, id DESC").Limit(page.Limit).Find(&models).Error; err != nil {
		return pagination.Result[domain.Event]{}, translate(err)
	}
	return repository.result(ctx, models)
}

// Claim claims due events.
func (repository Repository) Claim(
	ctx context.Context,
	workerID string,
	limit int,
	now time.Time,
	lockUntil time.Time,
) ([]domain.Event, error) {
	var models []EventModel
	db := repository.store.DB(ctx)
	err := db.Where("status IN ? AND available_at <= ?", []string{"pending", "failed"}, now).
		Order("available_at ASC, id ASC").
		Limit(limit).
		Find(&models).Error
	if err != nil {
		return nil, translate(err)
	}
	claimed := make([]EventModel, 0, len(models))
	for _, model := range models {
		updates := map[string]any{
			"status":        string(domain.StatusProcessing),
			"locked_by":     workerID,
			"locked_until":  lockUntil,
			"attempt_count": gorm.Expr("attempt_count + 1"),
			"updated_at":    now,
		}
		result := db.Model(&EventModel{}).
			Where("id = ? AND status IN ?", model.ID, []string{"pending", "failed"}).
			Updates(updates)
		if result.Error != nil {
			return nil, translate(result.Error)
		}
		if result.RowsAffected == 1 {
			model.Status = string(domain.StatusProcessing)
			model.LockedBy = workerID
			model.LockedUntil = &lockUntil
			model.AttemptCount++
			claimed = append(claimed, model)
		}
	}
	return repository.events(ctx, claimed)
}

// MarkProcessed marks an event processed.
func (repository Repository) MarkProcessed(ctx context.Context, id uuid.UUID, now time.Time) error {
	return repository.updateStatus(ctx, id, domain.StatusProcessed, now, map[string]any{"processed_at": now})
}

// MarkFailed marks an event failed and schedules retry.
func (repository Repository) MarkFailed(ctx context.Context, id uuid.UUID, message string, availableAt time.Time, now time.Time) error {
	return repository.updateStatus(ctx, id, domain.StatusFailed, now, map[string]any{"last_error": message, "available_at": availableAt})
}

// MarkDead marks an event dead-lettered.
func (repository Repository) MarkDead(ctx context.Context, id uuid.UUID, message string, now time.Time) error {
	return repository.updateStatus(ctx, id, domain.StatusDead, now, map[string]any{"last_error": message, "dead_at": now})
}

// Replay moves one event back to pending state.
func (repository Repository) Replay(ctx context.Context, id uuid.UUID, now time.Time) error {
	return repository.updateStatus(ctx, id, domain.StatusPending, now, map[string]any{"available_at": now, "last_error": ""})
}

// Cancel cancels one event.
func (repository Repository) Cancel(ctx context.Context, id uuid.UUID, now time.Time) error {
	return repository.updateStatus(ctx, id, domain.StatusCancelled, now, nil)
}

// updateStatus updates common status fields.
func (repository Repository) updateStatus(
	ctx context.Context,
	id uuid.UUID,
	status domain.Status,
	now time.Time,
	extra map[string]any,
) error {
	updates := map[string]any{"status": string(status), "locked_by": "", "locked_until": nil, "updated_at": now}
	for key, value := range extra {
		updates[key] = value
	}
	result := repository.store.DB(ctx).Model(&EventModel{}).Where("id = ?", id).Updates(updates)
	if result.Error != nil {
		return translate(result.Error)
	}
	if result.RowsAffected == 0 {
		return port.ErrNotFound
	}
	return nil
}

// filter applies list filters.
func (repository Repository) filter(db *gorm.DB, filter port.ListFilter) *gorm.DB {
	if filter.Status != "" {
		db = db.Where("status = ?", filter.Status)
	}
	if filter.Producer != "" {
		db = db.Where("producer = ?", filter.Producer)
	}
	if filter.EventKey != "" {
		db = db.Where("event_key = ?", filter.EventKey)
	}
	if filter.AggregateType != "" {
		db = db.Where("aggregate_type = ?", filter.AggregateType)
	}
	if filter.AggregateID != nil {
		db = db.Where("aggregate_id = ?", *filter.AggregateID)
	}
	return db
}

// result maps models into a paginated result.
func (repository Repository) result(ctx context.Context, models []EventModel) (pagination.Result[domain.Event], error) {
	events, err := repository.events(ctx, models)
	if err != nil {
		return pagination.Result[domain.Event]{}, err
	}
	return pagination.Result[domain.Event]{Items: events}, nil
}

// events maps event models with their scopes.
func (repository Repository) events(ctx context.Context, models []EventModel) ([]domain.Event, error) {
	ids := make([]uuid.UUID, 0, len(models))
	for _, model := range models {
		ids = append(ids, model.ID)
	}
	scopes, err := repository.scopes(ctx, ids)
	if err != nil {
		return nil, err
	}
	events := make([]domain.Event, 0, len(models))
	for _, model := range models {
		events = append(events, eventFromModel(model, scopes[model.ID]))
	}
	return events, nil
}

// scopes returns scopes grouped by event id.
func (repository Repository) scopes(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID][]ScopeModel, error) {
	var models []ScopeModel
	if len(ids) > 0 {
		if err := repository.store.DB(ctx).Where("event_id IN ?", ids).Find(&models).Error; err != nil {
			return nil, translate(err)
		}
	}
	grouped := map[uuid.UUID][]ScopeModel{}
	for _, model := range models {
		grouped[model.EventID] = append(grouped[model.EventID], model)
	}
	return grouped, nil
}

// translate maps persistence errors.
func translate(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(orm.TranslateError(err), orm.ErrNotFound):
		return port.ErrNotFound
	case errors.Is(orm.TranslateError(err), orm.ErrConflict):
		return port.ErrConflict
	default:
		return err
	}
}
