package http

import (
	"github.com/gofiber/fiber/v2"
	"github.com/realmkit/rk-backend/module/assets/port"
)

// Services contains assets application services used by handlers.
type Services struct {
	// Assets manages assets.
	Assets port.Service
}

// Register registers assets routes on router.
func Register(router fiber.Router, services Services) {
	handler := handler{services: services}
	group := router.Group("/assets")
	group.Post("/upload-intents", handler.createUploadIntent)
	group.Get("", handler.listAssets)
	group.Get("/", handler.listAssets)
	group.Get("/folders", handler.listFolders)
	group.Get("/:asset_id", handler.getAsset)
	group.Get("/:asset_id/url", handler.getAssetURL)
	group.Post("/:asset_id/complete", handler.completeUpload)
	group.Patch("/:asset_id", handler.updateAsset)
	group.Delete("/:asset_id", handler.deleteAsset)
}

// handler contains assets route dependencies.
type handler struct {
	services Services
}
