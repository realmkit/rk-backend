package http

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/module/groups/domain"
	"github.com/realmkit/rk-backend/module/groups/port"
)

// checkRequest is the permission check body.
type checkRequest struct {
	ActorUserID uuid.UUID         `json:"actor_user_id"`
	Action      domain.Action     `json:"action"`
	ScopeType   domain.ScopeType  `json:"scope_type"`
	ScopeID     uuid.UUID         `json:"scope_id"`
	Permission  domain.Permission `json:"permission"`
	ObjectType  domain.ObjectType `json:"object_type"`
	ObjectID    uuid.UUID         `json:"object_id"`
	Context     map[string]any    `json:"context,omitempty"`
}

// checkPermission checks a permission.
func (handler handler) checkPermission(ctx *fiber.Ctx) error {
	var request checkRequest
	if err := decodeJSON(ctx, &request); err != nil {
		return err
	}
	decision, err := handler.services.Checker.Check(
		ctx.UserContext(),
		port.CheckRequest{
			ActorUserID: request.ActorUserID,
			Action:      requestAction(request),
			ScopeType:   requestScopeType(request),
			ScopeID:     requestScopeID(request),
			Context:     request.Context,
		},
	)
	if err != nil {
		return handleError(ctx, err)
	}
	return writeJSON(ctx, fiber.StatusOK, decision)
}

// requestAction returns the new action field or legacy permission field.
func requestAction(request checkRequest) domain.Action {
	if request.Action != "" {
		return request.Action
	}
	return domain.Action(request.Permission)
}

// requestScopeType returns the new scope type field or legacy object type field.
func requestScopeType(request checkRequest) domain.ScopeType {
	if request.ScopeType != "" {
		return request.ScopeType
	}
	return domain.ScopeType(request.ObjectType)
}

// requestScopeID returns the new scope id field or legacy object id field.
func requestScopeID(request checkRequest) uuid.UUID {
	if request.ScopeID != uuid.Nil {
		return request.ScopeID
	}
	return request.ObjectID
}
