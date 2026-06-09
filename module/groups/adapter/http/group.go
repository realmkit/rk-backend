package http

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/niflaot/gamehub-go/module/groups/domain"
	"github.com/niflaot/gamehub-go/module/groups/port"
)

// groupRequest is the group create or update body.
type groupRequest struct {
	Key         domain.Key         `json:"key"`
	Name        string             `json:"name"`
	Description string             `json:"description"`
	Color       domain.Color       `json:"color"`
	Weight      int                `json:"weight"`
	Status      domain.GroupStatus `json:"status"`
	IconAssetID *uuid.UUID         `json:"icon_asset_id"`
}

// groupListResponse contains one group page.
type groupListResponse struct {
	Items         []domain.Group `json:"items"`
	NextPageToken string         `json:"next_page_token,omitempty"`
}

// createGroup creates a group.
func (handler handler) createGroup(ctx *fiber.Ctx) error {
	if err := requireIdempotency(ctx); err != nil {
		return err
	}
	var request groupRequest
	if err := decodeJSON(ctx, &request); err != nil {
		return err
	}
	group, err := handler.services.Groups.Create(ctx.Context(), port.CreateGroupCommand{Group: groupFromRequest(request)})
	if err != nil {
		return handleError(ctx, err)
	}
	setETag(ctx, group.Version)
	return writeJSON(ctx, fiber.StatusCreated, group)
}

// listGroups lists groups.
func (handler handler) listGroups(ctx *fiber.Ctx) error {
	page, err := pageFromQuery(ctx)
	if err != nil {
		return err
	}
	result, err := handler.services.Groups.List(ctx.Context(), port.GroupFilter{Status: domain.GroupStatus(ctx.Query("status"))}, page)
	if err != nil {
		return handleError(ctx, err)
	}
	return writeJSON(ctx, fiber.StatusOK, groupListResponse{Items: result.Items, NextPageToken: result.NextCursor})
}

// getGroup returns one group.
func (handler handler) getGroup(ctx *fiber.Ctx) error {
	id, err := idFromParam(ctx, "group_id")
	if err != nil {
		return err
	}
	group, err := handler.services.Groups.Get(ctx.Context(), id)
	if err != nil {
		return handleError(ctx, err)
	}
	setETag(ctx, group.Version)
	return writeJSON(ctx, fiber.StatusOK, group)
}

// updateGroup updates a group.
func (handler handler) updateGroup(ctx *fiber.Ctx) error {
	if err := requireIdempotency(ctx); err != nil {
		return err
	}
	id, err := idFromParam(ctx, "group_id")
	if err != nil {
		return err
	}
	version, err := expectedVersion(ctx)
	if err != nil {
		return err
	}
	var request groupRequest
	if err := decodeJSON(ctx, &request); err != nil {
		return err
	}
	group := groupFromRequest(request)
	group.ID = id
	updated, err := handler.services.Groups.Update(ctx.Context(), port.UpdateGroupCommand{Group: group, ExpectedVersion: version})
	if err != nil {
		return handleError(ctx, err)
	}
	setETag(ctx, updated.Version)
	return writeJSON(ctx, fiber.StatusOK, updated)
}

// deleteGroup deletes a group.
func (handler handler) deleteGroup(ctx *fiber.Ctx) error {
	if err := requireIdempotency(ctx); err != nil {
		return err
	}
	id, err := idFromParam(ctx, "group_id")
	if err != nil {
		return err
	}
	version, err := expectedVersion(ctx)
	if err != nil {
		return err
	}
	if err := handler.services.Groups.Delete(ctx.Context(), port.DeleteGroupCommand{ID: id, ExpectedVersion: version}); err != nil {
		return handleError(ctx, err)
	}
	return writeNoContent(ctx)
}

// groupFromRequest maps HTTP request to group.
func groupFromRequest(request groupRequest) domain.Group {
	return domain.Group{Key: request.Key, Name: request.Name, Description: request.Description, Color: request.Color, Weight: request.Weight, Status: request.Status, IconAssetID: request.IconAssetID}
}
