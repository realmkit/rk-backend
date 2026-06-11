package punishments_e2e

import (
	"context"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/niflaot/gamehub-go/e2e/harness"
	eventdomain "github.com/niflaot/gamehub-go/pkg/events/domain"
)

// TestEventsAndOpenAPIContract verifies punishment event and contract coverage.
func TestEventsAndOpenAPIContract(t *testing.T) {
	steps := harness.NewSteps(t)
	fixture := newPunishmentsFixture(t)
	actor := uuid.New()
	target := uuid.New()

	steps.Do("definition and punishment actions publish scoped events", func() {
		definition := fixture.createDefinition(t, actor, "event_restriction")
		definitionID := idFrom(t, definition, "id")
		punishment := fixture.issuePunishment(t, actor, definitionID, target, "event-restriction", nil)
		id := idFrom(t, punishment, "id")

		updated := fixture.do(t, configureRequest(
			harness.JSONRequest(fiber.MethodPatch, "/punishments/"+id.String(), `{"reason":"updated"}`),
			withPunishmentUser(actor),
			withPunishmentIfMatch(versionFrom(punishment)),
		))
		assertPunishmentStatus(t, updated, fiber.StatusOK)
		current := decodePunishmentObject(t, updated)

		revoked := fixture.do(t, configureRequest(
			harness.JSONRequest(fiber.MethodPost, "/punishments/"+id.String()+"/revoke", `{"reason":"done"}`),
			withPunishmentUser(actor),
			withPunishmentIdempotency("revoke-event-restriction"),
			withPunishmentIfMatch(versionFrom(current)),
		))
		assertPunishmentStatus(t, revoked, fiber.StatusNoContent)

		if _, err := fixture.service.RebuildRestrictions(context.Background()); err != nil {
			t.Fatalf("RebuildRestrictions() error = %v", err)
		}
		assertPunishmentEvent(t, fixture, "punishments.definition.created")
		assertPunishmentEvent(t, fixture, eventdomain.EventPunishmentsPunishmentIssued)
		assertPunishmentEvent(t, fixture, "punishments.punishment.updated")
		assertPunishmentEvent(t, fixture, "punishments.punishment.revoked")
		assertPunishmentEvent(t, fixture, "punishments.restrictions.rebuilt")
	})

	steps.Do("all punishment HTTP routes are present in OpenAPI", func() {
		for _, route := range []struct{ method, path string }{
			{fiber.MethodPost, "/punishment-definitions"},
			{fiber.MethodGet, "/punishment-definitions"},
			{fiber.MethodGet, "/punishment-definitions/{definition_id}"},
			{fiber.MethodPatch, "/punishment-definitions/{definition_id}"},
			{fiber.MethodDelete, "/punishment-definitions/{definition_id}"},
			{fiber.MethodPost, "/punishment-definitions/{definition_id}/actions/reorder"},
			{fiber.MethodPost, "/punishments"},
			{fiber.MethodGet, "/punishments"},
			{fiber.MethodGet, "/punishments/{punishment_id}"},
			{fiber.MethodPatch, "/punishments/{punishment_id}"},
			{fiber.MethodPost, "/punishments/{punishment_id}/revoke"},
			{fiber.MethodGet, "/users/{user_id}/punishments"},
			{fiber.MethodGet, "/users/{user_id}/punishments/active"},
			{fiber.MethodPost, "/punishments/restrictions/check"},
			{fiber.MethodGet, "/users/{user_id}/punishments/restrictions"},
		} {
			assertPunishmentOpenAPIRoute(t, route.method, route.path)
		}
	})
}
