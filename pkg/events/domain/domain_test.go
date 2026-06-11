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

// TestCatalogContainsPlannedModuleEvents verifies catalog coverage.
func TestCatalogContainsPlannedModuleEvents(t *testing.T) {
	required := map[EventKey]bool{
		EventUsersUserProvisioned:             false,
		EventAssetsAssetUploadCompleted:       false,
		EventGroupsMembershipAdded:            false,
		EventForumsThreadCreated:              false,
		EventPunishmentsPunishmentIssued:      false,
		EventTicketsTicketCreated:             false,
		EventTicketsMessageCreated:            false,
		EventCronjobRunCompleted:              false,
		EventNotificationsNotificationCreated: false,
		EventMessagesMessageSent:              false,
	}
	for _, descriptor := range Catalog() {
		if _, ok := required[descriptor.Key]; ok {
			required[descriptor.Key] = true
		}
	}
	for key, found := range required {
		if !found {
			t.Fatalf("catalog missing %s", key)
		}
	}
	if err := (Scope{Type: ScopeTicket, ID: "ticket-1"}).Validate(); err != nil {
		t.Fatalf("ticket scope Validate() error = %v", err)
	}
}
