// Package http exposes ticket routes.
package http

import (
	"github.com/gofiber/fiber/v2"
	"github.com/realmkit/rk-backend/module/tickets/port"
)

// Services contains ticket route dependencies.
type Services struct {
	Definitions  port.DefinitionService
	Tickets      port.TicketService
	Conversation port.ConversationService
	Operations   port.OperationsService
}

// Register registers ticket routes.
func Register(router fiber.Router, services Services) {
	handler := handler{services: services}
	router.Post("/ticket-definitions", handler.createDefinition)
	router.Get("/ticket-definitions", handler.listDefinitions)
	router.Get("/ticket-definitions/:definition_id", handler.getDefinition)
	router.Patch("/ticket-definitions/:definition_id", handler.updateDefinition)
	router.Delete("/ticket-definitions/:definition_id", handler.deleteDefinition)
	router.Post("/tickets", handler.createTicket)
	router.Get("/tickets", handler.listTickets)
	router.Post("/punishments/:punishment_id/appeals", handler.createAppeal)
	router.Get("/tickets/:ticket_id", handler.getTicket)
	router.Get("/tickets/:ticket_id/messages", handler.listMessages)
	router.Post("/tickets/:ticket_id/messages", handler.createMessage)
	router.Get("/tickets/:ticket_id/evidence", handler.listEvidence)
	router.Post("/tickets/:ticket_id/evidence", handler.addEvidence)
	router.Post("/tickets/:ticket_id/assign", handler.assignTicket)
	router.Post("/tickets/:ticket_id/escalate", handler.escalateTicket)
	router.Post("/tickets/:ticket_id/close", handler.closeTicket)
	router.Post("/tickets/:ticket_id/reopen", handler.reopenTicket)
	router.Post("/tickets/:ticket_id/appeal/accept", handler.acceptAppeal)
	router.Post("/tickets/:ticket_id/appeal/reject", handler.rejectAppeal)
	router.Post("/tickets/operations/stats/verify", handler.verifyStats)
	router.Post("/tickets/operations/stats/rebuild", handler.rebuildStats)
}

// handler contains ticket route dependencies.
type handler struct {
	services Services
}
