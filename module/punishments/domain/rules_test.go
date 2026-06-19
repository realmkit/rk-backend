package domain

import (
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

// TestPunishmentNormalizeValidateAndActiveBranches covers punishment validation branches.
func TestPunishmentNormalizeValidateAndActiveBranches(t *testing.T) {
	punishment := validPunishment()
	punishment.Reason = " reason "
	punishment.IssuerKey = " system "
	punishment = punishment.Normalize()

	if punishment.Reason != "reason" || punishment.IssuerKey != "system" {
		t.Fatalf("expected normalized text fields: %#v", punishment)
	}
	if err := punishment.Validate(); err != nil {
		t.Fatalf("expected punishment to validate: %v", err)
	}
	if !punishment.ActiveAt(punishment.StartsAt.Add(time.Second)) {
		t.Fatalf("expected active punishment during window")
	}
	if punishment.ActiveAt(punishment.StartsAt.Add(-time.Second)) {
		t.Fatalf("expected punishment not active before start")
	}
	punishment.Status = PunishmentRevoked
	if punishment.ActiveAt(time.Now().UTC()) {
		t.Fatalf("expected revoked punishment not active")
	}
}

// TestPunishmentValidateRejectsInvalidIssuerAndDates covers invalid punishment fields.
func TestPunishmentValidateRejectsInvalidIssuerAndDates(t *testing.T) {
	punishment := validPunishment()
	punishment.DefinitionID = uuid.Nil
	punishment.TargetUserID = uuid.Nil
	punishment.IssuerType = IssuerUser
	punishment.IssuerUserID = nil
	punishment.IssuerKey = ""
	punishment.Status = "paused"
	punishment.Reason = ""
	punishment.StartsAt = time.Time{}
	expiresAt := time.Now().UTC().Add(-time.Hour)
	punishment.ExpiresAt = &expiresAt
	punishment.Snapshots = []ActionSnapshot{{TargetSystem: "bad"}}

	if err := punishment.Validate(); err == nil {
		t.Fatalf("expected invalid punishment to fail validation")
	}
}

// TestDurationRulesCoverPermanentAndBounds covers duration policy branches.
func TestDurationRulesCoverPermanentAndBounds(t *testing.T) {
	definition := validDefinition()
	definition.AllowPermanent = false
	startsAt := time.Date(2026, 6, 10, 12, 0, 0, 0, time.UTC)
	if err := ValidateIssueDuration(definition, startsAt, nil); err == nil {
		t.Fatalf("expected permanent duration denial")
	}

	minimum := int64(60)
	maximum := int64(120)
	definition.AllowPermanent = true
	definition.MinDurationSeconds = &minimum
	definition.MaxDurationSeconds = &maximum
	tooLong := startsAt.Add(5 * time.Minute)
	if err := ValidateIssueDuration(definition, startsAt, &tooLong); err == nil {
		t.Fatalf("expected max duration denial")
	}
	valid := startsAt.Add(90 * time.Second)
	if err := ValidateIssueDuration(definition, startsAt, &valid); err != nil {
		t.Fatalf("expected valid bounded duration: %v", err)
	}
}

// TestActionSnapshotAndRestrictionRules covers snapshots, clones, and active windows.
func TestActionSnapshotAndRestrictionRules(t *testing.T) {
	punishment := validPunishment()
	webhook := validAction()
	webhook.TargetSystem = TargetWebhook
	webhook.ActionType = ActionWebhookDispatch
	snapshot := SnapshotFromTemplate(punishment.ID, webhook)
	if snapshot.ActionType != ActionWebhookDispatch {
		t.Fatalf("snapshot action type = %q, want webhook dispatch", snapshot.ActionType)
	}
	if _, ok := RestrictionFromSnapshot(punishment, snapshot); ok {
		t.Fatalf("webhook action should not create restriction")
	}
	if err := snapshot.Validate(); err != nil {
		t.Fatalf("expected snapshot to validate: %v", err)
	}

	restriction, ok := RestrictionFromSnapshot(punishment, SnapshotFromTemplate(punishment.ID, validAction()))
	if !ok {
		t.Fatalf("expected realmkit restrict action to create restriction")
	}
	if !restriction.ActiveAt(punishment.StartsAt.Add(time.Second)) {
		t.Fatalf("expected restriction active during window")
	}
	if restriction.ActiveAt(punishment.StartsAt.Add(-time.Second)) {
		t.Fatalf("expected restriction inactive before start")
	}
	expiresAt := punishment.StartsAt.Add(time.Minute)
	restriction.ExpiresAt = &expiresAt
	if restriction.ActiveAt(expiresAt.Add(time.Second)) {
		t.Fatalf("expected restriction inactive after expiration")
	}
}

// TestValidationHelpersCoverInvalidBranches covers enum and primitive validators.
func TestValidationHelpersCoverInvalidBranches(t *testing.T) {
	if got := NewValidationError([]Violation{{Field: "x", Message: "bad"}}).Error(); got != ErrValidation.Error() {
		t.Fatalf("validation error = %q, want %q", got, ErrValidation.Error())
	}
	if got := NewValidationError(nil); got != nil {
		t.Fatalf("expected nil validation error: %v", got)
	}
	invalidations := [][]Violation{
		ValidateKey("key", "Bad Key"),
		ValidateColor("color", "red"),
		ValidateActionKey("action_key", "bad-action"),
		ValidateDefinitionStatus("status", "paused"),
		ValidateTargetSystem("target", "matrix"),
		ValidateActionType("action_type", TargetRealmKit, ActionWebhookDispatch),
		ValidateIssuerType("issuer_type", "robot"),
		ValidatePunishmentStatus("status", "paused"),
	}
	for _, violations := range invalidations {
		if len(violations) == 0 {
			t.Fatalf("expected invalid value to fail")
		}
	}
}

// TestDefinitionValidationCoversDefaultsAndDurationErrors covers definition branches.
func TestDefinitionValidationCoversDefaultsAndDurationErrors(t *testing.T) {
	definition := validDefinition()
	definition.Name = "  Chat Ban  "
	definition.Color = ""
	definition.Status = ""
	definition.Version = 0
	definition = definition.Normalize()
	if definition.Name != "Chat Ban" || definition.Color == "" || definition.Version != 1 {
		t.Fatalf("expected normalized definition defaults: %#v", definition)
	}

	negative := int64(-1)
	definition.Key = "Bad Key"
	definition.Color = "red"
	definition.Name = ""
	definition.Status = "paused"
	definition.Severity = -1
	definition.MinDurationSeconds = &negative
	definition.Description = strings.Repeat("x", 10)
	if err := definition.Validate(); err == nil {
		t.Fatalf("expected invalid definition to fail")
	}
}
