package http

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/niflaot/gamehub-go/module/forums/domain"
	"github.com/niflaot/gamehub-go/module/forums/port"
)

// forumRequest is a forum write request.
type forumRequest struct {
	CategoryID                    uuid.UUID                   `json:"category_id"`
	ParentForumID                 *uuid.UUID                  `json:"parent_forum_id"`
	Kind                          domain.ForumKind            `json:"kind"`
	Key                           domain.Key                  `json:"key"`
	Slug                          domain.Slug                 `json:"slug"`
	Name                          string                      `json:"name"`
	Description                   string                      `json:"description"`
	DisplayOrder                  int                         `json:"display_order"`
	ExternalURL                   string                      `json:"external_url"`
	IconAssetID                   *uuid.UUID                  `json:"icon_asset_id"`
	ThreadVisibilityMode          domain.ThreadVisibilityMode `json:"thread_visibility_mode"`
	MaxStickyThreads              int                         `json:"max_sticky_threads"`
	DefaultThreadStatus           domain.ThreadStatus         `json:"default_thread_status"`
	AuthorPostEditWindowSeconds   int                         `json:"author_post_edit_window_seconds"`
	AuthorPostDeleteWindowSeconds int                         `json:"author_post_delete_window_seconds"`
	Status                        domain.ForumStatus          `json:"status"`
}

// moveForumRequest is a forum move request.
type moveForumRequest struct {
	CategoryID    uuid.UUID  `json:"category_id"`
	ParentForumID *uuid.UUID `json:"parent_forum_id"`
	DisplayOrder  int        `json:"display_order"`
}

// forumListResponse contains one forum page.
type forumListResponse struct {
	Items         []domain.Forum `json:"items"`
	NextPageToken string         `json:"next_page_token,omitempty"`
}

// tree returns the visible forum tree.
func (handler handler) tree(ctx *fiber.Ctx) error {
	actor, err := optionalUserID(ctx)
	if err != nil {
		return err
	}
	tree, err := handler.services.Forums.Tree(ctx.Context(), actor)
	if err != nil {
		return handleError(ctx, err)
	}
	return writeJSON(ctx, fiber.StatusOK, tree)
}

// createForum creates a forum.
func (handler handler) createForum(ctx *fiber.Ctx) error {
	if err := requireIdempotency(ctx); err != nil {
		return err
	}
	actor, err := currentUserID(ctx)
	if err != nil {
		return err
	}
	var request forumRequest
	if err := decodeJSON(ctx, &request); err != nil {
		return err
	}
	forum, err := handler.services.Forums.CreateForum(ctx.Context(), port.CreateForumCommand{ActorUserID: actor, Forum: forumFromRequest(uuid.Nil, request)})
	if err != nil {
		return handleError(ctx, err)
	}
	setETag(ctx, forum.Version)
	return writeJSON(ctx, fiber.StatusCreated, forum)
}

// getForum returns one forum.
func (handler handler) getForum(ctx *fiber.Ctx) error {
	id, err := idFromParam(ctx, "forum_id")
	if err != nil {
		return err
	}
	forum, err := handler.services.Forums.GetForum(ctx.Context(), id)
	if err != nil {
		return handleError(ctx, err)
	}
	setETag(ctx, forum.Version)
	return writeJSON(ctx, fiber.StatusOK, forum)
}

// listForums lists forums.
func (handler handler) listForums(ctx *fiber.Ctx) error {
	page, err := pageFromQuery(ctx)
	if err != nil {
		return err
	}
	result, err := handler.services.Forums.ListForums(ctx.Context(), forumFilterFromQuery(ctx), page)
	if err != nil {
		return handleError(ctx, err)
	}
	return writeJSON(ctx, fiber.StatusOK, forumListResponse{Items: result.Items, NextPageToken: result.NextCursor})
}

// updateForum updates a forum.
func (handler handler) updateForum(ctx *fiber.Ctx) error {
	if err := requireIdempotency(ctx); err != nil {
		return err
	}
	actor, err := currentUserID(ctx)
	if err != nil {
		return err
	}
	id, err := idFromParam(ctx, "forum_id")
	if err != nil {
		return err
	}
	version, err := expectedVersion(ctx)
	if err != nil {
		return err
	}
	var request forumRequest
	if err := decodeJSON(ctx, &request); err != nil {
		return err
	}
	forum, err := handler.services.Forums.UpdateForum(ctx.Context(), port.UpdateForumCommand{ActorUserID: actor, Forum: forumFromRequest(id, request), ExpectedVersion: version})
	if err != nil {
		return handleError(ctx, err)
	}
	setETag(ctx, forum.Version)
	return writeJSON(ctx, fiber.StatusOK, forum)
}

// getForumSettings returns admin forum settings.
func (handler handler) getForumSettings(ctx *fiber.Ctx) error {
	actor, err := currentUserID(ctx)
	if err != nil {
		return err
	}
	id, err := idFromParam(ctx, "forum_id")
	if err != nil {
		return err
	}
	settings, err := handler.services.Forums.GetForumSettings(ctx.Context(), actor, id)
	if err != nil {
		return handleError(ctx, err)
	}
	setETag(ctx, settings.Version)
	return writeJSON(ctx, fiber.StatusOK, settings)
}

// updateForumSettings updates admin forum settings.
func (handler handler) updateForumSettings(ctx *fiber.Ctx) error {
	if err := requireIdempotency(ctx); err != nil {
		return err
	}
	actor, err := currentUserID(ctx)
	if err != nil {
		return err
	}
	id, err := idFromParam(ctx, "forum_id")
	if err != nil {
		return err
	}
	version, err := expectedVersion(ctx)
	if err != nil {
		return err
	}
	var request domain.ForumSettings
	if err := decodeJSON(ctx, &request); err != nil {
		return err
	}
	request.ForumID = id
	settings, err := handler.services.Forums.UpdateForumSettings(ctx.Context(), port.UpdateForumSettingsCommand{ActorUserID: actor, Settings: request, ExpectedVersion: version})
	if err != nil {
		return handleError(ctx, err)
	}
	setETag(ctx, settings.Version)
	return writeJSON(ctx, fiber.StatusOK, settings)
}

// getForumPermissions returns forum permission grants.
func (handler handler) getForumPermissions(ctx *fiber.Ctx) error {
	actor, err := currentUserID(ctx)
	if err != nil {
		return err
	}
	id, err := idFromParam(ctx, "forum_id")
	if err != nil {
		return err
	}
	settings, err := handler.services.Forums.GetForumPermissionSettings(ctx.Context(), actor, id)
	if err != nil {
		return handleError(ctx, err)
	}
	return writeJSON(ctx, fiber.StatusOK, settings)
}

// updateForumPermissions updates forum permission grants.
func (handler handler) updateForumPermissions(ctx *fiber.Ctx) error {
	if err := requireIdempotency(ctx); err != nil {
		return err
	}
	actor, err := currentUserID(ctx)
	if err != nil {
		return err
	}
	id, err := idFromParam(ctx, "forum_id")
	if err != nil {
		return err
	}
	var request domain.ForumPermissionSettings
	if err := decodeJSON(ctx, &request); err != nil {
		return err
	}
	request.ForumID = id
	if err := handler.services.Forums.UpdateForumPermissionSettings(ctx.Context(), port.UpdateForumPermissionSettingsCommand{ActorUserID: actor, Settings: request}); err != nil {
		return handleError(ctx, err)
	}
	return writeNoContent(ctx)
}

// simulateForumPermission simulates one forum permission.
func (handler handler) simulateForumPermission(ctx *fiber.Ctx) error {
	actor, err := currentUserID(ctx)
	if err != nil {
		return err
	}
	id, err := idFromParam(ctx, "forum_id")
	if err != nil {
		return err
	}
	var request domain.ForumPermissionSimulationRequest
	if err := decodeJSON(ctx, &request); err != nil {
		return err
	}
	result, err := handler.services.Forums.SimulateForumPermission(ctx.Context(), port.SimulateForumPermissionCommand{ActorUserID: actor, ForumID: id, Request: request})
	if err != nil {
		return handleError(ctx, err)
	}
	return writeJSON(ctx, fiber.StatusOK, result)
}

// moveForum moves a forum.
func (handler handler) moveForum(ctx *fiber.Ctx) error {
	if err := requireIdempotency(ctx); err != nil {
		return err
	}
	actor, err := currentUserID(ctx)
	if err != nil {
		return err
	}
	id, err := idFromParam(ctx, "forum_id")
	if err != nil {
		return err
	}
	version, err := expectedVersion(ctx)
	if err != nil {
		return err
	}
	var request moveForumRequest
	if err := decodeJSON(ctx, &request); err != nil {
		return err
	}
	forum, err := handler.services.Forums.MoveForum(ctx.Context(), port.MoveForumCommand{ActorUserID: actor, ID: id, CategoryID: request.CategoryID, ParentForumID: request.ParentForumID, DisplayOrder: request.DisplayOrder, ExpectedVersion: version})
	if err != nil {
		return handleError(ctx, err)
	}
	setETag(ctx, forum.Version)
	return writeJSON(ctx, fiber.StatusOK, forum)
}

// deleteForum deletes a forum.
func (handler handler) deleteForum(ctx *fiber.Ctx) error {
	if err := requireIdempotency(ctx); err != nil {
		return err
	}
	actor, err := currentUserID(ctx)
	if err != nil {
		return err
	}
	id, err := idFromParam(ctx, "forum_id")
	if err != nil {
		return err
	}
	version, err := expectedVersion(ctx)
	if err != nil {
		return err
	}
	if err := handler.services.Forums.DeleteForum(ctx.Context(), port.DeleteForumCommand{ActorUserID: actor, ID: id, ExpectedVersion: version}); err != nil {
		return handleError(ctx, err)
	}
	return writeNoContent(ctx)
}

// reorderForums reorders forums.
func (handler handler) reorderForums(ctx *fiber.Ctx) error {
	if err := requireIdempotency(ctx); err != nil {
		return err
	}
	actor, err := currentUserID(ctx)
	if err != nil {
		return err
	}
	var request reorderRequest
	if err := decodeJSON(ctx, &request); err != nil {
		return err
	}
	if err := handler.services.Forums.ReorderForums(ctx.Context(), port.ReorderForumsCommand{ActorUserID: actor, Items: request.Items}); err != nil {
		return handleError(ctx, err)
	}
	return writeNoContent(ctx)
}

// forumFromRequest maps request data to domain.
func forumFromRequest(id uuid.UUID, request forumRequest) domain.Forum {
	return domain.Forum{ID: id, CategoryID: request.CategoryID, ParentForumID: request.ParentForumID, Kind: request.Kind, Key: request.Key, Slug: request.Slug, Name: request.Name, Description: request.Description, DisplayOrder: request.DisplayOrder, ExternalURL: request.ExternalURL, IconAssetID: request.IconAssetID, ThreadVisibilityMode: request.ThreadVisibilityMode, MaxStickyThreads: request.MaxStickyThreads, DefaultThreadStatus: request.DefaultThreadStatus, AuthorPostEditWindowSeconds: request.AuthorPostEditWindowSeconds, AuthorPostDeleteWindowSeconds: request.AuthorPostDeleteWindowSeconds, Status: request.Status}
}

// forumFilterFromQuery returns forum filters from query params.
func forumFilterFromQuery(ctx *fiber.Ctx) port.ForumFilter {
	var parentID *uuid.UUID
	if value := ctx.Query("parent_forum_id"); value != "" {
		parsed, err := uuid.Parse(value)
		if err == nil {
			parentID = &parsed
		}
	}
	categoryID, _ := uuid.Parse(ctx.Query("category_id"))
	return port.ForumFilter{CategoryID: categoryID, ParentForumID: parentID, Status: domain.ForumStatus(ctx.Query("status"))}
}
