package application

import (
	"context"

	"github.com/realmkit/rk-backend/pkg/cronjob/domain"
	eventdomain "github.com/realmkit/rk-backend/pkg/events/domain"
	"github.com/realmkit/rk-backend/pkg/events/emitter"
)

// publishRunEvent publishes one cron run event.
func (service Service) publishRunEvent(
	ctx context.Context,
	key eventdomain.EventKey,
	run domain.Run,
	result domain.Result,
) error {
	return emitter.Publish(ctx, service.events, eventdomain.Draft{
		Key:           key,
		SchemaVersion: 1,
		Producer:      eventdomain.ProducerCronjob,
		AggregateType: "cronjob_run",
		AggregateID:   emitter.UUID(run.ID),
		Payload: map[string]any{
			"id":              run.ID,
			"job_key":         run.JobKey,
			"status":          run.Status,
			"processed_count": result.ProcessedCount,
			"changed_count":   result.ChangedCount,
			"skipped_count":   result.SkippedCount,
			"worker_id":       run.WorkerID,
		},
		Scopes: []eventdomain.Scope{{Type: eventdomain.ScopeSystem}},
	})
}

// publishDefinitionEvent publishes one cron definition event.
func (service Service) publishDefinitionEvent(
	ctx context.Context,
	definition domain.Definition,
) error {
	return emitter.Publish(ctx, service.events, eventdomain.Draft{
		Key:           "cronjob.definition.updated",
		SchemaVersion: 1,
		Producer:      eventdomain.ProducerCronjob,
		AggregateType: "cronjob_definition",
		Payload: map[string]any{
			"key":           definition.Key,
			"name":          definition.Name,
			"enabled":       definition.Enabled,
			"schedule_kind": definition.ScheduleKind,
			"next_run_at":   definition.NextRunAt,
			"version":       definition.Version,
		},
		Scopes: []eventdomain.Scope{{Type: eventdomain.ScopeSystem}},
	})
}
