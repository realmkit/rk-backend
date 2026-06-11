package postgres

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/niflaot/gamehub-go/module/tickets/domain"
	"github.com/niflaot/gamehub-go/pkg/orm"
	"github.com/niflaot/gamehub-go/pkg/pagination"
	"gorm.io/gorm"
)

// AddMessage appends one message and updates ticket summaries.
func (repository TicketRepository) AddMessage(ctx context.Context, message domain.Message) (domain.Message, error) {
	db := repository.store.DB(ctx)
	sequence, err := repository.nextMessageSequence(ctx, message.TicketID)
	if err != nil {
		return domain.Message{}, err
	}
	message.Sequence = sequence
	model := messageModel(message)
	if err := db.Create(&model).Error; err != nil {
		return domain.Message{}, err
	}
	updates := map[string]any{
		"message_count":               gorm.Expr("message_count + 1"),
		"last_message_at":             model.CreatedAt,
		"last_message_author_user_id": model.AuthorUserID,
	}
	if message.AuthorRole == domain.RoleStaff {
		updates["staff_message_count"] = gorm.Expr("staff_message_count + 1")
		updates["first_staff_response_at"] = model.CreatedAt
	}
	if err := db.Model(&TicketModel{}).Where("id = ?", message.TicketID).
		Updates(updates).Error; err != nil {
		return domain.Message{}, err
	}
	return messageFromModel(model), nil
}

// ListMessages returns messages ordered by sequence.
func (repository TicketRepository) ListMessages(
	ctx context.Context,
	ticketID uuid.UUID,
	includeStaffOnly bool,
	page pagination.Page,
) (pagination.Result[domain.Message], error) {
	query := repository.store.DB(ctx).Model(&MessageModel{}).
		Where("ticket_id = ?", ticketID).
		Order("sequence asc").Limit(page.Limit + 1)
	if !includeStaffOnly {
		query = query.Where("visibility = ?", domain.VisibilityParticipants)
	}
	var models []MessageModel
	if err := query.Find(&models).Error; err != nil {
		return pagination.Result[domain.Message]{}, err
	}
	return messagePage(models, page.Limit), nil
}

// AddEvidence stores one evidence record and updates ticket counters.
func (repository TicketRepository) AddEvidence(ctx context.Context, evidence domain.Evidence) (domain.Evidence, error) {
	model := evidenceModel(evidence)
	db := repository.store.DB(ctx)
	if err := db.Create(&model).Error; err != nil {
		return domain.Evidence{}, err
	}
	if err := db.Model(&TicketModel{}).Where("id = ?", evidence.TicketID).
		Update("evidence_count", gorm.Expr("evidence_count + 1")).Error; err != nil {
		return domain.Evidence{}, err
	}
	return evidenceFromModel(model), nil
}

// ListEvidence returns ticket evidence.
func (repository TicketRepository) ListEvidence(ctx context.Context, ticketID uuid.UUID, includeStaffOnly bool) ([]domain.Evidence, error) {
	query := repository.store.DB(ctx).Model(&EvidenceModel{}).Where("ticket_id = ?", ticketID)
	if !includeStaffOnly {
		query = query.Where("visibility = ?", domain.VisibilityParticipants)
	}
	var models []EvidenceModel
	if err := query.Order("created_at asc, id asc").Find(&models).Error; err != nil {
		return nil, err
	}
	items := make([]domain.Evidence, 0, len(models))
	for _, model := range models {
		items = append(items, evidenceFromModel(model))
	}
	return items, nil
}

// AddAction stores one workflow action.
func (repository TicketRepository) AddAction(ctx context.Context, action domain.Action) (domain.Action, error) {
	model := actionModel(action)
	if err := repository.store.DB(ctx).Create(&model).Error; err != nil {
		return domain.Action{}, err
	}
	return actionFromModel(model), nil
}

// ListActions returns a ticket action page.
func (repository TicketRepository) ListActions(
	ctx context.Context,
	ticketID uuid.UUID,
	page pagination.Page,
) (pagination.Result[domain.Action], error) {
	var models []ActionModel
	err := repository.store.DB(ctx).Model(&ActionModel{}).Where("ticket_id = ?", ticketID).
		Order("created_at asc, id asc").Limit(page.Limit + 1).Find(&models).Error
	if err != nil {
		return pagination.Result[domain.Action]{}, err
	}
	return actionPage(models, page.Limit), nil
}

// nextMessageSequence allocates the next ticket-local sequence.
func (repository TicketRepository) nextMessageSequence(ctx context.Context, ticketID uuid.UUID) (int64, error) {
	var sequence int64
	err := repository.store.DB(ctx).Model(&MessageModel{}).Where("ticket_id = ?", ticketID).
		Select("COALESCE(MAX(sequence), 0)").Scan(&sequence).Error
	return sequence + 1, err
}

// evidenceModel maps domain evidence into persistence state.
func evidenceModel(evidence domain.Evidence) EvidenceModel {
	return EvidenceModel{
		ID:                orm.ID{ID: evidence.ID},
		TicketID:          evidence.TicketID,
		MessageID:         evidence.MessageID,
		AssetID:           evidence.AssetID,
		ExternalURL:       evidence.ExternalURL,
		Label:             evidence.Label,
		Description:       evidence.Description,
		Visibility:        string(evidence.Visibility),
		SubmittedByUserID: evidence.SubmittedByUserID,
		CreatedAt:         evidence.CreatedAt,
	}
}

// evidenceFromModel maps evidence rows into domain state.
func evidenceFromModel(model EvidenceModel) domain.Evidence {
	return domain.Evidence{
		ID:                model.ID.ID,
		TicketID:          model.TicketID,
		MessageID:         model.MessageID,
		AssetID:           model.AssetID,
		ExternalURL:       model.ExternalURL,
		Label:             model.Label,
		Description:       model.Description,
		Visibility:        domain.MessageVisibility(model.Visibility),
		SubmittedByUserID: model.SubmittedByUserID,
		CreatedAt:         model.CreatedAt,
	}
}

// actionModel maps domain action into persistence state.
func actionModel(action domain.Action) ActionModel {
	return ActionModel{
		ID:             orm.ID{ID: action.ID},
		TicketID:       action.TicketID,
		ActorUserID:    action.ActorUserID,
		Type:           string(action.Type),
		Status:         string(action.Status),
		PayloadJSON:    jsonObject(action.PayloadJSON),
		ResultJSON:     jsonObject(action.ResultJSON),
		IdempotencyKey: action.IdempotencyKey,
		CreatedAt:      action.CreatedAt,
		CompletedAt:    action.CompletedAt,
		FailedAt:       action.FailedAt,
		Error:          action.Error,
	}
}

// actionFromModel maps action rows into domain state.
func actionFromModel(model ActionModel) domain.Action {
	return domain.Action{
		ID:             model.ID.ID,
		TicketID:       model.TicketID,
		ActorUserID:    model.ActorUserID,
		Type:           domain.ActionType(model.Type),
		Status:         domain.ActionStatus(model.Status),
		PayloadJSON:    json.RawMessage(model.PayloadJSON),
		ResultJSON:     json.RawMessage(model.ResultJSON),
		IdempotencyKey: model.IdempotencyKey,
		CreatedAt:      model.CreatedAt,
		CompletedAt:    model.CompletedAt,
		FailedAt:       model.FailedAt,
		Error:          model.Error,
	}
}

// jsonObject returns a valid object JSON string.
func jsonObject(value json.RawMessage) string {
	if len(value) == 0 || !json.Valid(value) {
		return "{}"
	}
	return string(value)
}

// messagePage maps rows into a message page.
func messagePage(models []MessageModel, limit int) pagination.Result[domain.Message] {
	next := ""
	if len(models) > limit {
		next = models[limit-1].ID.ID.String()
		models = models[:limit]
	}
	items := make([]domain.Message, 0, len(models))
	for _, model := range models {
		items = append(items, messageFromModel(model))
	}
	return pagination.Result[domain.Message]{Items: items, NextCursor: next}
}

// actionPage maps rows into an action page.
func actionPage(models []ActionModel, limit int) pagination.Result[domain.Action] {
	next := ""
	if len(models) > limit {
		next = models[limit-1].ID.ID.String()
		models = models[:limit]
	}
	items := make([]domain.Action, 0, len(models))
	for _, model := range models {
		items = append(items, actionFromModel(model))
	}
	return pagination.Result[domain.Action]{Items: items, NextCursor: next}
}
