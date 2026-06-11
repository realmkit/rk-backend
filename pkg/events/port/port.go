package port

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/pkg/events/domain"
	"github.com/realmkit/rk-backend/pkg/pagination"
)

// Clock provides the current time.
type Clock interface {
	// Now returns the current time.
	Now() time.Time
}

// Repository stores and claims durable events.
type Repository interface {
	// Publish stores one event draft.
	Publish(ctx context.Context, draft domain.Draft, now time.Time) (domain.Event, error)

	// Get returns one event by id.
	Get(ctx context.Context, id uuid.UUID) (domain.Event, error)

	// List returns events matching filter.
	List(ctx context.Context, filter ListFilter, page pagination.Page) (pagination.Result[domain.Event], error)

	// Claim claims due events for worker.
	Claim(ctx context.Context, workerID string, limit int, now time.Time, lockUntil time.Time) ([]domain.Event, error)

	// MarkProcessed marks an event processed.
	MarkProcessed(ctx context.Context, id uuid.UUID, now time.Time) error

	// MarkFailed marks an event failed and schedules retry.
	MarkFailed(ctx context.Context, id uuid.UUID, message string, availableAt time.Time, now time.Time) error

	// MarkDead marks an event dead-lettered.
	MarkDead(ctx context.Context, id uuid.UUID, message string, now time.Time) error

	// Replay moves an event back to pending state.
	Replay(ctx context.Context, id uuid.UUID, now time.Time) error

	// Cancel cancels an event.
	Cancel(ctx context.Context, id uuid.UUID, now time.Time) error
}

// Broker sends dispatchable realtime envelopes.
type Broker interface {
	// Publish broadcasts one event.
	Publish(ctx context.Context, event domain.Event) error
}

// ScopeAuthorizer decides whether a principal can subscribe to a scope.
type ScopeAuthorizer interface {
	// CanSubscribe reports whether principal can subscribe to scope.
	CanSubscribe(ctx context.Context, principal Principal, scope domain.Scope) (bool, error)
}

// Principal contains subscription actor data.
type Principal struct {
	// UserID is the authenticated user ID.
	UserID uuid.UUID

	// Anonymous reports whether no authenticated user exists.
	Anonymous bool
}

// ListFilter filters event list queries.
type ListFilter struct {
	// Status filters by event status.
	Status domain.Status

	// Producer filters by producer.
	Producer domain.Producer

	// EventKey filters by key.
	EventKey domain.EventKey

	// AggregateType filters by aggregate type.
	AggregateType domain.AggregateType

	// AggregateID filters by aggregate id.
	AggregateID *uuid.UUID
}
