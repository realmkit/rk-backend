package application

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/module/tickets/domain"
	"github.com/realmkit/rk-backend/module/tickets/port"
	"github.com/realmkit/rk-backend/pkg/pagination"
)

// CreateMessage creates one ticket message.
func (service Service) CreateMessage(ctx context.Context, command port.MessageCommand) (domain.Message, error) {
	if err := service.requireAuthorizer(); err != nil {
		return domain.Message{}, err
	}
	if err := can(func() (bool, error) {
		return service.authorizer.CanReply(ctx, command.ActorUserID, command.TicketID)
	}); err != nil {
		return domain.Message{}, err
	}
	message := messageFromCommand(
		command.TicketID,
		command.ActorUserID,
		command.ContentDocumentJSON,
		command.ContentText,
	)
	message.Visibility = command.Visibility
	if message.Visibility == "" {
		message.Visibility = domain.VisibilityParticipants
	}
	if err := service.requirePrivateVisibility(ctx, command.ActorUserID, command.TicketID, message.Visibility); err != nil {
		return domain.Message{}, err
	}
	if err := message.Validate(); err != nil {
		return domain.Message{}, err
	}
	created, err := service.tickets.AddMessage(ctx, message)
	if err != nil {
		return domain.Message{}, err
	}
	_ = service.clearTicket(ctx, command.TicketID)
	return created, service.publishMessage(ctx, "tickets.message.created", created)
}

// ListMessages returns visible ticket messages.
func (service Service) ListMessages(
	ctx context.Context,
	ticketID uuid.UUID,
	actorUserID uuid.UUID,
	includeStaffOnly bool,
	page pagination.Page,
) (pagination.Result[domain.Message], error) {
	if err := service.requireAuthorizer(); err != nil {
		return pagination.Result[domain.Message]{}, err
	}
	if err := can(func() (bool, error) {
		return service.authorizer.CanView(ctx, actorUserID, ticketID)
	}); err != nil {
		return pagination.Result[domain.Message]{}, err
	}
	if includeStaffOnly {
		if err := service.requirePrivateVisibility(ctx, actorUserID, ticketID, domain.VisibilityStaffOnly); err != nil {
			return pagination.Result[domain.Message]{}, err
		}
	}
	return service.tickets.ListMessages(ctx, ticketID, includeStaffOnly, page)
}

// AddEvidence adds asset or external URL evidence.
func (service Service) AddEvidence(ctx context.Context, command port.EvidenceCommand) (domain.Evidence, error) {
	if err := service.requireAuthorizer(); err != nil {
		return domain.Evidence{}, err
	}
	if err := can(func() (bool, error) {
		return service.authorizer.CanReply(ctx, command.ActorUserID, command.TicketID)
	}); err != nil {
		return domain.Evidence{}, err
	}
	if command.AssetID != nil && service.assets != nil {
		exists, err := service.assets.AssetExists(ctx, *command.AssetID)
		if err != nil {
			return domain.Evidence{}, err
		}
		if !exists {
			return domain.Evidence{}, port.ErrNotFound
		}
	}
	evidence := domain.Evidence{
		ID:                uuid.New(),
		TicketID:          command.TicketID,
		MessageID:         command.MessageID,
		AssetID:           command.AssetID,
		ExternalURL:       command.ExternalURL,
		Label:             command.Label,
		Description:       command.Description,
		Visibility:        command.Visibility,
		SubmittedByUserID: &command.ActorUserID,
		CreatedAt:         time.Now().UTC(),
	}
	if evidence.Visibility == "" {
		evidence.Visibility = domain.VisibilityParticipants
	}
	if err := service.requirePrivateVisibility(ctx, command.ActorUserID, command.TicketID, evidence.Visibility); err != nil {
		return domain.Evidence{}, err
	}
	if err := evidence.Validate(); err != nil {
		return domain.Evidence{}, err
	}
	created, err := service.tickets.AddEvidence(ctx, evidence)
	if err != nil {
		return domain.Evidence{}, err
	}
	_ = service.clearTicket(ctx, command.TicketID)
	return created, service.publishEvidence(ctx, "tickets.evidence.added", created)
}

// ListEvidence returns ticket evidence.
func (service Service) ListEvidence(
	ctx context.Context,
	ticketID uuid.UUID,
	actorUserID uuid.UUID,
	includeStaffOnly bool,
) ([]domain.Evidence, error) {
	if err := service.requireAuthorizer(); err != nil {
		return nil, err
	}
	if err := can(func() (bool, error) {
		return service.authorizer.CanView(ctx, actorUserID, ticketID)
	}); err != nil {
		return nil, err
	}
	if includeStaffOnly {
		if err := service.requirePrivateVisibility(ctx, actorUserID, ticketID, domain.VisibilityStaffOnly); err != nil {
			return nil, err
		}
	}
	return service.tickets.ListEvidence(ctx, ticketID, includeStaffOnly)
}

// requirePrivateVisibility supports package behavior.
func (service Service) requirePrivateVisibility(
	ctx context.Context,
	actorUserID uuid.UUID,
	ticketID uuid.UUID,
	visibility domain.MessageVisibility,
) error {
	if visibility == "" || visibility == domain.VisibilityParticipants {
		return nil
	}
	return can(func() (bool, error) {
		return service.authorizer.CanStaffAction(ctx, actorUserID, ticketID)
	})
}

// messageFromCommand supports package behavior.
func messageFromCommand(ticketID uuid.UUID, actorUserID uuid.UUID, document json.RawMessage, text string) domain.Message {
	return domain.Message{
		ID:                  uuid.New(),
		TicketID:            ticketID,
		AuthorUserID:        &actorUserID,
		AuthorRole:          domain.RoleSubmitter,
		Visibility:          domain.VisibilityParticipants,
		ContentFormat:       "prosemirror_json",
		ContentDocumentJSON: ensureJSON(document),
		ContentText:         text,
		Version:             1,
		CreatedAt:           time.Now().UTC(),
		UpdatedAt:           time.Now().UTC(),
	}
}

// systemMessage supports package behavior.
func systemMessage(ticketID uuid.UUID, text string) domain.Message {
	now := time.Now().UTC()
	return domain.Message{
		ID:                  uuid.New(),
		TicketID:            ticketID,
		AuthorRole:          domain.RoleSystem,
		Visibility:          domain.VisibilityParticipants,
		ContentFormat:       "prosemirror_json",
		ContentDocumentJSON: json.RawMessage(`{"type":"doc","content":[]}`),
		ContentText:         text,
		Version:             1,
		CreatedAt:           now,
		UpdatedAt:           now,
	}
}

// ensureJSON supports package behavior.
func ensureJSON(document json.RawMessage) json.RawMessage {
	if len(document) == 0 || !json.Valid(document) {
		return json.RawMessage(`{"type":"doc","content":[]}`)
	}
	return document
}
