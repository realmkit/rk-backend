package application

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/pkg/events/domain"
	"github.com/realmkit/rk-backend/pkg/events/port"
	"github.com/realmkit/rk-backend/pkg/pagination"
)

// benchmarkEvent stores the publish benchmark result.
var benchmarkEvent domain.Event

// benchmarkDispatchResult stores the dispatch benchmark result.
var benchmarkDispatchResult DispatchResult

// BenchmarkPublish measures event draft validation and repository publish orchestration.
func BenchmarkPublish(b *testing.B) {
	service := NewService(Dependencies{Repository: benchmarkEventRepository{}, Clock: fixedClock{now: testNow()}})
	draft := testDraft()
	ctx := context.Background()

	b.ReportAllocs()
	for index := 0; index < b.N; index++ {
		event, err := service.Publish(ctx, draft)
		if err != nil {
			b.Fatalf("Publish() error = %v", err)
		}
		benchmarkEvent = event
	}
}

// BenchmarkDispatchOnce measures batch dispatch bookkeeping without broker I/O.
func BenchmarkDispatchOnce(b *testing.B) {
	repository := benchmarkEventRepository{events: benchmarkEvents(32)}
	service := NewService(Dependencies{
		Repository: repository,
		Broker:     &recordBroker{},
		Clock:      fixedClock{now: testNow()},
		BatchSize:  32,
	})
	ctx := context.Background()

	b.ReportAllocs()
	for index := 0; index < b.N; index++ {
		result, err := service.DispatchOnce(ctx, "worker")
		if err != nil {
			b.Fatalf("DispatchOnce() error = %v", err)
		}
		benchmarkDispatchResult = result
	}
}

// benchmarkEventRepository provides deterministic events without retaining state.
type benchmarkEventRepository struct {
	events []domain.Event
}

// Publish returns one deterministic benchmark event.
func (repository benchmarkEventRepository) Publish(_ context.Context, draft domain.Draft, now time.Time) (domain.Event, error) {
	return domain.Event{
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
	}, nil
}

// Get returns no benchmark event.
func (repository benchmarkEventRepository) Get(context.Context, uuid.UUID) (domain.Event, error) {
	return domain.Event{}, port.ErrNotFound
}

// List returns the configured benchmark events.
func (repository benchmarkEventRepository) List(
	context.Context,
	port.ListFilter,
	pagination.Page,
) (pagination.Result[domain.Event], error) {
	return pagination.Result[domain.Event]{Items: repository.events}, nil
}

// Claim returns up to limit configured benchmark events.
func (repository benchmarkEventRepository) Claim(
	context.Context,
	string,
	int,
	time.Time,
	time.Time,
) ([]domain.Event, error) {
	return repository.events, nil
}

// MarkProcessed records no benchmark state.
func (repository benchmarkEventRepository) MarkProcessed(context.Context, uuid.UUID, time.Time) error {
	return nil
}

// MarkFailed records no benchmark state.
func (repository benchmarkEventRepository) MarkFailed(context.Context, uuid.UUID, string, time.Time, time.Time) error {
	return nil
}

// MarkDead records no benchmark state.
func (repository benchmarkEventRepository) MarkDead(context.Context, uuid.UUID, string, time.Time) error {
	return nil
}

// Replay records no benchmark state.
func (repository benchmarkEventRepository) Replay(context.Context, uuid.UUID, time.Time) error {
	return nil
}

// Cancel records no benchmark state.
func (repository benchmarkEventRepository) Cancel(context.Context, uuid.UUID, time.Time) error {
	return nil
}

// benchmarkEvents returns dispatchable benchmark events.
func benchmarkEvents(count int) []domain.Event {
	events := make([]domain.Event, count)
	for index := range events {
		events[index] = domain.Event{
			ID:            uuid.New(),
			Key:           domain.EventForumsThreadCreated,
			SchemaVersion: 1,
			Producer:      domain.ProducerForums,
			AggregateType: "forum_thread",
			Status:        domain.StatusPending,
			Scopes:        []domain.Scope{{Type: domain.ScopeStaff}},
		}
	}
	return events
}
