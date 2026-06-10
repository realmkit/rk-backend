package structure

import (
	"github.com/gofiber/fiber/v2"
	"github.com/niflaot/gamehub-go/module/forums/adapter/http/shared"
	"github.com/niflaot/gamehub-go/module/forums/domain"
	"github.com/niflaot/gamehub-go/module/forums/port"
)

// getForumSettings returns admin forum settings.
func (handler handler) getForumSettings(ctx *fiber.Ctx) error {
	actor, err := shared.CurrentUserID(ctx)
	if err != nil {
		return err
	}
	id, err := shared.IDFromParam(ctx, "forum_id")
	if err != nil {
		return err
	}
	settings, err := handler.services.Admin.GetForumSettings(ctx.Context(), actor, id)
	if err != nil {
		return shared.HandleError(ctx, err)
	}
	shared.SetETag(ctx, settings.Version)
	return shared.WriteJSON(ctx, fiber.StatusOK, settings)
}

// updateForumSettings updates admin forum settings.
func (handler handler) updateForumSettings(ctx *fiber.Ctx) error {
	if err := shared.RequireIdempotency(ctx); err != nil {
		return err
	}
	actor, id, version, err := writeActorObjectVersion(ctx, "forum_id")
	if err != nil {
		return err
	}
	var request domain.ForumSettings
	if err := shared.DecodeJSON(ctx, &request); err != nil {
		return err
	}
	request.ForumID = id
	command := port.UpdateForumSettingsCommand{
		ActorUserID:     actor,
		Settings:        request,
		ExpectedVersion: version,
	}
	settings, err := handler.services.Admin.UpdateForumSettings(ctx.Context(), command)
	if err != nil {
		return shared.HandleError(ctx, err)
	}
	shared.SetETag(ctx, settings.Version)
	return shared.WriteJSON(ctx, fiber.StatusOK, settings)
}

// getForumPermissions returns forum permission grants.
func (handler handler) getForumPermissions(ctx *fiber.Ctx) error {
	actor, err := shared.CurrentUserID(ctx)
	if err != nil {
		return err
	}
	id, err := shared.IDFromParam(ctx, "forum_id")
	if err != nil {
		return err
	}
	settings, err := handler.services.Admin.GetForumPermissionSettings(ctx.Context(), actor, id)
	if err != nil {
		return shared.HandleError(ctx, err)
	}
	return shared.WriteJSON(ctx, fiber.StatusOK, settings)
}

// updateForumPermissions updates forum permission grants.
func (handler handler) updateForumPermissions(ctx *fiber.Ctx) error {
	if err := shared.RequireIdempotency(ctx); err != nil {
		return err
	}
	actor, err := shared.CurrentUserID(ctx)
	if err != nil {
		return err
	}
	id, err := shared.IDFromParam(ctx, "forum_id")
	if err != nil {
		return err
	}
	var request domain.ForumPermissionSettings
	if err := shared.DecodeJSON(ctx, &request); err != nil {
		return err
	}
	request.ForumID = id
	command := port.UpdateForumPermissionSettingsCommand{
		ActorUserID: actor,
		Settings:    request,
	}
	if err := handler.services.Admin.UpdateForumPermissionSettings(ctx.Context(), command); err != nil {
		return shared.HandleError(ctx, err)
	}
	return shared.WriteNoContent(ctx)
}

// simulateForumPermission simulates one forum permission.
func (handler handler) simulateForumPermission(ctx *fiber.Ctx) error {
	actor, err := shared.CurrentUserID(ctx)
	if err != nil {
		return err
	}
	id, err := shared.IDFromParam(ctx, "forum_id")
	if err != nil {
		return err
	}
	var request domain.ForumPermissionSimulationRequest
	if err := shared.DecodeJSON(ctx, &request); err != nil {
		return err
	}
	command := port.SimulateForumPermissionCommand{
		ActorUserID: actor,
		ForumID:     id,
		Request:     request,
	}
	result, err := handler.services.Admin.SimulateForumPermission(ctx.Context(), command)
	if err != nil {
		return shared.HandleError(ctx, err)
	}
	return shared.WriteJSON(ctx, fiber.StatusOK, result)
}
