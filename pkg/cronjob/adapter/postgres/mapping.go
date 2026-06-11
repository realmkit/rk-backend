package postgres

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/pkg/cronjob/domain"
)

// definitionToModel maps domain definition to model.
func definitionToModel(definition domain.Definition, now time.Time) DefinitionModel {
	createdAt := definition.CreatedAt
	if createdAt.IsZero() {
		createdAt = now
	}
	version := definition.Version
	if version == 0 {
		version = 1
	}
	return DefinitionModel{
		Key:                definition.Key,
		Name:               definition.Name,
		Description:        definition.Description,
		ScheduleKind:       string(definition.ScheduleKind),
		ScheduleExpression: definition.ScheduleExpression,
		Enabled:            definition.Enabled,
		ConcurrencyPolicy:  string(definition.ConcurrencyPolicy),
		NextRunAt:          definition.NextRunAt,
		LastRunAt:          definition.LastRunAt,
		LastStatus:         string(definition.LastStatus),
		LockedBy:           definition.LockedBy,
		LockedUntil:        definition.LockedUntil,
		Version:            version,
		CreatedAt:          createdAt,
		UpdatedAt:          now,
	}
}

// definitionFromModel maps model to domain definition.
func definitionFromModel(model DefinitionModel) domain.Definition {
	return domain.Definition{
		Key:                model.Key,
		Name:               model.Name,
		Description:        model.Description,
		ScheduleKind:       domain.ScheduleKind(model.ScheduleKind),
		ScheduleExpression: model.ScheduleExpression,
		Enabled:            model.Enabled,
		ConcurrencyPolicy:  domain.ConcurrencyPolicy(model.ConcurrencyPolicy),
		NextRunAt:          model.NextRunAt,
		LastRunAt:          model.LastRunAt,
		LastStatus:         domain.RunStatus(model.LastStatus),
		LockedBy:           model.LockedBy,
		LockedUntil:        model.LockedUntil,
		Version:            model.Version,
		CreatedAt:          model.CreatedAt,
		UpdatedAt:          model.UpdatedAt,
	}
}

// runFromModel maps model to domain run.
func runFromModel(model RunModel) domain.Run {
	return domain.Run{
		ID:                model.ID,
		JobKey:            model.JobKey,
		Status:            domain.RunStatus(model.Status),
		ScheduledFor:      model.ScheduledFor,
		StartedAt:         model.StartedAt,
		FinishedAt:        model.FinishedAt,
		DurationMS:        model.DurationMS,
		TriggerType:       domain.TriggerType(model.TriggerType),
		TriggeredByUserID: model.TriggeredByUserID,
		WorkerID:          model.WorkerID,
		ProcessedCount:    model.ProcessedCount,
		ChangedCount:      model.ChangedCount,
		SkippedCount:      model.SkippedCount,
		Metadata:          json.RawMessage(model.MetadataJSON),
		Error:             model.Error,
		CreatedAt:         model.CreatedAt,
		UpdatedAt:         model.UpdatedAt,
	}
}

// runningModel creates a running run model.
func runningModel(definition domain.Definition, trigger domain.TriggerType, workerID string, now time.Time) RunModel {
	return RunModel{
		ID:           uuid.New(),
		JobKey:       definition.Key,
		Status:       string(domain.RunRunning),
		ScheduledFor: definition.NextRunAt,
		StartedAt:    now,
		TriggerType:  string(trigger),
		WorkerID:     workerID,
		MetadataJSON: "{}",
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}
