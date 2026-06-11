package application

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/pkg/events/domain"
	"github.com/realmkit/rk-backend/pkg/events/port"
	"github.com/realmkit/rk-backend/pkg/pagination"
)

// Dependencies contains event service dependencies.
type Dependencies struct {
	// Repository stores durable events.
	Repository port.Repository

	// Broker broadcasts events after claim.
	Broker port.Broker

	// Clock provides the current time.
	Clock port.Clock

	// BatchSize controls dispatch batch size.
	BatchSize int

	// MaxAttempts controls when events become dead.
	MaxAttempts int

	// LockDuration controls claim lease duration.
	LockDuration time.Duration

	// RetryDelay controls failed event retry delay.
	RetryDelay time.Duration
}

// Service implements event use cases.
type Service struct {
	repository   port.Repository
	broker       port.Broker
	clock        port.Clock
	batchSize    int
	maxAttempts  int
	lockDuration time.Duration
	retryDelay   time.Duration
}

// NewService creates an event service.
func NewService(deps Dependencies) Service {
	service := Service{
		repository:   deps.Repository,
		broker:       deps.Broker,
		clock:        deps.Clock,
		batchSize:    deps.BatchSize,
		maxAttempts:  deps.MaxAttempts,
		lockDuration: deps.LockDuration,
		retryDelay:   deps.RetryDelay,
	}
	if service.clock == nil {
		service.clock = systemClock{}
	}
	if service.batchSize <= 0 {
		service.batchSize = 50
	}
	if service.maxAttempts <= 0 {
		service.maxAttempts = 5
	}
	if service.lockDuration <= 0 {
		service.lockDuration = time.Minute
	}
	if service.retryDelay <= 0 {
		service.retryDelay = time.Minute
	}
	return service
}

// Publish stores one event draft.
func (service Service) Publish(ctx context.Context, draft domain.Draft) (domain.Event, error) {
	if err := draft.Validate(); err != nil {
		return domain.Event{}, err
	}
	return service.repository.Publish(ctx, draft, service.clock.Now())
}

// Get returns one event.
func (service Service) Get(ctx context.Context, id uuid.UUID) (domain.Event, error) {
	return service.repository.Get(ctx, id)
}

// List returns matching events.
func (service Service) List(ctx context.Context, filter port.ListFilter, page pagination.Page) (pagination.Result[domain.Event], error) {
	return service.repository.List(ctx, filter, page)
}

// DispatchOnce dispatches one due batch.
func (service Service) DispatchOnce(ctx context.Context, workerID string) (DispatchResult, error) {
	now := service.clock.Now()
	events, err := service.repository.Claim(ctx, workerID, service.batchSize, now, now.Add(service.lockDuration))
	if err != nil {
		return DispatchResult{}, err
	}
	result := DispatchResult{Claimed: len(events)}
	for _, event := range events {
		if err := service.dispatch(ctx, event); err != nil {
			result.Failed++
			return result, err
		}
		result.Processed++
	}
	return result, nil
}

// Replay moves one event back to pending state.
func (service Service) Replay(ctx context.Context, id uuid.UUID) error {
	return service.repository.Replay(ctx, id, service.clock.Now())
}

// Cancel cancels one event.
func (service Service) Cancel(ctx context.Context, id uuid.UUID) error {
	return service.repository.Cancel(ctx, id, service.clock.Now())
}

// dispatch broadcasts and marks one event.
func (service Service) dispatch(ctx context.Context, event domain.Event) error {
	now := service.clock.Now()
	if service.broker != nil {
		if err := service.broker.Publish(ctx, event); err != nil {
			return errors.Join(err, service.fail(ctx, event, err.Error(), now))
		}
	}
	return service.repository.MarkProcessed(ctx, event.ID, now)
}

// fail records dispatch failure or dead-letter state.
func (service Service) fail(ctx context.Context, event domain.Event, message string, now time.Time) error {
	if event.AttemptCount+1 >= service.maxAttempts {
		return service.repository.MarkDead(ctx, event.ID, message, now)
	}
	return service.repository.MarkFailed(ctx, event.ID, message, now.Add(service.retryDelay), now)
}

// DispatchResult summarizes one dispatch batch.
type DispatchResult struct {
	// Claimed is the number of claimed events.
	Claimed int

	// Processed is the number of processed events.
	Processed int

	// Failed is the number of failed events.
	Failed int
}

// systemClock uses wall clock time.
type systemClock struct{}

// Now returns the current UTC time.
func (systemClock) Now() time.Time {
	return time.Now().UTC()
}
