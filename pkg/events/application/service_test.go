package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/niflaot/gamehub-go/pkg/events/domain"
	"github.com/niflaot/gamehub-go/pkg/events/port"
	"github.com/niflaot/gamehub-go/pkg/pagination"
)

// TestPublishValidatesAndStoresDraft verifies publish validation and storage.
func TestPublishValidatesAndStoresDraft(t *testing.T) {
	repo := &memoryEvents{}
	service := NewService(Dependencies{Repository: repo, Clock: fixedClock{now: testNow()}})

	event, err := service.Publish(context.Background(), testDraft())
	if err != nil {
		t.Fatalf("Publish() error = %v", err)
	}
	if event.Status != domain.StatusPending || len(repo.events) != 1 {
		t.Fatalf("event = %+v stored=%d, want pending stored event", event, len(repo.events))
	}
}

// TestDispatchOnceProcessesClaimedEvents verifies successful dispatch.
func TestDispatchOnceProcessesClaimedEvents(t *testing.T) {
	repo := &memoryEvents{}
	event, _ := repo.Publish(context.Background(), testDraft(), testNow())
	service := NewService(Dependencies{
		Repository: repo,
		Broker:     &recordBroker{},
		Clock:      fixedClock{now: testNow()},
	})

	result, err := service.DispatchOnce(context.Background(), "worker")
	if err != nil {
		t.Fatalf("DispatchOnce() error = %v", err)
	}
	if result.Processed != 1 || repo.events[event.ID].Status != domain.StatusProcessed {
		t.Fatalf("result = %+v status=%s, want processed", result, repo.events[event.ID].Status)
	}
}

// TestDispatchOnceMarksDeadAfterBrokerFailure verifies failed dispatch tracking.
func TestDispatchOnceMarksDeadAfterBrokerFailure(t *testing.T) {
	repo := &memoryEvents{}
	event, _ := repo.Publish(context.Background(), testDraft(), testNow())
	service := NewService(Dependencies{
		Repository:  repo,
		Broker:      failingBroker{},
		Clock:       fixedClock{now: testNow()},
		MaxAttempts: 1,
	})

	if _, err := service.DispatchOnce(context.Background(), "worker"); err == nil {
		t.Fatalf("DispatchOnce() error = nil, want failure")
	}
	if repo.events[event.ID].Status != domain.StatusDead {
		t.Fatalf("status = %s, want dead", repo.events[event.ID].Status)
	}
}

// TestReplayAndCancelUpdateStatus verifies operator state changes.
func TestReplayAndCancelUpdateStatus(t *testing.T) {
	repo := &memoryEvents{}
	event, _ := repo.Publish(context.Background(), testDraft(), testNow())
	service := NewService(Dependencies{Repository: repo, Clock: fixedClock{now: testNow()}})

	if err := service.Cancel(context.Background(), event.ID); err != nil {
		t.Fatalf("Cancel() error = %v", err)
	}
	if err := service.Replay(context.Background(), event.ID); err != nil {
		t.Fatalf("Replay() error = %v", err)
	}
	if repo.events[event.ID].Status != domain.StatusPending {
		t.Fatalf("status = %s, want pending", repo.events[event.ID].Status)
	}
}

// memoryEvents is an in-memory event repository.
type memoryEvents struct {
	events map[uuid.UUID]domain.Event
}

// Publish stores one event.
func (repo *memoryEvents) Publish(_ context.Context, draft domain.Draft, now time.Time) (domain.Event, error) {
	if repo.events == nil {
		repo.events = map[uuid.UUID]domain.Event{}
	}
	event := domain.Event{
		ID:            uuid.New(),
		Key:           draft.Key,
		SchemaVersion: draft.SchemaVersion,
		Producer:      draft.Producer,
		AggregateType: draft.AggregateType,
		Scopes:        draft.Scopes,
		Status:        domain.StatusPending,
		OccurredAt:    now,
		AvailableAt:   now,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	repo.events[event.ID] = event
	return event, nil
}

// Get returns one event.
func (repo *memoryEvents) Get(_ context.Context, id uuid.UUID) (domain.Event, error) {
	return repo.events[id], nil
}

// List returns all events.
func (repo *memoryEvents) List(context.Context, port.ListFilter, pagination.Page) (pagination.Result[domain.Event], error) {
	items := []domain.Event{}
	for _, event := range repo.events {
		items = append(items, event)
	}
	return pagination.Result[domain.Event]{Items: items}, nil
}

// Claim claims pending events.
func (repo *memoryEvents) Claim(_ context.Context, workerID string, _ int, _ time.Time, lockUntil time.Time) ([]domain.Event, error) {
	claimed := []domain.Event{}
	for id, event := range repo.events {
		if event.Status == domain.StatusPending || event.Status == domain.StatusFailed {
			event.Status = domain.StatusProcessing
			event.LockedBy = workerID
			event.LockedUntil = &lockUntil
			event.AttemptCount++
			repo.events[id] = event
			claimed = append(claimed, event)
		}
	}
	return claimed, nil
}

// MarkProcessed marks one event processed.
func (repo *memoryEvents) MarkProcessed(_ context.Context, id uuid.UUID, now time.Time) error {
	return repo.set(id, domain.StatusProcessed, now)
}

// MarkFailed marks one event failed.
func (repo *memoryEvents) MarkFailed(_ context.Context, id uuid.UUID, message string, _ time.Time, now time.Time) error {
	event := repo.events[id]
	event.LastError = message
	repo.events[id] = event
	return repo.set(id, domain.StatusFailed, now)
}

// MarkDead marks one event dead.
func (repo *memoryEvents) MarkDead(_ context.Context, id uuid.UUID, message string, now time.Time) error {
	event := repo.events[id]
	event.LastError = message
	repo.events[id] = event
	return repo.set(id, domain.StatusDead, now)
}

// Replay moves one event to pending.
func (repo *memoryEvents) Replay(_ context.Context, id uuid.UUID, now time.Time) error {
	return repo.set(id, domain.StatusPending, now)
}

// Cancel cancels one event.
func (repo *memoryEvents) Cancel(_ context.Context, id uuid.UUID, now time.Time) error {
	return repo.set(id, domain.StatusCancelled, now)
}

// set updates one event status.
func (repo *memoryEvents) set(id uuid.UUID, status domain.Status, now time.Time) error {
	event := repo.events[id]
	event.Status = status
	event.UpdatedAt = now
	repo.events[id] = event
	return nil
}

// recordBroker records successful broadcasts.
type recordBroker struct{}

// Publish records one event.
func (*recordBroker) Publish(context.Context, domain.Event) error {
	return nil
}

// failingBroker always fails.
type failingBroker struct{}

// Publish returns a broadcast failure.
func (failingBroker) Publish(context.Context, domain.Event) error {
	return errors.New("broadcast failed")
}

// fixedClock returns a fixed time.
type fixedClock struct {
	now time.Time
}

// Now returns a fixed time.
func (clock fixedClock) Now() time.Time {
	return clock.now
}

// testDraft returns a valid event draft.
func testDraft() domain.Draft {
	return domain.Draft{
		Key:           domain.EventForumsThreadCreated,
		SchemaVersion: 1,
		Producer:      domain.ProducerForums,
		AggregateType: "forum_thread",
		Payload:       map[string]any{"ok": true},
		Scopes:        []domain.Scope{{Type: domain.ScopeStaff}},
	}
}

// testNow returns a deterministic time.
func testNow() time.Time {
	return time.Unix(100, 0).UTC()
}
