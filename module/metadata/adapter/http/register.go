// Package http contains Fiber handlers for the metadata module.
package http

import (
	"github.com/gofiber/fiber/v2"
	"github.com/niflaot/gamehub-go/module/metadata/port"
)

// Services contains metadata application services used by handlers.
type Services struct {
	// Definitions manages metafield definitions.
	Definitions port.DefinitionService

	// Values manages owner metafield values.
	Values port.ValueService

	// Metaobjects manages metaobject definitions and entries.
	Metaobjects port.MetaobjectService
}

// Register registers metadata routes on router.
func Register(router fiber.Router, services Services) {
	handler := handler{services: services}
	group := router.Group("/metadata")
	group.Post("/metafield-definitions", handler.createDefinition)
	group.Get("/metafield-definitions", handler.listDefinitions)
	group.Get("/metafield-definitions/:definition_id", handler.getDefinition)
	group.Patch("/metafield-definitions/:definition_id", handler.updateDefinition)
	group.Delete("/metafield-definitions/:definition_id", handler.archiveDefinition)
	group.Put("/owners/:owner_type/:owner_id/metafields/:namespace/:key", handler.setValue)
	group.Get("/owners/:owner_type/:owner_id/metafields", handler.listValues)
	group.Get("/owners/:owner_type/:owner_id/metafields/:namespace/:key", handler.getValue)
	group.Delete("/owners/:owner_type/:owner_id/metafields/:namespace/:key", handler.deleteValue)
	group.Post("/metaobject-definitions", handler.createMetaobjectDefinition)
	group.Get("/metaobject-definitions", handler.listMetaobjectDefinitions)
	group.Get("/metaobject-definitions/:definition_id", handler.getMetaobjectDefinition)
	group.Patch("/metaobject-definitions/:definition_id", handler.updateMetaobjectDefinition)
	group.Delete("/metaobject-definitions/:definition_id", handler.archiveMetaobjectDefinition)
	group.Post("/metaobject-definitions/:definition_id/entries", handler.createMetaobjectEntry)
	group.Get("/metaobject-definitions/:definition_id/entries", handler.listMetaobjectEntries)
	group.Get("/metaobject-definitions/:definition_id/entries/:entry_id", handler.getMetaobjectEntry)
	group.Patch("/metaobject-definitions/:definition_id/entries/:entry_id", handler.updateMetaobjectEntry)
	group.Delete("/metaobject-definitions/:definition_id/entries/:entry_id", handler.deleteMetaobjectEntry)
}

// handler contains metadata route dependencies.
type handler struct {
	services Services
}
