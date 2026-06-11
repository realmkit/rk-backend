package domain

import (
	"testing"

	"github.com/google/uuid"
)

// TestDraftValidateAcceptsScopedEvent verifies valid event drafts.
func TestDraftValidateAcceptsScopedEvent(t *testing.T) {
	aggregateID := uuid.New()
	draft := Draft{
		Key:           EventForumsThreadCreated,
		SchemaVersion: 1,
		Producer:      ProducerForums,
		AggregateType: "forum_thread",
		AggregateID:   &aggregateID,
		Payload:       map[string]any{"thread_id": aggregateID.String()},
		Scopes:        []Scope{{Type: ScopeThread, ID: aggregateID.String()}},
	}

	if err := draft.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}

// TestDraftValidateRejectsInvalidKeyAndScope verifies invalid drafts.
func TestDraftValidateRejectsInvalidKeyAndScope(t *testing.T) {
	draft := Draft{
		Key:           "BadKey",
		SchemaVersion: 0,
		Producer:      ProducerForums,
		AggregateType: "forum_thread",
		Scopes:        []Scope{{Type: ScopeUser}},
	}

	err := draft.Validate()
	if err == nil {
		t.Fatalf("Validate() error = nil, want validation error")
	}
}

// TestDraftValidateRejectsMissingProducerAggregateAndScopes verifies required routing metadata.
func TestDraftValidateRejectsMissingProducerAggregateAndScopes(t *testing.T) {
	err := Draft{Key: EventTicketsTicketCreated, SchemaVersion: 1}.Validate()
	validation, ok := err.(ValidationError)
	if !ok {
		t.Fatalf("Validate() error = %v, want ValidationError", err)
	}
	if len(validation.Violations) != 3 {
		t.Fatalf("Violations = %+v, want producer, aggregate, and scopes violations", validation.Violations)
	}
	if validation.Error() == "" {
		t.Fatalf("ValidationError.Error() = empty")
	}
}

// TestScopeValidatePermissionRules verifies permission scope validation.
func TestScopeValidatePermissionRules(t *testing.T) {
	valid := Scope{Type: ScopePermission, Permission: "forums.view"}
	if err := valid.Validate(); err != nil {
		t.Fatalf("valid scope error = %v", err)
	}

	invalid := Scope{Type: ScopeStaff, Permission: "forums.view"}
	if err := invalid.Validate(); err == nil {
		t.Fatalf("invalid scope error = nil, want error")
	}
}

// TestStatusAndScopeTypeValidation verifies small enum validators.
func TestStatusAndScopeTypeValidation(t *testing.T) {
	if violations := ValidateStatus("status", StatusDead); len(violations) != 0 {
		t.Fatalf("ValidateStatus(dead) = %+v, want none", violations)
	}
	if violations := ValidateStatus("status", "lost"); len(violations) != 1 {
		t.Fatalf("ValidateStatus(lost) = %+v, want one violation", violations)
	}
	if violations := ValidateScopeType("scope", ScopeAsset); len(violations) != 0 {
		t.Fatalf("ValidateScopeType(asset) = %+v, want none", violations)
	}
	if err := (Scope{Type: ScopeAsset}).Validate(); err == nil {
		t.Fatalf("asset scope without id Validate() error = nil, want validation")
	}
	if err := (Scope{Type: ScopeGlobal, Permission: "forums.view"}).Validate(); err == nil {
		t.Fatalf("global scope with permission Validate() error = nil, want validation")
	}
}

// TestSharedEventVocabularyValidates verifies stable producer event keys.
func TestSharedEventVocabularyValidates(t *testing.T) {
	keys := []EventKey{
		EventUsersUserProvisioned,
		EventAssetsAssetUploadCompleted,
		EventGroupsMembershipAdded,
		EventForumsThreadCreated,
		EventPunishmentsPunishmentIssued,
		EventTicketsTicketCreated,
		EventTicketsMessageCreated,
		EventCronjobRunCompleted,
		EventNotificationsNotificationCreated,
		EventMessagesMessageSent,
	}
	for _, key := range keys {
		if violations := ValidateEventKey("key", key); len(violations) > 0 {
			t.Fatalf("%s ValidateEventKey() violations = %+v", key, violations)
		}
	}
	if err := (Scope{Type: ScopeTicket, ID: "ticket-1"}).Validate(); err != nil {
		t.Fatalf("ticket scope Validate() error = %v", err)
	}
}
