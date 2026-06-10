// Package port defines forum application contracts.
package port

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/niflaot/gamehub-go/module/forums/domain"
	"github.com/niflaot/gamehub-go/pkg/pagination"
)

// CreateThreadCommand creates a thread and opener post.
type CreateThreadCommand struct {
	// ActorUserID is the user performing the command.
	ActorUserID uuid.UUID

	// ForumID is the target forum.
	ForumID uuid.UUID

	// Title is the thread title.
	Title string

	// Slug is the thread slug.
	Slug domain.Slug

	// ContentDocumentJSON is the opener content document.
	ContentDocumentJSON []byte

	// ContentText is the opener extracted text.
	ContentText string

	// ContentChecksum is the opener content checksum.
	ContentChecksum string
}

// UpdateThreadTitleCommand updates a thread title.
type UpdateThreadTitleCommand struct {
	// ActorUserID is the user performing the command.
	ActorUserID uuid.UUID

	// ThreadID is the target thread.
	ThreadID uuid.UUID

	// Title is the replacement title.
	Title string

	// Slug is the replacement slug.
	Slug domain.Slug

	// ExpectedVersion is the required current version.
	ExpectedVersion uint64
}

// DeleteThreadCommand soft deletes a thread.
type DeleteThreadCommand struct {
	// ActorUserID is the user performing the command.
	ActorUserID uuid.UUID

	// ThreadID is the target thread.
	ThreadID uuid.UUID

	// ExpectedVersion is the required current version.
	ExpectedVersion uint64
}

// CreateReplyCommand creates a reply post.
type CreateReplyCommand struct {
	// ActorUserID is the user performing the command.
	ActorUserID uuid.UUID

	// ThreadID is the target thread.
	ThreadID uuid.UUID

	// ContentDocumentJSON is the reply content document.
	ContentDocumentJSON []byte

	// ContentText is the reply extracted text.
	ContentText string

	// ContentChecksum is the reply content checksum.
	ContentChecksum string

	// References are structured post references.
	References []domain.PostReference
}

// UpdatePostCommand updates a post and writes a revision.
type UpdatePostCommand struct {
	// ActorUserID is the user performing the command.
	ActorUserID uuid.UUID

	// PostID is the target post.
	PostID uuid.UUID

	// ContentDocumentJSON is the replacement content document.
	ContentDocumentJSON []byte

	// ContentText is the replacement extracted text.
	ContentText string

	// ContentChecksum is the replacement content checksum.
	ContentChecksum string

	// EditReason explains the edit.
	EditReason string

	// ExpectedVersion is the required current version.
	ExpectedVersion uint64
}

// DeletePostCommand soft deletes a post.
type DeletePostCommand struct {
	// ActorUserID is the user performing the command.
	ActorUserID uuid.UUID

	// PostID is the target post.
	PostID uuid.UUID

	// ExpectedVersion is the required current version.
	ExpectedVersion uint64
}

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

// ThreadFilter filters thread lists.
type ThreadFilter struct {
	// ForumID filters by forum.
	ForumID uuid.UUID

	// Status filters by thread status.
	Status domain.ThreadStatus

	// Section filters sticky or normal sections.
	Section string
}

// PostFilter filters post lists.
type PostFilter struct {
	// ThreadID filters by thread.
	ThreadID uuid.UUID

	// IncludeHidden includes hidden or pending posts when allowed.
	IncludeHidden bool
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

// ThreadRepository stores threads.
type ThreadRepository interface {
	// Create stores a thread.
	Create(ctx context.Context, thread domain.Thread) (domain.Thread, error)

	// FindByID returns one thread.
	FindByID(ctx context.Context, id uuid.UUID) (domain.Thread, error)

	// List returns matching threads.
	List(ctx context.Context, filter ThreadFilter, page pagination.Page) (pagination.Result[domain.Thread], error)

	// UpdateTitle updates thread title fields.
	UpdateTitle(ctx context.Context, thread domain.Thread, expectedVersion uint64) (domain.Thread, error)

	// Delete soft deletes a thread.
	Delete(ctx context.Context, id uuid.UUID, expectedVersion uint64) error
}

// PostRepository stores posts, revisions, and references.
type PostRepository interface {
	// Create stores a post with references.
	Create(ctx context.Context, post domain.Post, references []domain.PostReference) (domain.Post, error)

	// FindByID returns one post.
	FindByID(ctx context.Context, id uuid.UUID) (domain.Post, error)

	// List returns matching posts.
	List(ctx context.Context, filter PostFilter, page pagination.Page) (pagination.Result[domain.Post], error)

	// NextSequence returns the next post sequence for a thread.
	NextSequence(ctx context.Context, threadID uuid.UUID) (int64, error)

	// UpdateWithRevision updates a post and writes a revision.
	UpdateWithRevision(ctx context.Context, post domain.Post, revision domain.PostRevision, expectedVersion uint64) (domain.Post, error)

	// Delete soft deletes one post.
	Delete(ctx context.Context, id uuid.UUID, expectedVersion uint64) error

	// ListRevisions returns post revisions.
	ListRevisions(ctx context.Context, postID uuid.UUID, page pagination.Page) (pagination.Result[domain.PostRevision], error)

	// ListReferences returns references for posts.
	ListReferences(ctx context.Context, postIDs []uuid.UUID) (map[uuid.UUID][]domain.PostReference, error)
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

// AssetResolver validates attachment references against the assets module.
type AssetResolver interface {
	// AssetExists reports whether an attachment asset exists.
	AssetExists(ctx context.Context, id uuid.UUID) (bool, error)
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
