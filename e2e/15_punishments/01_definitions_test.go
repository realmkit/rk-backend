package punishments_e2e

import (
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/e2e/harness"
)

// TestDefinitionLifecycle verifies punishment definition administration.
func TestDefinitionLifecycle(t *testing.T) {
	steps := harness.NewSteps(t)
	fixture := newPunishmentsFixture(t)
	actor := uuid.New()

	steps.Do("anonymous definition creation is rejected", func() {
		response := fixture.do(t, harness.JSONRequest(
			fiber.MethodPost,
			"/punishment-definitions",
			definitionBody("anon_ban"),
		))
		assertPunishmentStatus(t, response, fiber.StatusUnauthorized)
	})

	steps.Do("definition creation requires idempotency", func() {
		request := harness.JSONRequest(fiber.MethodPost, "/punishment-definitions", definitionBody("missing_key"))
		response := fixture.do(t, configureRequest(request, withPunishmentUser(actor)))
		assertPunishmentStatus(t, response, fiber.StatusBadRequest)
	})

	steps.Do("definition can be created, listed, updated, reordered, and deleted", func() {
		created := fixture.createDefinition(t, actor, "forum_mute", "realmkit.forums.reply", "realmkit.forums.update_thread")
		id := idFrom(t, created, "id")
		version := versionFrom(created)
		if created["color"] != "#ff5555" {
			t.Fatalf("color = %v, want #ff5555", created["color"])
		}

		get := fixture.do(t, configureRequest(
			harness.JSONRequest(fiber.MethodGet, "/punishment-definitions/"+id.String(), ""),
			withPunishmentUser(actor),
		))
		assertPunishmentStatus(t, get, fiber.StatusOK)

		list := fixture.do(t, configureRequest(
			harness.JSONRequest(fiber.MethodGet, "/punishment-definitions?status=active", ""),
			withPunishmentUser(actor),
		))
		assertPunishmentStatus(t, list, fiber.StatusOK)

		missingMatch := fixture.do(t, configureRequest(
			harness.JSONRequest(fiber.MethodPatch, "/punishment-definitions/"+id.String(), definitionBody("forum_mute")),
			withPunishmentUser(actor),
		))
		assertPunishmentStatus(t, missingMatch, fiber.StatusPreconditionRequired)

		updated := fixture.do(t, configureRequest(
			harness.JSONRequest(fiber.MethodPatch, "/punishment-definitions/"+id.String(), definitionBody("forum_mute")),
			withPunishmentUser(actor),
			withPunishmentIfMatch(version),
		))
		assertPunishmentStatus(t, updated, fiber.StatusOK)
		current := decodePunishmentObject(t, updated)

		actions := current["actions"].([]any)
		reorderBody := `{"ids":["` + actions[1].(map[string]any)["id"].(string) +
			`","` + actions[0].(map[string]any)["id"].(string) + `"]}`
		reordered := fixture.do(t, configureRequest(
			harness.JSONRequest(fiber.MethodPost, "/punishment-definitions/"+id.String()+"/actions/reorder", reorderBody),
			withPunishmentUser(actor),
			withPunishmentIdempotency("reorder-forum-mute"),
		))
		assertPunishmentStatus(t, reordered, fiber.StatusNoContent)

		stale := fixture.do(t, configureRequest(
			harness.JSONRequest(fiber.MethodPatch, "/punishment-definitions/"+id.String(), definitionBody("forum_mute")),
			withPunishmentUser(actor),
			withPunishmentIfMatch(version),
		))
		assertPunishmentStatus(t, stale, fiber.StatusPreconditionFailed)

		deleted := fixture.do(t, configureRequest(
			harness.JSONRequest(fiber.MethodDelete, "/punishment-definitions/"+id.String(), ""),
			withPunishmentUser(actor),
			withPunishmentIfMatch(versionFrom(current)),
		))
		assertPunishmentStatus(t, deleted, fiber.StatusNoContent)
	})

	steps.Do("duplicate active definition keys conflict", func() {
		_ = fixture.createDefinition(t, actor, "duplicate_mute")
		response := fixture.do(t, configureRequest(
			harness.JSONRequest(fiber.MethodPost, "/punishment-definitions", definitionBody("duplicate_mute")),
			withPunishmentUser(actor),
			withPunishmentIdempotency("definition-duplicate-mute-2"),
		))
		assertPunishmentStatus(t, response, fiber.StatusConflict)
	})

	steps.Do("invalid definitions and missing resources map to problems", func() {
		invalid := fixture.do(t, configureRequest(
			harness.JSONRequest(
				fiber.MethodPost,
				"/punishment-definitions",
				`{"key":"Bad Key","name":"","color":"red","status":"active","actions":[]}`,
			),
			withPunishmentUser(actor),
			withPunishmentIdempotency("invalid-definition"),
		))
		assertPunishmentStatus(t, invalid, fiber.StatusUnprocessableEntity)

		unknownField := fixture.do(t, configureRequest(
			harness.JSONRequest(
				fiber.MethodPost,
				"/punishment-definitions",
				`{"key":"unknown_field","name":"Unknown","status":"active","actions":[],"extra":true}`,
			),
			withPunishmentUser(actor),
			withPunishmentIdempotency("unknown-field"),
		))
		assertPunishmentStatus(t, unknownField, fiber.StatusBadRequest)

		missing := fixture.do(t, configureRequest(
			harness.JSONRequest(fiber.MethodGet, "/punishment-definitions/"+uuid.NewString(), ""),
			withPunishmentUser(actor),
		))
		assertPunishmentStatus(t, missing, fiber.StatusNotFound)
	})
}
