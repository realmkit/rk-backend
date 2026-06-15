package tickets_e2e

import (
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/e2e/harness"
	"github.com/realmkit/rk-backend/module/groups/domain"
)

// TestTicketLifecycle verifies intake, idempotency, and read authorization.
func TestTicketLifecycle(t *testing.T) {
	steps := harness.NewSteps(t)
	fixture := newTicketsFixture(t)
	submitter := uuid.New()
	staff := uuid.New()
	outsider := uuid.New()
	definition := fixture.createTicketDefinition(t, staff, "lifecycle_support", "support")
	definitionID := ticketIDFrom(t, definition, "id")
	fixture.grantTicketRelation(t, definitionID, domain.RelationCreator, submitter)

	steps.Do("ticket creation requires idempotency", func() {
		response := fixture.do(t, configureTicketRequest(
			harness.JSONRequest(fiber.MethodPost, "/tickets", ticketBody(definitionID, "Missing key")),
			withTicketUser(submitter),
		))
		assertTicketStatus(t, response, fiber.StatusBadRequest)
	})

	steps.Do("ticket can be created and idempotently replayed", func() {
		first := fixture.createTicket(t, submitter, staff, definitionID, "ticket-lifecycle")
		secondResponse := fixture.do(t, configureTicketRequest(
			harness.JSONRequest(fiber.MethodPost, "/tickets", ticketBody(definitionID, "E2E Ticket")),
			withTicketUser(submitter),
			withTicketIdempotency("ticket-lifecycle"),
		))
		assertTicketStatus(t, secondResponse, fiber.StatusCreated)
		second := decodeTicketObject(t, secondResponse)
		if ticketIDFrom(t, second, "id") != ticketIDFrom(t, first, "id") {
			t.Fatalf("idempotent ticket id changed")
		}
		if first["status"] != "open" || first["message_count"].(float64) != 1 {
			t.Fatalf("ticket = %+v", first)
		}
	})

	steps.Do("ticket reads require a matching relation", func() {
		ticket := fixture.createTicket(t, submitter, staff, definitionID, "ticket-read")
		ticketID := ticketIDFrom(t, ticket, "id")

		ok := fixture.do(t, configureTicketRequest(
			harness.JSONRequest(fiber.MethodGet, "/tickets/"+ticketID.String(), ""),
			withTicketUser(submitter),
		))
		assertTicketStatus(t, ok, fiber.StatusOK)

		denied := fixture.do(t, configureTicketRequest(
			harness.JSONRequest(fiber.MethodGet, "/tickets/"+ticketID.String(), ""),
			withTicketUser(outsider),
		))
		assertTicketStatus(t, denied, fiber.StatusForbidden)

		list := fixture.do(t, configureTicketRequest(
			harness.JSONRequest(fiber.MethodGet, "/tickets?submitter_user_id="+submitter.String(), ""),
			withTicketUser(staff),
		))
		assertTicketStatus(t, list, fiber.StatusOK)
	})

	steps.Do("ticket intake rejects missing permission and required fields", func() {
		noGrantDefinition := fixture.createTicketDefinition(t, staff, "no_grant_support", "support")
		noGrantID := ticketIDFrom(t, noGrantDefinition, "id")
		noGrantUser := uuid.New()
		forbidden := fixture.do(t, configureTicketRequest(
			harness.JSONRequest(fiber.MethodPost, "/tickets", ticketBody(noGrantID, "No grant")),
			withTicketUser(noGrantUser),
			withTicketIdempotency("ticket-no-grant"),
		))
		assertTicketStatus(t, forbidden, fiber.StatusForbidden)

		reportBody := `{"key":"player_report","name":"Report","kind":"report",` +
			`"status":"active","requires_target_user":true}`
		report := fixture.do(t, configureTicketRequest(
			harness.JSONRequest(fiber.MethodPost, "/ticket-definitions", reportBody),
			withTicketUser(staff),
		))
		assertTicketStatus(t, report, fiber.StatusCreated)
		reportID := ticketIDFrom(t, decodeTicketObject(t, report), "id")
		fixture.grantTicketRelation(t, reportID, domain.RelationCreator, submitter)

		missingTarget := fixture.do(t, configureTicketRequest(
			harness.JSONRequest(fiber.MethodPost, "/tickets", ticketBody(reportID, "Missing target")),
			withTicketUser(submitter),
			withTicketIdempotency("ticket-missing-target"),
		))
		assertTicketStatus(t, missingTarget, fiber.StatusUnprocessableEntity)

		assetID := uuid.New()
		evidenceBody := ticketBody(reportID, "Missing asset")
		evidenceBody = evidenceBody[:len(evidenceBody)-1] + `,"target_user_id":"` +
			uuid.NewString() + `","evidence_asset_ids":["` + assetID.String() + `"]}`
		missingAsset := fixture.do(t, configureTicketRequest(
			harness.JSONRequest(fiber.MethodPost, "/tickets", evidenceBody),
			withTicketUser(submitter),
			withTicketIdempotency("ticket-missing-asset"),
		))
		assertTicketStatus(t, missingAsset, fiber.StatusNotFound)
	})
}
