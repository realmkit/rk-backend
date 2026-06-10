package interaction

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/niflaot/gamehub-go/module/forums/domain"
	"github.com/niflaot/gamehub-go/module/forums/port"
	eventdomain "github.com/niflaot/gamehub-go/pkg/events/domain"
	"github.com/niflaot/gamehub-go/pkg/pagination"
)

// LikePost likes one post idempotently.
func (service Service) LikePost(
	ctx context.Context,
	command port.LikePostCommand,
) (domain.PostLikeSummary, error) {
	if command.ActorUserID == uuid.Nil {
		return domain.PostLikeSummary{}, port.ErrForbidden
	}
	post, err := service.posts.FindByID(ctx, command.PostID)
	if err != nil {
		return domain.PostLikeSummary{}, err
	}
	if err := service.requireLikePosts(ctx, command.ActorUserID, post.ForumID); err != nil {
		return domain.PostLikeSummary{}, err
	}
	like := newPostLike(post, command.ActorUserID)
	if err := like.Validate(); err != nil {
		return domain.PostLikeSummary{}, err
	}
	changed, err := service.interactions.LikePost(ctx, like)
	if err != nil {
		return domain.PostLikeSummary{}, err
	}
	if changed {
		_ = service.clearInteractionCaches(ctx)
		if err := service.publishPostInteraction(
			ctx,
			"forums.post.liked",
			post,
			command.ActorUserID,
		); err != nil {
			return domain.PostLikeSummary{}, err
		}
	}
	return service.postLikeSummary(ctx, post.ID, true)
}

// UnlikePost unlikes one post idempotently.
func (service Service) UnlikePost(
	ctx context.Context,
	command port.UnlikePostCommand,
) (domain.PostLikeSummary, error) {
	if command.ActorUserID == uuid.Nil {
		return domain.PostLikeSummary{}, port.ErrForbidden
	}
	post, err := service.posts.FindByID(ctx, command.PostID)
	if err != nil {
		return domain.PostLikeSummary{}, err
	}
	if err := service.requireLikePosts(ctx, command.ActorUserID, post.ForumID); err != nil {
		return domain.PostLikeSummary{}, err
	}
	changed, err := service.interactions.UnlikePost(ctx, post.ID, command.ActorUserID)
	if err != nil {
		return domain.PostLikeSummary{}, err
	}
	if changed {
		_ = service.clearInteractionCaches(ctx)
		if err := service.publishPostInteraction(
			ctx,
			"forums.post.unliked",
			post,
			command.ActorUserID,
		); err != nil {
			return domain.PostLikeSummary{}, err
		}
	}
	return service.postLikeSummary(ctx, post.ID, false)
}

// ListLatestPosts lists latest posts across visible forums.
func (service Service) ListLatestPosts(
	ctx context.Context,
	actorUserID uuid.UUID,
	forumID uuid.UUID,
	page pagination.Page,
) (pagination.Result[domain.LatestPostSummary], error) {
	forumIDs, err := service.visibleForumIDs(ctx, actorUserID, forumID)
	if err != nil {
		return pagination.Result[domain.LatestPostSummary]{}, err
	}
	key := latestPostsCacheKey(actorUserID, forumID, page)
	if service.cache != nil {
		if cached, ok, err := service.cache.GetLatestPosts(ctx, key); err == nil && ok {
			return cached, nil
		}
	}
	result, err := service.interactions.ListLatestPosts(ctx, port.LatestPostFilter{ForumIDs: forumIDs}, page)
	if err != nil {
		return pagination.Result[domain.LatestPostSummary]{}, err
	}
	if service.cache != nil {
		_ = service.cache.SetLatestPosts(ctx, key, result, widgetCacheTTL)
	}
	return result, nil
}

// ListMostLikedPosts lists most-liked posts for one visible forum.
func (service Service) ListMostLikedPosts(
	ctx context.Context,
	actorUserID uuid.UUID,
	forumID uuid.UUID,
	page pagination.Page,
) (pagination.Result[domain.MostLikedPost], error) {
	forumIDs, err := service.visibleForumIDs(ctx, actorUserID, forumID)
	if err != nil {
		return pagination.Result[domain.MostLikedPost]{}, err
	}
	if len(forumIDs) == 0 {
		return pagination.Result[domain.MostLikedPost]{}, port.ErrForbidden
	}
	key := mostLikedCacheKey(actorUserID, forumID, page)
	if service.cache != nil {
		if cached, ok, err := service.cache.GetMostLikedPosts(ctx, key); err == nil && ok {
			return cached, nil
		}
	}
	result, err := service.interactions.ListMostLikedPosts(ctx, port.MostLikedFilter{ForumID: forumIDs[0]}, page)
	if err != nil {
		return pagination.Result[domain.MostLikedPost]{}, err
	}
	if service.cache != nil {
		_ = service.cache.SetMostLikedPosts(ctx, key, result, widgetCacheTTL)
	}
	return result, nil
}

// MarkThreadRead stores read state for one thread.
func (service Service) MarkThreadRead(
	ctx context.Context,
	command port.MarkThreadReadCommand,
) (domain.ThreadReadState, error) {
	if command.ActorUserID == uuid.Nil {
		return domain.ThreadReadState{}, port.ErrForbidden
	}
	thread, err := service.threads.FindByID(ctx, command.ThreadID)
	if err != nil {
		return domain.ThreadReadState{}, err
	}
	if err := service.requireThreadView(ctx, command.ActorUserID, thread); err != nil {
		return domain.ThreadReadState{}, err
	}
	state := threadReadState(command, thread)
	if err := state.Validate(); err != nil {
		return domain.ThreadReadState{}, err
	}
	if err := service.interactions.MarkThreadRead(ctx, state); err != nil {
		return domain.ThreadReadState{}, err
	}
	return state, service.publishReadEvent(
		ctx,
		"forums.thread.read",
		command.ActorUserID,
		thread.ID,
		map[string]any{
			"thread_id":               thread.ID,
			"forum_id":                thread.ForumID,
			"last_read_post_sequence": state.LastReadPostSequence,
		},
		[]eventdomain.Scope{
			{Type: eventdomain.ScopeUser, ID: command.ActorUserID.String()},
			{Type: eventdomain.ScopeThread, ID: thread.ID.String()},
		},
	)
}

// MarkForumRead stores read state for visible threads in one forum.
func (service Service) MarkForumRead(
	ctx context.Context,
	command port.MarkForumReadCommand,
) error {
	if command.ActorUserID == uuid.Nil {
		return port.ErrForbidden
	}
	forumIDs, err := service.visibleForumIDs(ctx, command.ActorUserID, command.ForumID)
	if err != nil {
		return err
	}
	if len(forumIDs) == 0 {
		return port.ErrForbidden
	}
	if err := service.interactions.MarkForumRead(ctx, command.ActorUserID, forumIDs[0], time.Now().UTC()); err != nil {
		return err
	}
	return service.publishReadEvent(
		ctx,
		"forums.forum.read",
		command.ActorUserID,
		forumIDs[0],
		map[string]any{"forum_id": forumIDs[0]},
		[]eventdomain.Scope{
			{Type: eventdomain.ScopeUser, ID: command.ActorUserID.String()},
			{Type: eventdomain.ScopeForum, ID: forumIDs[0].String()},
		},
	)
}

// GetUnreadSummary returns unread totals for visible forums.
func (service Service) GetUnreadSummary(
	ctx context.Context,
	actorUserID uuid.UUID,
) (domain.UnreadSummary, error) {
	if actorUserID == uuid.Nil {
		return domain.UnreadSummary{}, port.ErrForbidden
	}
	forumIDs, err := service.visibleForumIDs(ctx, actorUserID, uuid.Nil)
	if err != nil {
		return domain.UnreadSummary{}, err
	}
	return service.interactions.UnreadSummary(ctx, actorUserID, forumIDs)
}
