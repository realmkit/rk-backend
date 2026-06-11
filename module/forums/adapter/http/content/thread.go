package content

import (
	"github.com/gofiber/fiber/v2"
	"github.com/niflaot/gamehub-go/module/forums/adapter/http/shared"
	"github.com/niflaot/gamehub-go/module/forums/domain"
	"github.com/niflaot/gamehub-go/module/forums/port"
)

// createThread creates a thread.
func (handler handler) createThread(ctx *fiber.Ctx) error {
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
	var request threadCreateRequest
	if err := shared.DecodeJSON(ctx, &request); err != nil {
		return err
	}
	thread, post, err := handler.services.Content.CreateThread(ctx.UserContext(), port.CreateThreadCommand{
		ActorUserID:         actor,
		ForumID:             forumID,
		Title:               request.Title,
		Slug:                request.Slug,
		ContentDocumentJSON: request.ContentDocumentJSON,
		ContentText:         request.ContentText,
		ContentChecksum:     request.ContentChecksum,
	})
	if err != nil {
		return shared.HandleError(ctx, err)
	}
	shared.SetETag(ctx, thread.Version)
	return shared.WriteJSON(ctx, fiber.StatusCreated, threadCreateResponse{
		Thread: thread,
		Post:   post,
	})
}

// listThreads lists forum threads.
func (handler handler) listThreads(ctx *fiber.Ctx) error {
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
	filter := port.ThreadFilter{
		ForumID: forumID,
		Status:  domain.ThreadStatus(ctx.Query("status")),
		Section: ctx.Query("section"),
	}
	result, err := handler.services.Content.ListThreads(ctx.UserContext(), actor, filter, page)
	if err != nil {
		return shared.HandleError(ctx, err)
	}
	return shared.WriteJSON(ctx, fiber.StatusOK, threadListResponse{
		Items:         result.Items,
		NextPageToken: result.NextCursor,
	})
}

// getThread returns one thread.
func (handler handler) getThread(ctx *fiber.Ctx) error {
	actor, err := shared.OptionalUserID(ctx)
	if err != nil {
		return err
	}
	id, err := shared.IDFromParam(ctx, "thread_id")
	if err != nil {
		return err
	}
	thread, err := handler.services.Content.GetThread(ctx.UserContext(), actor, id)
	if err != nil {
		return shared.HandleError(ctx, err)
	}
	shared.SetETag(ctx, thread.Version)
	return shared.WriteJSON(ctx, fiber.StatusOK, thread)
}

// updateThread updates thread title.
func (handler handler) updateThread(ctx *fiber.Ctx) error {
	if err := shared.RequireIdempotency(ctx); err != nil {
		return err
	}
	actor, err := shared.CurrentUserID(ctx)
	if err != nil {
		return err
	}
	id, err := shared.IDFromParam(ctx, "thread_id")
	if err != nil {
		return err
	}
	version, err := shared.ExpectedVersion(ctx)
	if err != nil {
		return err
	}
	var request threadUpdateRequest
	if err := shared.DecodeJSON(ctx, &request); err != nil {
		return err
	}
	command := port.UpdateThreadTitleCommand{
		ActorUserID:     actor,
		ThreadID:        id,
		Title:           request.Title,
		Slug:            request.Slug,
		ExpectedVersion: version,
	}
	thread, err := handler.services.Content.UpdateThreadTitle(ctx.UserContext(), command)
	if err != nil {
		return shared.HandleError(ctx, err)
	}
	shared.SetETag(ctx, thread.Version)
	return shared.WriteJSON(ctx, fiber.StatusOK, thread)
}

// deleteThread deletes one thread.
func (handler handler) deleteThread(ctx *fiber.Ctx) error {
	if err := shared.RequireIdempotency(ctx); err != nil {
		return err
	}
	actor, err := shared.CurrentUserID(ctx)
	if err != nil {
		return err
	}
	id, err := shared.IDFromParam(ctx, "thread_id")
	if err != nil {
		return err
	}
	version, err := shared.ExpectedVersion(ctx)
	if err != nil {
		return err
	}
	command := port.DeleteThreadCommand{
		ActorUserID:     actor,
		ThreadID:        id,
		ExpectedVersion: version,
	}
	if err := handler.services.Content.DeleteThread(ctx.UserContext(), command); err != nil {
		return shared.HandleError(ctx, err)
	}
	return shared.WriteNoContent(ctx)
}
