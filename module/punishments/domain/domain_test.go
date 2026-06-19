package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

// TestDefinitionValidationAllowsActionlessDefinitions verifies definitions can be query-only.
func TestDefinitionValidationAllowsActionlessDefinitions(t *testing.T) {
	definition := validDefinition()
	definition.Actions = nil

	if err := definition.Normalize().Validate(); err != nil {
		t.Fatalf("Validate() error = %v, want nil", err)
	}
}

// TestActionValidationRejectsMismatchedTargetAction verifies target-owned actions.
func TestActionValidationRejectsMismatchedTargetAction(t *testing.T) {
	action := validAction()
	action.TargetSystem = TargetWebhook

	if err := action.Normalize().Validate(); err == nil {
		t.Fatalf("Validate() error = nil, want validation error")
	}
}

// TestIssueDurationValidationEnforcesDefinitionLimits verifies duration policy.
func TestIssueDurationValidationEnforcesDefinitionLimits(t *testing.T) {
	minimum := int64(60)
	maximum := int64(120)
	startsAt := time.Date(2026, 6, 10, 12, 0, 0, 0, time.UTC)
	expiresAt := startsAt.Add(30 * time.Second)
	definition := validDefinition()
	definition.MinDurationSeconds = &minimum
	definition.MaxDurationSeconds = &maximum

	if err := ValidateIssueDuration(definition, startsAt, &expiresAt); err == nil {
		t.Fatalf("ValidateIssueDuration() error = nil, want validation error")
	}
}

// TestRestrictionFromSnapshotBuildsRealmKitProjection verifies active restrictions.
func TestRestrictionFromSnapshotBuildsRealmKitProjection(t *testing.T) {
	punishment := validPunishment()
	snapshot := SnapshotFromTemplate(punishment.ID, validAction())

	restriction, ok := RestrictionFromSnapshot(punishment, snapshot)
	if !ok {
		t.Fatalf("RestrictionFromSnapshot() ok = false, want true")
	}
	if restriction.TargetUserID != punishment.TargetUserID {
		t.Fatalf("TargetUserID = %s, want %s", restriction.TargetUserID, punishment.TargetUserID)
	}
}

// TestPunishmentActiveAtHonorsExpiration verifies expired punishments stop applying.
func TestPunishmentActiveAtHonorsExpiration(t *testing.T) {
	punishment := validPunishment()
	expiresAt := punishment.StartsAt.Add(time.Minute)
	punishment.ExpiresAt = &expiresAt

	if punishment.ActiveAt(expiresAt.Add(time.Second)) {
		t.Fatalf("ActiveAt() = true after expiration, want false")
	}
}

func validDefinition() Definition {
	return Definition{
		ID:             uuid.New(),
		Key:            "chat_ban",
		Name:           "Chat Ban",
		Color:          "#ff5555",
		Severity:       10,
		Status:         DefinitionActive,
		AllowPermanent: true,
		Actions:        []ActionTemplate{validAction()},
		Version:        1,
	}.Normalize()
}

func validAction() ActionTemplate {
	return ActionTemplate{
		ID:                uuid.New(),
		DefinitionID:      uuid.New(),
		TargetSystem:      TargetRealmKit,
		ActionType:        ActionForumsReply,
		ConfigurationJSON: []byte(`{}`),
		Status:            DefinitionActive,
	}.Normalize()
}

func validPunishment() Punishment {
	return Punishment{
		ID:           uuid.New(),
		DefinitionID: uuid.New(),
		TargetUserID: uuid.New(),
		IssuerType:   IssuerSystem,
		IssuerKey:    "test",
		Reason:       "Testing",
		Status:       PunishmentActive,
		StartsAt:     time.Date(2026, 6, 10, 12, 0, 0, 0, time.UTC),
		Version:      1,
	}.Normalize()
}
