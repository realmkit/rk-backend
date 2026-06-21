package http

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/module/groups/adapter/httpguard"
	"github.com/realmkit/rk-backend/module/groups/domain"
	"github.com/realmkit/rk-backend/module/groups/port"
	userport "github.com/realmkit/rk-backend/module/user/port"
)

// Services contains groups application services used by handlers.
type Services struct {
	// Groups manages groups.
	Groups port.GroupService

	// Memberships manages group memberships.
	Memberships port.MembershipService

	// Grants manages permission grants.
	Grants port.PermissionGrantService

	// Checker checks permissions.
	Checker port.Checker

	// Users resolves local user summaries for membership displays.
	Users userport.Service
}

// Register registers groups and permissions routes on router.
func Register(router fiber.Router, services Services) {
	handler := handler{services: services}
	groups := router.Group("/groups")
	groups.Post("", handler.createGroup)
	groups.Get("", handler.listGroups)
	groups.Get("/:group_id", handler.getGroup)
	groups.Patch("/:group_id", handler.updateGroup)
	groups.Delete("/:group_id", handler.deleteGroup)
	groups.Get("/:group_id/permission-grants", handler.listGroupPermissionGrants)
	groups.Post("/:group_id/permission-grants", handler.createGroupPermissionGrant)
	groups.Delete("/:group_id/permission-grants/:grant_id", handler.deleteGroupPermissionGrant)
	groups.Get("/:group_id/members", handler.listGroupMembers)
	groups.Put("/:group_id/members/:user_id", handler.assignMembership)
	groups.Delete("/:group_id/members/:user_id", handler.removeMembership)
	router.Get("/users/me/groups", handler.listCurrentUserGroups)
	router.Get("/users/:user_id/groups", handler.listUserGroups)
	router.Get("/permission-actions", handler.listPermissionActions)
	router.Post("/permissions/check", handler.checkPermission)
}

// handler contains groups route dependencies.
type handler struct {
	services Services // services stores the services value.
}

// checkRequest is the permission check body.
type checkRequest struct {
	ActorUserID uuid.UUID         `json:"actor_user_id"`     // ActorUserID stores the actor user i d value.
	Action      domain.Action     `json:"action"`            // Action stores the action value.
	ScopeType   domain.ScopeType  `json:"scope_type"`        // ScopeType stores the scope type value.
	ScopeID     uuid.UUID         `json:"scope_id"`          // ScopeID stores the scope i d value.
	Permission  domain.Permission `json:"permission"`        // Permission stores the permission value.
	ObjectType  domain.ObjectType `json:"object_type"`       // ObjectType stores the object type value.
	ObjectID    uuid.UUID         `json:"object_id"`         // ObjectID stores the object i d value.
	Context     map[string]any    `json:"context,omitempty"` // Context stores the context value.
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
