package punishments_e2e

import (
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/niflaot/gamehub-go/e2e/harness"
)

// TestIssueUpdateAndRevoke verifies active punishment lifecycle behavior.
func TestIssueUpdateAndRevoke(t *testing.T) {
	steps := harness.NewSteps(t)
	fixture := newPunishmentsFixture(t)
	actor := uuid.New()
	target := uuid.New()
	definition := fixture.createDefinition(t, actor, "reply_restriction")
	definitionID := idFrom(t, definition, "id")

	steps.Do("issuing requires idempotency", func() {
		response := fixture.do(t, configureRequest(
			harness.JSONRequest(fiber.MethodPost, "/punishments", issueBody(definitionID, target, nil)),
			withPunishmentUser(actor),
		))
		assertPunishmentStatus(t, response, fiber.StatusBadRequest)
	})

	steps.Do("active punishment can be issued and idempotently replayed", func() {
		expiresAt := time.Now().UTC().Add(time.Hour)
		issued := fixture.issuePunishment(t, actor, definitionID, target, "reply-restriction", &expiresAt)
		id := idFrom(t, issued, "id")
		if issued["status"] != "active" || issued["target_user_id"] != target.String() {
			t.Fatalf("issued punishment = %+v", issued)
		}

		replay := fixture.issuePunishment(t, actor, definitionID, target, "reply-restriction", &expiresAt)
		if idFrom(t, replay, "id") != id {
			t.Fatalf("idempotent replay id = %s, want %s", idFrom(t, replay, "id"), id)
		}

		get := fixture.do(t, configureRequest(
			harness.JSONRequest(fiber.MethodGet, "/punishments/"+id.String(), ""),
			withPunishmentUser(actor),
		))
		assertPunishmentStatus(t, get, fiber.StatusOK)

		list := fixture.do(t, configureRequest(
			harness.JSONRequest(fiber.MethodGet, "/users/"+target.String()+"/punishments/active", ""),
			withPunishmentUser(actor),
		))
		assertPunishmentStatus(t, list, fiber.StatusOK)
	})

	steps.Do("punishment notes update and revoke use If-Match", func() {
		revokeTarget := uuid.New()
		punishment := fixture.issuePunishment(t, actor, definitionID, revokeTarget, "reply-revoke", nil)
		id := idFrom(t, punishment, "id")
		version := versionFrom(punishment)

		missingMatch := fixture.do(t, configureRequest(
			harness.JSONRequest(fiber.MethodPatch, "/punishments/"+id.String(), `{"reason":"updated"}`),
			withPunishmentUser(actor),
		))
		assertPunishmentStatus(t, missingMatch, fiber.StatusPreconditionRequired)

		updated := fixture.do(t, configureRequest(
			harness.JSONRequest(fiber.MethodPatch, "/punishments/"+id.String(), `{"reason":"updated"}`),
			withPunishmentUser(actor),
			withPunishmentIfMatch(version),
		))
		assertPunishmentStatus(t, updated, fiber.StatusOK)
		current := decodePunishmentObject(t, updated)

		missingRevokeMatch := fixture.do(t, configureRequest(
			harness.JSONRequest(fiber.MethodPost, "/punishments/"+id.String()+"/revoke", `{"reason":"appealed"}`),
			withPunishmentUser(actor),
			withPunishmentIdempotency("revoke-missing-match"),
		))
		assertPunishmentStatus(t, missingRevokeMatch, fiber.StatusPreconditionRequired)

		revoked := fixture.do(t, configureRequest(
			harness.JSONRequest(fiber.MethodPost, "/punishments/"+id.String()+"/revoke", `{"reason":"appealed"}`),
			withPunishmentUser(actor),
			withPunishmentIdempotency("revoke-reply-restriction"),
			withPunishmentIfMatch(versionFrom(current)),
		))
		assertPunishmentStatus(t, revoked, fiber.StatusNoContent)

		check := fixture.do(t, configureRequest(
			harness.JSONRequest(
				fiber.MethodPost,
				"/punishments/restrictions/check",
				`{"user_id":"`+revokeTarget.String()+`","action_key":"gamehub.forums.reply"}`,
			),
			withPunishmentUser(actor),
		))
		assertPunishmentStatus(t, check, fiber.StatusOK)
		if allowed := decodePunishmentObject(t, check)["allowed"]; allowed != true {
			t.Fatalf("allowed = %v, want true after revoke", allowed)
		}
	})
}
