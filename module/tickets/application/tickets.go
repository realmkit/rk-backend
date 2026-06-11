package application

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/niflaot/gamehub-go/module/tickets/domain"
	"github.com/niflaot/gamehub-go/module/tickets/port"
	"github.com/niflaot/gamehub-go/pkg/pagination"
)

// CreateTicket opens one ticket with an opener message and evidence.
func (service Service) CreateTicket(ctx context.Context, command port.CreateTicketCommand) (domain.Ticket, error) {
	if command.IdempotencyKey != "" {
		if existing, err := service.tickets.FindByIdempotencyKey(ctx, command.IdempotencyKey); err == nil {
			return existing, nil
		}
	}
	definition, err := service.definitions.FindByID(ctx, command.DefinitionID)
	if err != nil {
		return domain.Ticket{}, err
	}
	if definition.Status != domain.DefinitionActive {
		return domain.Ticket{}, port.ErrConflict
	}
	if err := service.validateIntake(ctx, command, definition); err != nil {
		return domain.Ticket{}, err
	}
	ticket, opener, evidence := service.prepareTicket(command, definition)
	var created domain.Ticket
	err = service.withinTx(ctx, func(ctx context.Context) error {
		stored, err := service.tickets.Create(ctx, ticket, opener, evidence)
		if err != nil {
			return err
		}
		created = stored
		return service.publishTicket(ctx, "tickets.ticket.created", stored)
	})
	return created, err
}

// GetTicket returns one ticket if actor may view it.
func (service Service) GetTicket(ctx context.Context, ticketID uuid.UUID, actorUserID uuid.UUID) (domain.Ticket, error) {
	if service.authorizer != nil {
		if err := can(func() (bool, error) {
			return service.authorizer.CanView(ctx, actorUserID, ticketID)
		}); err != nil {
			return domain.Ticket{}, err
		}
	}
	return service.tickets.FindByID(ctx, ticketID)
}

// ListTickets returns a queue or user ticket list.
func (service Service) ListTickets(ctx context.Context, filter port.TicketFilter, page pagination.Page) (pagination.Result[domain.Ticket], error) {
	return service.tickets.List(ctx, filter, page)
}

func (service Service) validateIntake(ctx context.Context, command port.CreateTicketCommand, definition domain.Definition) error {
	if service.authorizer != nil {
		if err := can(func() (bool, error) {
			return service.authorizer.CanCreate(ctx, command.ActorUserID, definition.ID)
		}); err != nil {
			return err
		}
	}
	if definition.RequiresTargetUser && command.TargetUserID == nil {
		return domain.NewValidationError([]domain.Violation{{Field: "target_user_id", Message: "is required"}})
	}
	if definition.RequiresPunishment && command.PunishmentID == nil {
		return domain.NewValidationError([]domain.Violation{{Field: "punishment_id", Message: "is required"}})
	}
	if definition.RequiresEvidence && len(command.EvidenceAssetIDs) == 0 {
		return domain.NewValidationError([]domain.Violation{{Field: "evidence", Message: "is required"}})
	}
	return service.validatePunishmentAndAssets(ctx, command, definition)
}

func (service Service) validatePunishmentAndAssets(ctx context.Context, command port.CreateTicketCommand, definition domain.Definition) error {
	if command.PunishmentID != nil && service.punishments != nil {
		punishment, err := service.punishments.GetPunishment(ctx, *command.PunishmentID)
		if err != nil {
			return err
		}
		if definition.Kind == domain.KindAppeal &&
			command.SubmitterUserID != nil &&
			punishment.TargetUserID != *command.SubmitterUserID {
			return port.ErrForbidden
		}
	}
	for _, assetID := range command.EvidenceAssetIDs {
		if service.assets == nil {
			continue
		}
		exists, err := service.assets.AssetExists(ctx, assetID)
		if err != nil {
			return err
		}
		if !exists {
			return port.ErrNotFound
		}
	}
	return nil
}

func (service Service) prepareTicket(command port.CreateTicketCommand, definition domain.Definition) (domain.Ticket, domain.Message, []domain.Evidence) {
	now := time.Now().UTC()
	firstSLA, resolutionSLA := domain.SLADueAt(now, definition)
	assignee := definition.DefaultAssigneeUserID
	team := definition.DefaultTeamGroupID
	ticketID := uuid.New()
	ticket := domain.Ticket{
		ID:                    ticketID,
		DefinitionID:          definition.ID,
		Key:                   command.IdempotencyKey,
		Title:                 command.Title,
		Kind:                  definition.Kind,
		Status:                domain.StatusOpen,
		Priority:              domain.PriorityNormal,
		SubmitterUserID:       command.SubmitterUserID,
		TargetUserID:          command.TargetUserID,
		PunishmentID:          command.PunishmentID,
		CurrentTeamGroupID:    team,
		AssigneeUserID:        assignee,
		OpenedAt:              now,
		SLAFirstResponseDueAt: firstSLA,
		SLAResolutionDueAt:    resolutionSLA,
		MessageCount:          1,
		EvidenceCount:         int64(len(command.EvidenceAssetIDs)),
		Version:               1,
	}.Normalize()
	opener := messageFromCommand(ticketID, command.ActorUserID, command.ContentDocumentJSON, command.ContentText)
	evidence := make([]domain.Evidence, 0, len(command.EvidenceAssetIDs))
	for _, assetID := range command.EvidenceAssetIDs {
		id := assetID
		evidence = append(evidence, domain.Evidence{
			ID:                uuid.New(),
			TicketID:          ticketID,
			AssetID:           &id,
			Visibility:        domain.VisibilityParticipants,
			SubmittedByUserID: &command.ActorUserID,
			CreatedAt:         now,
		})
	}
	return ticket, opener, evidence
}
