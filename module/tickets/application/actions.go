package application

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/module/tickets/domain"
	"github.com/realmkit/rk-backend/module/tickets/port"
)

// AssignTicket assigns a ticket to a user.
func (service Service) AssignTicket(ctx context.Context, command port.StaffCommand) (domain.Ticket, error) {
	return service.staffTransition(ctx, command, domain.ActionAssign, func(ticket domain.Ticket) domain.Ticket {
		ticket.AssigneeUserID = command.AssigneeUserID
		return ticket
	})
}

// EscalateTicket escalates a ticket to another group.
func (service Service) EscalateTicket(ctx context.Context, command port.StaffCommand) (domain.Ticket, error) {
	return service.staffTransition(ctx, command, domain.ActionEscalate, func(ticket domain.Ticket) domain.Ticket {
		ticket.CurrentTeamGroupID = command.TeamGroupID
		ticket.AssigneeUserID = nil
		ticket.EscalationLevel++
		ticket.Status = domain.StatusEscalated
		return ticket
	})
}

// CloseTicket closes a ticket.
func (service Service) CloseTicket(ctx context.Context, command port.StaffCommand) (domain.Ticket, error) {
	now := time.Now().UTC()
	return service.staffTransition(ctx, command, domain.ActionClose, func(ticket domain.Ticket) domain.Ticket {
		ticket.Status = domain.StatusClosed
		ticket.ClosedAt = &now
		ticket.ClosedByUserID = &command.ActorUserID
		ticket.CloseReason = command.Reason
		return ticket
	})
}

// ReopenTicket reopens a ticket.
func (service Service) ReopenTicket(ctx context.Context, command port.StaffCommand) (domain.Ticket, error) {
	return service.staffTransition(ctx, command, domain.ActionReopen, func(ticket domain.Ticket) domain.Ticket {
		ticket.Status = domain.StatusOpen
		ticket.ClosedAt = nil
		ticket.ClosedByUserID = nil
		ticket.CloseReason = ""
		return ticket
	})
}

// AcceptAppeal accepts an appeal and optionally revokes its punishment.
func (service Service) AcceptAppeal(ctx context.Context, command port.AppealDecisionCommand) (domain.Ticket, error) {
	if command.RevokePunishment {
		current, err := service.tickets.FindByID(ctx, command.TicketID)
		if err != nil {
			return domain.Ticket{}, err
		}
		if current.PunishmentID != nil {
			if err := service.requireAppealPunishmentRevoke(ctx, command, *current.PunishmentID); err != nil {
				return domain.Ticket{}, err
			}
		}
	}
	ticket, err := service.staffTransition(ctx, port.StaffCommand{
		ActorUserID:     command.ActorUserID,
		TicketID:        command.TicketID,
		Reason:          command.Reason,
		ExpectedVersion: command.ExpectedVersion,
		IdempotencyKey:  command.IdempotencyKey,
	}, domain.ActionAcceptAppeal, func(ticket domain.Ticket) domain.Ticket {
		ticket.Status = domain.StatusAccepted
		ticket.Resolution = command.Reason
		return ticket
	})
	if err != nil || !command.RevokePunishment || ticket.PunishmentID == nil || service.punishments == nil {
		return ticket, err
	}
	err = service.punishments.RevokePunishment(
		ctx,
		*ticket.PunishmentID,
		command.ActorUserID,
		command.Reason,
		0,
	)
	return ticket, err
}

// RejectAppeal rejects an appeal.
func (service Service) RejectAppeal(ctx context.Context, command port.AppealDecisionCommand) (domain.Ticket, error) {
	return service.staffTransition(ctx, port.StaffCommand{
		ActorUserID:     command.ActorUserID,
		TicketID:        command.TicketID,
		Reason:          command.Reason,
		ExpectedVersion: command.ExpectedVersion,
		IdempotencyKey:  command.IdempotencyKey,
	}, domain.ActionRejectAppeal, func(ticket domain.Ticket) domain.Ticket {
		ticket.Status = domain.StatusRejected
		ticket.Resolution = command.Reason
		return ticket
	})
}

func (service Service) staffTransition(
	ctx context.Context,
	command port.StaffCommand,
	actionType domain.ActionType,
	apply func(domain.Ticket) domain.Ticket,
) (domain.Ticket, error) {
	if err := service.requireAuthorizer(); err != nil {
		return domain.Ticket{}, err
	}
	if err := can(func() (bool, error) {
		return service.authorizer.CanStaffAction(ctx, command.ActorUserID, command.TicketID)
	}); err != nil {
		return domain.Ticket{}, err
	}
	current, err := service.tickets.FindByID(ctx, command.TicketID)
	if err != nil {
		return domain.Ticket{}, err
	}
	next := apply(current).Normalize()
	if !domain.CanTransition(current.Status, next.Status) {
		return domain.Ticket{}, port.ErrConflict
	}
	var updated domain.Ticket
	err = service.withinTx(ctx, func(ctx context.Context) error {
		stored, err := service.tickets.Update(ctx, next, command.ExpectedVersion)
		if err != nil {
			return err
		}
		updated = stored
		if _, err := service.tickets.AddAction(ctx, action(command, actionType)); err != nil {
			return err
		}
		_, _ = service.tickets.AddMessage(ctx, systemMessage(command.TicketID, string(actionType)))
		return service.clearTicket(ctx, command.TicketID)
	})
	if err != nil {
		return domain.Ticket{}, err
	}
	return updated, service.publishTicket(ctx, "tickets.ticket.status_changed", updated)
}

func (service Service) requireAppealPunishmentRevoke(
	ctx context.Context,
	command port.AppealDecisionCommand,
	punishmentID uuid.UUID,
) error {
	if err := service.requireAuthorizer(); err != nil {
		return err
	}
	return can(func() (bool, error) {
		return service.authorizer.CanRevokePunishmentFromAppeal(
			ctx,
			command.ActorUserID,
			command.TicketID,
			punishmentID,
		)
	})
}

func action(command port.StaffCommand, actionType domain.ActionType) domain.Action {
	payload, _ := json.Marshal(map[string]string{"reason": command.Reason})
	now := time.Now().UTC()
	return domain.Action{
		ID:             uuid.New(),
		TicketID:       command.TicketID,
		ActorUserID:    &command.ActorUserID,
		Type:           actionType,
		Status:         domain.ActionCompleted,
		PayloadJSON:    payload,
		ResultJSON:     json.RawMessage(`{}`),
		IdempotencyKey: command.IdempotencyKey,
		CreatedAt:      now,
		CompletedAt:    &now,
	}
}
