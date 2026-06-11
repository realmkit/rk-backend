package punishments_e2e

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/e2e/harness"
	punishmentsport "github.com/realmkit/rk-backend/module/punishments/port"
)

// TestOperationsExpireVerifyAndRebuild verifies operator punishment paths.
func TestOperationsExpireVerifyAndRebuild(t *testing.T) {
	steps := harness.NewSteps(t)
	fixture := newPunishmentsFixture(t)
	actor := uuid.New()
	target := uuid.New()
	definition := fixture.createDefinition(t, actor, "expiring_restriction")

	steps.Do("expired punishments are closed by the operation", func() {
		startsAt := time.Now().UTC().Add(-2 * time.Hour)
		expiresAt := time.Now().UTC().Add(-time.Hour)
		_, err := fixture.service.IssuePunishment(context.Background(), issueCommand(
			actor,
			idFrom(t, definition, "id"),
			target,
			startsAt,
			&expiresAt,
			"expired-operation",
		))
		if err != nil {
			t.Fatalf("IssuePunishment() error = %v", err)
		}
		count, err := fixture.service.ExpirePunishments(context.Background())
		if err != nil {
			t.Fatalf("ExpirePunishments() error = %v", err)
		}
		if count != 1 {
			t.Fatalf("expired count = %d, want 1", count)
		}
	})

	steps.Do("restriction projections verify and rebuild without drift", func() {
		_ = fixture.issuePunishment(t, actor, idFrom(t, definition, "id"), uuid.New(), "verify-rebuild", nil)
		report, err := fixture.service.VerifyRestrictions(context.Background())
		if err != nil {
			t.Fatalf("VerifyRestrictions() error = %v", err)
		}
		if len(report.Mismatches) != 0 || report.Repaired {
			t.Fatalf("verify report = %+v", report)
		}
		rebuilt, err := fixture.service.RebuildRestrictions(context.Background())
		if err != nil {
			t.Fatalf("RebuildRestrictions() error = %v", err)
		}
		if len(rebuilt.Mismatches) != 0 || !rebuilt.Repaired {
			t.Fatalf("rebuild report = %+v", rebuilt)
		}
		if err := fixture.service.ClearRestrictionCache(context.Background()); err != nil {
			t.Fatalf("ClearRestrictionCache() error = %v", err)
		}
	})
}

// issueCommand returns a service-level issue command for operation tests.
func issueCommand(
	actor uuid.UUID,
	definitionID uuid.UUID,
	target uuid.UUID,
	startsAt time.Time,
	expiresAt *time.Time,
	key string,
) punishmentsport.IssueCommand {
	issuer := actor
	return punishmentsport.IssueCommand{
		ActorUserID:    actor,
		DefinitionID:   definitionID,
		TargetUserID:   target,
		IssuerType:     "user",
		IssuerUserID:   &issuer,
		Reason:         "E2E operation",
		StartsAt:       startsAt,
		ExpiresAt:      expiresAt,
		Source:         "e2e",
		IdempotencyKey: key,
	}
}
