package http

import (
	"github.com/gofiber/fiber/v2"
	groupsport "github.com/realmkit/rk-backend/module/groups/port"
	userport "github.com/realmkit/rk-backend/module/user/port"
)

// Services contains user route dependencies.
type Services struct {
	// Users manages local users.
	Users userport.Service

	// Groups returns group summaries when available.
	Groups groupsport.MembershipService

	// Checker checks group-backed permissions.
	Checker groupsport.Checker
}

// Register registers user routes on router.
func Register(router fiber.Router, services Services, authenticate fiber.Handler) {
	handler := handler{services: services}
	group := router.Group("/users", authenticate)
	group.Get("", handler.listUsers)
	group.Get("/", handler.listUsers)
	group.Get("/me", handler.currentUser)
	group.Patch("/me", handler.updateCurrentUser)
	group.Get("/me/identity/account-url", handler.accountURL)
}

// handler contains user route dependencies.
type handler struct {
	services Services
}
