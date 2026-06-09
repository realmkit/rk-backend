package port

import (
	"context"

	"github.com/google/uuid"
	"github.com/niflaot/gamehub-go/module/forums/domain"
	"github.com/niflaot/gamehub-go/pkg/pagination"
)

// CategoryRepository stores forum categories.
type CategoryRepository interface {
	// Create stores a category.
	Create(ctx context.Context, category domain.ForumCategory) (domain.ForumCategory, error)

	// Update stores mutable category fields.
	Update(ctx context.Context, category domain.ForumCategory, expectedVersion uint64) (domain.ForumCategory, error)

	// FindByID returns one category.
	FindByID(ctx context.Context, id uuid.UUID) (domain.ForumCategory, error)

	// List returns matching categories.
	List(ctx context.Context, filter CategoryFilter, page pagination.Page) (pagination.Result[domain.ForumCategory], error)

	// Delete soft deletes one category.
	Delete(ctx context.Context, id uuid.UUID, expectedVersion uint64) error

	// Reorder updates category display order.
	Reorder(ctx context.Context, items []ReorderItem) error
}

// ForumRepository stores forums and forum tree data.
type ForumRepository interface {
	// Create stores a forum and creates its stats row.
	Create(ctx context.Context, forum domain.Forum) (domain.Forum, error)

	// Update stores mutable forum fields.
	Update(ctx context.Context, forum domain.Forum, expectedVersion uint64) (domain.Forum, error)

	// FindByID returns one forum.
	FindByID(ctx context.Context, id uuid.UUID) (domain.Forum, error)

	// List returns matching forums.
	List(ctx context.Context, filter ForumFilter, page pagination.Page) (pagination.Result[domain.Forum], error)

	// ListTreeForums returns forums used by tree reads.
	ListTreeForums(ctx context.Context) ([]domain.Forum, error)

	// ListStats returns stats for forum ids.
	ListStats(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]domain.ForumStats, error)

	// Move changes a forum path and descendant paths.
	Move(ctx context.Context, forum domain.Forum, oldPath string, expectedVersion uint64) (domain.Forum, error)

	// Delete soft deletes one forum.
	Delete(ctx context.Context, id uuid.UUID, expectedVersion uint64) error

	// Reorder updates forum display order.
	Reorder(ctx context.Context, items []ReorderItem) error
}

// VisibilityAuthorizer checks forum visibility and management permissions.
type VisibilityAuthorizer interface {
	// VisibleForums returns visible forum IDs for actor.
	VisibleForums(ctx context.Context, actorUserID uuid.UUID, forumIDs []uuid.UUID) (map[uuid.UUID]bool, error)

	// CanManageForum reports whether actor can manage target forum.
	CanManageForum(ctx context.Context, actorUserID uuid.UUID, forumID uuid.UUID) (bool, error)

	// CanCreateThread reports whether actor can create a thread in forum.
	CanCreateThread(ctx context.Context, actorUserID uuid.UUID, forumID uuid.UUID) (bool, error)

	// CanReply reports whether actor can reply in forum.
	CanReply(ctx context.Context, actorUserID uuid.UUID, forumID uuid.UUID) (bool, error)

	// CanLikePosts reports whether actor can like posts in forum.
	CanLikePosts(ctx context.Context, actorUserID uuid.UUID, forumID uuid.UUID) (bool, error)

	// CanManageThreads reports whether actor can manage threads in forum.
	CanManageThreads(ctx context.Context, actorUserID uuid.UUID, forumID uuid.UUID) (bool, error)

	// CanManagePosts reports whether actor can manage posts in forum.
	CanManagePosts(ctx context.Context, actorUserID uuid.UUID, forumID uuid.UUID) (bool, error)
}
