package port

import (
	"github.com/google/uuid"
	"github.com/niflaot/gamehub-go/module/forums/domain"
	"github.com/niflaot/gamehub-go/pkg/pagination"
)

// CreateCategoryCommand creates a forum category.
type CreateCategoryCommand struct {
	// ActorUserID is the user performing the command.
	ActorUserID uuid.UUID

	// Category is the category to create.
	Category domain.ForumCategory
}

// UpdateCategoryCommand updates a forum category.
type UpdateCategoryCommand struct {
	// ActorUserID is the user performing the command.
	ActorUserID uuid.UUID

	// Category is the replacement category.
	Category domain.ForumCategory

	// ExpectedVersion is the required current version.
	ExpectedVersion uint64
}

// DeleteCategoryCommand deletes a forum category.
type DeleteCategoryCommand struct {
	// ActorUserID is the user performing the command.
	ActorUserID uuid.UUID

	// ID is the category identifier.
	ID uuid.UUID

	// ExpectedVersion is the required current version.
	ExpectedVersion uint64
}

// ReorderItem changes one resource display order.
type ReorderItem struct {
	// ID is the reordered resource identifier.
	ID uuid.UUID `json:"id"`

	// DisplayOrder is the replacement display order.
	DisplayOrder int `json:"display_order"`
}

// ReorderCategoriesCommand reorders categories.
type ReorderCategoriesCommand struct {
	// ActorUserID is the user performing the command.
	ActorUserID uuid.UUID

	// Items contains order changes.
	Items []ReorderItem
}

// CreateForumCommand creates a forum.
type CreateForumCommand struct {
	// ActorUserID is the user performing the command.
	ActorUserID uuid.UUID

	// Forum is the forum to create.
	Forum domain.Forum
}

// UpdateForumCommand updates a forum.
type UpdateForumCommand struct {
	// ActorUserID is the user performing the command.
	ActorUserID uuid.UUID

	// Forum is the replacement forum.
	Forum domain.Forum

	// ExpectedVersion is the required current version.
	ExpectedVersion uint64
}

// MoveForumCommand moves one forum in the tree.
type MoveForumCommand struct {
	// ActorUserID is the user performing the command.
	ActorUserID uuid.UUID

	// ID is the moved forum identifier.
	ID uuid.UUID

	// CategoryID is the replacement category.
	CategoryID uuid.UUID

	// ParentForumID is the replacement parent forum.
	ParentForumID *uuid.UUID

	// DisplayOrder is the replacement display order.
	DisplayOrder int

	// ExpectedVersion is the required current version.
	ExpectedVersion uint64
}

// DeleteForumCommand deletes a forum.
type DeleteForumCommand struct {
	// ActorUserID is the user performing the command.
	ActorUserID uuid.UUID

	// ID is the forum identifier.
	ID uuid.UUID

	// ExpectedVersion is the required current version.
	ExpectedVersion uint64
}

// ReorderForumsCommand reorders sibling forums.
type ReorderForumsCommand struct {
	// ActorUserID is the user performing the command.
	ActorUserID uuid.UUID

	// Items contains order changes.
	Items []ReorderItem
}

// CategoryFilter filters categories.
type CategoryFilter struct {
	// Status filters by category status.
	Status domain.CategoryStatus
}

// ForumFilter filters forums.
type ForumFilter struct {
	// CategoryID filters by category.
	CategoryID uuid.UUID

	// ParentForumID filters by parent forum.
	ParentForumID *uuid.UUID

	// Status filters by forum status.
	Status domain.ForumStatus
}

// Page aliases the shared pagination page.
type Page = pagination.Page

// Result aliases the shared pagination result.
type Result[T any] = pagination.Result[T]
