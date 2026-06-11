package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/realmkit/rk-backend/pkg/cronjob/domain"
	"github.com/realmkit/rk-backend/pkg/cronjob/port"
	"github.com/realmkit/rk-backend/pkg/orm"
	"github.com/realmkit/rk-backend/pkg/pagination"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// TestRepositoryDefinitionLifecycle verifies cron definition persistence.
func TestRepositoryDefinitionLifecycle(t *testing.T) {
	repo := newCronRepository(t)
	definition, err := repo.UpsertDefinition(context.Background(), testDefinition())
	if err != nil {
		t.Fatalf("UpsertDefinition() error = %v", err)
	}
	got, err := repo.GetDefinition(context.Background(), definition.Key)
	if err != nil {
		t.Fatalf("GetDefinition() error = %v", err)
	}
	if got.Key != definition.Key {
		t.Fatalf("definition = %+v, want persisted definition", got)
	}
	list, err := repo.ListDefinitions(context.Background(), pagination.Page{Limit: 10})
	if err != nil || len(list.Items) != 1 {
		t.Fatalf("ListDefinitions() = %+v, %v, want one item", list, err)
	}
}

// TestRepositoryListsAreBounded verifies cron operator lists honor page limits.
func TestRepositoryListsAreBounded(t *testing.T) {
	repo := newCronRepository(t)
	definition := testDefinition()
	second := testDefinition()
	second.Key = domain.JobForumsVerifyStats
	second.Name = "Verify forum stats"
	for _, item := range []domain.Definition{definition, second} {
		if _, err := repo.UpsertDefinition(context.Background(), item); err != nil {
			t.Fatalf("UpsertDefinition(%s) error = %v", item.Key, err)
		}
	}

	list, err := repo.ListDefinitions(context.Background(), pagination.Page{Limit: 1})
	if err != nil {
		t.Fatalf("ListDefinitions() error = %v", err)
	}
	if len(list.Items) != 1 {
		t.Fatalf("ListDefinitions() items = %d, want bounded limit 1", len(list.Items))
	}

	claimed, ok, err := repo.ClaimDue(context.Background(), "worker", testNow(), testNow().Add(time.Minute))
	if err != nil || !ok {
		t.Fatalf("ClaimDue() = %+v, %v, %v, want claimed", claimed, ok, err)
	}
	if _, err := repo.StartRun(context.Background(), claimed, domain.TriggerManual, "worker", testNow()); err != nil {
		t.Fatalf("StartRun first error = %v", err)
	}
	if _, err := repo.StartRun(context.Background(), claimed, domain.TriggerManual, "worker", testNow().Add(time.Second)); err != nil {
		t.Fatalf("StartRun second error = %v", err)
	}
	runs, err := repo.ListRuns(context.Background(), claimed.Key, pagination.Page{Limit: 1})
	if err != nil {
		t.Fatalf("ListRuns() error = %v", err)
	}
	if len(runs.Items) != 1 {
		t.Fatalf("ListRuns() items = %d, want bounded limit 1", len(runs.Items))
	}
}

// TestRepositoryClaimRunCompleteAndRepair verifies run lifecycle.
func TestRepositoryClaimRunCompleteAndRepair(t *testing.T) {
	repo := newCronRepository(t)
	definition, _ := repo.UpsertDefinition(context.Background(), testDefinition())
	claimed, ok, err := repo.ClaimDue(context.Background(), "worker", testNow(), testNow().Add(time.Minute))
	if err != nil || !ok {
		t.Fatalf("ClaimDue() = %+v, %v, %v, want claimed", claimed, ok, err)
	}
	run, err := repo.StartRun(context.Background(), claimed, domain.TriggerSchedule, "worker", testNow())
	if err != nil {
		t.Fatalf("StartRun() error = %v", err)
	}
	if err := repo.CompleteRun(context.Background(), run, domain.Result{ProcessedCount: 2}, testNow().Add(time.Second), definition.NextAfter(testNow())); err != nil {
		t.Fatalf("CompleteRun() error = %v", err)
	}
	runs, err := repo.ListRuns(context.Background(), definition.Key, pagination.Page{Limit: 10})
	if err != nil || runs.Items[0].Status != domain.RunSucceeded {
		t.Fatalf("ListRuns() = %+v, %v, want succeeded run", runs, err)
	}
}

// TestRepositoryPausePrecondition verifies optimistic version checks.
func TestRepositoryPausePrecondition(t *testing.T) {
	repo := newCronRepository(t)
	definition, _ := repo.UpsertDefinition(context.Background(), testDefinition())
	if err := repo.Pause(context.Background(), definition.Key, definition.Version+1); err != port.ErrPreconditionFailed {
		t.Fatalf("Pause() error = %v, want precondition", err)
	}
}

// newCronRepository creates a SQLite-backed cron repository.
func newCronRepository(t *testing.T) Repository {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite error = %v", err)
	}
	if err := db.AutoMigrate(&DefinitionModel{}, &RunModel{}); err != nil {
		t.Fatalf("AutoMigrate() error = %v", err)
	}
	return NewRepository(orm.NewStore(db))
}

// testDefinition returns a due cron definition.
func testDefinition() domain.Definition {
	next := testNow()
	return domain.Definition{
		Key:                domain.JobEventsDispatchPending,
		Name:               "Dispatch events",
		ScheduleKind:       domain.ScheduleInterval,
		ScheduleExpression: time.Minute.String(),
		Enabled:            true,
		ConcurrencyPolicy:  domain.ConcurrencyForbid,
		NextRunAt:          &next,
		Version:            1,
	}
}

// testNow returns deterministic time.
func testNow() time.Time {
	return time.Unix(400, 0).UTC()
}
