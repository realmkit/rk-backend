package interaction

import (
	"context"

	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/module/forums/domain"
	"github.com/realmkit/rk-backend/module/forums/port"
)

// requireLikePosts verifies like permission.
func (service Service) requireLikePosts(
	ctx context.Context,
	actorUserID uuid.UUID,
	forumID uuid.UUID,
) error {
	allowed, err := service.authorizer.CanLikePosts(ctx, actorUserID, forumID)
	return decisionError(allowed, err)
}

// requireManageThreads verifies thread-management permission.
func (service Service) requireManageThreads(
	ctx context.Context,
	actorUserID uuid.UUID,
	forumID uuid.UUID,
) error {
	allowed, err := service.authorizer.CanManageThreads(ctx, actorUserID, forumID)
	return decisionError(allowed, err)
}

// requireThreadView verifies thread visibility.
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

// decisionError maps authorization decisions to use-case errors.
func decisionError(allowed bool, err error) error {
	if err != nil {
		return err
	}
	if !allowed {
		return port.ErrForbidden
	}
	return nil
}

// requireUnrestricted verifies punishment restrictions do not block the action.
func (service Service) requireUnrestricted(
	ctx context.Context,
	actorUserID uuid.UUID,
	actionKey string,
) error {
	if service.restrictions == nil || actorUserID == uuid.Nil {
		return nil
	}
	restricted, err := service.restrictions.Restricted(ctx, actorUserID, actionKey)
	if err != nil {
		return err
	}
	if restricted {
		return port.ErrForbidden
	}
	return nil
}
