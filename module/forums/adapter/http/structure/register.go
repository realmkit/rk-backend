package structure

import (
	"github.com/gofiber/fiber/v2"
	"github.com/niflaot/gamehub-go/module/forums/port"
)

// Services contains structure and admin application services.
type Services struct {
	// Structure manages categories, forums, and trees.
	Structure port.StructureService

	// Admin manages forum settings and permission configuration.
	Admin port.AdminService
}

// Register registers forum structure and admin routes.
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
	forums.Get("/:forum_id/settings", handler.getForumSettings)
	forums.Patch("/:forum_id/settings", handler.updateForumSettings)
	forums.Get("/:forum_id/permissions", handler.getForumPermissions)
	forums.Put("/:forum_id/permissions", handler.updateForumPermissions)
	forums.Post("/:forum_id/permissions/simulate", handler.simulateForumPermission)
	forums.Get("/:forum_id", handler.getForum)
	forums.Patch("/:forum_id", handler.updateForum)
	forums.Delete("/:forum_id", handler.deleteForum)
	forums.Post("/:forum_id/move", handler.moveForum)
	forums.Post("/reorder", handler.reorderForums)
}

// handler contains route dependencies.
type handler struct {
	services Services
}
