package application

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/pkg/cronjob/domain"
	"github.com/realmkit/rk-backend/pkg/pagination"
)

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
func (repo *memoryCron) StartRun(
	_ context.Context,
	definition domain.Definition,
	trigger domain.TriggerType,
	workerID string,
	now time.Time,
) (domain.Run, error) {
	run := domain.Run{
		ID:          uuid.New(),
		JobKey:      definition.Key,
		Status:      domain.RunRunning,
		TriggerType: trigger,
		WorkerID:    workerID,
		StartedAt:   now,
	}
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
