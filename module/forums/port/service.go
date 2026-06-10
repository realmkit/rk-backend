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

	// CreateThread creates a thread with its opener post.
	CreateThread(ctx context.Context, command CreateThreadCommand) (domain.Thread, domain.Post, error)

	// GetThread returns one thread.
	GetThread(ctx context.Context, actorUserID uuid.UUID, id uuid.UUID) (domain.Thread, error)

	// ListThreads lists threads.
	ListThreads(ctx context.Context, actorUserID uuid.UUID, filter ThreadFilter, page pagination.Page) (pagination.Result[domain.Thread], error)

	// UpdateThreadTitle updates thread title fields.
	UpdateThreadTitle(ctx context.Context, command UpdateThreadTitleCommand) (domain.Thread, error)

	// DeleteThread deletes a thread.
	DeleteThread(ctx context.Context, command DeleteThreadCommand) error

	// CreateReply creates a reply post.
	CreateReply(ctx context.Context, command CreateReplyCommand) (domain.Post, error)

	// ListPosts lists posts for a thread.
	ListPosts(ctx context.Context, actorUserID uuid.UUID, filter PostFilter, page pagination.Page) (pagination.Result[domain.Post], error)

	// GetPost returns one post.
	GetPost(ctx context.Context, actorUserID uuid.UUID, id uuid.UUID) (domain.Post, error)

	// UpdatePost updates one post.
	UpdatePost(ctx context.Context, command UpdatePostCommand) (domain.Post, error)

	// DeletePost deletes one post.
	DeletePost(ctx context.Context, command DeletePostCommand) error

	// ListPostRevisions lists post revisions.
	ListPostRevisions(ctx context.Context, actorUserID uuid.UUID, postID uuid.UUID, page pagination.Page) (pagination.Result[domain.PostRevision], error)

	// LikePost likes one post.
	LikePost(ctx context.Context, command LikePostCommand) (domain.PostLikeSummary, error)

	// UnlikePost unlikes one post.
	UnlikePost(ctx context.Context, command UnlikePostCommand) (domain.PostLikeSummary, error)

	// ListLatestPosts lists latest posts for visible forums.
	ListLatestPosts(ctx context.Context, actorUserID uuid.UUID, forumID uuid.UUID, page pagination.Page) (pagination.Result[domain.LatestPostSummary], error)

	// ListMostLikedPosts lists most-liked posts for one forum.
	ListMostLikedPosts(ctx context.Context, actorUserID uuid.UUID, forumID uuid.UUID, page pagination.Page) (pagination.Result[domain.MostLikedPost], error)

	// MarkThreadRead marks one thread read.
	MarkThreadRead(ctx context.Context, command MarkThreadReadCommand) (domain.ThreadReadState, error)

	// MarkForumRead marks visible forum threads read.
	MarkForumRead(ctx context.Context, command MarkForumReadCommand) error

	// GetUnreadSummary returns unread counts for the actor.
	GetUnreadSummary(ctx context.Context, actorUserID uuid.UUID) (domain.UnreadSummary, error)

	// Search searches visible forum content.
	Search(ctx context.Context, command SearchCommand, page pagination.Page) (pagination.Result[domain.SearchResult], error)

	// VerifyStats reports stats counter drift.
	VerifyStats(ctx context.Context) (domain.CounterDriftReport, error)

	// RebuildStats repairs stats counters.
	RebuildStats(ctx context.Context) (domain.CounterDriftReport, error)

	// VerifyLikes reports like counter drift.
	VerifyLikes(ctx context.Context) (domain.CounterDriftReport, error)

	// RebuildLikes repairs like counters.
	RebuildLikes(ctx context.Context) (domain.CounterDriftReport, error)

	// FlushThreadViews persists buffered view counters.
	FlushThreadViews(ctx context.Context) (int64, error)

	// ClearReadCache clears forum read caches.
	ClearReadCache(ctx context.Context) error
}
