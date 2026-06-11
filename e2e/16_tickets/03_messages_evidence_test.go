package tickets_e2e

import (
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/niflaot/gamehub-go/e2e/harness"
	"github.com/niflaot/gamehub-go/module/groups/domain"
)

// TestMessagesAndEvidence verifies ticket conversation and evidence behavior.
func TestMessagesAndEvidence(t *testing.T) {
	steps := harness.NewSteps(t)
	fixture := newTicketsFixture(t)
	submitter := uuid.New()
	staff := uuid.New()
	definition := fixture.createTicketDefinition(t, staff, "conversation_support", "support")
	definitionID := ticketIDFrom(t, definition, "id")
	fixture.grantTicketRelation(t, definitionID, domain.RelationCreator, submitter)
	ticket := fixture.createTicket(t, submitter, staff, definitionID, "conversation-ticket")
	ticketID := ticketIDFrom(t, ticket, "id")

	steps.Do("messages require idempotency and are visible by audience", func() {
		missingKey := fixture.do(t, configureTicketRequest(
			harness.JSONRequest(fiber.MethodPost, "/tickets/"+ticketID.String()+"/messages", messageBody("hello", "")),
			withTicketUser(submitter),
		))
		assertTicketStatus(t, missingKey, fiber.StatusBadRequest)

		public := fixture.do(t, configureTicketRequest(
			harness.JSONRequest(fiber.MethodPost, "/tickets/"+ticketID.String()+"/messages", messageBody("hello", "")),
			withTicketUser(submitter),
			withTicketIdempotency("message-public"),
		))
		assertTicketStatus(t, public, fiber.StatusCreated)

		private := fixture.do(t, configureTicketRequest(
			harness.JSONRequest(
				fiber.MethodPost,
				"/tickets/"+ticketID.String()+"/messages",
				messageBody("staff note", "staff_only"),
			),
			withTicketUser(staff),
			withTicketIdempotency("message-staff"),
		))
		assertTicketStatus(t, private, fiber.StatusCreated)

		visibleToSubmitter := fixture.do(t, configureTicketRequest(
			harness.JSONRequest(fiber.MethodGet, "/tickets/"+ticketID.String()+"/messages", ""),
			withTicketUser(submitter),
		))
		assertTicketStatus(t, visibleToSubmitter, fiber.StatusOK)

		visibleToStaff := fixture.do(t, configureTicketRequest(
			harness.JSONRequest(
				fiber.MethodGet,
				"/tickets/"+ticketID.String()+"/messages?include_staff_only=true",
				"",
			),
			withTicketUser(staff),
		))
		assertTicketStatus(t, visibleToStaff, fiber.StatusOK)
	})

	steps.Do("evidence supports URLs and validates assets", func() {
		external := fixture.do(t, configureTicketRequest(
			harness.JSONRequest(
				fiber.MethodPost,
				"/tickets/"+ticketID.String()+"/evidence",
				`{"external_url":"https://example.test/proof","label":"proof"}`,
			),
			withTicketUser(submitter),
			withTicketIdempotency("evidence-url"),
		))
		assertTicketStatus(t, external, fiber.StatusCreated)

		missingAssetID := uuid.New()
		missingAsset := fixture.do(t, configureTicketRequest(
			harness.JSONRequest(
				fiber.MethodPost,
				"/tickets/"+ticketID.String()+"/evidence",
				`{"asset_id":"`+missingAssetID.String()+`","label":"missing"}`,
			),
			withTicketUser(submitter),
			withTicketIdempotency("evidence-missing-asset"),
		))
		assertTicketStatus(t, missingAsset, fiber.StatusNotFound)

		assetID := uuid.New()
		fixture.resolver.assets[assetID] = true
		assetEvidence := fixture.do(t, configureTicketRequest(
			harness.JSONRequest(
				fiber.MethodPost,
				"/tickets/"+ticketID.String()+"/evidence",
				`{"asset_id":"`+assetID.String()+`","label":"asset proof"}`,
			),
			withTicketUser(staff),
			withTicketIdempotency("evidence-asset"),
		))
		assertTicketStatus(t, assetEvidence, fiber.StatusCreated)

		list := fixture.do(t, configureTicketRequest(
			harness.JSONRequest(fiber.MethodGet, "/tickets/"+ticketID.String()+"/evidence", ""),
			withTicketUser(submitter),
		))
		assertTicketStatus(t, list, fiber.StatusOK)
	})
}

// messageBody returns a valid message body.
func messageBody(text string, visibility string) string {
	prefix := `{"content_document_json":{"type":"doc"},"content_text":"` + text + `"`
	if visibility != "" {
		prefix += `,"visibility":"` + visibility + `"`
	}
	return prefix + `}`
}
