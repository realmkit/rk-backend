package http

import (
	"github.com/gofiber/fiber/v2"
	"github.com/niflaot/gamehub-go/module/forums/port"
)

// Services contains forum application services used by handlers.
type Services struct {
	// Forums manages forum structure.
	Forums port.Service
}

// Register registers forum routes on router.
func Register(router fiber.Router, services Services) {
	handler := handler{services: services}
	router.Get("/forums/tree", handler.tree)
	categories := router.Group("/forum-categories")
	categories.Post("", handler.createCategory)
	categories.Get("", handler.listCategories)
	categories.Get("/:category_id", handler.getCategory)
	categories.Patch("/:category_id", handler.updateCategory)
	categories.Delete("/:category_id", handler.deleteCategory)
	categories.Post("/reorder", handler.reorderCategories)
	forums := router.Group("/forums")
	forums.Post("", handler.createForum)
	forums.Get("", handler.listForums)
	forums.Get("/latest-posts", handler.listLatestPosts)
	forums.Get("/search", handler.searchForums)
	forums.Get("/unread-summary", handler.unreadSummary)
	forums.Get("/:forum_id/threads", handler.listThreads)
	forums.Post("/:forum_id/threads", handler.createThread)
	forums.Get("/:forum_id/latest-posts", handler.listForumLatestPosts)
	forums.Get("/:forum_id/posts/most-liked", handler.listMostLikedPosts)
	forums.Get("/:forum_id/search", handler.searchForum)
	forums.Get("/:forum_id/settings", handler.getForumSettings)
	forums.Patch("/:forum_id/settings", handler.updateForumSettings)
	forums.Get("/:forum_id/permissions", handler.getForumPermissions)
	forums.Put("/:forum_id/permissions", handler.updateForumPermissions)
	forums.Post("/:forum_id/permissions/simulate", handler.simulateForumPermission)
	forums.Get("/:forum_id", handler.getForum)
	forums.Patch("/:forum_id", handler.updateForum)
	forums.Delete("/:forum_id", handler.deleteForum)
	forums.Post("/:forum_id/move", handler.moveForum)
	forums.Post("/:forum_id/read", handler.markForumRead)
	forums.Post("/reorder", handler.reorderForums)
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

// handler contains forum route dependencies.
type handler struct {
	services Services
}
