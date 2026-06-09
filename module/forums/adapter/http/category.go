package http

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/niflaot/gamehub-go/module/forums/domain"
	"github.com/niflaot/gamehub-go/module/forums/port"
)

// categoryRequest is a category write request.
type categoryRequest struct {
	Key          domain.Key            `json:"key"`
	Name         string                `json:"name"`
	Description  string                `json:"description"`
	DisplayOrder int                   `json:"display_order"`
	Status       domain.CategoryStatus `json:"status"`
}

// reorderRequest is a display-order request.
type reorderRequest struct {
	Items []port.ReorderItem `json:"items"`
}

// categoryListResponse contains one category page.
type categoryListResponse struct {
	Items         []domain.ForumCategory `json:"items"`
	NextPageToken string                 `json:"next_page_token,omitempty"`
}

// createCategory creates a category.
func (handler handler) createCategory(ctx *fiber.Ctx) error {
	if err := requireIdempotency(ctx); err != nil {
		return err
	}
	actor, err := currentUserID(ctx)
	if err != nil {
		return err
	}
	var request categoryRequest
	if err := decodeJSON(ctx, &request); err != nil {
		return err
	}
	category, err := handler.services.Forums.CreateCategory(ctx.Context(), port.CreateCategoryCommand{ActorUserID: actor, Category: categoryFromRequest(uuid.Nil, request)})
	if err != nil {
		return handleError(ctx, err)
	}
	setETag(ctx, category.Version)
	return writeJSON(ctx, fiber.StatusCreated, category)
}

// getCategory returns one category.
func (handler handler) getCategory(ctx *fiber.Ctx) error {
	id, err := idFromParam(ctx, "category_id")
	if err != nil {
		return err
	}
	category, err := handler.services.Forums.GetCategory(ctx.Context(), id)
	if err != nil {
		return handleError(ctx, err)
	}
	setETag(ctx, category.Version)
	return writeJSON(ctx, fiber.StatusOK, category)
}

// listCategories lists categories.
func (handler handler) listCategories(ctx *fiber.Ctx) error {
	page, err := pageFromQuery(ctx)
	if err != nil {
		return err
	}
	result, err := handler.services.Forums.ListCategories(ctx.Context(), port.CategoryFilter{Status: domain.CategoryStatus(ctx.Query("status"))}, page)
	if err != nil {
		return handleError(ctx, err)
	}
	return writeJSON(ctx, fiber.StatusOK, categoryListResponse{Items: result.Items, NextPageToken: result.NextCursor})
}

// updateCategory updates a category.
func (handler handler) updateCategory(ctx *fiber.Ctx) error {
	if err := requireIdempotency(ctx); err != nil {
		return err
	}
	actor, err := currentUserID(ctx)
	if err != nil {
		return err
	}
	id, err := idFromParam(ctx, "category_id")
	if err != nil {
		return err
	}
	version, err := expectedVersion(ctx)
	if err != nil {
		return err
	}
	var request categoryRequest
	if err := decodeJSON(ctx, &request); err != nil {
		return err
	}
	category, err := handler.services.Forums.UpdateCategory(ctx.Context(), port.UpdateCategoryCommand{ActorUserID: actor, Category: categoryFromRequest(id, request), ExpectedVersion: version})
	if err != nil {
		return handleError(ctx, err)
	}
	setETag(ctx, category.Version)
	return writeJSON(ctx, fiber.StatusOK, category)
}

// deleteCategory deletes a category.
func (handler handler) deleteCategory(ctx *fiber.Ctx) error {
	if err := requireIdempotency(ctx); err != nil {
		return err
	}
	actor, err := currentUserID(ctx)
	if err != nil {
		return err
	}
	id, err := idFromParam(ctx, "category_id")
	if err != nil {
		return err
	}
	version, err := expectedVersion(ctx)
	if err != nil {
		return err
	}
	if err := handler.services.Forums.DeleteCategory(ctx.Context(), port.DeleteCategoryCommand{ActorUserID: actor, ID: id, ExpectedVersion: version}); err != nil {
		return handleError(ctx, err)
	}
	return writeNoContent(ctx)
}

// reorderCategories reorders categories.
func (handler handler) reorderCategories(ctx *fiber.Ctx) error {
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
	if err := handler.services.Forums.ReorderCategories(ctx.Context(), port.ReorderCategoriesCommand{ActorUserID: actor, Items: request.Items}); err != nil {
		return handleError(ctx, err)
	}
	return writeNoContent(ctx)
}

// categoryFromRequest maps request data to domain.
func categoryFromRequest(id uuid.UUID, request categoryRequest) domain.ForumCategory {
	return domain.ForumCategory{ID: id, Key: request.Key, Name: request.Name, Description: request.Description, DisplayOrder: request.DisplayOrder, Status: request.Status}
}
