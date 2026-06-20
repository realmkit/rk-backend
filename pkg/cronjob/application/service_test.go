package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/realmkit/rk-backend/pkg/cronjob/domain"
	"github.com/realmkit/rk-backend/pkg/cronjob/port"
	eventdomain "github.com/realmkit/rk-backend/pkg/events/domain"
	eventtesting "github.com/realmkit/rk-backend/pkg/events/testing"
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

// TestRunOnceBoundsHandlerContext verifies handlers receive a run deadline.
func TestRunOnceBoundsHandlerContext(t *testing.T) {
	repo := newMemoryCron(testCronDefinition())
	service := NewService(Dependencies{
		Repository: repo,
		Clock:      cronClock{now: testCronNow()},
		RunTimeout: time.Nanosecond,
	}, map[string]port.Handler{
		domain.JobEventsDispatchPending: port.HandlerFunc(func(ctx context.Context, _ port.RunContext) (domain.Result, error) {
			if _, ok := ctx.Deadline(); !ok {
				t.Fatalf("handler context has no deadline")
			}
			<-ctx.Done()
			return domain.Result{}, ctx.Err()
		}),
	})

	_, err := service.RunOnce(context.Background())
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("RunOnce() error = %v, want deadline exceeded", err)
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
