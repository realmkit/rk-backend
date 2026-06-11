package port

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/niflaot/gamehub-go/module/forums/domain"
	"github.com/niflaot/gamehub-go/pkg/pagination"
)

// ErrNotFound reports that a forum resource was not found.
var ErrNotFound = errors.New("forum resource not found")

// ErrConflict reports a conflicting forum state.
var ErrConflict = errors.New("forum resource conflict")

// ErrPreconditionFailed reports a stale optimistic version.
var ErrPreconditionFailed = errors.New("forum precondition failed")

// ErrForbidden reports a denied forum permission.
var ErrForbidden = errors.New("forum permission denied")

// ErrInvalidMove reports an invalid forum tree move.
var ErrInvalidMove = errors.New("invalid forum move")

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

// PermissionAdmin manages forum permission settings and simulations.
type PermissionAdmin interface {
	// ForumPermissionSettings returns permission grants for a forum.
	ForumPermissionSettings(ctx context.Context, forumID uuid.UUID) (domain.ForumPermissionSettings, error)

	// UpdateForumPermissionSettings replaces permission grants for a forum.
	UpdateForumPermissionSettings(ctx context.Context, actorUserID uuid.UUID, settings domain.ForumPermissionSettings) error

	// SimulateForumPermission explains a forum permission decision.
	SimulateForumPermission(
		ctx context.Context,
		forumID uuid.UUID,
		request domain.ForumPermissionSimulationRequest,
	) (domain.ForumPermissionSimulationResult, error)
}

// ReadCache caches visible forum read paths.
type ReadCache interface {
	// GetTree returns a cached tree when present.
	GetTree(ctx context.Context, key string) (domain.ForumTree, bool, error)

	// SetTree stores a tree for ttl.
	SetTree(ctx context.Context, key string, tree domain.ForumTree, ttl time.Duration) error

	// ClearTree removes forum tree cache entries.
	ClearTree(ctx context.Context) error

	// GetLatestPosts returns a cached latest-post page when present.
	GetLatestPosts(ctx context.Context, key string) (pagination.Result[domain.LatestPostSummary], bool, error)

	// SetLatestPosts stores a latest-post page for ttl.
	SetLatestPosts(ctx context.Context, key string, result pagination.Result[domain.LatestPostSummary], ttl time.Duration) error

	// ClearLatestPosts removes latest-post cache entries.
	ClearLatestPosts(ctx context.Context) error

	// GetMostLikedPosts returns a cached most-liked page when present.
	GetMostLikedPosts(ctx context.Context, key string) (pagination.Result[domain.MostLikedPost], bool, error)

	// SetMostLikedPosts stores a most-liked page for ttl.
	SetMostLikedPosts(ctx context.Context, key string, result pagination.Result[domain.MostLikedPost], ttl time.Duration) error

	// ClearMostLikedPosts removes most-liked cache entries.
	ClearMostLikedPosts(ctx context.Context) error

	// IncrementThreadView buffers one thread view.
	IncrementThreadView(ctx context.Context, threadID string) error

	// DrainThreadViews atomically returns and clears buffered thread views.
	DrainThreadViews(ctx context.Context) (map[string]int64, error)

	// ClearAll removes all forum read-cache keys.
	ClearAll(ctx context.Context) error
}
