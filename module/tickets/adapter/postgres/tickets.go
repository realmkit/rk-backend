package postgres

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/niflaot/gamehub-go/module/tickets/domain"
	"github.com/niflaot/gamehub-go/module/tickets/port"
	"github.com/niflaot/gamehub-go/pkg/orm"
	"github.com/niflaot/gamehub-go/pkg/pagination"
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
func (repository TicketRepository) Create(ctx context.Context, ticket domain.Ticket, opener domain.Message, evidence []domain.Evidence) (domain.Ticket, error) {
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
func (repository TicketRepository) List(ctx context.Context, filter port.TicketFilter, page pagination.Page) (pagination.Result[domain.Ticket], error) {
	query := repository.store.DB(ctx).Model(&TicketModel{}).
		Order("updated_at desc, id desc").Limit(page.Limit + 1)
	query = ticketFilter(query, filter)
	var models []TicketModel
	if err := query.Find(&models).Error; err != nil {
		return pagination.Result[domain.Ticket]{}, err
	}
	return ticketPage(models, page.Limit), nil
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
	return query
}

// ticketUpdates returns mutable update columns.
func ticketUpdates(model TicketModel, expectedVersion uint64) map[string]any {
	return map[string]any{
		"title":                       model.Title,
		"status":                      model.Status,
		"priority":                    model.Priority,
		"target_user_id":              model.TargetUserID,
		"punishment_id":               model.PunishmentID,
		"current_team_group_id":       model.CurrentTeamGroupID,
		"assignee_user_id":            model.AssigneeUserID,
		"first_staff_response_at":     model.FirstStaffResponseAt,
		"last_message_at":             model.LastMessageAt,
		"last_message_author_user_id": model.LastMessageAuthorUserID,
		"closed_at":                   model.ClosedAt,
		"closed_by_user_id":           model.ClosedByUserID,
		"close_reason":                model.CloseReason,
		"resolution":                  model.Resolution,
		"escalation_level":            model.EscalationLevel,
		"sla_first_response_due_at":   model.SLAFirstResponseDueAt,
		"sla_resolution_due_at":       model.SLAResolutionDueAt,
		"message_count":               model.MessageCount,
		"staff_message_count":         model.StaffMessageCount,
		"evidence_count":              model.EvidenceCount,
		"version":                     expectedVersion + 1,
	}
}

// ticketPage maps ticket models into a page.
func ticketPage(models []TicketModel, limit int) pagination.Result[domain.Ticket] {
	next := ""
	if len(models) > limit {
		next = models[limit-1].ID.ID.String()
		models = models[:limit]
	}
	items := make([]domain.Ticket, 0, len(models))
	for _, model := range models {
		items = append(items, ticketFromModel(model))
	}
	return pagination.Result[domain.Ticket]{Items: items, NextCursor: next}
}

// pointer returns a pointer to value for GORM create calls.
func pointer[T any](value T) *T { return &value }

// now returns UTC wall clock time for persistence summaries.
func now() time.Time { return time.Now().UTC() }
