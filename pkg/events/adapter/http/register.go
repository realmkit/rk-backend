package http

import (
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	groupsport "github.com/realmkit/rk-backend/module/groups/port"
	"github.com/realmkit/rk-backend/pkg/events/application"
	"github.com/realmkit/rk-backend/pkg/events/port"
)

// Services contains event HTTP dependencies.
type Services struct {
	// Events manages durable events.
	Events application.Service

	// Hub manages local WebSocket clients.
	Hub *Hub

	// ScopeAuthorizer checks non-public websocket subscriptions.
	ScopeAuthorizer port.ScopeAuthorizer

	// Checker checks group-backed permissions.
	Checker groupsport.Checker
}

// Register registers event routes.
func Register(router fiber.Router, services Services) {
	handler := handler{services: services}
	group := router.Group("/events")
	group.Get("", handler.listEvents)
	group.Get("/ws", websocket.New(handler.webSocket))
	group.Get("/:event_id", handler.getEvent)
	group.Post("/:event_id/replay", handler.replayEvent)
	group.Post("/:event_id/cancel", handler.cancelEvent)
}

// handler contains event route dependencies.
type handler struct {
	services Services
}
