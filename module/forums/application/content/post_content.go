package content

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/module/forums/domain"
	"github.com/realmkit/rk-backend/module/forums/port"
	"github.com/realmkit/rk-backend/pkg/pagination"
)

// CreateReply creates a reply post.
func (service Service) CreateReply(
	ctx context.Context,
	command port.CreateReplyCommand,
) (domain.Post, error) {
	thread, err := service.threads.FindByID(ctx, command.ThreadID)
	if err != nil {
		return domain.Post{}, err
	}
	if !thread.Replyable() {
		return domain.Post{}, port.ErrConflict
	}
	if err := service.requireReply(ctx, command.ActorUserID, thread.ForumID); err != nil {
		return domain.Post{}, err
	}
	post, err := service.replyPost(ctx, thread, command)
	if err != nil {
		return domain.Post{}, err
	}
	references, err := service.replyReferences(ctx, command.ActorUserID, post, command)
	if err != nil {
		return domain.Post{}, err
	}
	var created domain.Post
	err = service.transactions.WithinTx(ctx, func(ctx context.Context) error {
		stored, err := service.posts.Create(ctx, post, references)
		if err != nil {
			return err
		}
		created = stored
		if err := service.clearTree(ctx); err != nil {
			return err
		}
		return service.publishPostEvent(
			ctx,
			"forums.post.created",
			stored,
			command.ActorUserID,
		)
	})
	return created, err
}

// ListPosts lists posts for a visible thread.
func (service Service) ListPosts(
	ctx context.Context,
	actorUserID uuid.UUID,
	filter port.PostFilter,
	page pagination.Page,
) (pagination.Result[domain.Post], error) {
	thread, err := service.threads.FindByID(ctx, filter.ThreadID)
	if err != nil {
		return pagination.Result[domain.Post]{}, err
	}
	if filter.IncludeHidden {
		if err := service.requireManagePosts(ctx, actorUserID, thread.ForumID); err != nil {
			return pagination.Result[domain.Post]{}, err
		}
	} else if err := service.requireThreadView(ctx, actorUserID, thread); err != nil {
		return pagination.Result[domain.Post]{}, err
	}
	return service.posts.List(ctx, filter, page)
}

// GetPost returns one post.
func (service Service) GetPost(
	ctx context.Context,
	actorUserID uuid.UUID,
	id uuid.UUID,
) (domain.Post, error) {
	post, err := service.posts.FindByID(ctx, id)
	if err != nil {
		return domain.Post{}, err
	}
	if !post.Visible() {
		if err := service.requireManagePosts(ctx, actorUserID, post.ForumID); err != nil {
			return domain.Post{}, err
		}
		return post, nil
	}
	thread, err := service.threads.FindByID(ctx, post.ThreadID)
	if err != nil {
		return domain.Post{}, err
	}
	if err := service.requireThreadView(ctx, actorUserID, thread); err != nil {
		return domain.Post{}, err
	}
	return post, nil
}

// UpdatePost updates one post and records a revision.
func (service Service) UpdatePost(
	ctx context.Context,
	command port.UpdatePostCommand,
) (domain.Post, error) {
	current, err := service.posts.FindByID(ctx, command.PostID)
	if err != nil {
		return domain.Post{}, err
	}
	if err := service.requirePostUpdate(ctx, command.ActorUserID, current); err != nil {
		return domain.Post{}, err
	}
	updated := updatedPost(current, command)
	if err := updated.Validate(); err != nil {
		return domain.Post{}, err
	}
	revision := domain.PostRevision{
		ID:                          uuid.New(),
		PostID:                      current.ID,
		EditedByUserID:              command.ActorUserID,
		PreviousContentDocumentJSON: current.ContentDocumentJSON,
		PreviousContentText:         current.ContentText,
		EditReason:                  strings.TrimSpace(command.EditReason),
	}
	stored, err := service.posts.UpdateWithRevision(ctx, updated, revision, command.ExpectedVersion)
	if err != nil {
		return domain.Post{}, err
	}
	return stored, service.publishPostEvent(
		ctx,
		"forums.post.updated",
		stored,
		command.ActorUserID,
	)
}

// DeletePost deletes one post.
func (service Service) DeletePost(
	ctx context.Context,
	command port.DeletePostCommand,
) error {
	post, err := service.posts.FindByID(ctx, command.PostID)
	if err != nil {
		return err
	}
	if err := service.requirePostDelete(ctx, command.ActorUserID, post); err != nil {
		return err
	}
	if err := service.posts.Delete(ctx, command.PostID, command.ExpectedVersion); err != nil {
		return err
	}
	if err := service.clearTree(ctx); err != nil {
		return err
	}
	return service.publishPostEvent(
		ctx,
		"forums.post.deleted",
		post,
		command.ActorUserID,
	)
}

// ListPostRevisions lists post revisions for moderators.
func (service Service) ListPostRevisions(
	ctx context.Context,
	actorUserID uuid.UUID,
	postID uuid.UUID,
	page pagination.Page,
) (pagination.Result[domain.PostRevision], error) {
	post, err := service.posts.FindByID(ctx, postID)
	if err != nil {
		return pagination.Result[domain.PostRevision]{}, err
	}
	if err := service.requireManagePosts(ctx, actorUserID, post.ForumID); err != nil {
		return pagination.Result[domain.PostRevision]{}, err
	}
	return service.posts.ListRevisions(ctx, postID, page)
}

// replyPost supports package behavior.
func (service Service) replyPost(
	ctx context.Context,
	thread domain.Thread,
	command port.CreateReplyCommand,
) (domain.Post, error) {
	sequence, err := service.posts.NextSequence(ctx, thread.ID)
	if err != nil {
		return domain.Post{}, err
	}
	post := domain.Post{
		ID:                  uuid.New(),
		ThreadID:            thread.ID,
		ForumID:             thread.ForumID,
		AuthorUserID:        command.ActorUserID,
		Sequence:            sequence,
		ContentDocumentJSON: command.ContentDocumentJSON,
		ContentText:         contentText(command.ContentText, command.ContentDocumentJSON),
		ContentChecksum:     checksum(command.ContentChecksum, command.ContentDocumentJSON),
		Version:             1,
	}.Normalize()
	if err := post.Validate(); err != nil {
		return domain.Post{}, err
	}
	return post, nil
}

// replyReferences supports package behavior.
func (service Service) replyReferences(
	ctx context.Context,
	actorUserID uuid.UUID,
	post domain.Post,
	command port.CreateReplyCommand,
) ([]domain.PostReference, error) {
	combined := append(extractReferences(command.ContentDocumentJSON), command.References...)
	references := prepareReferences(post.ID, combined)
	for _, reference := range references {
		if err := reference.Validate(); err != nil {
			return nil, err
		}
	}
	if err := service.validateReferences(ctx, actorUserID, references); err != nil {
		return nil, err
	}
	return references, nil
}

// updatedPost supports package behavior.
func updatedPost(current domain.Post, command port.UpdatePostCommand) domain.Post {
	updated := current
	updated.ContentDocumentJSON = command.ContentDocumentJSON
	updated.ContentText = contentText(command.ContentText, command.ContentDocumentJSON)
	updated.ContentChecksum = checksum(command.ContentChecksum, command.ContentDocumentJSON)
	updated.EditCount++
	now := time.Now().UTC()
	updated.EditedAt = &now
	updated.EditedByUserID = &command.ActorUserID
	return updated.Normalize()
}
