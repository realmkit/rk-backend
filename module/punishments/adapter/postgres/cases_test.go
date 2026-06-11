package postgres

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/module/punishments/domain"
	"github.com/realmkit/rk-backend/module/punishments/port"
	"github.com/realmkit/rk-backend/pkg/pagination"
)

// TestDefinitionRepositoryListReorderAndConflicts covers list, reorder, and stale versions.
func TestDefinitionRepositoryListReorderAndConflicts(t *testing.T) {
	definitions, _, _ := newRepositories(t)
	first, err := definitions.Create(context.Background(), testDefinition())
	if err != nil {
		t.Fatalf("Create first error = %v", err)
	}
	secondDefinition := testDefinition()
	secondDefinition.ID = uuid.New()
	secondDefinition.Key = "voice_ban"
	secondDefinition.Actions[0].ID = uuid.New()
	second, err := definitions.Create(context.Background(), secondDefinition)
	if err != nil {
		t.Fatalf("Create second error = %v", err)
	}

	list, err := definitions.List(
		context.Background(),
		port.DefinitionFilter{Status: domain.DefinitionActive},
		pagination.Page{Limit: 1},
	)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(list.Items) != 1 || list.NextCursor == "" {
		t.Fatalf("list = %+v, want one item and next cursor", list)
	}

	if err := definitions.ReorderActions(context.Background(), first.ID, []uuid.UUID{first.Actions[0].ID}); err != nil {
		t.Fatalf("ReorderActions() error = %v", err)
	}
	if err := definitions.ReorderActions(context.Background(), first.ID, []uuid.UUID{second.Actions[0].ID}); !errors.Is(err, port.ErrNotFound) {
		t.Fatalf("ReorderActions missing error = %v, want not found", err)
	}

	first.Name = "Updated"
	if _, err := definitions.Update(context.Background(), first, first.Version+99); !errors.Is(err, port.ErrPreconditionFailed) {
		t.Fatalf("Update stale error = %v, want precondition", err)
	}
	if err := definitions.Delete(context.Background(), first.ID, first.Version+99); !errors.Is(err, port.ErrPreconditionFailed) {
		t.Fatalf("Delete stale error = %v, want precondition", err)
	}
	if _, err := definitions.FindByID(context.Background(), uuid.New()); !errors.Is(err, port.ErrNotFound) {
		t.Fatalf("FindByID missing error = %v, want not found", err)
	}
}

// TestCaseRepositoryUpdateListRevokeExpireAndIdempotency covers case repository workflows.
func TestCaseRepositoryUpdateListRevokeExpireAndIdempotency(t *testing.T) {
	definitions, cases, _ := newRepositories(t)
	definition, err := definitions.Create(context.Background(), testDefinition())
	if err != nil {
		t.Fatalf("Create definition error = %v", err)
	}
	punishment, restrictions := testPunishment(definition)
	punishment.IdempotencyKey = "issue-1"
	stored, err := cases.Issue(context.Background(), punishment, restrictions)
	if err != nil {
		t.Fatalf("Issue() error = %v", err)
	}

	replayed, err := cases.FindByIdempotencyKey(context.Background(), "issue-1")
	if err != nil {
		t.Fatalf("FindByIdempotencyKey() error = %v", err)
	}
	if replayed.ID != stored.ID {
		t.Fatalf("idempotent id = %s, want %s", replayed.ID, stored.ID)
	}
	if _, err := cases.FindByIdempotencyKey(context.Background(), "missing"); !errors.Is(err, port.ErrNotFound) {
		t.Fatalf("missing idempotency error = %v, want not found", err)
	}

	stored.Reason = "updated"
	updated, err := cases.Update(context.Background(), stored, stored.Version)
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	if updated.Version != stored.Version+1 || updated.Reason != "updated" {
		t.Fatalf("updated = %+v, want incremented version and reason", updated)
	}
	if _, err := cases.Update(context.Background(), stored, stored.Version); !errors.Is(err, port.ErrPreconditionFailed) {
		t.Fatalf("stale update error = %v, want precondition", err)
	}

	list, err := cases.List(
		context.Background(),
		port.PunishmentFilter{TargetUserID: stored.TargetUserID, Status: domain.PunishmentActive},
		pagination.Page{Limit: 10},
	)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(list.Items) != 1 {
		t.Fatalf("list count = %d, want 1", len(list.Items))
	}
	other, otherRestrictions := testPunishment(definition)
	other.ID = uuid.New()
	other.IdempotencyKey = "issue-2"
	if _, err := cases.Issue(context.Background(), other, otherRestrictions); err != nil {
		t.Fatalf("Issue second error = %v", err)
	}
	page, err := cases.List(context.Background(), port.PunishmentFilter{}, pagination.Page{Limit: 1})
	if err != nil {
		t.Fatalf("List bounded error = %v", err)
	}
	if len(page.Items) != 1 || page.NextCursor == "" {
		t.Fatalf("bounded page = %+v, want one item and next cursor", page)
	}

	active, err := cases.ListActiveRestrictions(context.Background(), stored.TargetUserID, time.Now().UTC())
	if err != nil {
		t.Fatalf("ListActiveRestrictions() error = %v", err)
	}
	if len(active) != 1 {
		t.Fatalf("active restriction count = %d, want 1", len(active))
	}

	stored.Status = domain.PunishmentRevoked
	now := time.Now().UTC()
	stored.RevokedAt = &now
	if err := cases.Revoke(context.Background(), stored, updated.Version+99); !errors.Is(err, port.ErrPreconditionFailed) {
		t.Fatalf("stale revoke error = %v, want precondition", err)
	}
	if err := cases.Revoke(context.Background(), stored, updated.Version); err != nil {
		t.Fatalf("Revoke() error = %v", err)
	}
	if _, _, err := cases.ActiveRestriction(
		context.Background(),
		stored.TargetUserID,
		domain.ActionForumsReply,
		time.Now().UTC(),
	); !errors.Is(err, port.ErrNotFound) {
		t.Fatalf("revoked ActiveRestriction error = %v, want not found", err)
	}
}

// TestCaseRepositoryExpireDueRemovesRestrictions covers natural expiration.
func TestCaseRepositoryExpireDueRemovesRestrictions(t *testing.T) {
	definitions, cases, _ := newRepositories(t)
	definition, err := definitions.Create(context.Background(), testDefinition())
	if err != nil {
		t.Fatalf("Create definition error = %v", err)
	}
	punishment, restrictions := testPunishment(definition)
	expiresAt := time.Now().UTC().Add(-time.Minute)
	punishment.ExpiresAt = &expiresAt
	if _, err := cases.Issue(context.Background(), punishment, restrictions); err != nil {
		t.Fatalf("Issue() error = %v", err)
	}

	expired, err := cases.ExpireDue(context.Background(), time.Now().UTC())
	if err != nil {
		t.Fatalf("ExpireDue() error = %v", err)
	}
	if expired != 1 {
		t.Fatalf("expired = %d, want 1", expired)
	}
	if _, _, err := cases.ActiveRestriction(
		context.Background(),
		punishment.TargetUserID,
		domain.ActionForumsReply,
		time.Now().UTC(),
	); !errors.Is(err, port.ErrNotFound) {
		t.Fatalf("expired ActiveRestriction error = %v, want not found", err)
	}
}
