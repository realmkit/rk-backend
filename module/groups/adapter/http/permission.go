package http

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/niflaot/gamehub-go/module/groups/domain"
	"github.com/niflaot/gamehub-go/module/groups/port"
)

// checkRequest is the permission check body.
type checkRequest struct {
	ActorUserID uuid.UUID         `json:"actor_user_id"`
	Permission  domain.Permission `json:"permission"`
	ObjectType  domain.ObjectType `json:"object_type"`
	ObjectID    uuid.UUID         `json:"object_id"`
}

// checkPermission checks a permission.
func (handler handler) checkPermission(ctx *fiber.Ctx) error {
	var request checkRequest
	if err := decodeJSON(ctx, &request); err != nil {
		return err
	}
	decision, err := handler.services.Checker.Check(ctx.Context(), port.CheckRequest{ActorUserID: request.ActorUserID, Permission: request.Permission, ObjectType: request.ObjectType, ObjectID: request.ObjectID})
	if err != nil {
		return handleError(ctx, err)
	}
	return writeJSON(ctx, fiber.StatusOK, decision)
}
