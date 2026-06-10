package port

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/niflaot/gamehub-go/module/forums/domain"
	"github.com/niflaot/gamehub-go/pkg/pagination"
)

// LikePostCommand likes a post.
type LikePostCommand struct {
	// ActorUserID is the liking user.
	ActorUserID uuid.UUID

	// PostID is the target post.
	PostID uuid.UUID
}

// UnlikePostCommand unlikes a post.
type UnlikePostCommand struct {
	// ActorUserID is the unliking user.
	ActorUserID uuid.UUID

	// PostID is the target post.
	PostID uuid.UUID
}

// MarkThreadReadCommand marks one thread as read.
type MarkThreadReadCommand struct {
	// ActorUserID is the reading user.
	ActorUserID uuid.UUID

	// ThreadID is the target thread.
	ThreadID uuid.UUID

	// LastReadPostSequence is the highest read post sequence.
	LastReadPostSequence int64
}

// MarkForumReadCommand marks visible threads in one forum as read.
type MarkForumReadCommand struct {
	// ActorUserID is the reading user.
	ActorUserID uuid.UUID

	// ForumID is the target forum.
	ForumID uuid.UUID
}

// SearchCommand searches forum content.
type SearchCommand struct {
	// ActorUserID is the searching actor.
	ActorUserID uuid.UUID

	// ForumID optionally scopes search to one forum.
	ForumID uuid.UUID

	// Query is the search text.
	Query string
}

// LatestPostFilter filters latest-post widgets.
type LatestPostFilter struct {
	// ForumIDs limits results to visible forums.
	ForumIDs []uuid.UUID
}

// MostLikedFilter filters most-liked widgets.
type MostLikedFilter struct {
	// ForumID filters by forum.
	ForumID uuid.UUID
}

// SearchFilter filters forum search.
type SearchFilter struct {
	// ForumIDs limits search to visible forums.
	ForumIDs []uuid.UUID

	// Query is the normalized search text.
	Query string
}

// InteractionRepository stores likes, widgets, and read state.
type InteractionRepository interface {
	// LikePost creates or restores an active like and returns whether counters changed.
	LikePost(ctx context.Context, like domain.PostLike) (bool, error)

	// UnlikePost removes an active like and returns whether counters changed.
	UnlikePost(ctx context.Context, postID uuid.UUID, userID uuid.UUID) (bool, error)

	// LikedByUser reports whether user currently likes post.
	LikedByUser(ctx context.Context, postID uuid.UUID, userID uuid.UUID) (bool, error)

	// ListLatestPosts returns latest visible post summaries.
	ListLatestPosts(ctx context.Context, filter LatestPostFilter, page pagination.Page) (pagination.Result[domain.LatestPostSummary], error)

	// ListMostLikedPosts returns most-liked visible posts.
	ListMostLikedPosts(ctx context.Context, filter MostLikedFilter, page pagination.Page) (pagination.Result[domain.MostLikedPost], error)

	// MarkThreadRead stores one thread read state.
	MarkThreadRead(ctx context.Context, state domain.ThreadReadState) error

	// MarkForumRead stores read states for every visible thread in a forum.
	MarkForumRead(ctx context.Context, userID uuid.UUID, forumID uuid.UUID, readAt time.Time) error

	// UnreadSummary returns unread counts for visible forums.
	UnreadSummary(ctx context.Context, userID uuid.UUID, forumIDs []uuid.UUID) (domain.UnreadSummary, error)
}

// OperationsRepository runs forum search, repairs, and counter flushes.
type OperationsRepository interface {
	// Search returns visible search results from PostgreSQL.
	Search(ctx context.Context, filter SearchFilter, page pagination.Page) (pagination.Result[domain.SearchResult], error)

	// VerifyStats reports counter drift without mutating rows.
	VerifyStats(ctx context.Context) (domain.CounterDriftReport, error)

	// RebuildStats repairs stats and post/thread counters from source rows.
	RebuildStats(ctx context.Context) (domain.CounterDriftReport, error)

	// VerifyLikes reports like counter drift without mutating rows.
	VerifyLikes(ctx context.Context) (domain.CounterDriftReport, error)

	// RebuildLikes repairs like counters from active like rows.
	RebuildLikes(ctx context.Context) (domain.CounterDriftReport, error)

	// ApplyThreadViews flushes buffered view increments into threads.
	ApplyThreadViews(ctx context.Context, increments map[uuid.UUID]int64) error
}
