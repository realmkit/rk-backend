package tickets_e2e

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/niflaot/gamehub-go/e2e/harness"
	groupsdomain "github.com/niflaot/gamehub-go/module/groups/domain"
	groupsport "github.com/niflaot/gamehub-go/module/groups/port"
	punishmentsdomain "github.com/niflaot/gamehub-go/module/punishments/domain"
	punishmentsport "github.com/niflaot/gamehub-go/module/punishments/port"
	"github.com/niflaot/gamehub-go/pkg/api/headers"
	"github.com/niflaot/gamehub-go/pkg/api/openapi"
	eventdomain "github.com/niflaot/gamehub-go/pkg/events/domain"
)

// createTicketDefinition creates one definition.
func (fixture ticketsFixture) createTicketDefinition(t *testing.T, actor uuid.UUID, key string, kind string) map[string]any {
	t.Helper()
	body := `{"key":"` + key + `","name":"` + key + `","kind":"` + kind + `","status":"active"}`
	if kind == "appeal" {
		body = `{"key":"` + key + `","name":"` + key + `","kind":"appeal","status":"active","requires_punishment":true}`
	}
	response := fixture.do(t, configureTicketRequest(
		harness.JSONRequest(fiber.MethodPost, "/ticket-definitions", body),
		withTicketUser(actor),
	))
	assertTicketStatus(t, response, fiber.StatusCreated)
	return decodeTicketObject(t, response)
}

// grantTicketRelation grants one user relation on a ticket object.
func (fixture ticketsFixture) grantTicketRelation(
	t *testing.T,
	objectID uuid.UUID,
	relation groupsdomain.Relation,
	userID uuid.UUID,
) {
	t.Helper()
	_, err := fixture.groups.CreateTuple(context.Background(), groupsport.CreateTupleCommand{
		Tuple: groupsdomain.RelationTuple{
			ObjectType:  groupsdomain.ObjectTicket,
			ObjectID:    objectID,
			Relation:    relation,
			SubjectType: groupsdomain.SubjectUser,
			SubjectID:   userID,
		},
	})
	if err != nil {
		t.Fatalf("CreateTuple(%s) error = %v", relation, err)
	}
}

// grantPunishmentRevoke grants a staff actor punishment revoke permission.
func (fixture ticketsFixture) grantPunishmentRevoke(t *testing.T, punishmentID uuid.UUID, actor uuid.UUID) {
	t.Helper()
	_, err := fixture.groups.CreateTuple(context.Background(), groupsport.CreateTupleCommand{
		Tuple: groupsdomain.RelationTuple{
			ObjectType:  groupsdomain.ObjectPunishment,
			ObjectID:    punishmentID,
			Relation:    groupsdomain.RelationModerator,
			SubjectType: groupsdomain.SubjectUser,
			SubjectID:   actor,
		},
	})
	if err != nil {
		t.Fatalf("CreateTuple(punishment) error = %v", err)
	}
}

// issueAppealablePunishment creates one punishment through the service.
func (fixture ticketsFixture) issueAppealablePunishment(t *testing.T, actor uuid.UUID, target uuid.UUID) uuid.UUID {
	t.Helper()
	definition, err := fixture.punishments.CreateDefinition(context.Background(), appealPunishmentDefinition())
	if err != nil {
		t.Fatalf("CreateDefinition() error = %v", err)
	}
	issuer := actor
	punishment, err := fixture.punishments.IssuePunishment(context.Background(), punishmentsport.IssueCommand{
		ActorUserID: actor, DefinitionID: definition.ID, TargetUserID: target,
		IssuerType: punishmentsdomain.IssuerUser, IssuerUserID: &issuer,
		Reason: "appealable", Source: "e2e", IdempotencyKey: "appealable-" + target.String(),
	})
	if err != nil {
		t.Fatalf("IssuePunishment() error = %v", err)
	}
	return punishment.ID
}

// appealPunishmentDefinition returns one punishment definition for appeal tests.
func appealPunishmentDefinition() punishmentsdomain.Definition {
	return punishmentsdomain.Definition{
		ID: uuid.New(), Key: "ticket_appeal_ban", Name: "Ticket Appeal Ban",
		Color: "#aa0000", Severity: 10, Status: punishmentsdomain.DefinitionActive,
		AllowPermanent: true, RequiresReason: true,
		Actions: []punishmentsdomain.ActionTemplate{{
			ID: uuid.New(), TargetSystem: punishmentsdomain.TargetGameHub,
			ActionKey: punishmentsdomain.ActionForumsReply, Effect: punishmentsdomain.EffectRestrict,
			Status: punishmentsdomain.DefinitionActive,
		}},
	}
}

// configureTicketRequest applies request mutations.
func configureTicketRequest(request *http.Request, configs ...func(*http.Request)) *http.Request {
	for _, config := range configs {
		config(request)
	}
	return request
}

// withTicketUser adds the current-user header.
func withTicketUser(userID uuid.UUID) func(*http.Request) {
	return func(request *http.Request) {
		request.Header.Set("X-GameHub-User-Id", userID.String())
	}
}

// withTicketIdempotency adds an idempotency key.
func withTicketIdempotency(key string) func(*http.Request) {
	return func(request *http.Request) { request.Header.Set(headers.IdempotencyKey, key) }
}

// withTicketIfMatch adds an If-Match header.
func withTicketIfMatch(version uint64) func(*http.Request) {
	return func(request *http.Request) {
		request.Header.Set(headers.IfMatch, `"`+strconv.FormatUint(version, 10)+`"`)
	}
}

// decodeTicketObject decodes one JSON object.
func decodeTicketObject(t *testing.T, response *http.Response) map[string]any {
	t.Helper()
	var payload map[string]any
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		t.Fatalf("Decode() error = %v", err)
	}
	return payload
}

// assertTicketStatus verifies response status.
func assertTicketStatus(t *testing.T, response *http.Response, want int) {
	t.Helper()
	if response.StatusCode != want {
		t.Fatalf("StatusCode = %d, want %d body = %q", response.StatusCode, want, harness.ResponseBody(t, response))
	}
}

// ticketIDFrom extracts an ID field.
func ticketIDFrom(t *testing.T, payload map[string]any, field string) uuid.UUID {
	t.Helper()
	id, err := uuid.Parse(payload[field].(string))
	if err != nil {
		t.Fatalf("Parse(%s) error = %v", field, err)
	}
	return id
}

// ticketVersionFrom extracts the root version.
func ticketVersionFrom(payload map[string]any) uint64 {
	return uint64(payload["version"].(float64))
}

// ticketBody returns a valid ticket intake body.
func ticketBody(definitionID uuid.UUID, title string) string {
	return `{"definition_id":"` + definitionID.String() + `","title":"` + title +
		`","content_document_json":{"type":"doc"},"content_text":"hello"}`
}

// createTicket creates one ticket and grants participant/staff relations.
func (fixture ticketsFixture) createTicket(
	t *testing.T,
	submitter uuid.UUID,
	staff uuid.UUID,
	definitionID uuid.UUID,
	key string,
) map[string]any {
	t.Helper()
	response := fixture.do(t, configureTicketRequest(
		harness.JSONRequest(fiber.MethodPost, "/tickets", ticketBody(definitionID, "E2E Ticket")),
		withTicketUser(submitter),
		withTicketIdempotency(key),
	))
	assertTicketStatus(t, response, fiber.StatusCreated)
	ticket := decodeTicketObject(t, response)
	id := ticketIDFrom(t, ticket, "id")
	fixture.grantTicketRelation(t, id, groupsdomain.RelationSubmitter, submitter)
	fixture.grantTicketRelation(t, id, groupsdomain.RelationManager, staff)
	return ticket
}

// assertTicketEvent verifies that an event key was published.
func assertTicketEvent(t *testing.T, fixture ticketsFixture, key eventdomain.EventKey) {
	t.Helper()
	for _, draft := range fixture.events.Drafts() {
		if draft.Key == key && draft.Producer == eventdomain.ProducerTickets {
			return
		}
	}
	t.Fatalf("event %s was not published", key)
}

// assertTicketOpenAPIRoute verifies an OpenAPI operation exists.
func assertTicketOpenAPIRoute(t *testing.T, method string, path string) {
	t.Helper()
	ok, err := openapi.OperationExists(method, path)
	if err != nil {
		t.Fatalf("OperationExists() error = %v", err)
	}
	if !ok {
		t.Fatalf("%s %s missing OpenAPI operation", method, path)
	}
}
