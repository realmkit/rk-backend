package structure

import (
	"context"

	"github.com/google/uuid"
	"github.com/niflaot/gamehub-go/module/forums/domain"
	"github.com/niflaot/gamehub-go/module/forums/port"
	"github.com/niflaot/gamehub-go/pkg/pagination"
)

// CreateCategory creates a category.
func (service Service) CreateCategory(
	ctx context.Context,
	command port.CreateCategoryCommand,
) (domain.ForumCategory, error) {
	if err := service.requireManage(ctx, command.ActorUserID, domain.RootForumObjectID()); err != nil {
		return domain.ForumCategory{}, err
	}
	category := command.Category.Normalize()
	if category.ID == uuid.Nil {
		category.ID = uuid.New()
	}
	if err := category.Validate(); err != nil {
		return domain.ForumCategory{}, err
	}
	created, err := service.categories.Create(ctx, category)
	if err != nil {
		return domain.ForumCategory{}, err
	}
	if err := service.clearTree(ctx); err != nil {
		return domain.ForumCategory{}, err
	}
	return created, service.publishCategoryEvent(
		ctx,
		"forums.category.created",
		created,
		command.ActorUserID,
	)
}

// UpdateCategory updates a category.
func (service Service) UpdateCategory(
	ctx context.Context,
	command port.UpdateCategoryCommand,
) (domain.ForumCategory, error) {
	if err := service.requireManage(ctx, command.ActorUserID, domain.RootForumObjectID()); err != nil {
		return domain.ForumCategory{}, err
	}
	category := command.Category.Normalize()
	if err := category.Validate(); err != nil {
		return domain.ForumCategory{}, err
	}
	updated, err := service.categories.Update(ctx, category, command.ExpectedVersion)
	if err != nil {
		return domain.ForumCategory{}, err
	}
	if err := service.clearTree(ctx); err != nil {
		return domain.ForumCategory{}, err
	}
	return updated, service.publishCategoryEvent(
		ctx,
		"forums.category.updated",
		updated,
		command.ActorUserID,
	)
}

// GetCategory returns one category.
func (service Service) GetCategory(
	ctx context.Context,
	id uuid.UUID,
) (domain.ForumCategory, error) {
	return service.categories.FindByID(ctx, id)
}

// ListCategories lists categories.
func (service Service) ListCategories(
	ctx context.Context,
	filter port.CategoryFilter,
	page pagination.Page,
) (pagination.Result[domain.ForumCategory], error) {
	return service.categories.List(ctx, filter, page)
}

// DeleteCategory deletes a category.
func (service Service) DeleteCategory(
	ctx context.Context,
	command port.DeleteCategoryCommand,
) error {
	if err := service.requireManage(ctx, command.ActorUserID, domain.RootForumObjectID()); err != nil {
		return err
	}
	if err := service.categories.Delete(ctx, command.ID, command.ExpectedVersion); err != nil {
		return err
	}
	if err := service.clearTree(ctx); err != nil {
		return err
	}
	return service.publishCategoryEvent(
		ctx,
		"forums.category.deleted",
		domain.ForumCategory{ID: command.ID},
		command.ActorUserID,
	)
}

// ReorderCategories reorders categories.
func (service Service) ReorderCategories(
	ctx context.Context,
	command port.ReorderCategoriesCommand,
) error {
	if err := service.requireManage(ctx, command.ActorUserID, domain.RootForumObjectID()); err != nil {
		return err
	}
	if err := service.validateReorder(command.Items); err != nil {
		return err
	}
	if err := service.categories.Reorder(ctx, command.Items); err != nil {
		return err
	}
	return service.clearTree(ctx)
}
