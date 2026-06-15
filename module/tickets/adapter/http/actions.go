package http

import (
	"context"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	groupsdomain "github.com/realmkit/rk-backend/module/groups/domain"
	"github.com/realmkit/rk-backend/module/tickets/port"
)

// staffRequest is the common staff workflow DTO.
type staffRequest struct {
	AssigneeUserID *uuid.UUID `json:"assignee_user_id"`
	TeamGroupID    *uuid.UUID `json:"team_group_id"`
	Reason         string     `json:"reason"`
}

// appealDecisionRequest is the appeal decision DTO.
type appealDecisionRequest struct {
	Reason           string `json:"reason"`
	RevokePunishment bool   `json:"revoke_punishment"`
}

// assignTicket handles ticket assignment.
func (handler handler) assignTicket(ctx *fiber.Ctx) error {
	return handler.staffAction(ctx, func(ctx context.Context, command port.StaffCommand) (any, error) {
		return handler.services.Tickets.AssignTicket(ctx, command)
	})
}

// escalateTicket handles ticket escalation.
func (handler handler) escalateTicket(ctx *fiber.Ctx) error {
	return handler.staffAction(ctx, func(ctx context.Context, command port.StaffCommand) (any, error) {
		return handler.services.Tickets.EscalateTicket(ctx, command)
	})
}

// closeTicket handles ticket closure.
func (handler handler) closeTicket(ctx *fiber.Ctx) error {
	return handler.staffAction(ctx, func(ctx context.Context, command port.StaffCommand) (any, error) {
		return handler.services.Tickets.CloseTicket(ctx, command)
	})
}

// reopenTicket handles ticket reopen.
func (handler handler) reopenTicket(ctx *fiber.Ctx) error {
	return handler.staffAction(ctx, func(ctx context.Context, command port.StaffCommand) (any, error) {
		return handler.services.Tickets.ReopenTicket(ctx, command)
	})
}

// acceptAppeal handles appeal acceptance.
func (handler handler) acceptAppeal(ctx *fiber.Ctx) error {
	return handler.appealAction(ctx, func(ctx context.Context, command port.AppealDecisionCommand) (any, error) {
		return handler.services.Tickets.AcceptAppeal(ctx, command)
	})
}

// rejectAppeal handles appeal rejection.
func (handler handler) rejectAppeal(ctx *fiber.Ctx) error {
	return handler.appealAction(ctx, func(ctx context.Context, command port.AppealDecisionCommand) (any, error) {
		return handler.services.Tickets.RejectAppeal(ctx, command)
	})
}

// verifyStats handles ticket stat verification.
func (handler handler) verifyStats(ctx *fiber.Ctx) error {
	if err := requireTicket(ctx, handler.services.Checker, groupsdomain.PermissionTicketsManage, uuid.Nil); err != nil {
		return err
	}
	report, err := handler.services.Operations.VerifyStats(ctx.UserContext())
	if err != nil {
		return handleError(ctx, err)
	}
	return writeJSON(ctx, fiber.StatusOK, report)
}

// rebuildStats handles ticket stat rebuild.
func (handler handler) rebuildStats(ctx *fiber.Ctx) error {
	if err := requireTicket(ctx, handler.services.Checker, groupsdomain.PermissionTicketsManage, uuid.Nil); err != nil {
		return err
	}
	report, err := handler.services.Operations.RebuildStats(ctx.UserContext())
	if err != nil {
		return handleError(ctx, err)
	}
	return writeJSON(ctx, fiber.StatusOK, report)
}

// staffAction maps a common staff command.
func (handler handler) staffAction(ctx *fiber.Ctx, run func(context.Context, port.StaffCommand) (any, error)) error {
	ticketID, actor, version, key, err := ticketActorVersionKey(ctx)
	if err != nil {
		return err
	}
	var request staffRequest
	if err := decodeJSON(ctx, &request); err != nil {
		return err
	}
	result, err := run(ctx.UserContext(), port.StaffCommand{
		ActorUserID:     actor,
		TicketID:        ticketID,
		AssigneeUserID:  request.AssigneeUserID,
		TeamGroupID:     request.TeamGroupID,
		Reason:          request.Reason,
		ExpectedVersion: version,
		IdempotencyKey:  key,
	})
	if err != nil {
		return handleError(ctx, err)
	}
	return writeJSON(ctx, fiber.StatusOK, result)
}

// appealAction maps an appeal decision command.
func (handler handler) appealAction(ctx *fiber.Ctx, run func(context.Context, port.AppealDecisionCommand) (any, error)) error {
	ticketID, actor, version, key, err := ticketActorVersionKey(ctx)
	if err != nil {
		return err
	}
	var request appealDecisionRequest
	if err := decodeJSON(ctx, &request); err != nil {
		return err
	}
	result, err := run(ctx.UserContext(), port.AppealDecisionCommand{
		ActorUserID:      actor,
		TicketID:         ticketID,
		Reason:           request.Reason,
		RevokePunishment: request.RevokePunishment,
		ExpectedVersion:  version,
		IdempotencyKey:   key,
	})
	if err != nil {
		return handleError(ctx, err)
	}
	return writeJSON(ctx, fiber.StatusOK, result)
}

// ticketActorVersionKey parses common workflow headers.
func ticketActorVersionKey(ctx *fiber.Ctx) (uuid.UUID, uuid.UUID, uint64, string, error) {
	ticketID, actor, key, err := ticketActorAndKey(ctx)
	if err != nil {
		return uuid.Nil, uuid.Nil, 0, "", err
	}
	version, err := expectedVersion(ctx)
	if err != nil {
		return uuid.Nil, uuid.Nil, 0, "", err
	}
	return ticketID, actor, version, key, nil
}
