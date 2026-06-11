package http

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/niflaot/gamehub-go/module/punishments/domain"
	"github.com/niflaot/gamehub-go/module/punishments/port"
)

func (handler handler) issuePunishment(ctx *fiber.Ctx) error {
	if err := requireIdempotency(ctx); err != nil {
		return err
	}
	actor, err := currentUserID(ctx)
	if err != nil {
		return err
	}
	var request issueRequest
	if err := decodeJSON(ctx, &request); err != nil {
		return err
	}
	issued, err := handler.services.Punishments.IssuePunishment(ctx.UserContext(), issueCommand(ctx, actor, request))
	if err != nil {
		return handleError(ctx, err)
	}
	setETag(ctx, issued.Version)
	return writeJSON(ctx, fiber.StatusCreated, issued)
}

func (handler handler) listPunishments(ctx *fiber.Ctx) error {
	if _, err := currentUserID(ctx); err != nil {
		return err
	}
	page, err := pageFromQuery(ctx)
	if err != nil {
		return err
	}
	result, err := handler.services.Punishments.ListPunishments(ctx.UserContext(), punishmentFilter(ctx), page)
	if err != nil {
		return handleError(ctx, err)
	}
	return writeJSON(ctx, fiber.StatusOK, result)
}

func (handler handler) getPunishment(ctx *fiber.Ctx) error {
	if _, err := currentUserID(ctx); err != nil {
		return err
	}
	id, err := idFromParam(ctx, "punishment_id")
	if err != nil {
		return err
	}
	punishment, err := handler.services.Punishments.GetPunishment(ctx.UserContext(), id)
	if err != nil {
		return handleError(ctx, err)
	}
	setETag(ctx, punishment.Version)
	return writeJSON(ctx, fiber.StatusOK, punishment)
}

func (handler handler) updatePunishment(ctx *fiber.Ctx) error {
	actor, err := currentUserID(ctx)
	if err != nil {
		return err
	}
	id, err := idFromParam(ctx, "punishment_id")
	if err != nil {
		return err
	}
	version, err := expectedVersion(ctx)
	if err != nil {
		return err
	}
	var request struct {
		Reason        string `json:"reason"`
		PrivateReason string `json:"private_reason"`
	}
	if err := decodeJSON(ctx, &request); err != nil {
		return err
	}
	updated, err := handler.services.Punishments.UpdatePunishment(ctx.UserContext(), port.UpdateCommand{
		ActorUserID:     actor,
		PunishmentID:    id,
		Reason:          request.Reason,
		PrivateReason:   request.PrivateReason,
		ExpectedVersion: version,
	})
	if err != nil {
		return handleError(ctx, err)
	}
	setETag(ctx, updated.Version)
	return writeJSON(ctx, fiber.StatusOK, updated)
}

func (handler handler) revokePunishment(ctx *fiber.Ctx) error {
	if err := requireIdempotency(ctx); err != nil {
		return err
	}
	actor, err := currentUserID(ctx)
	if err != nil {
		return err
	}
	id, err := idFromParam(ctx, "punishment_id")
	if err != nil {
		return err
	}
	version, err := expectedVersion(ctx)
	if err != nil {
		return err
	}
	var request struct {
		Reason string `json:"reason"`
	}
	if err := decodeJSON(ctx, &request); err != nil {
		return err
	}
	err = handler.services.Punishments.RevokePunishment(ctx.UserContext(), port.RevokeCommand{
		ActorUserID: actor, PunishmentID: id, Reason: request.Reason,
		ExpectedVersion: version,
	})
	if err != nil {
		return handleError(ctx, err)
	}
	return ctx.SendStatus(fiber.StatusNoContent)
}

func (handler handler) listUserPunishments(ctx *fiber.Ctx) error {
	if _, err := currentUserID(ctx); err != nil {
		return err
	}
	userID, err := idFromParam(ctx, "user_id")
	if err != nil {
		return err
	}
	page, err := pageFromQuery(ctx)
	if err != nil {
		return err
	}
	filter := port.PunishmentFilter{TargetUserID: userID}
	if activePunishmentPath(ctx.Path()) {
		filter.Status = domain.PunishmentActive
	}
	result, err := handler.services.Punishments.ListPunishments(ctx.UserContext(), filter, page)
	if err != nil {
		return handleError(ctx, err)
	}
	return writeJSON(ctx, fiber.StatusOK, result)
}

func (handler handler) checkRestriction(ctx *fiber.Ctx) error {
	if _, err := currentUserID(ctx); err != nil {
		return err
	}
	var request struct {
		UserID    uuid.UUID `json:"user_id"`
		ActionKey string    `json:"action_key"`
	}
	if err := decodeJSON(ctx, &request); err != nil {
		return err
	}
	result, err := handler.services.Punishments.CheckRestriction(ctx.UserContext(), port.CheckCommand{
		UserID:    request.UserID,
		ActionKey: request.ActionKey,
	})
	if err != nil {
		return handleError(ctx, err)
	}
	return writeJSON(ctx, fiber.StatusOK, result)
}

func (handler handler) listRestrictions(ctx *fiber.Ctx) error {
	if _, err := currentUserID(ctx); err != nil {
		return err
	}
	userID, err := idFromParam(ctx, "user_id")
	if err != nil {
		return err
	}
	restrictions, err := handler.services.Punishments.ListActiveRestrictions(ctx.UserContext(), userID)
	if err != nil {
		return handleError(ctx, err)
	}
	return writeJSON(ctx, fiber.StatusOK, restrictions)
}

// punishmentFilter maps list query parameters to a repository filter.
func punishmentFilter(ctx *fiber.Ctx) port.PunishmentFilter {
	var userID uuid.UUID
	if value := ctx.Query("target_user_id"); value != "" {
		userID, _ = uuid.Parse(value)
	}
	return port.PunishmentFilter{
		TargetUserID: userID,
		Status:       domain.PunishmentStatus(ctx.Query("status")),
	}
}

// activePunishmentPath reports whether a user punishment list requests active rows.
func activePunishmentPath(path string) bool {
	return len(path) >= len("/active") && path[len(path)-len("/active"):] == "/active"
}
