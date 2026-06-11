package http

import (
	"github.com/gofiber/fiber/v2"
	contentroutes "github.com/realmkit/rk-backend/module/forums/adapter/http/content"
	structureroutes "github.com/realmkit/rk-backend/module/forums/adapter/http/structure"
	"github.com/realmkit/rk-backend/module/forums/port"
)

// Services contains forum application services used by handlers.
type Services struct {
	// Structure manages categories, forums, and trees.
	Structure port.StructureService

	// Content manages threads, posts, and revisions.
	Content port.ContentService

	// Interaction manages likes, widgets, and read state.
	Interaction port.InteractionService

	// Operations manages search, cache, and repair operations.
	Operations port.OperationsService

	// Admin manages forum settings and permission configuration.
	Admin port.AdminService
}

// Register registers forum routes on router.
func Register(router fiber.Router, services Services) {
	contentroutes.Register(router, contentroutes.Services{
		Content:     services.Content,
		Interaction: services.Interaction,
		Operations:  services.Operations,
	})
	structureroutes.Register(router, structureroutes.Services{
		Structure: services.Structure,
		Admin:     services.Admin,
	})
}
