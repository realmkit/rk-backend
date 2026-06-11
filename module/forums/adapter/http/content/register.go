// Package content adapts forum content and interaction use cases to Fiber routes.
package content

import (
	"github.com/gofiber/fiber/v2"
	"github.com/realmkit/rk-backend/module/forums/port"
)

// Services contains content route dependencies.
type Services struct {
	// Content manages threads, posts, and revisions.
	Content port.ContentService

	// Interaction manages likes, widgets, and read state.
	Interaction port.InteractionService

	// Operations manages search, cache, and repair operations.
	Operations port.OperationsService
}

// Register registers content, interaction, and search routes.
func Register(router fiber.Router, services Services) {
	handler := handler{services: services}
	forums := router.Group("/forums")
	forums.Get("/latest-posts", handler.listLatestPosts)
	forums.Get("/search", handler.searchForums)
	forums.Get("/unread-summary", handler.unreadSummary)
	forums.Get("/:forum_id/threads", handler.listThreads)
	forums.Post("/:forum_id/threads", handler.createThread)
	forums.Get("/:forum_id/latest-posts", handler.listForumLatestPosts)
	forums.Get("/:forum_id/posts/most-liked", handler.listMostLikedPosts)
	forums.Get("/:forum_id/search", handler.searchForum)
	forums.Post("/:forum_id/read", handler.markForumRead)

	threads := router.Group("/threads")
	threads.Get("/:thread_id", handler.getThread)
	threads.Patch("/:thread_id", handler.updateThread)
	threads.Delete("/:thread_id", handler.deleteThread)
	threads.Get("/:thread_id/posts", handler.listPosts)
	threads.Post("/:thread_id/posts", handler.createReply)
	threads.Post("/:thread_id/read", handler.markThreadRead)

	posts := router.Group("/posts")
	posts.Get("/:post_id", handler.getPost)
	posts.Patch("/:post_id", handler.updatePost)
	posts.Delete("/:post_id", handler.deletePost)
	posts.Put("/:post_id/like", handler.likePost)
	posts.Delete("/:post_id/like", handler.unlikePost)
	posts.Get("/:post_id/revisions", handler.listPostRevisions)
}

// handler contains content route dependencies.
type handler struct {
	services Services
}
