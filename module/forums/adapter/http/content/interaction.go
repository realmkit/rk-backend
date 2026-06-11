package content

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/module/forums/adapter/http/shared"
	"github.com/realmkit/rk-backend/module/forums/port"
)

// likePost likes one post.
func (handler handler) likePost(ctx *fiber.Ctx) error {
	if err := shared.RequireIdempotency(ctx); err != nil {
		return err
	}
	actor, err := shared.CurrentUserID(ctx)
	if err != nil {
		return err
	}
	postID, err := shared.IDFromParam(ctx, "post_id")
	if err != nil {
		return err
	}
	command := port.LikePostCommand{
		ActorUserID: actor,
		PostID:      postID,
	}
	summary, err := handler.services.Interaction.LikePost(ctx.UserContext(), command)
	if err != nil {
		return shared.HandleError(ctx, err)
	}
	return shared.WriteJSON(ctx, fiber.StatusOK, summary)
}

// unlikePost unlikes one post.
func (handler handler) unlikePost(ctx *fiber.Ctx) error {
	if err := shared.RequireIdempotency(ctx); err != nil {
		return err
	}
	actor, err := shared.CurrentUserID(ctx)
	if err != nil {
		return err
	}
	postID, err := shared.IDFromParam(ctx, "post_id")
	if err != nil {
		return err
	}
	command := port.UnlikePostCommand{
		ActorUserID: actor,
		PostID:      postID,
	}
	summary, err := handler.services.Interaction.UnlikePost(ctx.UserContext(), command)
	if err != nil {
		return shared.HandleError(ctx, err)
	}
	return shared.WriteJSON(ctx, fiber.StatusOK, summary)
}

// listLatestPosts lists global latest-post widget rows.
func (handler handler) listLatestPosts(ctx *fiber.Ctx) error {
	return handler.latestPosts(ctx, uuid.Nil)
}

// listForumLatestPosts lists forum latest-post widget rows.
func (handler handler) listForumLatestPosts(ctx *fiber.Ctx) error {
	forumID, err := shared.IDFromParam(ctx, "forum_id")
	if err != nil {
		return err
	}
	return handler.latestPosts(ctx, forumID)
}

// latestPosts lists latest-post widget rows.
func (handler handler) latestPosts(ctx *fiber.Ctx, forumID uuid.UUID) error {
	actor, err := shared.OptionalUserID(ctx)
	if err != nil {
		return err
	}
	page, err := shared.PageFromQuery(ctx)
	if err != nil {
		return err
	}
	result, err := handler.services.Interaction.ListLatestPosts(ctx.UserContext(), actor, forumID, page)
	if err != nil {
		return shared.HandleError(ctx, err)
	}
	return shared.WriteJSON(ctx, fiber.StatusOK, latestPostListResponse{
		Items:         result.Items,
		NextPageToken: result.NextCursor,
	})
}

// listMostLikedPosts lists most-liked posts for one forum.
func (handler handler) listMostLikedPosts(ctx *fiber.Ctx) error {
	actor, err := shared.OptionalUserID(ctx)
	if err != nil {
		return err
	}
	forumID, err := shared.IDFromParam(ctx, "forum_id")
	if err != nil {
		return err
	}
	page, err := shared.PageFromQuery(ctx)
	if err != nil {
		return err
	}
	result, err := handler.services.Interaction.ListMostLikedPosts(ctx.UserContext(), actor, forumID, page)
	if err != nil {
		return shared.HandleError(ctx, err)
	}
	return shared.WriteJSON(ctx, fiber.StatusOK, mostLikedPostListResponse{
		Items:         result.Items,
		NextPageToken: result.NextCursor,
	})
}

// markThreadRead marks one thread read.
func (handler handler) markThreadRead(ctx *fiber.Ctx) error {
	if err := shared.RequireIdempotency(ctx); err != nil {
		return err
	}
	actor, err := shared.CurrentUserID(ctx)
	if err != nil {
		return err
	}
	threadID, err := shared.IDFromParam(ctx, "thread_id")
	if err != nil {
		return err
	}
	var request readThreadRequest
	if len(ctx.Body()) > 0 {
		if err := shared.DecodeJSON(ctx, &request); err != nil {
			return err
		}
	}
	state, err := handler.services.Interaction.MarkThreadRead(ctx.UserContext(), port.MarkThreadReadCommand{
		ActorUserID:          actor,
		ThreadID:             threadID,
		LastReadPostSequence: request.LastReadPostSequence,
	})
	if err != nil {
		return shared.HandleError(ctx, err)
	}
	return shared.WriteJSON(ctx, fiber.StatusOK, state)
}

// markForumRead marks visible forum threads read.
func (handler handler) markForumRead(ctx *fiber.Ctx) error {
	if err := shared.RequireIdempotency(ctx); err != nil {
		return err
	}
	actor, err := shared.CurrentUserID(ctx)
	if err != nil {
		return err
	}
	forumID, err := shared.IDFromParam(ctx, "forum_id")
	if err != nil {
		return err
	}
	command := port.MarkForumReadCommand{
		ActorUserID: actor,
		ForumID:     forumID,
	}
	if err := handler.services.Interaction.MarkForumRead(ctx.UserContext(), command); err != nil {
		return shared.HandleError(ctx, err)
	}
	return shared.WriteNoContent(ctx)
}

// unreadSummary returns unread forum counts.
func (handler handler) unreadSummary(ctx *fiber.Ctx) error {
	actor, err := shared.CurrentUserID(ctx)
	if err != nil {
		return err
	}
	summary, err := handler.services.Interaction.GetUnreadSummary(ctx.UserContext(), actor)
	if err != nil {
		return shared.HandleError(ctx, err)
	}
	return shared.WriteJSON(ctx, fiber.StatusOK, summary)
}
