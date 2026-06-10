package port

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/niflaot/gamehub-go/pkg/cronjob/domain"
	"github.com/niflaot/gamehub-go/pkg/pagination"
)

// Clock provides current time.
type Clock interface {
	// Now returns current time.
	Now() time.Time
}

// Handler runs one cron job.
type Handler interface {
	// Run executes one cron job.
	Run(ctx context.Context, run RunContext) (domain.Result, error)
}

// HandlerFunc adapts a function to Handler.
type HandlerFunc func(ctx context.Context, run RunContext) (domain.Result, error)

// Run executes one cron job.
func (fn HandlerFunc) Run(ctx context.Context, run RunContext) (domain.Result, error) {
	return fn(ctx, run)
}

// RunContext describes one run.
type RunContext struct {
	// RunID is the run identifier.
	RunID uuid.UUID

	// JobKey is the job key.
	JobKey string

	// ScheduledFor is the scheduled due time.
	ScheduledFor *time.Time

	// WorkerID is the executing worker.
	WorkerID string
}

// Repository stores cron definitions and runs.
type Repository interface {
	// UpsertDefinition inserts or updates a definition.
	UpsertDefinition(ctx context.Context, definition domain.Definition) (domain.Definition, error)

	// GetDefinition returns one definition.
	GetDefinition(ctx context.Context, key string) (domain.Definition, error)

	// ListDefinitions returns all definitions.
	ListDefinitions(ctx context.Context, page pagination.Page) (pagination.Result[domain.Definition], error)

	// ClaimDue claims one due definition.
	ClaimDue(ctx context.Context, workerID string, now time.Time, lockUntil time.Time) (domain.Definition, bool, error)

	// StartRun creates a running run record.
	StartRun(ctx context.Context, definition domain.Definition, trigger domain.TriggerType, workerID string, now time.Time) (domain.Run, error)

	// CompleteRun marks a run complete and advances definition.
	CompleteRun(ctx context.Context, run domain.Run, result domain.Result, now time.Time, nextRunAt *time.Time) error

	// FailRun marks a run failed and advances definition.
	FailRun(ctx context.Context, run domain.Run, message string, now time.Time, nextRunAt *time.Time) error

	// Trigger returns a definition for manual execution.
	Trigger(ctx context.Context, key string) (domain.Definition, error)

	// Pause disables one definition.
	Pause(ctx context.Context, key string, expectedVersion uint64) error

	// Resume enables one definition.
	Resume(ctx context.Context, key string, expectedVersion uint64) error

	// ListRuns returns runs for one job.
	ListRuns(ctx context.Context, key string, page pagination.Page) (pagination.Result[domain.Run], error)

	// RepairLocks clears stale locks.
	RepairLocks(ctx context.Context, now time.Time) (int64, error)
}
