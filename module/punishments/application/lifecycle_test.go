package application

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/module/punishments/domain"
	"github.com/realmkit/rk-backend/module/punishments/port"
	"github.com/realmkit/rk-backend/pkg/pagination"
)

// TestDefinitionLifecyclePublishesEvents covers create, update, delete, list, and reorder.
func TestDefinitionLifecyclePublishesEvents(t *testing.T) {
	definitions := newDefinitionFake(testDefinition())
	events := &eventFake{}
	service := NewService(Dependencies{
		Definitions: definitions,
		Cases:       newCaseFake(),
		Events:      events,
	})

	created, err := service.CreateDefinition(context.Background(), testDefinition())
	if err != nil {
		t.Fatalf("CreateDefinition() error = %v", err)
	}
	if created.Version != 1 {
		t.Fatalf("created version = %d, want 1", created.Version)
	}

	created.Name = "Updated Ban"
	updated, err := service.UpdateDefinition(context.Background(), created, created.Version)
	if err != nil {
		t.Fatalf("UpdateDefinition() error = %v", err)
	}
	if updated.Name != "Updated Ban" {
		t.Fatalf("updated name = %q", updated.Name)
	}

	if _, err := service.GetDefinition(context.Background(), created.ID); err != nil {
		t.Fatalf("GetDefinition() error = %v", err)
	}
	if _, err := service.ListDefinitions(context.Background(), port.DefinitionFilter{}, pagination.Page{}); err != nil {
		t.Fatalf("ListDefinitions() error = %v", err)
	}

	actionIDs := []uuid.UUID{uuid.New(), uuid.New()}
	if err := service.ReorderDefinitionActions(context.Background(), created.ID, actionIDs); err != nil {
		t.Fatalf("ReorderDefinitionActions() error = %v", err)
	}
	if len(definitions.reordered) != len(actionIDs) {
		t.Fatalf("reordered action count = %d, want %d", len(definitions.reordered), len(actionIDs))
	}

	if err := service.DeleteDefinition(context.Background(), created.ID, created.Version); err != nil {
		t.Fatalf("DeleteDefinition() error = %v", err)
	}
	if definitions.deletedID != created.ID {
		t.Fatalf("deleted id = %s, want %s", definitions.deletedID, created.ID)
	}
	if got := len(events.types); got != 4 {
		t.Fatalf("event count = %d, want 4", got)
	}
}

// TestPunishmentLifecycleReadUpdateAndRestrictionPaths covers case reads and updates.
func TestPunishmentLifecycleReadUpdateAndRestrictionPaths(t *testing.T) {
	definitions := newDefinitionFake(testDefinition())
	cases := newCaseFake()
	cache := &cacheFake{}
	service := NewService(Dependencies{
		Definitions: definitions,
		Cases:       cases,
		Cache:       cache,
	})
	targetID := uuid.New()
	issued, err := service.IssuePunishment(context.Background(), port.IssueCommand{
		DefinitionID: definitions.definition.ID,
		TargetUserID: targetID,
		IssuerType:   domain.IssuerSystem,
		IssuerKey:    "system",
		Reason:       "spam",
	})
	if err != nil {
		t.Fatalf("IssuePunishment() error = %v", err)
	}

	updated, err := service.UpdatePunishment(context.Background(), port.UpdateCommand{
		PunishmentID:    issued.ID,
		Reason:          "updated",
		PrivateReason:   "staff notes",
		ExpectedVersion: issued.Version,
	})
	if err != nil {
		t.Fatalf("UpdatePunishment() error = %v", err)
	}
	if updated.Reason != "updated" || updated.PrivateReason != "staff notes" {
		t.Fatalf("unexpected updated punishment: %#v", updated)
	}

	if _, err := service.GetPunishment(context.Background(), issued.ID); err != nil {
		t.Fatalf("GetPunishment() error = %v", err)
	}
	list, err := service.ListPunishments(context.Background(), port.PunishmentFilter{}, pagination.Page{})
	if err != nil {
		t.Fatalf("ListPunishments() error = %v", err)
	}
	if len(list.Items) != 1 {
		t.Fatalf("punishment list count = %d, want 1", len(list.Items))
	}

	restricted, err := service.Restricted(context.Background(), targetID, domain.ActionForumsReply)
	if err != nil {
		t.Fatalf("Restricted() error = %v", err)
	}
	if !restricted {
		t.Fatalf("Restricted() = false, want true")
	}

	active, err := service.ListActiveRestrictions(context.Background(), targetID)
	if err != nil {
		t.Fatalf("ListActiveRestrictions() error = %v", err)
	}
	if len(active) != 1 {
		t.Fatalf("active restriction count = %d, want 1", len(active))
	}
	if len(cache.clearedUsers) == 0 {
		t.Fatalf("expected issue to clear target restriction cache")
	}
}

// TestOperationsClearCacheAndPublishEvents covers expiration and repair workflows.
func TestOperationsClearCacheAndPublishEvents(t *testing.T) {
	cases := newCaseFake()
	cases.expireCount = 2
	cases.verifyReport = domain.DriftReport{
		Mismatches: []domain.CounterDrift{
			{PunishmentID: uuid.New(), ActionKey: domain.ActionForumsReply, Expected: true},
		},
	}
	cases.rebuildReport = cases.verifyReport
	cache := &cacheFake{}
	events := &eventFake{}
	service := NewService(Dependencies{
		Definitions: newDefinitionFake(testDefinition()),
		Cases:       cases,
		Cache:       cache,
		Events:      events,
	})

	expired, err := service.ExpirePunishments(context.Background())
	if err != nil {
		t.Fatalf("ExpirePunishments() error = %v", err)
	}
	if expired != 2 {
		t.Fatalf("expired count = %d, want 2", expired)
	}

	verify, err := service.VerifyRestrictions(context.Background())
	if err != nil {
		t.Fatalf("VerifyRestrictions() error = %v", err)
	}
	if len(verify.Mismatches) != 1 {
		t.Fatalf("verify mismatch count = %d, want 1", len(verify.Mismatches))
	}

	rebuild, err := service.RebuildRestrictions(context.Background())
	if err != nil {
		t.Fatalf("RebuildRestrictions() error = %v", err)
	}
	if !rebuild.Repaired {
		t.Fatalf("expected rebuild report to be marked repaired")
	}
	if cache.clearAllHits < 2 {
		t.Fatalf("clear all hits = %d, want at least 2", cache.clearAllHits)
	}
	if len(events.types) != 2 {
		t.Fatalf("operation events = %d, want 2", len(events.types))
	}

	if err := service.ClearRestrictionCache(context.Background()); err != nil {
		t.Fatalf("ClearRestrictionCache() error = %v", err)
	}
}

// TestIssuePunishmentReturnsIdempotentExistingCase covers idempotent replay behavior.
func TestIssuePunishmentReturnsIdempotentExistingCase(t *testing.T) {
	definitions := newDefinitionFake(testDefinition())
	cases := newCaseFake()
	service := NewService(Dependencies{
		Definitions: definitions,
		Cases:       cases,
	})
	command := port.IssueCommand{
		DefinitionID:   definitions.definition.ID,
		TargetUserID:   uuid.New(),
		IssuerType:     domain.IssuerSystem,
		IssuerKey:      "system",
		Reason:         "spam",
		IdempotencyKey: "same-request",
	}

	first, err := service.IssuePunishment(context.Background(), command)
	if err != nil {
		t.Fatalf("first IssuePunishment() error = %v", err)
	}
	second, err := service.IssuePunishment(context.Background(), command)
	if err != nil {
		t.Fatalf("second IssuePunishment() error = %v", err)
	}
	if first.ID != second.ID {
		t.Fatalf("idempotent replay id = %s, want %s", second.ID, first.ID)
	}
}

// TestCheckRestrictionAllowsAnonymousAndCachesMisses covers permissive check branches.
func TestCheckRestrictionAllowsAnonymousAndCachesMisses(t *testing.T) {
	cache := &cacheFake{}
	service := NewService(Dependencies{
		Definitions: newDefinitionFake(testDefinition()),
		Cases:       newCaseFake(),
		Cache:       cache,
	})

	anonymous, err := service.CheckRestriction(context.Background(), port.CheckCommand{
		ActionKey: domain.ActionForumsReply,
	})
	if err != nil {
		t.Fatalf("anonymous CheckRestriction() error = %v", err)
	}
	if !anonymous.Allowed {
		t.Fatalf("anonymous restriction check should be allowed")
	}

	userID := uuid.New()
	result, err := service.CheckRestriction(context.Background(), port.CheckCommand{
		UserID:    userID,
		ActionKey: domain.ActionForumsUpdateThread,
	})
	if err != nil {
		t.Fatalf("miss CheckRestriction() error = %v", err)
	}
	if !result.Allowed {
		t.Fatalf("missing restriction should allow")
	}
	if cached, ok := cache.values[domain.ActionForumsUpdateThread]; !ok || !cached.Allowed {
		t.Fatalf("expected allowed miss to be cached")
	}
}
