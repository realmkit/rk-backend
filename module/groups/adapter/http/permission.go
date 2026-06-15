package http

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/module/groups/adapter/httpguard"
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
	actor, err := currentUserID(ctx)
	if err != nil {
		return err
	}
	if request.ActorUserID == uuid.Nil {
		request.ActorUserID = actor
	}
	if request.ActorUserID != actor {
		if err := httpguard.Check(ctx, handler.services.Checker, actor, checkManagementTarget(request)); err != nil {
			return err
		}
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

// checkManagementTarget returns the guard target for checking another actor.
func checkManagementTarget(request checkRequest) httpguard.Target {
	scopeType := requestScopeType(request)
	scopeID := requestScopeID(request)
	if scopeType == domain.ObjectGroup && scopeID != uuid.Nil {
		return httpguard.Object(domain.PermissionGroupsManagePermissions, domain.ObjectGroup, scopeID)
	}
	return httpguard.All(domain.PermissionGroupsManagePermissions, domain.ObjectGroup)
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
