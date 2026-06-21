package application

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/pkg/cronjob/domain"
	"github.com/realmkit/rk-backend/pkg/cronjob/port"
	"github.com/realmkit/rk-backend/pkg/pagination"
)

// benchmarkCronSummary stores the cron run benchmark result.
var benchmarkCronSummary RunSummary

// BenchmarkRunOnce measures the scheduler claim, handler, completion, and event-free run path.
func BenchmarkRunOnce(b *testing.B) {
	repository := benchmarkCronRepository{
		definition: testCronDefinition(),
		runID:      uuid.New(),
	}
	service := NewService(Dependencies{Repository: repository, Clock: cronClock{now: testCronNow()}}, map[string]port.Handler{
		domain.JobEventsDispatchPending: port.HandlerFunc(func(context.Context, port.RunContext) (domain.Result, error) {
			return domain.Result{ProcessedCount: 3, ChangedCount: 2}, nil
		}),
	})
	ctx := context.Background()

	b.ReportAllocs()
	for index := 0; index < b.N; index++ {
		summary, err := service.RunOnce(ctx)
		if err != nil {
			b.Fatalf("RunOnce() error = %v", err)
		}
		benchmarkCronSummary = summary
	}
}

// benchmarkCronRepository returns deterministic cron rows without retaining run history.
type benchmarkCronRepository struct {
	definition domain.Definition
	runID      uuid.UUID
}

// UpsertDefinition stores the benchmark definition.
func (repository benchmarkCronRepository) UpsertDefinition(
	context.Context,
	domain.Definition,
) (domain.Definition, error) {
	return repository.definition, nil
}

// GetDefinition returns the benchmark definition.
func (repository benchmarkCronRepository) GetDefinition(context.Context, string) (domain.Definition, error) {
	return repository.definition, nil
}

// ListDefinitions returns the benchmark definition.
func (repository benchmarkCronRepository) ListDefinitions(
	context.Context,
	pagination.Page,
) (pagination.Result[domain.Definition], error) {
	return pagination.Result[domain.Definition]{Items: []domain.Definition{repository.definition}}, nil
}

// ClaimDue returns the benchmark definition as due.
func (repository benchmarkCronRepository) ClaimDue(
	context.Context,
	string,
	time.Time,
	time.Time,
) (domain.Definition, bool, error) {
	return repository.definition, true, nil
}

// StartRun returns a deterministic benchmark run.
func (repository benchmarkCronRepository) StartRun(
	_ context.Context,
	definition domain.Definition,
	trigger domain.TriggerType,
	workerID string,
	now time.Time,
) (domain.Run, error) {
	return domain.Run{
		ID:          repository.runID,
		JobKey:      definition.Key,
		Status:      domain.RunRunning,
		TriggerType: trigger,
		WorkerID:    workerID,
		StartedAt:   now,
	}, nil
}

// CompleteRun records no benchmark state.
func (repository benchmarkCronRepository) CompleteRun(
	context.Context,
	domain.Run,
	domain.Result,
	time.Time,
	*time.Time,
) error {
	return nil
}

// FailRun records no benchmark state.
func (repository benchmarkCronRepository) FailRun(context.Context, domain.Run, string, time.Time, *time.Time) error {
	return nil
}

// Trigger returns the benchmark definition.
func (repository benchmarkCronRepository) Trigger(context.Context, string) (domain.Definition, error) {
	return repository.definition, nil
}

// Pause records no benchmark state.
func (repository benchmarkCronRepository) Pause(context.Context, string, uint64) error {
	return nil
}

// Resume records no benchmark state.
func (repository benchmarkCronRepository) Resume(context.Context, string, uint64) error {
	return nil
}

// ListRuns returns no benchmark run history.
func (repository benchmarkCronRepository) ListRuns(
	context.Context,
	string,
	pagination.Page,
) (pagination.Result[domain.Run], error) {
	return pagination.Result[domain.Run]{}, nil
}

// RepairLocks reports no repaired benchmark locks.
func (repository benchmarkCronRepository) RepairLocks(context.Context, time.Time) (int64, error) {
	return 0, nil
}
