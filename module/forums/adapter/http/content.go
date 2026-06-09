// Package http adapts forum use cases to Fiber routes.
package http

import (
	"encoding/json"

	"github.com/gofiber/fiber/v2"
	"github.com/niflaot/gamehub-go/module/forums/domain"
	"github.com/niflaot/gamehub-go/module/forums/port"
)

// contentRequest is a post content write request.
type contentRequest struct {
	ContentDocumentJSON json.RawMessage        `json:"content_document_json"`
	ContentText         string                 `json:"content_text"`
	ContentChecksum     string                 `json:"content_checksum"`
	EditReason          string                 `json:"edit_reason"`
	References          []domain.PostReference `json:"references"`
}

// threadCreateRequest is a thread creation request.
type threadCreateRequest struct {
	Title               string          `json:"title"`
	Slug                domain.Slug     `json:"slug"`
	ContentDocumentJSON json.RawMessage `json:"content_document_json"`
	ContentText         string          `json:"content_text"`
	ContentChecksum     string          `json:"content_checksum"`
}

// threadUpdateRequest is a thread title update request.
type threadUpdateRequest struct {
	Title string      `json:"title"`
	Slug  domain.Slug `json:"slug"`
}

// threadCreateResponse returns created thread and opener post.
type threadCreateResponse struct {
	Thread domain.Thread `json:"thread"`
	Post   domain.Post   `json:"post"`
}

// threadListResponse contains one thread page.
type threadListResponse struct {
	Items         []domain.Thread `json:"items"`
	NextPageToken string          `json:"next_page_token,omitempty"`
}

// postListResponse contains one post page.
type postListResponse struct {
	Items         []domain.Post `json:"items"`
	NextPageToken string        `json:"next_page_token,omitempty"`
}

// revisionListResponse contains one revision page.
type revisionListResponse struct {
	Items         []domain.PostRevision `json:"items"`
	NextPageToken string                `json:"next_page_token,omitempty"`
}

// createThread creates a thread.
func (handler handler) createThread(ctx *fiber.Ctx) error {
	if err := requireIdempotency(ctx); err != nil {
		return err
	}
	actor, err := currentUserID(ctx)
	if err != nil {
		return err
	}
	forumID, err := idFromParam(ctx, "forum_id")
	if err != nil {
		return err
	}
	var request threadCreateRequest
	if err := decodeJSON(ctx, &request); err != nil {
		return err
	}
	thread, post, err := handler.services.Forums.CreateThread(ctx.Context(), port.CreateThreadCommand{ActorUserID: actor, ForumID: forumID, Title: request.Title, Slug: request.Slug, ContentDocumentJSON: request.ContentDocumentJSON, ContentText: request.ContentText, ContentChecksum: request.ContentChecksum})
	if err != nil {
		return handleError(ctx, err)
	}
	setETag(ctx, thread.Version)
	return writeJSON(ctx, fiber.StatusCreated, threadCreateResponse{Thread: thread, Post: post})
}

// listThreads lists forum threads.
func (handler handler) listThreads(ctx *fiber.Ctx) error {
	actor, err := optionalUserID(ctx)
	if err != nil {
		return err
	}
	forumID, err := idFromParam(ctx, "forum_id")
	if err != nil {
		return err
	}
	page, err := pageFromQuery(ctx)
	if err != nil {
		return err
	}
	result, err := handler.services.Forums.ListThreads(ctx.Context(), actor, port.ThreadFilter{ForumID: forumID, Status: domain.ThreadStatus(ctx.Query("status")), Section: ctx.Query("section")}, page)
	if err != nil {
		return handleError(ctx, err)
	}
	return writeJSON(ctx, fiber.StatusOK, threadListResponse{Items: result.Items, NextPageToken: result.NextCursor})
}

// getThread returns one thread.
func (handler handler) getThread(ctx *fiber.Ctx) error {
	actor, err := optionalUserID(ctx)
	if err != nil {
		return err
	}
	id, err := idFromParam(ctx, "thread_id")
	if err != nil {
		return err
	}
	thread, err := handler.services.Forums.GetThread(ctx.Context(), actor, id)
	if err != nil {
		return handleError(ctx, err)
	}
	setETag(ctx, thread.Version)
	return writeJSON(ctx, fiber.StatusOK, thread)
}

// updateThread updates thread title.
func (handler handler) updateThread(ctx *fiber.Ctx) error {
	if err := requireIdempotency(ctx); err != nil {
		return err
	}
	actor, err := currentUserID(ctx)
	if err != nil {
		return err
	}
	id, err := idFromParam(ctx, "thread_id")
	if err != nil {
		return err
	}
	version, err := expectedVersion(ctx)
	if err != nil {
		return err
	}
	var request threadUpdateRequest
	if err := decodeJSON(ctx, &request); err != nil {
		return err
	}
	thread, err := handler.services.Forums.UpdateThreadTitle(ctx.Context(), port.UpdateThreadTitleCommand{ActorUserID: actor, ThreadID: id, Title: request.Title, Slug: request.Slug, ExpectedVersion: version})
	if err != nil {
		return handleError(ctx, err)
	}
	setETag(ctx, thread.Version)
	return writeJSON(ctx, fiber.StatusOK, thread)
}

// deleteThread deletes one thread.
func (handler handler) deleteThread(ctx *fiber.Ctx) error {
	if err := requireIdempotency(ctx); err != nil {
		return err
	}
	actor, err := currentUserID(ctx)
	if err != nil {
		return err
	}
	id, err := idFromParam(ctx, "thread_id")
	if err != nil {
		return err
	}
	version, err := expectedVersion(ctx)
	if err != nil {
		return err
	}
	if err := handler.services.Forums.DeleteThread(ctx.Context(), port.DeleteThreadCommand{ActorUserID: actor, ThreadID: id, ExpectedVersion: version}); err != nil {
		return handleError(ctx, err)
	}
	return writeNoContent(ctx)
}

// createReply creates a reply.
func (handler handler) createReply(ctx *fiber.Ctx) error {
	if err := requireIdempotency(ctx); err != nil {
		return err
	}
	actor, err := currentUserID(ctx)
	if err != nil {
		return err
	}
	threadID, err := idFromParam(ctx, "thread_id")
	if err != nil {
		return err
	}
	var request contentRequest
	if err := decodeJSON(ctx, &request); err != nil {
		return err
	}
	post, err := handler.services.Forums.CreateReply(ctx.Context(), port.CreateReplyCommand{ActorUserID: actor, ThreadID: threadID, ContentDocumentJSON: request.ContentDocumentJSON, ContentText: request.ContentText, ContentChecksum: request.ContentChecksum, References: request.References})
	if err != nil {
		return handleError(ctx, err)
	}
	setETag(ctx, post.Version)
	return writeJSON(ctx, fiber.StatusCreated, post)
}

// listPosts lists thread posts.
func (handler handler) listPosts(ctx *fiber.Ctx) error {
	actor, err := optionalUserID(ctx)
	if err != nil {
		return err
	}
	threadID, err := idFromParam(ctx, "thread_id")
	if err != nil {
		return err
	}
	page, err := pageFromQuery(ctx)
	if err != nil {
		return err
	}
	result, err := handler.services.Forums.ListPosts(ctx.Context(), actor, port.PostFilter{ThreadID: threadID, IncludeHidden: ctx.QueryBool("include_hidden")}, page)
	if err != nil {
		return handleError(ctx, err)
	}
	return writeJSON(ctx, fiber.StatusOK, postListResponse{Items: result.Items, NextPageToken: result.NextCursor})
}

// getPost returns one post.
func (handler handler) getPost(ctx *fiber.Ctx) error {
	actor, err := optionalUserID(ctx)
	if err != nil {
		return err
	}
	id, err := idFromParam(ctx, "post_id")
	if err != nil {
		return err
	}
	post, err := handler.services.Forums.GetPost(ctx.Context(), actor, id)
	if err != nil {
		return handleError(ctx, err)
	}
	setETag(ctx, post.Version)
	return writeJSON(ctx, fiber.StatusOK, post)
}

// updatePost updates one post.
func (handler handler) updatePost(ctx *fiber.Ctx) error {
	if err := requireIdempotency(ctx); err != nil {
		return err
	}
	actor, err := currentUserID(ctx)
	if err != nil {
		return err
	}
	id, err := idFromParam(ctx, "post_id")
	if err != nil {
		return err
	}
	version, err := expectedVersion(ctx)
	if err != nil {
		return err
	}
	var request contentRequest
	if err := decodeJSON(ctx, &request); err != nil {
		return err
	}
	post, err := handler.services.Forums.UpdatePost(ctx.Context(), port.UpdatePostCommand{ActorUserID: actor, PostID: id, ContentDocumentJSON: request.ContentDocumentJSON, ContentText: request.ContentText, ContentChecksum: request.ContentChecksum, EditReason: request.EditReason, ExpectedVersion: version})
	if err != nil {
		return handleError(ctx, err)
	}
	setETag(ctx, post.Version)
	return writeJSON(ctx, fiber.StatusOK, post)
}

// deletePost deletes one post.
func (handler handler) deletePost(ctx *fiber.Ctx) error {
	if err := requireIdempotency(ctx); err != nil {
		return err
	}
	actor, err := currentUserID(ctx)
	if err != nil {
		return err
	}
	id, err := idFromParam(ctx, "post_id")
	if err != nil {
		return err
	}
	version, err := expectedVersion(ctx)
	if err != nil {
		return err
	}
	if err := handler.services.Forums.DeletePost(ctx.Context(), port.DeletePostCommand{ActorUserID: actor, PostID: id, ExpectedVersion: version}); err != nil {
		return handleError(ctx, err)
	}
	return writeNoContent(ctx)
}

// listPostRevisions lists post revisions.
func (handler handler) listPostRevisions(ctx *fiber.Ctx) error {
	actor, err := currentUserID(ctx)
	if err != nil {
		return err
	}
	postID, err := idFromParam(ctx, "post_id")
	if err != nil {
		return err
	}
	page, err := pageFromQuery(ctx)
	if err != nil {
		return err
	}
	result, err := handler.services.Forums.ListPostRevisions(ctx.Context(), actor, postID, page)
	if err != nil {
		return handleError(ctx, err)
	}
	return writeJSON(ctx, fiber.StatusOK, revisionListResponse{Items: result.Items, NextPageToken: result.NextCursor})
}
