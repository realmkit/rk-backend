package content

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/niflaot/gamehub-go/module/forums/domain"
	"github.com/niflaot/gamehub-go/module/forums/port"
)

// authorPostEditWindow is the default author self-edit window before admin configuration exists.
const authorPostEditWindow = 10 * time.Minute

// authorPostDeleteWindow is the default author self-delete window before admin configuration exists.
const authorPostDeleteWindow = 5 * time.Minute

func (service Service) validateReferences(
	ctx context.Context,
	actorUserID uuid.UUID,
	references []domain.PostReference,
) error {
	for _, reference := range references {
		if reference.TargetPostID != nil {
			if _, err := service.GetPost(ctx, actorUserID, *reference.TargetPostID); err != nil {
				return err
			}
		}
		if reference.TargetAssetID != nil && service.assets != nil {
			exists, err := service.assets.AssetExists(ctx, *reference.TargetAssetID)
			if err != nil {
				return err
			}
			if !exists {
				return port.ErrNotFound
			}
		}
	}
	return nil
}

func (service Service) requirePostUpdate(
	ctx context.Context,
	actorUserID uuid.UUID,
	post domain.Post,
) error {
	if actorUserID != post.AuthorUserID {
		return service.requireManagePosts(ctx, actorUserID, post.ForumID)
	}
	allowed, err := service.authorCanUpdatePost(ctx, post)
	if err != nil {
		return err
	}
	if !allowed {
		return port.ErrForbidden
	}
	return nil
}

func (service Service) requirePostDelete(
	ctx context.Context,
	actorUserID uuid.UUID,
	post domain.Post,
) error {
	if actorUserID != post.AuthorUserID {
		return service.requireManagePosts(ctx, actorUserID, post.ForumID)
	}
	allowed, err := service.authorCanDeletePost(ctx, post)
	if err != nil {
		return err
	}
	if !allowed {
		return port.ErrForbidden
	}
	return nil
}

func (service Service) authorCanUpdatePost(
	ctx context.Context,
	post domain.Post,
) (bool, error) {
	forum, err := service.forums.FindByID(ctx, post.ForumID)
	if err != nil {
		return false, err
	}
	if !insideAuthorWindow(post.CreatedAt, forum.AuthorPostEditWindowSeconds, authorPostEditWindow) {
		return false, nil
	}
	thread, err := service.threads.FindByID(ctx, post.ThreadID)
	if err != nil {
		return false, err
	}
	return thread.Replyable(), nil
}

func (service Service) authorCanDeletePost(
	ctx context.Context,
	post domain.Post,
) (bool, error) {
	forum, err := service.forums.FindByID(ctx, post.ForumID)
	if err != nil {
		return false, err
	}
	return insideAuthorWindow(
		post.CreatedAt,
		forum.AuthorPostDeleteWindowSeconds,
		authorPostDeleteWindow,
	), nil
}

func insideAuthorWindow(
	createdAt time.Time,
	configuredSeconds int,
	fallback time.Duration,
) bool {
	if createdAt.IsZero() || configuredSeconds < 0 {
		return false
	}
	window := time.Duration(configuredSeconds) * time.Second
	if configuredSeconds == 0 {
		window = fallback
	}
	return time.Since(createdAt) <= window
}

func (service Service) requireThreadCreate(
	ctx context.Context,
	actorUserID uuid.UUID,
	forumID uuid.UUID,
) error {
	allowed, err := service.authorizer.CanCreateThread(ctx, actorUserID, forumID)
	return decisionError(allowed, err)
}

func (service Service) requireReply(
	ctx context.Context,
	actorUserID uuid.UUID,
	forumID uuid.UUID,
) error {
	allowed, err := service.authorizer.CanReply(ctx, actorUserID, forumID)
	return decisionError(allowed, err)
}

func (service Service) requireLikePosts(
	ctx context.Context,
	actorUserID uuid.UUID,
	forumID uuid.UUID,
) error {
	allowed, err := service.authorizer.CanLikePosts(ctx, actorUserID, forumID)
	return decisionError(allowed, err)
}

func (service Service) requireManageThreads(
	ctx context.Context,
	actorUserID uuid.UUID,
	forumID uuid.UUID,
) error {
	allowed, err := service.authorizer.CanManageThreads(ctx, actorUserID, forumID)
	return decisionError(allowed, err)
}

func (service Service) requireManagePosts(
	ctx context.Context,
	actorUserID uuid.UUID,
	forumID uuid.UUID,
) error {
	allowed, err := service.authorizer.CanManagePosts(ctx, actorUserID, forumID)
	return decisionError(allowed, err)
}

func (service Service) requireThreadView(
	ctx context.Context,
	actorUserID uuid.UUID,
	thread domain.Thread,
) error {
	visible, err := service.authorizer.VisibleForums(ctx, actorUserID, []uuid.UUID{thread.ForumID})
	if err != nil {
		return err
	}
	if visible[thread.ForumID] && thread.Visible() {
		return nil
	}
	if actorUserID == thread.AuthorUserID {
		return nil
	}
	return service.requireManageThreads(ctx, actorUserID, thread.ForumID)
}

func decisionError(allowed bool, err error) error {
	if err != nil {
		return err
	}
	if !allowed {
		return port.ErrForbidden
	}
	return nil
}
