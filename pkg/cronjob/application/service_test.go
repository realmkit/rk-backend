package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/niflaot/gamehub-go/pkg/cronjob/domain"
	"github.com/niflaot/gamehub-go/pkg/cronjob/port"
	eventdomain "github.com/niflaot/gamehub-go/pkg/events/domain"
	eventtesting "github.com/niflaot/gamehub-go/pkg/events/testing"
	"github.com/niflaot/gamehub-go/pkg/pagination"
)

// TestRunOnceExecutesDueHandler verifies due job execution.
func TestRunOnceExecutesDueHandler(t *testing.T) {
	repo := newMemoryCron(testCronDefinition())
	service := NewService(Dependencies{Repository: repo, Clock: cronClock{now: testCronNow()}}, map[string]port.Handler{
		domain.JobEventsDispatchPending: port.HandlerFunc(func(context.Context, port.RunContext) (domain.Result, error) {
			return domain.Result{ProcessedCount: 3}, nil
		}),
	})

	summary, err := service.RunOnce(context.Background())
	if err != nil {
		t.Fatalf("RunOnce() error = %v", err)
	}
	if summary.ProcessedCount != 3 || repo.runs[0].Status != domain.RunSucceeded {
		t.Fatalf("summary=%+v run=%+v, want successful run", summary, repo.runs[0])
	}
}

// TestServicePublishesCronLifecycleEvents verifies scheduler events are emitted.
func TestServicePublishesCronLifecycleEvents(t *testing.T) {
	events := &eventtesting.PublisherRecorder{}
	repo := newMemoryCron(testCronDefinition())
	service := NewService(Dependencies{
		Repository: repo,
		Clock:      cronClock{now: testCronNow()},
		Events:     events,
	}, map[string]port.Handler{
		domain.JobEventsDispatchPending: port.HandlerFunc(func(context.Context, port.RunContext) (domain.Result, error) {
			return domain.Result{ProcessedCount: 3}, nil
		}),
	})

	if err := service.EnsureDefinitions(context.Background(), []domain.Definition{testCronDefinition()}); err != nil {
		t.Fatalf("EnsureDefinitions() error = %v", err)
	}
	if _, err := service.RunOnce(context.Background()); err != nil {
		t.Fatalf("RunOnce() error = %v", err)
	}
	assertCronEventKeys(t, events.Drafts(), []string{
		"cronjob.definition.updated",
		"cronjob.run.started",
		"cronjob.run.completed",
	})
}

// assertCronEventKeys verifies event draft key order.
func assertCronEventKeys(t *testing.T, drafts []eventdomain.Draft, want []string) {
	t.Helper()
	if len(drafts) != len(want) {
		t.Fatalf("event count = %d, want %d", len(drafts), len(want))
	}
	for index, key := range want {
		if string(drafts[index].Key) != key {
			t.Fatalf("event[%d] = %s, want %s", index, drafts[index].Key, key)
		}
	}
}

// TestRunOnceRecordsHandlerFailure verifies failed handlers persist runs.
func TestRunOnceRecordsHandlerFailure(t *testing.T) {
	repo := newMemoryCron(testCronDefinition())
	service := NewService(Dependencies{Repository: repo, Clock: cronClock{now: testCronNow()}}, map[string]port.Handler{
		domain.JobEventsDispatchPending: port.HandlerFunc(func(context.Context, port.RunContext) (domain.Result, error) {
			return domain.Result{}, errors.New("boom")
		}),
	})

	if _, err := service.RunOnce(context.Background()); err == nil {
		t.Fatalf("RunOnce() error = nil, want handler error")
	}
	if repo.runs[0].Status != domain.RunFailed {
		t.Fatalf("status = %s, want failed", repo.runs[0].Status)
	}
}

// TestRunOnceNoDueJob verifies empty schedules.
func TestRunOnceNoDueJob(t *testing.T) {
	repo := newMemoryCron()
	service := NewService(Dependencies{Repository: repo, Clock: cronClock{now: testCronNow()}}, nil)

	if _, err := service.RunOnce(context.Background()); !errors.Is(err, port.ErrNoDueJob) {
		t.Fatalf("RunOnce() error = %v, want no due job", err)
	}
}

// memoryCron is an in-memory cron repository.
type memoryCron struct {
	definitions map[string]domain.Definition
	runs        []domain.Run
}

// newMemoryCron creates a memory cron repository.
func newMemoryCron(definitions ...domain.Definition) *memoryCron {
	repo := &memoryCron{definitions: map[string]domain.Definition{}}
	for _, definition := range definitions {
		repo.definitions[definition.Key] = definition
	}
	return repo
}

// UpsertDefinition stores one definition.
func (repo *memoryCron) UpsertDefinition(_ context.Context, definition domain.Definition) (domain.Definition, error) {
	repo.definitions[definition.Key] = definition
	return definition, nil
}

// GetDefinition returns one definition.
func (repo *memoryCron) GetDefinition(_ context.Context, key string) (domain.Definition, error) {
	return repo.definitions[key], nil
}

// ListDefinitions returns definitions.
func (repo *memoryCron) ListDefinitions(context.Context, pagination.Page) (pagination.Result[domain.Definition], error) {
	items := []domain.Definition{}
	for _, definition := range repo.definitions {
		items = append(items, definition)
	}
	return pagination.Result[domain.Definition]{Items: items}, nil
}

// ClaimDue claims one due definition.
func (repo *memoryCron) ClaimDue(_ context.Context, workerID string, now time.Time, lockUntil time.Time) (domain.Definition, bool, error) {
	for key, definition := range repo.definitions {
		if definition.NextRunAt != nil && !definition.NextRunAt.After(now) && definition.Enabled {
			definition.LockedBy = workerID
			definition.LockedUntil = &lockUntil
			repo.definitions[key] = definition
			return definition, true, nil
		}
	}
	return domain.Definition{}, false, nil
}

// StartRun creates a run.
func (repo *memoryCron) StartRun(_ context.Context, definition domain.Definition, trigger domain.TriggerType, workerID string, now time.Time) (domain.Run, error) {
	run := domain.Run{ID: uuid.New(), JobKey: definition.Key, Status: domain.RunRunning, TriggerType: trigger, WorkerID: workerID, StartedAt: now}
	repo.runs = append(repo.runs, run)
	return run, nil
}

// CompleteRun marks a run successful.
func (repo *memoryCron) CompleteRun(_ context.Context, run domain.Run, result domain.Result, _ time.Time, _ *time.Time) error {
	repo.runs[0].Status = domain.RunSucceeded
	repo.runs[0].ProcessedCount = result.ProcessedCount
	return nil
}

// FailRun marks a run failed.
func (repo *memoryCron) FailRun(_ context.Context, _ domain.Run, message string, _ time.Time, _ *time.Time) error {
	repo.runs[0].Status = domain.RunFailed
	repo.runs[0].Error = message
	return nil
}

// Trigger returns one definition.
func (repo *memoryCron) Trigger(_ context.Context, key string) (domain.Definition, error) {
	return repo.definitions[key], nil
}

// Pause disables a definition.
func (repo *memoryCron) Pause(_ context.Context, key string, _ uint64) error {
	definition := repo.definitions[key]
	definition.Enabled = false
	repo.definitions[key] = definition
	return nil
}

// Resume enables a definition.
func (repo *memoryCron) Resume(_ context.Context, key string, _ uint64) error {
	definition := repo.definitions[key]
	definition.Enabled = true
	repo.definitions[key] = definition
	return nil
}

// ListRuns returns runs.
func (repo *memoryCron) ListRuns(context.Context, string, pagination.Page) (pagination.Result[domain.Run], error) {
	return pagination.Result[domain.Run]{Items: repo.runs}, nil
}

// RepairLocks repairs locks.
func (repo *memoryCron) RepairLocks(context.Context, time.Time) (int64, error) {
	return 0, nil
}

// cronClock is a fixed test clock.
type cronClock struct {
	now time.Time
}

// Now returns the fixed time.
func (clock cronClock) Now() time.Time {
	return clock.now
}

// testCronDefinition returns a due definition.
func testCronDefinition() domain.Definition {
	now := testCronNow()
	return domain.Definition{
		Key:                domain.JobEventsDispatchPending,
		Name:               "Dispatch events",
		ScheduleKind:       domain.ScheduleInterval,
		ScheduleExpression: time.Minute.String(),
		Enabled:            true,
		ConcurrencyPolicy:  domain.ConcurrencyForbid,
		NextRunAt:          &now,
		Version:            1,
	}
}

// testCronNow returns deterministic time.
func testCronNow() time.Time {
	return time.Unix(200, 0).UTC()
}
