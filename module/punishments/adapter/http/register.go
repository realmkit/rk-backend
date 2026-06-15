// Package http exposes punishment routes.
package http

import (
	"github.com/gofiber/fiber/v2"
	groupsport "github.com/realmkit/rk-backend/module/groups/port"
	"github.com/realmkit/rk-backend/module/punishments/port"
)

// Services contains punishment HTTP dependencies.
type Services struct {
	// Punishments manages punishment definitions and cases.
	Punishments port.Service

	// Checker checks group-backed permissions.
	Checker groupsport.Checker
}

// Register registers punishment routes.
func Register(router fiber.Router, services Services) {
	handler := handler{services: services}
	router.Post("/punishment-definitions", handler.createDefinition)
	router.Get("/punishment-definitions", handler.listDefinitions)
	router.Get("/punishment-definitions/:definition_id", handler.getDefinition)
	router.Patch("/punishment-definitions/:definition_id", handler.updateDefinition)
	router.Delete("/punishment-definitions/:definition_id", handler.deleteDefinition)
	router.Post("/punishment-definitions/:definition_id/actions/reorder", handler.reorderActions)
	router.Post("/punishments", handler.issuePunishment)
	router.Get("/punishments", handler.listPunishments)
	router.Get("/punishments/:punishment_id", handler.getPunishment)
	router.Patch("/punishments/:punishment_id", handler.updatePunishment)
	router.Post("/punishments/:punishment_id/revoke", handler.revokePunishment)
	router.Get("/users/:user_id/punishments", handler.listUserPunishments)
	router.Get("/users/:user_id/punishments/active", handler.listUserPunishments)
	router.Post("/punishments/restrictions/check", handler.checkRestriction)
	router.Get("/users/:user_id/punishments/restrictions", handler.listRestrictions)
}

type handler struct {
	services Services
}
