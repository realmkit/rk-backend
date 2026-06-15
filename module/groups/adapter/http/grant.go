package http

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/module/groups/adapter/httpguard"
	"github.com/realmkit/rk-backend/module/groups/domain"
	"github.com/realmkit/rk-backend/module/groups/port"
)

// permissionActionListResponse contains grantable actions.
type permissionActionListResponse struct {
	// Items contains available permission actions.
	Items []domain.PermissionAction `json:"items"`
}

// permissionGrantRequest is the create grant body for a group.
type permissionGrantRequest struct {
	// Action is the granted dotted permission key.
	Action domain.Action `json:"action"`

	// ScopeType is the resource type this grant applies to.
	ScopeType domain.ScopeType `json:"scope_type"`

	// ScopeID is the resource identifier this grant applies to.
	ScopeID uuid.UUID `json:"scope_id"`

	// Inherit reports whether descendant scopes inherit this grant.
	Inherit bool `json:"inherit"`

	// ConditionKey references an optional named runtime condition.
	ConditionKey string `json:"condition_key,omitempty"`
}

// permissionGrantListResponse contains one grant page.
type permissionGrantListResponse struct {
	// Items contains permission grants in the current page.
	Items []domain.PermissionGrant `json:"items"`

	// NextPageToken is the cursor for the next page when present.
	NextPageToken string `json:"next_page_token,omitempty"`
}

// listPermissionActions lists grantable permission actions.
func (handler handler) listPermissionActions(ctx *fiber.Ctx) error {
	if _, err := httpguard.Require(
		ctx,
		handler.services.Checker,
		httpguard.All(domain.PermissionGroupsManagePermissions, domain.ObjectGroup),
	); err != nil {
		return err
	}
	actions, err := handler.services.Grants.ListPermissionActions(ctx.UserContext())
	if err != nil {
		return handleError(ctx, err)
	}
	return writeJSON(ctx, fiber.StatusOK, permissionActionListResponse{Items: actions})
}

// listGroupPermissionGrants lists global permission grants assigned to a group.
func (handler handler) listGroupPermissionGrants(ctx *fiber.Ctx) error {
	groupID, err := idFromParam(ctx, "group_id")
	if err != nil {
		return err
	}
	if _, err := httpguard.Require(
		ctx,
		handler.services.Checker,
		httpguard.Object(domain.PermissionGroupsManagePermissions, domain.ObjectGroup, groupID),
	); err != nil {
		return err
	}
	page, err := pageFromQuery(ctx)
	if err != nil {
		return err
	}
	result, err := handler.services.Grants.ListPermissionGrants(
		ctx.UserContext(),
		port.PermissionGrantFilter{GroupID: groupID},
		page,
	)
	if err != nil {
		return handleError(ctx, err)
	}
	return writeJSON(ctx, fiber.StatusOK, permissionGrantListResponse{
		Items:         result.Items,
		NextPageToken: result.NextCursor,
	})
}

// createGroupPermissionGrant assigns a global permission grant to a group.
func (handler handler) createGroupPermissionGrant(ctx *fiber.Ctx) error {
	if err := requireIdempotency(ctx); err != nil {
		return err
	}
	groupID, err := idFromParam(ctx, "group_id")
	if err != nil {
		return err
	}
	if _, err := httpguard.Require(
		ctx,
		handler.services.Checker,
		httpguard.Object(domain.PermissionGroupsManagePermissions, domain.ObjectGroup, groupID),
	); err != nil {
		return err
	}
	var request permissionGrantRequest
	if err := decodeJSON(ctx, &request); err != nil {
		return err
	}
	grant, err := handler.services.Grants.CreatePermissionGrant(
		ctx.UserContext(),
		port.CreatePermissionGrantCommand{
			GroupID: groupID,
			Grant:   permissionGrantFromRequest(request),
		},
	)
	if err != nil {
		return handleError(ctx, err)
	}
	return writeJSON(ctx, fiber.StatusCreated, grant)
}

// deleteGroupPermissionGrant deletes one permission grant.
func (handler handler) deleteGroupPermissionGrant(ctx *fiber.Ctx) error {
	if err := requireIdempotency(ctx); err != nil {
		return err
	}
	grantID, err := idFromParam(ctx, "grant_id")
	if err != nil {
		return err
	}
	groupID, err := idFromParam(ctx, "group_id")
	if err != nil {
		return err
	}
	if _, err := httpguard.Require(
		ctx,
		handler.services.Checker,
		httpguard.Object(domain.PermissionGroupsManagePermissions, domain.ObjectGroup, groupID),
	); err != nil {
		return err
	}
	if err := handler.services.Grants.DeletePermissionGrant(
		ctx.UserContext(),
		port.DeletePermissionGrantCommand{GroupID: groupID, ID: grantID},
	); err != nil {
		return handleError(ctx, err)
	}
	return writeNoContent(ctx)
}

// permissionGrantFromRequest maps a grant request to global domain state.
func permissionGrantFromRequest(request permissionGrantRequest) domain.PermissionGrant {
	return domain.PermissionGrant{
		Action:       request.Action,
		ScopeType:    request.ScopeType,
		ScopeID:      request.ScopeID,
		Inherit:      request.Inherit,
		ConditionKey: request.ConditionKey,
	}
}
