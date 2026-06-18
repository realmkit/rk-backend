package structure

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/module/forums/adapter/http/shared"
	"github.com/realmkit/rk-backend/module/forums/domain"
	"github.com/realmkit/rk-backend/module/forums/port"
	"github.com/realmkit/rk-backend/pkg/search"
)

// createCategory creates a category.
func (handler handler) createCategory(ctx *fiber.Ctx) error {
	if err := shared.RequireIdempotency(ctx); err != nil {
		return err
	}
	actor, err := shared.CurrentUserID(ctx)
	if err != nil {
		return err
	}
	var request categoryRequest
	if err := shared.DecodeJSON(ctx, &request); err != nil {
		return err
	}
	command := port.CreateCategoryCommand{
		ActorUserID: actor,
		Category:    categoryFromRequest(uuid.Nil, request),
	}
	category, err := handler.services.Structure.CreateCategory(ctx.UserContext(), command)
	if err != nil {
		return shared.HandleError(ctx, err)
	}
	shared.SetETag(ctx, category.Version)
	return shared.WriteJSON(ctx, fiber.StatusCreated, category)
}

// getCategory returns one category.
func (handler handler) getCategory(ctx *fiber.Ctx) error {
	id, err := shared.IDFromParam(ctx, "category_id")
	if err != nil {
		return err
	}
	category, err := handler.services.Structure.GetCategory(ctx.UserContext(), id)
	if err != nil {
		return shared.HandleError(ctx, err)
	}
	shared.SetETag(ctx, category.Version)
	return shared.WriteJSON(ctx, fiber.StatusOK, category)
}

// listCategories lists categories.
func (handler handler) listCategories(ctx *fiber.Ctx) error {
	page, err := shared.PageFromQuery(ctx)
	if err != nil {
		return err
	}
	query, err := search.NewTextQuery(ctx.Query("q"), search.QueryOptions{})
	if err != nil {
		return shared.InvalidQuery(ctx)
	}
	filter := port.CategoryFilter{
		Query:  query,
		Status: domain.CategoryStatus(ctx.Query("status")),
	}
	result, err := handler.services.Structure.ListCategories(ctx.UserContext(), filter, page)
	if err != nil {
		return shared.HandleError(ctx, err)
	}
	return shared.WriteJSON(ctx, fiber.StatusOK, categoryListResponse{
		Items:         result.Items,
		NextPageToken: result.NextCursor,
	})
}

// updateCategory updates a category.
func (handler handler) updateCategory(ctx *fiber.Ctx) error {
	if err := shared.RequireIdempotency(ctx); err != nil {
		return err
	}
	actor, id, version, err := writeActorObjectVersion(ctx, "category_id")
	if err != nil {
		return err
	}
	var request categoryRequest
	if err := shared.DecodeJSON(ctx, &request); err != nil {
		return err
	}
	command := port.UpdateCategoryCommand{
		ActorUserID:     actor,
		Category:        categoryFromRequest(id, request),
		ExpectedVersion: version,
	}
	category, err := handler.services.Structure.UpdateCategory(ctx.UserContext(), command)
	if err != nil {
		return shared.HandleError(ctx, err)
	}
	shared.SetETag(ctx, category.Version)
	return shared.WriteJSON(ctx, fiber.StatusOK, category)
}

// deleteCategory deletes a category.
func (handler handler) deleteCategory(ctx *fiber.Ctx) error {
	if err := shared.RequireIdempotency(ctx); err != nil {
		return err
	}
	actor, id, version, err := writeActorObjectVersion(ctx, "category_id")
	if err != nil {
		return err
	}
	command := port.DeleteCategoryCommand{
		ActorUserID:     actor,
		ID:              id,
		ExpectedVersion: version,
	}
	if err := handler.services.Structure.DeleteCategory(ctx.UserContext(), command); err != nil {
		return shared.HandleError(ctx, err)
	}
	return shared.WriteNoContent(ctx)
}

// reorderCategories reorders categories.
func (handler handler) reorderCategories(ctx *fiber.Ctx) error {
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
	command := port.ReorderCategoriesCommand{
		ActorUserID: actor,
		Items:       request.Items,
	}
	if err := handler.services.Structure.ReorderCategories(ctx.UserContext(), command); err != nil {
		return shared.HandleError(ctx, err)
	}
	return shared.WriteNoContent(ctx)
}

// categoryFromRequest maps request data to domain.
func categoryFromRequest(id uuid.UUID, request categoryRequest) domain.ForumCategory {
	return domain.ForumCategory{
		ID:           id,
		Key:          request.Key,
		Name:         request.Name,
		Description:  request.Description,
		DisplayOrder: request.DisplayOrder,
		Status:       request.Status,
	}
}
