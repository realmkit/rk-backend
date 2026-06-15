package http

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	groupsdomain "github.com/realmkit/rk-backend/module/groups/domain"
	"github.com/realmkit/rk-backend/module/punishments/domain"
	"github.com/realmkit/rk-backend/module/punishments/port"
	"github.com/realmkit/rk-backend/pkg/search"
)

func (handler handler) issuePunishment(ctx *fiber.Ctx) error {
	if err := requireIdempotency(ctx); err != nil {
		return err
	}
	actor, err := currentUserID(ctx)
	if err != nil {
		return err
	}
	if err := requirePunishment(ctx, handler.services.Checker, groupsdomain.PermissionPunishmentsIssue, uuid.Nil); err != nil {
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
	if err := requirePunishment(ctx, handler.services.Checker, groupsdomain.PermissionPunishmentsView, uuid.Nil); err != nil {
		return err
	}
	page, err := pageFromQuery(ctx)
	if err != nil {
		return err
	}
	filter, err := punishmentFilter(ctx)
	if err != nil {
		return err
	}
	result, err := handler.services.Punishments.ListPunishments(ctx.UserContext(), filter, page)
	if err != nil {
		return handleError(ctx, err)
	}
	return writeJSON(ctx, fiber.StatusOK, result)
}

func (handler handler) getPunishment(ctx *fiber.Ctx) error {
	id, err := idFromParam(ctx, "punishment_id")
	if err != nil {
		return err
	}
	if err := requirePunishment(ctx, handler.services.Checker, groupsdomain.PermissionPunishmentsView, id); err != nil {
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
	if err := requirePunishment(ctx, handler.services.Checker, groupsdomain.PermissionPunishmentsUpdate, id); err != nil {
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
	if err := requirePunishment(ctx, handler.services.Checker, groupsdomain.PermissionPunishmentsRevoke, id); err != nil {
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
	userID, err := idFromParam(ctx, "user_id")
	if err != nil {
		return err
	}
	if err := requireSelfOrPunishment(ctx, handler.services.Checker, userID, groupsdomain.PermissionPunishmentsView); err != nil {
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
	if err := requirePunishment(ctx, handler.services.Checker, groupsdomain.PermissionPunishmentsView, uuid.Nil); err != nil {
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
	userID, err := idFromParam(ctx, "user_id")
	if err != nil {
		return err
	}
	if err := requireSelfOrPunishment(ctx, handler.services.Checker, userID, groupsdomain.PermissionPunishmentsView); err != nil {
		return err
	}
	restrictions, err := handler.services.Punishments.ListActiveRestrictions(ctx.UserContext(), userID)
	if err != nil {
		return handleError(ctx, err)
	}
	return writeJSON(ctx, fiber.StatusOK, restrictions)
}

// punishmentFilter maps list query parameters to a repository filter.
func punishmentFilter(ctx *fiber.Ctx) (port.PunishmentFilter, error) {
	var userID uuid.UUID
	if value := ctx.Query("target_user_id"); value != "" {
		userID, _ = uuid.Parse(value)
	}
	query, err := search.NewTextQuery(ctx.Query("q"), search.QueryOptions{})
	if err != nil {
		return port.PunishmentFilter{}, searchProblem(err)
	}
	sort, err := search.NewSort(ctx.Query("sort"), ctx.Query("direction"), port.DefaultPunishmentSort(), port.AllowedPunishmentSorts())
	if err != nil {
		return port.PunishmentFilter{}, searchProblem(err)
	}
	return port.PunishmentFilter{
		TargetUserID: userID,
		Status:       domain.PunishmentStatus(ctx.Query("status")),
		Query:        query,
		Sort:         sort,
	}, nil
}

// activePunishmentPath reports whether a user punishment list requests active rows.
func activePunishmentPath(path string) bool {
	return len(path) >= len("/active") && path[len(path)-len("/active"):] == "/active"
}
