package tickets_e2e

import (
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/niflaot/gamehub-go/e2e/harness"
)

// TestTicketDefinitionLifecycle verifies ticket definition administration.
func TestTicketDefinitionLifecycle(t *testing.T) {
	steps := harness.NewSteps(t)
	fixture := newTicketsFixture(t)
	actor := uuid.New()

	steps.Do("anonymous definition creation is rejected", func() {
		response := fixture.do(t, harness.JSONRequest(
			fiber.MethodPost,
			"/ticket-definitions",
			`{"key":"anon","name":"Anon","kind":"support","status":"active"}`,
		))
		assertTicketStatus(t, response, fiber.StatusUnauthorized)
	})

	steps.Do("definition can be created without core color", func() {
		created := fixture.createTicketDefinition(t, actor, "support_help", "support")
		id := ticketIDFrom(t, created, "id")
		if _, ok := created["color"]; ok {
			t.Fatalf("ticket definition should not expose core color: %+v", created)
		}

		get := fixture.do(t, configureTicketRequest(
			harness.JSONRequest(fiber.MethodGet, "/ticket-definitions/"+id.String(), ""),
			withTicketUser(actor),
		))
		assertTicketStatus(t, get, fiber.StatusOK)

		list := fixture.do(t, configureTicketRequest(
			harness.JSONRequest(fiber.MethodGet, "/ticket-definitions?kind=support&status=active", ""),
			withTicketUser(actor),
		))
		assertTicketStatus(t, list, fiber.StatusOK)

		missingMatch := fixture.do(t, configureTicketRequest(
			harness.JSONRequest(fiber.MethodPatch, "/ticket-definitions/"+id.String(), definitionUpdateBody("support_help")),
			withTicketUser(actor),
		))
		assertTicketStatus(t, missingMatch, fiber.StatusPreconditionRequired)

		updated := fixture.do(t, configureTicketRequest(
			harness.JSONRequest(fiber.MethodPatch, "/ticket-definitions/"+id.String(), definitionUpdateBody("support_help")),
			withTicketUser(actor),
			withTicketIfMatch(ticketVersionFrom(created)),
		))
		assertTicketStatus(t, updated, fiber.StatusOK)

		stale := fixture.do(t, configureTicketRequest(
			harness.JSONRequest(fiber.MethodPatch, "/ticket-definitions/"+id.String(), definitionUpdateBody("support_help")),
			withTicketUser(actor),
			withTicketIfMatch(ticketVersionFrom(created)),
		))
		assertTicketStatus(t, stale, fiber.StatusPreconditionFailed)

		current := decodeTicketObject(t, updated)
		deleted := fixture.do(t, configureTicketRequest(
			harness.JSONRequest(fiber.MethodDelete, "/ticket-definitions/"+id.String(), ""),
			withTicketUser(actor),
			withTicketIfMatch(ticketVersionFrom(current)),
		))
		assertTicketStatus(t, deleted, fiber.StatusNoContent)
	})

	steps.Do("appeal definitions must require punishments", func() {
		response := fixture.do(t, configureTicketRequest(
			harness.JSONRequest(
				fiber.MethodPost,
				"/ticket-definitions",
				`{"key":"bad_appeal","name":"Bad","kind":"appeal","status":"active"}`,
			),
			withTicketUser(actor),
		))
		assertTicketStatus(t, response, fiber.StatusUnprocessableEntity)
	})
}

// definitionUpdateBody returns a valid update body.
func definitionUpdateBody(key string) string {
	return `{"key":"` + key + `","name":"Updated ` + key + `","kind":"support","status":"active"}`
}
