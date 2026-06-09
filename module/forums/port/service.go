package port

import (
	"context"

	"github.com/google/uuid"
	"github.com/niflaot/gamehub-go/module/forums/domain"
	"github.com/niflaot/gamehub-go/pkg/pagination"
)

// Service manages forum structure.
type Service interface {
	// CreateCategory creates a category.
	CreateCategory(ctx context.Context, command CreateCategoryCommand) (domain.ForumCategory, error)

	// UpdateCategory updates a category.
	UpdateCategory(ctx context.Context, command UpdateCategoryCommand) (domain.ForumCategory, error)

	// GetCategory returns one category.
	GetCategory(ctx context.Context, id uuid.UUID) (domain.ForumCategory, error)

	// ListCategories lists categories.
	ListCategories(ctx context.Context, filter CategoryFilter, page pagination.Page) (pagination.Result[domain.ForumCategory], error)

	// DeleteCategory deletes a category.
	DeleteCategory(ctx context.Context, command DeleteCategoryCommand) error

	// ReorderCategories reorders categories.
	ReorderCategories(ctx context.Context, command ReorderCategoriesCommand) error

	// CreateForum creates a forum.
	CreateForum(ctx context.Context, command CreateForumCommand) (domain.Forum, error)

	// UpdateForum updates a forum.
	UpdateForum(ctx context.Context, command UpdateForumCommand) (domain.Forum, error)

	// MoveForum moves a forum.
	MoveForum(ctx context.Context, command MoveForumCommand) (domain.Forum, error)

	// GetForum returns one forum.
	GetForum(ctx context.Context, id uuid.UUID) (domain.Forum, error)

	// ListForums lists forums.
	ListForums(ctx context.Context, filter ForumFilter, page pagination.Page) (pagination.Result[domain.Forum], error)

	// DeleteForum deletes a forum.
	DeleteForum(ctx context.Context, command DeleteForumCommand) error

	// ReorderForums reorders forums.
	ReorderForums(ctx context.Context, command ReorderForumsCommand) error

	// Tree returns the visible forum tree.
	Tree(ctx context.Context, actorUserID uuid.UUID) (domain.ForumTree, error)
}
