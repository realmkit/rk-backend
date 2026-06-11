package operations

import (
	"context"

	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/module/forums/domain"
	"github.com/realmkit/rk-backend/module/forums/port"
)

// visibleForumIDs returns forum IDs visible to the actor.
func (service Service) visibleForumIDs(
	ctx context.Context,
	actorUserID uuid.UUID,
	forumID uuid.UUID,
) ([]uuid.UUID, error) {
	forumIDs, err := service.candidateForumIDs(ctx, forumID)
	if err != nil {
		return nil, err
	}
	visible, err := service.authorizer.VisibleForums(ctx, actorUserID, forumIDs)
	if err != nil {
		return nil, err
	}
	result := make([]uuid.UUID, 0, len(forumIDs))
	for _, id := range forumIDs {
		if visible[id] {
			result = append(result, id)
		}
	}
	return result, nil
}

// candidateForumIDs returns either one scoped forum ID or all active forums.
func (service Service) candidateForumIDs(
	ctx context.Context,
	forumID uuid.UUID,
) ([]uuid.UUID, error) {
	if forumID != uuid.Nil {
		return []uuid.UUID{forumID}, nil
	}
	forums, err := service.forums.List(
		ctx,
		port.ForumFilter{Status: domain.ForumStatusActive},
		port.Page{Limit: 1000},
	)
	if err != nil {
		return nil, err
	}
	forumIDs := make([]uuid.UUID, 0, len(forums.Items))
	for _, forum := range forums.Items {
		forumIDs = append(forumIDs, forum.ID)
	}
	return forumIDs, nil
}
