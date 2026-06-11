package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/niflaot/gamehub-go/module/punishments/domain"
	"github.com/niflaot/gamehub-go/pkg/orm"
	"github.com/niflaot/gamehub-go/pkg/postgres/migrations"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// TestDefinitionRepositoryLifecycle verifies persisted definition actions.
func TestDefinitionRepositoryLifecycle(t *testing.T) {
	definitions, _, _ := newRepositories(t)
	created, err := definitions.Create(context.Background(), testDefinition())
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if len(created.Actions) != 1 {
		t.Fatalf("actions = %d, want 1", len(created.Actions))
	}

	created.Name = "Muted"
	updated, err := definitions.Update(context.Background(), created, created.Version)
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	if updated.Version != created.Version+1 {
		t.Fatalf("version = %d, want incremented", updated.Version)
	}
	if err := definitions.Delete(context.Background(), updated.ID, updated.Version); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
}

// TestCaseRepositoryIssueAndRestriction verifies case and restriction persistence.
func TestCaseRepositoryIssueAndRestriction(t *testing.T) {
	definitions, cases, _ := newRepositories(t)
	definition, err := definitions.Create(context.Background(), testDefinition())
	if err != nil {
		t.Fatalf("Create definition error = %v", err)
	}
	punishment, restrictions := testPunishment(definition)

	stored, err := cases.Issue(context.Background(), punishment, restrictions)
	if err != nil {
		t.Fatalf("Issue() error = %v", err)
	}
	if len(stored.Snapshots) != 1 {
		t.Fatalf("snapshots = %d, want 1", len(stored.Snapshots))
	}
	restriction, summary, err := cases.ActiveRestriction(
		context.Background(),
		punishment.TargetUserID,
		domain.ActionForumsReply,
		time.Now().UTC(),
	)
	if err != nil {
		t.Fatalf("ActiveRestriction() error = %v", err)
	}
	if summary == nil || restriction.PunishmentID != stored.ID {
		t.Fatalf("restriction = %+v summary = %+v, want stored punishment", restriction, summary)
	}
}

// TestCaseRepositoryRebuildRestrictionsRepairsDrift verifies projection repair.
func TestCaseRepositoryRebuildRestrictionsRepairsDrift(t *testing.T) {
	definitions, cases, db := newRepositories(t)
	definition, err := definitions.Create(context.Background(), testDefinition())
	if err != nil {
		t.Fatalf("Create definition error = %v", err)
	}
	punishment, restrictions := testPunishment(definition)
	if _, err := cases.Issue(context.Background(), punishment, restrictions); err != nil {
		t.Fatalf("Issue() error = %v", err)
	}
	if err := db.Where("1 = 1").Delete(&RestrictionModel{}).Error; err != nil {
		t.Fatalf("Delete restrictions error = %v", err)
	}

	report, err := cases.RebuildRestrictions(context.Background(), time.Now().UTC())
	if err != nil {
		t.Fatalf("RebuildRestrictions() error = %v", err)
	}
	if !report.Repaired || len(report.Mismatches) != 1 {
		t.Fatalf("report = %+v, want one repaired mismatch", report)
	}
}

func newRepositories(t *testing.T) (DefinitionRepository, CaseRepository, *gorm.DB) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("gorm.Open() error = %v", err)
	}
	if _, err := migrations.NewRunner(db, migrations.DefaultSource()).Up(context.Background()); err != nil {
		t.Fatalf("migrate Up() error = %v", err)
	}
	store := orm.NewStore(db)
	return NewDefinitionRepository(store), NewCaseRepository(store), db
}

func testDefinition() domain.Definition {
	action := domain.ActionTemplate{
		ID:                uuid.New(),
		TargetSystem:      domain.TargetGameHub,
		ActionKey:         domain.ActionForumsReply,
		Effect:            domain.EffectRestrict,
		ConfigurationJSON: []byte(`{}`),
		Status:            domain.DefinitionActive,
	}.Normalize()
	return domain.Definition{
		ID:             uuid.New(),
		Key:            "chat_ban",
		Name:           "Chat Ban",
		Color:          "#ff5555",
		Status:         domain.DefinitionActive,
		AllowPermanent: true,
		RequiresReason: true,
		Actions:        []domain.ActionTemplate{action},
		Version:        1,
	}.Normalize()
}

func testPunishment(definition domain.Definition) (domain.Punishment, []domain.ActiveRestriction) {
	now := time.Now().UTC()
	punishment := domain.Punishment{
		ID:           uuid.New(),
		DefinitionID: definition.ID,
		TargetUserID: uuid.New(),
		IssuerType:   domain.IssuerSystem,
		IssuerKey:    "test",
		Reason:       "spam",
		Status:       domain.PunishmentActive,
		StartsAt:     now.Add(-time.Minute),
		Version:      1,
	}.Normalize()
	snapshot := domain.SnapshotFromTemplate(punishment.ID, definition.Actions[0])
	punishment.Snapshots = []domain.ActionSnapshot{snapshot}
	restriction, _ := domain.RestrictionFromSnapshot(punishment, snapshot)
	return punishment, []domain.ActiveRestriction{restriction}
}
