package postgres

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/module/tickets/domain"
	"github.com/realmkit/rk-backend/module/tickets/port"
	"github.com/realmkit/rk-backend/pkg/orm"
	"github.com/realmkit/rk-backend/pkg/pagination"
	"github.com/realmkit/rk-backend/pkg/search"
	"gorm.io/gorm"
)

// TicketRepository stores ticket cases and their child records.
type TicketRepository struct {
	store orm.Store
}

// NewTicketRepository creates a ticket repository.
func NewTicketRepository(store orm.Store) TicketRepository {
	return TicketRepository{store: store}
}

// Create stores one ticket with its opener and initial evidence.
func (repository TicketRepository) Create(
	ctx context.Context,
	ticket domain.Ticket,
	opener domain.Message,
	evidence []domain.Evidence,
) (domain.Ticket, error) {
	db := repository.store.DB(ctx)
	model := ticketModel(ticket.Normalize())
	opener.Sequence = 1
	if err := db.Create(&model).Error; err != nil {
		return domain.Ticket{}, port.ErrConflict
	}
	opener.TicketID = model.ID.ID
	if err := db.Create(pointer(messageModel(opener))).Error; err != nil {
		return domain.Ticket{}, err
	}
	for _, item := range evidence {
		item.TicketID = model.ID.ID
		if err := db.Create(pointer(evidenceModel(item))).Error; err != nil {
			return domain.Ticket{}, err
		}
	}
	return ticketFromModel(model), nil
}

// Update stores mutable ticket fields using optimistic concurrency.
func (repository TicketRepository) Update(ctx context.Context, ticket domain.Ticket, expectedVersion uint64) (domain.Ticket, error) {
	model := ticketModel(ticket.Normalize())
	updates := ticketUpdates(model, expectedVersion)
	result := repository.store.DB(ctx).Model(&TicketModel{}).
		Where("id = ? AND version = ?", ticket.ID, expectedVersion).
		Updates(updates)
	if result.Error != nil {
		return domain.Ticket{}, result.Error
	}
	if result.RowsAffected == 0 {
		return domain.Ticket{}, port.ErrPreconditionFailed
	}
	return repository.FindByID(ctx, ticket.ID)
}

// FindByID returns one ticket.
func (repository TicketRepository) FindByID(ctx context.Context, id uuid.UUID) (domain.Ticket, error) {
	var model TicketModel
	if err := repository.store.DB(ctx).First(&model, "id = ?", id).Error; err != nil {
		return domain.Ticket{}, mapError(err)
	}
	return ticketFromModel(model), nil
}

// FindByIdempotencyKey returns a ticket opened by the same idempotency key.
func (repository TicketRepository) FindByIdempotencyKey(ctx context.Context, key string) (domain.Ticket, error) {
	var model TicketModel
	if err := repository.store.DB(ctx).First(&model, "key = ?", key).Error; err != nil {
		return domain.Ticket{}, mapError(err)
	}
	return ticketFromModel(model), nil
}

// List returns a filtered ticket page.
func (repository TicketRepository) List(
	ctx context.Context,
	filter port.TicketFilter,
	page pagination.Page,
) (pagination.Result[domain.Ticket], error) {
	sort := filter.Sort
	if sort.Key == "" {
		sort, _ = search.NewSort("", "", port.DefaultTicketSort(), port.AllowedTicketSorts())
	}
	filterHash := ticketFilterHash(filter, sort)
	cursor, hasCursor, err := search.RequireCursor(page.Cursor, filterHash, sort)
	if err != nil {
		return pagination.Result[domain.Ticket]{}, err
	}
	query := ticketFilter(repository.store.DB(ctx).Model(&TicketModel{}), filter)
	query, err = applyTicketCursor(query, cursor, hasCursor, sort)
	if err != nil {
		return pagination.Result[domain.Ticket]{}, err
	}
	query = query.Order(ticketOrder(sort)).Limit(page.Limit + 1)
	var models []TicketModel
	if err := query.Find(&models).Error; err != nil {
		return pagination.Result[domain.Ticket]{}, err
	}
	return ticketPage(models, page.Limit, filterHash, sort)
}

// ticketFilter applies ticket filters.
func ticketFilter(query *gorm.DB, filter port.TicketFilter) *gorm.DB {
	if filter.SubmitterUserID != uuid.Nil {
		query = query.Where("submitter_user_id = ?", filter.SubmitterUserID)
	}
	if filter.TargetUserID != uuid.Nil {
		query = query.Where("target_user_id = ?", filter.TargetUserID)
	}
	if filter.PunishmentID != uuid.Nil {
		query = query.Where("punishment_id = ?", filter.PunishmentID)
	}
	if filter.CurrentTeamGroupID != uuid.Nil {
		query = query.Where("current_team_group_id = ?", filter.CurrentTeamGroupID)
	}
	if filter.AssigneeUserID != uuid.Nil {
		query = query.Where("assignee_user_id = ?", filter.AssigneeUserID)
	}
	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}
	if filter.Kind != "" {
		query = query.Where("kind = ?", filter.Kind)
	}
	if !filter.Query.Empty() {
		if query.Dialector.Name() == "postgres" {
			query = query.Where(ticketPostgresSearchCondition(), filter.Query.String(), filter.Query.LowerLike())
		} else {
			like := filter.Query.LowerLike()
			query = query.Where(ticketSearchCondition(), like, like, like, like)
		}
	}
	return query
}

// ticketPage maps ticket models into a page.
func ticketPage(models []TicketModel, limit int, filterHash string, sort search.Sort) (pagination.Result[domain.Ticket], error) {
	next := ""
	if len(models) > limit {
		cursor, err := ticketCursor(models[limit-1], filterHash, sort)
		if err != nil {
			return pagination.Result[domain.Ticket]{}, err
		}
		next = cursor
		models = models[:limit]
	}
	items := make([]domain.Ticket, 0, len(models))
	for _, model := range models {
		items = append(items, ticketFromModel(model))
	}
	return pagination.Result[domain.Ticket]{Items: items, NextCursor: next}, nil
}

// applyTicketCursor applies keyset cursor filtering.
func applyTicketCursor(query *gorm.DB, cursor search.Cursor, ok bool, sort search.Sort) (*gorm.DB, error) {
	if !ok || len(cursor.Values) == 0 {
		return query, nil
	}
	id, err := uuid.Parse(cursor.ID)
	if err != nil {
		return nil, search.ErrInvalidCursor
	}
	column := ticketSortColumn(sort.Key)
	value := ticketCursorValue(cursor.Values[0], sort.Key)
	if sort.Desc() {
		return query.Where(column+" < ? OR ("+column+" = ? AND id > ?)", value, value, id), nil
	}
	return query.Where(column+" > ? OR ("+column+" = ? AND id > ?)", value, value, id), nil
}

// ticketOrder returns deterministic ticket ordering.
func ticketOrder(sort search.Sort) string {
	direction := "ASC"
	if sort.Desc() {
		direction = "DESC"
	}
	return ticketSortColumn(sort.Key) + " " + direction + ", id ASC"
}

// ticketSortColumn maps public sort keys to columns.
func ticketSortColumn(key string) string {
	switch key {
	case "created_at":
		return "created_at"
	case "priority":
		return "priority"
	case "title":
		return "title"
	default:
		return "updated_at"
	}
}

// ticketCursor returns an encoded ticket cursor.
func ticketCursor(model TicketModel, filterHash string, sort search.Sort) (string, error) {
	return search.EncodeCursor(search.Cursor{
		FilterHash: filterHash,
		Sort:       sort.Key,
		Direction:  sort.Direction,
		Values:     []string{ticketModelSortValue(model, sort.Key)},
		ID:         model.ID.ID.String(),
	})
}

// ticketModelSortValue returns the cursor value.
func ticketModelSortValue(model TicketModel, key string) string {
	switch key {
	case "created_at":
		return model.CreatedAt.Format(time.RFC3339Nano)
	case "priority":
		return model.Priority
	case "title":
		return model.Title
	default:
		return model.UpdatedAt.Format(time.RFC3339Nano)
	}
}

// ticketCursorValue converts cursor text to the matching SQL type.
func ticketCursorValue(value string, key string) any {
	if key == "created_at" || key == "updated_at" || key == "" {
		parsed, _ := time.Parse(time.RFC3339Nano, value)
		return parsed
	}
	return value
}

// pointer returns a pointer to value for GORM create calls.
func pointer[T any](value T) *T { return &value }

// now returns UTC wall clock time for persistence summaries.
func now() time.Time { return time.Now().UTC() }
