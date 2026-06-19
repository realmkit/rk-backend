package punishments_e2e

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/e2e/harness"
	"github.com/realmkit/rk-backend/module/punishments/domain"
)

// TestRestrictionChecks verifies active RealmKit restrictions are enforced.
func TestRestrictionChecks(t *testing.T) {
	steps := harness.NewSteps(t)
	fixture := newPunishmentsFixture(t)
	actor := uuid.New()
	target := uuid.New()
	definition := fixture.createDefinition(
		t,
		actor,
		"community_lock",
		domain.ActionForumsCreateThread,
		domain.ActionForumsReply,
	)
	_ = fixture.issuePunishment(t, actor, idFrom(t, definition, "id"), target, "community-lock", nil)

	steps.Do("matching action is denied with punishment summary", func() {
		result := restrictionCheck(t, fixture, actor, target, domain.ActionForumsReply)
		if result["allowed"] != false || result["punishment"] == nil || result["restriction"] == nil {
			t.Fatalf("restriction result = %+v", result)
		}
	})

	steps.Do("unrelated action remains allowed", func() {
		result := restrictionCheck(t, fixture, actor, target, domain.ActionForumsUpdateThread)
		if result["allowed"] != true {
			t.Fatalf("allowed = %v, want true", result["allowed"])
		}
	})

	steps.Do("active restriction list exposes all current projections", func() {
		response := fixture.do(t, configureRequest(
			harness.JSONRequest(fiber.MethodGet, "/users/"+target.String()+"/punishments/restrictions", ""),
			withPunishmentUser(actor),
		))
		assertPunishmentStatus(t, response, fiber.StatusOK)
		var restrictions []map[string]any
		if err := json.NewDecoder(response.Body).Decode(&restrictions); err != nil {
			t.Fatalf("Decode() error = %v", err)
		}
		if len(restrictions) != 2 {
			t.Fatalf("len(restrictions) = %d, want 2", len(restrictions))
		}
	})
}

// restrictionCheck checks one restriction route.
func restrictionCheck(
	t *testing.T,
	fixture punishmentsFixture,
	actor uuid.UUID,
	target uuid.UUID,
	action string,
) map[string]any {
	t.Helper()
	response := fixture.do(t, configureRequest(
		harness.JSONRequest(
			http.MethodPost,
			"/punishments/restrictions/check",
			`{"user_id":"`+target.String()+`","action_key":"`+action+`"}`,
		),
		withPunishmentUser(actor),
	))
	assertPunishmentStatus(t, response, fiber.StatusOK)
	return decodePunishmentObject(t, response)
}
