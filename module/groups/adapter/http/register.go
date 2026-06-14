package http

import (
	"github.com/gofiber/fiber/v2"
	"github.com/realmkit/rk-backend/module/groups/port"
	userport "github.com/realmkit/rk-backend/module/user/port"
)

// Services contains groups application services used by handlers.
type Services struct {
	// Groups manages groups.
	Groups port.GroupService

	// Memberships manages group memberships.
	Memberships port.MembershipService

	// Grants manages permission grants.
	Grants port.PermissionGrantService

	// Checker checks permissions.
	Checker port.Checker

	// Users resolves local user summaries for membership displays.
	Users userport.Service
}

// Register registers groups and permissions routes on router.
func Register(router fiber.Router, services Services) {
	handler := handler{services: services}
	groups := router.Group("/groups")
	groups.Post("", handler.createGroup)
	groups.Get("", handler.listGroups)
	groups.Get("/:group_id", handler.getGroup)
	groups.Patch("/:group_id", handler.updateGroup)
	groups.Delete("/:group_id", handler.deleteGroup)
	groups.Get("/:group_id/members", handler.listGroupMembers)
	groups.Put("/:group_id/members/:user_id", handler.assignMembership)
	groups.Delete("/:group_id/members/:user_id", handler.removeMembership)
	router.Get("/users/me/groups", handler.listCurrentUserGroups)
	router.Get("/users/:user_id/groups", handler.listUserGroups)
	router.Post("/permissions/check", handler.checkPermission)
}

// handler contains groups route dependencies.
type handler struct {
	services Services
}
