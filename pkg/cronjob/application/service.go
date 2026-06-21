package application

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/pkg/cronjob/domain"
	"github.com/realmkit/rk-backend/pkg/cronjob/port"
	eventdomain "github.com/realmkit/rk-backend/pkg/events/domain"
	"github.com/realmkit/rk-backend/pkg/events/emitter"
	"github.com/realmkit/rk-backend/pkg/pagination"
)

// Dependencies contains scheduler dependencies.
type Dependencies struct {
	// Repository stores cron state.
	Repository port.Repository

	// Clock provides the current time.
	Clock port.Clock

	// WorkerID identifies this scheduler worker.
	WorkerID string

	// LockDuration controls lease duration.
	LockDuration time.Duration

	// RunTimeout bounds one handler execution.
	RunTimeout time.Duration

	// Events publishes cron lifecycle events.
	Events emitter.Publisher
}

// Service coordinates cron jobs.
type Service struct {
	repository   port.Repository         // repository stores the repository value.
	clock        port.Clock              // clock stores the clock value.
	handlers     map[string]port.Handler // handlers stores the handlers value.
	workerID     string                  // workerID stores the worker i d value.
	lockDuration time.Duration           // lockDuration stores the lock duration value.
	runTimeout   time.Duration           // runTimeout stores the run timeout value.
	events       emitter.Publisher       // events stores the events value.
}

// NewService creates a cron service.
func NewService(deps Dependencies, handlers map[string]port.Handler) Service {
	service := Service{
		repository:   deps.Repository,
		clock:        deps.Clock,
		handlers:     handlers,
		workerID:     deps.WorkerID,
		lockDuration: deps.LockDuration,
		runTimeout:   deps.RunTimeout,
		events:       deps.Events,
	}
	if service.clock == nil {
		service.clock = systemClock{}
	}
	if service.workerID == "" {
		service.workerID = "realmkit-cron"
	}
	if service.lockDuration <= 0 {
		service.lockDuration = 5 * time.Minute
	}
	if service.runTimeout <= 0 {
		service.runTimeout = 5 * time.Minute
	}
	return service
}

// EnsureDefinitions upserts code-owned definitions.
func (service Service) EnsureDefinitions(ctx context.Context, definitions []domain.Definition) error {
	for _, definition := range definitions {
		if err := definition.Validate(); err != nil {
			return err
		}
		stored, err := service.repository.UpsertDefinition(ctx, definition)
		if err != nil {
			return err
		}
		if err := service.publishDefinitionEvent(ctx, stored); err != nil {
			return err
		}
	}
	return nil
}

// RunOnce claims and executes one due job.
func (service Service) RunOnce(ctx context.Context) (RunSummary, error) {
	now := service.clock.Now()
	definition, ok, err := service.repository.ClaimDue(ctx, service.workerID, now, now.Add(service.lockDuration))
	if err != nil {
		return RunSummary{}, err
	}
	if !ok {
		return RunSummary{}, port.ErrNoDueJob
	}
	return service.run(ctx, definition, domain.TriggerSchedule)
}

// Trigger runs one job manually.
func (service Service) Trigger(ctx context.Context, key string) (RunSummary, error) {
	definition, err := service.repository.Trigger(ctx, key)
	if err != nil {
		return RunSummary{}, err
	}
	return service.run(ctx, definition, domain.TriggerManual)
}

// ListDefinitions returns cron definitions.
func (service Service) ListDefinitions(ctx context.Context, page pagination.Page) (pagination.Result[domain.Definition], error) {
	return service.repository.ListDefinitions(ctx, page)
}

// GetDefinition returns one cron definition.
func (service Service) GetDefinition(ctx context.Context, key string) (domain.Definition, error) {
	return service.repository.GetDefinition(ctx, key)
}

// Pause disables one cron definition.
func (service Service) Pause(ctx context.Context, key string, expectedVersion uint64) error {
	if err := service.repository.Pause(ctx, key, expectedVersion); err != nil {
		return err
	}
	definition, err := service.repository.GetDefinition(ctx, key)
	if err != nil {
		return err
	}
	return service.publishDefinitionEvent(ctx, definition)
}

// Resume enables one cron definition.
func (service Service) Resume(ctx context.Context, key string, expectedVersion uint64) error {
	if err := service.repository.Resume(ctx, key, expectedVersion); err != nil {
		return err
	}
	definition, err := service.repository.GetDefinition(ctx, key)
	if err != nil {
		return err
	}
	return service.publishDefinitionEvent(ctx, definition)
}

// ListRuns returns run history.
func (service Service) ListRuns(ctx context.Context, key string, page pagination.Page) (pagination.Result[domain.Run], error) {
	return service.repository.ListRuns(ctx, key, page)
}

// RepairLocks clears stale locks.
func (service Service) RepairLocks(ctx context.Context) (int64, error) {
	return service.repository.RepairLocks(ctx, service.clock.Now())
}

// run executes one claimed or triggered job.
func (service Service) run(ctx context.Context, definition domain.Definition, trigger domain.TriggerType) (RunSummary, error) {
	handler, ok := service.handlers[definition.Key]
	if !ok || handler == nil {
		return RunSummary{JobKey: definition.Key}, port.ErrHandlerMissing
	}
	now := service.clock.Now()
	run, err := service.repository.StartRun(ctx, definition, trigger, service.workerID, now)
	if err != nil {
		return RunSummary{}, err
	}
	if err := service.publishRunEvent(ctx, "cronjob.run.started", run, domain.Result{}); err != nil {
		return RunSummary{}, err
	}
	runCtx, cancel := context.WithTimeout(ctx, service.runTimeout)
	defer cancel()
	result, err := handler.Run(runCtx, port.RunContext{
		RunID:        run.ID,
		JobKey:       run.JobKey,
		ScheduledFor: run.ScheduledFor,
		WorkerID:     service.workerID,
	})
	next := definition.NextAfter(service.clock.Now())
	if err != nil {
		failErr := service.repository.FailRun(ctx, run, err.Error(), service.clock.Now(), next)
		eventErr := service.publishRunEvent(ctx, "cronjob.run.failed", run, domain.Result{})
		return RunSummary{RunID: run.ID, JobKey: run.JobKey, Failed: true}, join(err, join(failErr, eventErr))
	}
	if err := service.repository.CompleteRun(ctx, run, result, service.clock.Now(), next); err != nil {
		return RunSummary{}, err
	}
	eventKey := eventdomain.EventKey("cronjob.run.completed")
	if result.SkippedCount > 0 && result.ProcessedCount == 0 && result.ChangedCount == 0 {
		eventKey = "cronjob.run.skipped"
	}
	if err := service.publishRunEvent(ctx, eventKey, run, result); err != nil {
		return RunSummary{}, err
	}
	return RunSummary{RunID: run.ID, JobKey: run.JobKey, ProcessedCount: result.ProcessedCount}, nil
}

// RunSummary summarizes one run.
type RunSummary struct {
	// RunID is the run identifier.
	RunID uuid.UUID

	// JobKey is the job key.
	JobKey string

	// ProcessedCount is the processed item count.
	ProcessedCount int64

	// Failed reports whether the run failed.
	Failed bool
}

// systemClock uses UTC wall clock time.
type systemClock struct{}

// Now returns the current time.
func (systemClock) Now() time.Time {
	return time.Now().UTC()
}

// join joins errors without importing errors in the main flow.
func join(first error, second error) error {
	if second != nil {
		return errors.Join(first, second)
	}
	return first
}
