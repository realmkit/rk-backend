package tickets_e2e

import (
	"context"
	"strconv"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/niflaot/gamehub-go/e2e/harness"
	"github.com/niflaot/gamehub-go/module/groups/domain"
	punishmentsdomain "github.com/niflaot/gamehub-go/module/punishments/domain"
	eventdomain "github.com/niflaot/gamehub-go/pkg/events/domain"
)

// TestAppealsAndStaffWorkflows verifies moderation ticket actions.
func TestAppealsAndStaffWorkflows(t *testing.T) {
	steps := harness.NewSteps(t)
	fixture := newTicketsFixture(t)
	submitter := uuid.New()
	staff := uuid.New()

	steps.Do("staff can assign, escalate, close, and reopen tickets", func() {
		outsider := uuid.New()
		definition := fixture.createTicketDefinition(t, staff, "workflow_support", "support")
		definitionID := ticketIDFrom(t, definition, "id")
		fixture.grantTicketRelation(t, definitionID, domain.RelationCreator, submitter)
		ticket := fixture.createTicket(t, submitter, staff, definitionID, "workflow-ticket")
		ticketID := ticketIDFrom(t, ticket, "id")

		missingMatch := fixture.do(t, configureTicketRequest(
			harness.JSONRequest(fiber.MethodPost, "/tickets/"+ticketID.String()+"/assign", `{"reason":"missing"}`),
			withTicketUser(staff),
			withTicketIdempotency("workflow-missing-match"),
		))
		assertTicketStatus(t, missingMatch, fiber.StatusPreconditionRequired)

		denied := fixture.do(t, configureTicketRequest(
			harness.JSONRequest(fiber.MethodPost, "/tickets/"+ticketID.String()+"/assign", `{"reason":"denied"}`),
			withTicketUser(outsider),
			withTicketIdempotency("workflow-outsider"),
			withTicketIfMatch(ticketVersionFrom(ticket)),
		))
		assertTicketStatus(t, denied, fiber.StatusForbidden)

		assignBody := `{"assignee_user_id":"` + staff.String() + `","reason":"take"}`
		assigned := workflow(t, fixture, staff, ticketID, "assign", ticketVersionFrom(ticket), assignBody)
		escalated := workflow(
			t,
			fixture,
			staff,
			ticketID,
			"escalate",
			ticketVersionFrom(assigned),
			`{"team_group_id":"`+uuid.NewString()+`","reason":"higher rank"}`,
		)
		closed := workflow(t, fixture, staff, ticketID, "close", ticketVersionFrom(escalated), `{"reason":"done"}`)
		reopened := workflow(t, fixture, staff, ticketID, "reopen", ticketVersionFrom(closed), `{"reason":"more info"}`)
		if reopened["status"] != "open" {
			t.Fatalf("reopened status = %v, want open", reopened["status"])
		}

		stale := fixture.do(t, configureTicketRequest(
			harness.JSONRequest(fiber.MethodPost, "/tickets/"+ticketID.String()+"/close", `{"reason":"stale"}`),
			withTicketUser(staff),
			withTicketIdempotency("workflow-stale"),
			withTicketIfMatch(ticketVersionFrom(closed)),
		))
		assertTicketStatus(t, stale, fiber.StatusPreconditionFailed)
	})

	steps.Do("appeal acceptance can revoke linked punishment", func() {
		wrongTargetPunishment := fixture.issueAppealablePunishment(t, staff, uuid.New())
		wrongDefinition := fixture.createTicketDefinition(t, staff, "wrong_ban_appeal", "appeal")
		wrongDefinitionID := ticketIDFrom(t, wrongDefinition, "id")
		fixture.grantTicketRelation(t, wrongDefinitionID, domain.RelationCreator, submitter)
		wrongAppeal := fixture.do(t, configureTicketRequest(
			harness.JSONRequest(
				fiber.MethodPost,
				"/punishments/"+wrongTargetPunishment.String()+"/appeals",
				ticketBody(wrongDefinitionID, "Wrong target"),
			),
			withTicketUser(submitter),
			withTicketIdempotency("wrong-target-appeal"),
		))
		assertTicketStatus(t, wrongAppeal, fiber.StatusForbidden)

		punishmentID := fixture.issueAppealablePunishment(t, staff, submitter)
		definition := fixture.createTicketDefinition(t, staff, "ban_appeal", "appeal")
		definitionID := ticketIDFrom(t, definition, "id")
		fixture.grantTicketRelation(t, definitionID, domain.RelationCreator, submitter)

		appealResponse := fixture.do(t, configureTicketRequest(
			harness.JSONRequest(
				fiber.MethodPost,
				"/punishments/"+punishmentID.String()+"/appeals",
				ticketBody(definitionID, "Please review"),
			),
			withTicketUser(submitter),
			withTicketIdempotency("appeal-ticket"),
		))
		assertTicketStatus(t, appealResponse, fiber.StatusCreated)
		appeal := decodeTicketObject(t, appealResponse)
		ticketID := ticketIDFrom(t, appeal, "id")
		fixture.grantTicketRelation(t, ticketID, domain.RelationSubmitter, submitter)
		fixture.grantTicketRelation(t, ticketID, domain.RelationManager, staff)
		fixture.grantPunishmentRevoke(t, punishmentID, staff)

		accepted := fixture.do(t, configureTicketRequest(
			harness.JSONRequest(
				fiber.MethodPost,
				"/tickets/"+ticketID.String()+"/appeal/accept",
				`{"reason":"valid appeal","revoke_punishment":true}`,
			),
			withTicketUser(staff),
			withTicketIdempotency("appeal-accept"),
			withTicketIfMatch(ticketVersionFrom(appeal)),
		))
		assertTicketStatus(t, accepted, fiber.StatusOK)
		punishment, err := fixture.punishments.GetPunishment(context.Background(), punishmentID)
		if err != nil {
			t.Fatalf("GetPunishment() error = %v", err)
		}
		if punishment.Status != punishmentsdomain.PunishmentRevoked {
			t.Fatalf("punishment status = %s, want revoked", punishment.Status)
		}
	})
}

// workflow runs one staff action and decodes the ticket.
func workflow(
	t *testing.T,
	fixture ticketsFixture,
	actor uuid.UUID,
	ticketID uuid.UUID,
	action string,
	version uint64,
	body string,
) map[string]any {
	t.Helper()
	response := fixture.do(t, configureTicketRequest(
		harness.JSONRequest(fiber.MethodPost, "/tickets/"+ticketID.String()+"/"+action, body),
		withTicketUser(actor),
		withTicketIdempotency("workflow-"+action+"-"+strconv.FormatUint(version, 10)),
		withTicketIfMatch(version),
	))
	assertTicketStatus(t, response, fiber.StatusOK)
	return decodeTicketObject(t, response)
}

// TestOperationsEventsAndOpenAPI verifies ticket operations and contracts.
func TestOperationsEventsAndOpenAPI(t *testing.T) {
	steps := harness.NewSteps(t)
	fixture := newTicketsFixture(t)
	submitter := uuid.New()
	staff := uuid.New()

	steps.Do("ticket actions publish events and stats operations respond", func() {
		definition := fixture.createTicketDefinition(t, staff, "event_support", "support")
		definitionID := ticketIDFrom(t, definition, "id")
		fixture.grantTicketRelation(t, definitionID, domain.RelationCreator, submitter)
		ticket := fixture.createTicket(t, submitter, staff, definitionID, "event-ticket")
		ticketID := ticketIDFrom(t, ticket, "id")

		message := fixture.do(t, configureTicketRequest(
			harness.JSONRequest(fiber.MethodPost, "/tickets/"+ticketID.String()+"/messages", messageBody("event", "")),
			withTicketUser(submitter),
			withTicketIdempotency("event-message"),
		))
		assertTicketStatus(t, message, fiber.StatusCreated)

		evidence := fixture.do(t, configureTicketRequest(
			harness.JSONRequest(
				fiber.MethodPost,
				"/tickets/"+ticketID.String()+"/evidence",
				`{"external_url":"https://example.test/event","label":"event"}`,
			),
			withTicketUser(submitter),
			withTicketIdempotency("event-evidence"),
		))
		assertTicketStatus(t, evidence, fiber.StatusCreated)

		rebuild := fixture.do(t, configureTicketRequest(
			harness.JSONRequest(fiber.MethodPost, "/tickets/operations/stats/rebuild", ""),
			withTicketUser(staff),
		))
		assertTicketStatus(t, rebuild, fiber.StatusOK)

		verify := fixture.do(t, configureTicketRequest(
			harness.JSONRequest(fiber.MethodPost, "/tickets/operations/stats/verify", ""),
			withTicketUser(staff),
		))
		assertTicketStatus(t, verify, fiber.StatusOK)

		assertTicketEvent(t, fixture, "tickets.definition.created")
		assertTicketEvent(t, fixture, eventdomain.EventTicketsTicketCreated)
		assertTicketEvent(t, fixture, eventdomain.EventTicketsMessageCreated)
		assertTicketEvent(t, fixture, "tickets.evidence.added")
		assertTicketEvent(t, fixture, "tickets.stats.rebuilt")
	})

	steps.Do("all ticket HTTP routes are present in OpenAPI", func() {
		for _, route := range ticketOpenAPIRoutes() {
			assertTicketOpenAPIRoute(t, route.method, route.path)
		}
	})
}

// ticketOpenAPIRoutes returns ticket route contract entries.
func ticketOpenAPIRoutes() []struct{ method, path string } {
	return []struct{ method, path string }{
		{fiber.MethodPost, "/ticket-definitions"},
		{fiber.MethodGet, "/ticket-definitions"},
		{fiber.MethodGet, "/ticket-definitions/{definition_id}"},
		{fiber.MethodPatch, "/ticket-definitions/{definition_id}"},
		{fiber.MethodDelete, "/ticket-definitions/{definition_id}"},
		{fiber.MethodPost, "/tickets"},
		{fiber.MethodGet, "/tickets"},
		{fiber.MethodPost, "/punishments/{punishment_id}/appeals"},
		{fiber.MethodGet, "/tickets/{ticket_id}"},
		{fiber.MethodGet, "/tickets/{ticket_id}/messages"},
		{fiber.MethodPost, "/tickets/{ticket_id}/messages"},
		{fiber.MethodGet, "/tickets/{ticket_id}/evidence"},
		{fiber.MethodPost, "/tickets/{ticket_id}/evidence"},
		{fiber.MethodPost, "/tickets/{ticket_id}/assign"},
		{fiber.MethodPost, "/tickets/{ticket_id}/escalate"},
		{fiber.MethodPost, "/tickets/{ticket_id}/close"},
		{fiber.MethodPost, "/tickets/{ticket_id}/reopen"},
		{fiber.MethodPost, "/tickets/{ticket_id}/appeal/accept"},
		{fiber.MethodPost, "/tickets/{ticket_id}/appeal/reject"},
		{fiber.MethodPost, "/tickets/operations/stats/verify"},
		{fiber.MethodPost, "/tickets/operations/stats/rebuild"},
	}
}
