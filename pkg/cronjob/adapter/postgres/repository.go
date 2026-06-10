package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/niflaot/gamehub-go/pkg/cronjob/domain"
	"github.com/niflaot/gamehub-go/pkg/cronjob/port"
	"github.com/niflaot/gamehub-go/pkg/orm"
	"github.com/niflaot/gamehub-go/pkg/pagination"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Repository stores cron state.
type Repository struct {
	store orm.Store
}

// NewRepository creates a cron repository.
func NewRepository(store orm.Store) Repository {
	return Repository{store: store}
}

// UpsertDefinition inserts or updates a definition.
func (repository Repository) UpsertDefinition(ctx context.Context, definition domain.Definition) (domain.Definition, error) {
	now := time.Now().UTC()
	model := definitionToModel(definition, now)
	err := repository.store.DB(ctx).Save(&model).Error
	return definitionFromModel(model), translate(err)
}

// GetDefinition returns one definition.
func (repository Repository) GetDefinition(ctx context.Context, key string) (domain.Definition, error) {
	var model DefinitionModel
	err := repository.store.DB(ctx).Where("key = ?", key).First(&model).Error
	return definitionFromModel(model), translate(err)
}

// ListDefinitions returns all definitions.
func (repository Repository) ListDefinitions(ctx context.Context, page pagination.Page) (pagination.Result[domain.Definition], error) {
	var models []DefinitionModel
	err := repository.store.DB(ctx).Order("key ASC").Limit(page.Limit).Find(&models).Error
	if err != nil {
		return pagination.Result[domain.Definition]{}, translate(err)
	}
	definitions := make([]domain.Definition, 0, len(models))
	for _, model := range models {
		definitions = append(definitions, definitionFromModel(model))
	}
	return pagination.Result[domain.Definition]{Items: definitions}, nil
}

// ClaimDue claims one due definition.
func (repository Repository) ClaimDue(ctx context.Context, workerID string, now time.Time, lockUntil time.Time) (domain.Definition, bool, error) {
	var model DefinitionModel
	db := repository.store.DB(ctx)
	err := db.Clauses(clause.Locking{Strength: "UPDATE", Options: "SKIP LOCKED"}).
		Where("enabled = ? AND next_run_at IS NOT NULL AND next_run_at <= ?", true, now).
		Where("(locked_until IS NULL OR locked_until <= ?)", now).
		Order("next_run_at ASC, key ASC").
		First(&model).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return domain.Definition{}, false, nil
	}
	if err != nil {
		return domain.Definition{}, false, translate(err)
	}
	updates := map[string]any{"locked_by": workerID, "locked_until": lockUntil, "updated_at": now}
	if err := db.Model(&DefinitionModel{}).Where("key = ?", model.Key).Updates(updates).Error; err != nil {
		return domain.Definition{}, false, translate(err)
	}
	model.LockedBy = workerID
	model.LockedUntil = &lockUntil
	return definitionFromModel(model), true, nil
}

// StartRun creates a running run.
func (repository Repository) StartRun(ctx context.Context, definition domain.Definition, trigger domain.TriggerType, workerID string, now time.Time) (domain.Run, error) {
	model := runningModel(definition, trigger, workerID, now)
	err := repository.store.DB(ctx).Create(&model).Error
	return runFromModel(model), translate(err)
}

// CompleteRun marks a run complete and advances definition.
func (repository Repository) CompleteRun(ctx context.Context, run domain.Run, result domain.Result, now time.Time, nextRunAt *time.Time) error {
	metadata, _ := json.Marshal(result.Metadata)
	runUpdates := map[string]any{
		"status": domain.RunSucceeded, "finished_at": now, "duration_ms": now.Sub(run.StartedAt).Milliseconds(),
		"processed_count": result.ProcessedCount, "changed_count": result.ChangedCount,
		"skipped_count": result.SkippedCount, "metadata_json": string(metadata), "updated_at": now,
	}
	if err := repository.store.DB(ctx).Model(&RunModel{}).Where("id = ?", run.ID).Updates(runUpdates).Error; err != nil {
		return translate(err)
	}
	return repository.releaseDefinition(ctx, run.JobKey, domain.RunSucceeded, now, nextRunAt)
}

// FailRun marks a run failed and advances definition.
func (repository Repository) FailRun(ctx context.Context, run domain.Run, message string, now time.Time, nextRunAt *time.Time) error {
	updates := map[string]any{
		"status": domain.RunFailed, "finished_at": now, "duration_ms": now.Sub(run.StartedAt).Milliseconds(),
		"error": message, "updated_at": now,
	}
	if err := repository.store.DB(ctx).Model(&RunModel{}).Where("id = ?", run.ID).Updates(updates).Error; err != nil {
		return translate(err)
	}
	return repository.releaseDefinition(ctx, run.JobKey, domain.RunFailed, now, nextRunAt)
}

// Trigger returns a definition for manual execution.
func (repository Repository) Trigger(ctx context.Context, key string) (domain.Definition, error) {
	return repository.GetDefinition(ctx, key)
}

// Pause disables one definition.
func (repository Repository) Pause(ctx context.Context, key string, expectedVersion uint64) error {
	return repository.setEnabled(ctx, key, expectedVersion, false)
}

// Resume enables one definition.
func (repository Repository) Resume(ctx context.Context, key string, expectedVersion uint64) error {
	return repository.setEnabled(ctx, key, expectedVersion, true)
}

// ListRuns returns runs for one job.
func (repository Repository) ListRuns(ctx context.Context, key string, page pagination.Page) (pagination.Result[domain.Run], error) {
	var models []RunModel
	err := repository.store.DB(ctx).Where("job_key = ?", key).Order("started_at DESC, id DESC").Limit(page.Limit).Find(&models).Error
	if err != nil {
		return pagination.Result[domain.Run]{}, translate(err)
	}
	runs := make([]domain.Run, 0, len(models))
	for _, model := range models {
		runs = append(runs, runFromModel(model))
	}
	return pagination.Result[domain.Run]{Items: runs}, nil
}

// RepairLocks clears stale locks.
func (repository Repository) RepairLocks(ctx context.Context, now time.Time) (int64, error) {
	result := repository.store.DB(ctx).Model(&DefinitionModel{}).
		Where("locked_until IS NOT NULL AND locked_until <= ?", now).
		Updates(map[string]any{"locked_by": "", "locked_until": nil, "updated_at": now})
	return result.RowsAffected, translate(result.Error)
}

// releaseDefinition updates definition state after a run.
func (repository Repository) releaseDefinition(ctx context.Context, key string, status domain.RunStatus, now time.Time, nextRunAt *time.Time) error {
	updates := map[string]any{"last_run_at": now, "last_status": status, "next_run_at": nextRunAt, "locked_by": "", "locked_until": nil, "updated_at": now}
	return translate(repository.store.DB(ctx).Model(&DefinitionModel{}).Where("key = ?", key).Updates(updates).Error)
}

// setEnabled changes enabled state with optimistic version.
func (repository Repository) setEnabled(ctx context.Context, key string, version uint64, enabled bool) error {
	result := repository.store.DB(ctx).Model(&DefinitionModel{}).
		Where("key = ? AND version = ?", key, version).
		Updates(map[string]any{"enabled": enabled, "version": gorm.Expr("version + 1"), "updated_at": time.Now().UTC()})
	if result.Error != nil {
		return translate(result.Error)
	}
	if result.RowsAffected == 0 {
		return port.ErrPreconditionFailed
	}
	return nil
}

// translate maps persistence errors.
func translate(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(orm.TranslateError(err), orm.ErrNotFound):
		return port.ErrNotFound
	default:
		return err
	}
}
