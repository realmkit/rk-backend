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
	forums.Get("/:forum_id", handler.getForum)
	forums.Patch("/:forum_id", handler.updateForum)
	forums.Delete("/:forum_id", handler.deleteForum)
	forums.Post("/:forum_id/move", handler.moveForum)
	forums.Post("/reorder", handler.reorderForums)
}

// handler contains forum route dependencies.
type handler struct {
	services Services
}
