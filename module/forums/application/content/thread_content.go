package content

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/module/forums/domain"
	"github.com/realmkit/rk-backend/module/forums/port"
	"github.com/realmkit/rk-backend/pkg/pagination"
)

// CreateThread creates a thread and opener post.
func (service Service) CreateThread(
	ctx context.Context,
	command port.CreateThreadCommand,
) (domain.Thread, domain.Post, error) {
	forum, err := service.forums.FindByID(ctx, command.ForumID)
	if err != nil {
		return domain.Thread{}, domain.Post{}, err
	}
	if !forum.Discussion() {
		return domain.Thread{}, domain.Post{}, port.ErrConflict
	}
	if err := service.requireThreadCreate(ctx, command.ActorUserID, forum.ID); err != nil {
		return domain.Thread{}, domain.Post{}, err
	}
	thread, post := openerThreadAndPost(forum, command, time.Now().UTC())
	if err := thread.Validate(); err != nil {
		return domain.Thread{}, domain.Post{}, err
	}
	if err := post.Validate(); err != nil {
		return domain.Thread{}, domain.Post{}, err
	}
	var createdThread domain.Thread
	var createdPost domain.Post
	err = service.transactions.WithinTx(ctx, func(ctx context.Context) error {
		storedThread, err := service.threads.Create(ctx, thread)
		if err != nil {
			return err
		}
		storedPost, err := service.posts.Create(ctx, post, nil)
		if err != nil {
			return err
		}
		createdThread = storedThread
		createdPost = storedPost
		if err := service.clearTree(ctx); err != nil {
			return err
		}
		if err := service.publishThreadEvent(
			ctx,
			"forums.thread.created",
			storedThread,
			command.ActorUserID,
		); err != nil {
			return err
		}
		return service.publishPostEvent(
			ctx,
			"forums.post.created",
			storedPost,
			command.ActorUserID,
		)
	})
	return createdThread, createdPost, err
}

// GetThread returns one visible thread.
func (service Service) GetThread(
	ctx context.Context,
	actorUserID uuid.UUID,
	id uuid.UUID,
) (domain.Thread, error) {
	thread, err := service.threads.FindByID(ctx, id)
	if err != nil {
		return domain.Thread{}, err
	}
	if err := service.requireThreadView(ctx, actorUserID, thread); err != nil {
		return domain.Thread{}, err
	}
	if service.cache != nil {
		_ = service.cache.IncrementThreadView(ctx, thread.ID.String())
	}
	return thread, nil
}

// ListThreads lists visible threads.
func (service Service) ListThreads(
	ctx context.Context,
	actorUserID uuid.UUID,
	filter port.ThreadFilter,
	page pagination.Page,
) (pagination.Result[domain.Thread], error) {
	forum, err := service.forums.FindByID(ctx, filter.ForumID)
	if err != nil {
		return pagination.Result[domain.Thread]{}, err
	}
	visible, err := service.authorizer.VisibleForums(ctx, actorUserID, []uuid.UUID{forum.ID})
	if err != nil {
		return pagination.Result[domain.Thread]{}, err
	}
	if !visible[forum.ID] {
		return pagination.Result[domain.Thread]{}, port.ErrForbidden
	}
	return service.threads.List(ctx, filter, page)
}

// UpdateThreadTitle updates one thread title.
func (service Service) UpdateThreadTitle(
	ctx context.Context,
	command port.UpdateThreadTitleCommand,
) (domain.Thread, error) {
	thread, err := service.threads.FindByID(ctx, command.ThreadID)
	if err != nil {
		return domain.Thread{}, err
	}
	if command.ActorUserID != thread.AuthorUserID {
		if err := service.requireManageThreads(ctx, command.ActorUserID, thread.ForumID); err != nil {
			return domain.Thread{}, err
		}
	}
	if err := service.requireUnrestricted(ctx, command.ActorUserID, "realmkit.forums.update_thread"); err != nil {
		return domain.Thread{}, err
	}
	thread.Title = command.Title
	thread.Slug = command.Slug
	thread = thread.Normalize()
	if err := thread.Validate(); err != nil {
		return domain.Thread{}, err
	}
	updated, err := service.threads.UpdateTitle(ctx, thread, command.ExpectedVersion)
	if err != nil {
		return domain.Thread{}, err
	}
	return updated, service.publishThreadEvent(
		ctx,
		"forums.thread.updated",
		updated,
		command.ActorUserID,
	)
}

// DeleteThread deletes one thread.
func (service Service) DeleteThread(
	ctx context.Context,
	command port.DeleteThreadCommand,
) error {
	thread, err := service.threads.FindByID(ctx, command.ThreadID)
	if err != nil {
		return err
	}
	if command.ActorUserID != thread.AuthorUserID {
		if err := service.requireManageThreads(ctx, command.ActorUserID, thread.ForumID); err != nil {
			return err
		}
	}
	if err := service.threads.Delete(ctx, command.ThreadID, command.ExpectedVersion); err != nil {
		return err
	}
	if err := service.clearTree(ctx); err != nil {
		return err
	}
	return service.publishThreadEvent(
		ctx,
		"forums.thread.deleted",
		thread,
		command.ActorUserID,
	)
}

func openerThreadAndPost(
	forum domain.Forum,
	command port.CreateThreadCommand,
	now time.Time,
) (domain.Thread, domain.Post) {
	threadID := uuid.New()
	postID := uuid.New()
	thread := domain.Thread{
		ID:                     threadID,
		ForumID:                forum.ID,
		AuthorUserID:           command.ActorUserID,
		OpenerPostID:           postID,
		LatestPostID:           postID,
		LatestPostAuthorUserID: command.ActorUserID,
		LatestPostAt:           now,
		Title:                  command.Title,
		Slug:                   command.Slug,
		Status:                 forum.DefaultThreadStatus,
		StickyState:            domain.StickyStateNormal,
		PostCount:              1,
		VisiblePostCount:       1,
		Version:                1,
	}.Normalize()
	post := domain.Post{
		ID:                  postID,
		ThreadID:            threadID,
		ForumID:             forum.ID,
		AuthorUserID:        command.ActorUserID,
		Sequence:            1,
		ContentDocumentJSON: command.ContentDocumentJSON,
		ContentText:         contentText(command.ContentText, command.ContentDocumentJSON),
		ContentChecksum:     checksum(command.ContentChecksum, command.ContentDocumentJSON),
		Version:             1,
	}.Normalize()
	return thread, post
}
