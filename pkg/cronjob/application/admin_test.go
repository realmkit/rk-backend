package application

import (
	"context"
	"errors"
	"testing"

	"github.com/niflaot/gamehub-go/pkg/cronjob/domain"
	"github.com/niflaot/gamehub-go/pkg/cronjob/port"
	"github.com/niflaot/gamehub-go/pkg/pagination"
)

// TestAdminUseCasesCoverDefinitionAndRunOperations verifies admin-facing service methods.
func TestAdminUseCasesCoverDefinitionAndRunOperations(t *testing.T) {
	repo := newMemoryCron(testCronDefinition())
	service := NewService(Dependencies{Repository: repo, Clock: cronClock{now: testCronNow()}}, map[string]port.Handler{
		domain.JobEventsDispatchPending: port.HandlerFunc(func(context.Context, port.RunContext) (domain.Result, error) {
			return domain.Result{ProcessedCount: 7}, nil
		}),
	})

	summary, err := service.Trigger(context.Background(), domain.JobEventsDispatchPending)
	if err != nil {
		t.Fatalf("Trigger() error = %v", err)
	}
	if summary.ProcessedCount != 7 {
		t.Fatalf("manual processed count = %d, want 7", summary.ProcessedCount)
	}

	list, err := service.ListDefinitions(context.Background(), pagination.Page{Limit: 10})
	if err != nil {
		t.Fatalf("ListDefinitions() error = %v", err)
	}
	if len(list.Items) != 1 {
		t.Fatalf("definition count = %d, want 1", len(list.Items))
	}
	definition, err := service.GetDefinition(context.Background(), domain.JobEventsDispatchPending)
	if err != nil {
		t.Fatalf("GetDefinition() error = %v", err)
	}
	if definition.Key != domain.JobEventsDispatchPending {
		t.Fatalf("definition key = %q", definition.Key)
	}

	if err := service.Pause(context.Background(), domain.JobEventsDispatchPending, definition.Version); err != nil {
		t.Fatalf("Pause() error = %v", err)
	}
	if repo.definitions[domain.JobEventsDispatchPending].Enabled {
		t.Fatalf("expected definition to be paused")
	}
	if err := service.Resume(context.Background(), domain.JobEventsDispatchPending, definition.Version); err != nil {
		t.Fatalf("Resume() error = %v", err)
	}
	if !repo.definitions[domain.JobEventsDispatchPending].Enabled {
		t.Fatalf("expected definition to be resumed")
	}

	runs, err := service.ListRuns(context.Background(), domain.JobEventsDispatchPending, pagination.Page{Limit: 10})
	if err != nil {
		t.Fatalf("ListRuns() error = %v", err)
	}
	if len(runs.Items) != 1 {
		t.Fatalf("run count = %d, want 1", len(runs.Items))
	}
	repaired, err := service.RepairLocks(context.Background())
	if err != nil {
		t.Fatalf("RepairLocks() error = %v", err)
	}
	if repaired != 0 {
		t.Fatalf("repaired locks = %d, want 0", repaired)
	}
}

// TestTriggerMissingHandlerReturnsConflict verifies manual runs fail without handlers.
func TestTriggerMissingHandlerReturnsConflict(t *testing.T) {
	repo := newMemoryCron(testCronDefinition())
	service := NewService(Dependencies{Repository: repo, Clock: cronClock{now: testCronNow()}}, nil)

	if _, err := service.Trigger(context.Background(), domain.JobEventsDispatchPending); !errors.Is(err, port.ErrHandlerMissing) {
		t.Fatalf("Trigger() error = %v, want handler missing", err)
	}
}

// TestDefaultServiceSettingsUseSystemClock covers default worker and clock branches.
func TestDefaultServiceSettingsUseSystemClock(t *testing.T) {
	service := NewService(Dependencies{Repository: newMemoryCron()}, nil)
	if service.workerID != "gamehub-cron" {
		t.Fatalf("workerID = %q, want default", service.workerID)
	}
	if service.lockDuration == 0 {
		t.Fatalf("expected default lock duration")
	}
	if service.clock.Now().IsZero() {
		t.Fatalf("expected system clock to return a time")
	}
}
