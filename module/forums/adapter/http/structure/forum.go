package structure

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/module/forums/adapter/http/shared"
	"github.com/realmkit/rk-backend/module/forums/domain"
	"github.com/realmkit/rk-backend/module/forums/port"
	"github.com/realmkit/rk-backend/pkg/search"
)

// tree returns the visible forum tree.
func (handler handler) tree(ctx *fiber.Ctx) error {
	actor, err := shared.OptionalUserID(ctx)
	if err != nil {
		return err
	}
	tree, err := handler.services.Structure.Tree(ctx.UserContext(), actor)
	if err != nil {
		return shared.HandleError(ctx, err)
	}
	return shared.WriteJSON(ctx, fiber.StatusOK, tree)
}

// createForum creates a forum.
func (handler handler) createForum(ctx *fiber.Ctx) error {
	if err := shared.RequireIdempotency(ctx); err != nil {
		return err
	}
	actor, err := shared.CurrentUserID(ctx)
	if err != nil {
		return err
	}
	var request forumRequest
	if err := shared.DecodeJSON(ctx, &request); err != nil {
		return err
	}
	command := port.CreateForumCommand{
		ActorUserID: actor,
		Forum:       forumFromRequest(uuid.Nil, request),
	}
	forum, err := handler.services.Structure.CreateForum(ctx.UserContext(), command)
	if err != nil {
		return shared.HandleError(ctx, err)
	}
	shared.SetETag(ctx, forum.Version)
	return shared.WriteJSON(ctx, fiber.StatusCreated, forum)
}

// getForum returns one forum.
func (handler handler) getForum(ctx *fiber.Ctx) error {
	id, err := shared.IDFromParam(ctx, "forum_id")
	if err != nil {
		return err
	}
	forum, err := handler.services.Structure.GetForum(ctx.UserContext(), id)
	if err != nil {
		return shared.HandleError(ctx, err)
	}
	shared.SetETag(ctx, forum.Version)
	return shared.WriteJSON(ctx, fiber.StatusOK, forum)
}

// listForums lists forums.
func (handler handler) listForums(ctx *fiber.Ctx) error {
	page, err := shared.PageFromQuery(ctx)
	if err != nil {
		return err
	}
	filter, err := forumFilterFromQuery(ctx)
	if err != nil {
		return err
	}
	result, err := handler.services.Structure.ListForums(ctx.UserContext(), filter, page)
	if err != nil {
		return shared.HandleError(ctx, err)
	}
	return shared.WriteJSON(ctx, fiber.StatusOK, forumListResponse{
		Items:         result.Items,
		NextPageToken: result.NextCursor,
	})
}

// updateForum updates a forum.
func (handler handler) updateForum(ctx *fiber.Ctx) error {
	if err := shared.RequireIdempotency(ctx); err != nil {
		return err
	}
	actor, id, version, err := writeActorObjectVersion(ctx, "forum_id")
	if err != nil {
		return err
	}
	var request forumRequest
	if err := shared.DecodeJSON(ctx, &request); err != nil {
		return err
	}
	command := port.UpdateForumCommand{
		ActorUserID:     actor,
		Forum:           forumFromRequest(id, request),
		ExpectedVersion: version,
	}
	forum, err := handler.services.Structure.UpdateForum(ctx.UserContext(), command)
	if err != nil {
		return shared.HandleError(ctx, err)
	}
	shared.SetETag(ctx, forum.Version)
	return shared.WriteJSON(ctx, fiber.StatusOK, forum)
}

// moveForum moves a forum.
func (handler handler) moveForum(ctx *fiber.Ctx) error {
	if err := shared.RequireIdempotency(ctx); err != nil {
		return err
	}
	actor, id, version, err := writeActorObjectVersion(ctx, "forum_id")
	if err != nil {
		return err
	}
	var request moveForumRequest
	if err := shared.DecodeJSON(ctx, &request); err != nil {
		return err
	}
	command := port.MoveForumCommand{
		ActorUserID:     actor,
		ID:              id,
		CategoryID:      request.CategoryID,
		ParentForumID:   request.ParentForumID,
		DisplayOrder:    request.DisplayOrder,
		ExpectedVersion: version,
	}
	forum, err := handler.services.Structure.MoveForum(ctx.UserContext(), command)
	if err != nil {
		return shared.HandleError(ctx, err)
	}
	shared.SetETag(ctx, forum.Version)
	return shared.WriteJSON(ctx, fiber.StatusOK, forum)
}

// deleteForum deletes a forum.
func (handler handler) deleteForum(ctx *fiber.Ctx) error {
	if err := shared.RequireIdempotency(ctx); err != nil {
		return err
	}
	actor, id, version, err := writeActorObjectVersion(ctx, "forum_id")
	if err != nil {
		return err
	}
	command := port.DeleteForumCommand{
		ActorUserID:     actor,
		ID:              id,
		ExpectedVersion: version,
	}
	if err := handler.services.Structure.DeleteForum(ctx.UserContext(), command); err != nil {
		return shared.HandleError(ctx, err)
	}
	return shared.WriteNoContent(ctx)
}

// reorderForums reorders forums.
func (handler handler) reorderForums(ctx *fiber.Ctx) error {
	if err := shared.RequireIdempotency(ctx); err != nil {
		return err
	}
	actor, err := shared.CurrentUserID(ctx)
	if err != nil {
		return err
	}
	var request reorderRequest
	if err := shared.DecodeJSON(ctx, &request); err != nil {
		return err
	}
	command := port.ReorderForumsCommand{
		ActorUserID: actor,
		Items:       request.Items,
	}
	if err := handler.services.Structure.ReorderForums(ctx.UserContext(), command); err != nil {
		return shared.HandleError(ctx, err)
	}
	return shared.WriteNoContent(ctx)
}

// forumFilterFromQuery returns forum filters from query params.
func forumFilterFromQuery(ctx *fiber.Ctx) (port.ForumFilter, error) {
	var parentID *uuid.UUID
	if value := ctx.Query("parent_forum_id"); value != "" {
		parsed, err := uuid.Parse(value)
		if err == nil {
			parentID = &parsed
		}
	}
	categoryID, _ := uuid.Parse(ctx.Query("category_id"))
	query, err := search.NewTextQuery(ctx.Query("q"), search.QueryOptions{})
	if err != nil {
		return port.ForumFilter{}, shared.InvalidQuery(ctx)
	}
	return port.ForumFilter{
		CategoryID:    categoryID,
		ParentForumID: parentID,
		Status:        domain.ForumStatus(ctx.Query("status")),
		Query:         query,
	}, nil
}
