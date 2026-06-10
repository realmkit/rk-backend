package content

import (
	"github.com/gofiber/fiber/v2"
	"github.com/niflaot/gamehub-go/module/forums/adapter/http/shared"
	"github.com/niflaot/gamehub-go/module/forums/port"
)

// createReply creates a reply.
func (handler handler) createReply(ctx *fiber.Ctx) error {
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
	var request contentRequest
	if err := shared.DecodeJSON(ctx, &request); err != nil {
		return err
	}
	command := port.CreateReplyCommand{
		ActorUserID:         actor,
		ThreadID:            threadID,
		ContentDocumentJSON: request.ContentDocumentJSON,
		ContentText:         request.ContentText,
		ContentChecksum:     request.ContentChecksum,
		References:          request.References,
	}
	post, err := handler.services.Content.CreateReply(ctx.Context(), command)
	if err != nil {
		return shared.HandleError(ctx, err)
	}
	shared.SetETag(ctx, post.Version)
	return shared.WriteJSON(ctx, fiber.StatusCreated, post)
}

// listPosts lists thread posts.
func (handler handler) listPosts(ctx *fiber.Ctx) error {
	actor, err := shared.OptionalUserID(ctx)
	if err != nil {
		return err
	}
	threadID, err := shared.IDFromParam(ctx, "thread_id")
	if err != nil {
		return err
	}
	page, err := shared.PageFromQuery(ctx)
	if err != nil {
		return err
	}
	filter := port.PostFilter{
		ThreadID:      threadID,
		IncludeHidden: ctx.QueryBool("include_hidden"),
	}
	result, err := handler.services.Content.ListPosts(ctx.Context(), actor, filter, page)
	if err != nil {
		return shared.HandleError(ctx, err)
	}
	return shared.WriteJSON(ctx, fiber.StatusOK, postListResponse{
		Items:         result.Items,
		NextPageToken: result.NextCursor,
	})
}

// getPost returns one post.
func (handler handler) getPost(ctx *fiber.Ctx) error {
	actor, err := shared.OptionalUserID(ctx)
	if err != nil {
		return err
	}
	id, err := shared.IDFromParam(ctx, "post_id")
	if err != nil {
		return err
	}
	post, err := handler.services.Content.GetPost(ctx.Context(), actor, id)
	if err != nil {
		return shared.HandleError(ctx, err)
	}
	shared.SetETag(ctx, post.Version)
	return shared.WriteJSON(ctx, fiber.StatusOK, post)
}

// updatePost updates one post.
func (handler handler) updatePost(ctx *fiber.Ctx) error {
	if err := shared.RequireIdempotency(ctx); err != nil {
		return err
	}
	actor, err := shared.CurrentUserID(ctx)
	if err != nil {
		return err
	}
	id, err := shared.IDFromParam(ctx, "post_id")
	if err != nil {
		return err
	}
	version, err := shared.ExpectedVersion(ctx)
	if err != nil {
		return err
	}
	var request contentRequest
	if err := shared.DecodeJSON(ctx, &request); err != nil {
		return err
	}
	command := port.UpdatePostCommand{
		ActorUserID:         actor,
		PostID:              id,
		ContentDocumentJSON: request.ContentDocumentJSON,
		ContentText:         request.ContentText,
		ContentChecksum:     request.ContentChecksum,
		EditReason:          request.EditReason,
		ExpectedVersion:     version,
	}
	post, err := handler.services.Content.UpdatePost(ctx.Context(), command)
	if err != nil {
		return shared.HandleError(ctx, err)
	}
	shared.SetETag(ctx, post.Version)
	return shared.WriteJSON(ctx, fiber.StatusOK, post)
}

// deletePost deletes one post.
func (handler handler) deletePost(ctx *fiber.Ctx) error {
	if err := shared.RequireIdempotency(ctx); err != nil {
		return err
	}
	actor, err := shared.CurrentUserID(ctx)
	if err != nil {
		return err
	}
	id, err := shared.IDFromParam(ctx, "post_id")
	if err != nil {
		return err
	}
	version, err := shared.ExpectedVersion(ctx)
	if err != nil {
		return err
	}
	command := port.DeletePostCommand{
		ActorUserID:     actor,
		PostID:          id,
		ExpectedVersion: version,
	}
	if err := handler.services.Content.DeletePost(ctx.Context(), command); err != nil {
		return shared.HandleError(ctx, err)
	}
	return shared.WriteNoContent(ctx)
}

// listPostRevisions lists post revisions.
func (handler handler) listPostRevisions(ctx *fiber.Ctx) error {
	actor, err := shared.CurrentUserID(ctx)
	if err != nil {
		return err
	}
	postID, err := shared.IDFromParam(ctx, "post_id")
	if err != nil {
		return err
	}
	page, err := shared.PageFromQuery(ctx)
	if err != nil {
		return err
	}
	result, err := handler.services.Content.ListPostRevisions(ctx.Context(), actor, postID, page)
	if err != nil {
		return shared.HandleError(ctx, err)
	}
	return shared.WriteJSON(ctx, fiber.StatusOK, revisionListResponse{
		Items:         result.Items,
		NextPageToken: result.NextCursor,
	})
}
