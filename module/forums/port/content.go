// Package port defines forum application contracts.
package port

import (
	"context"

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

// AssetResolver validates attachment references against the assets module.
type AssetResolver interface {
	// AssetExists reports whether an attachment asset exists.
	AssetExists(ctx context.Context, id uuid.UUID) (bool, error)
}
