package operations

import (
	"context"

	"github.com/realmkit/rk-backend/module/forums/domain"
	eventdomain "github.com/realmkit/rk-backend/pkg/events/domain"
	"github.com/realmkit/rk-backend/pkg/events/emitter"
)

// publishOperationEvent publishes one private forum operations event.
func (service Service) publishOperationEvent(
	ctx context.Context,
	key eventdomain.EventKey,
	payload any,
) error {
	return emitter.Publish(ctx, service.events, eventdomain.Draft{
		Key:           key,
		SchemaVersion: 1,
		Producer:      eventdomain.ProducerForums,
		AggregateType: "forum_stats",
		Payload:       payload,
		Scopes:        []eventdomain.Scope{{Type: eventdomain.ScopeSystem}},
	})
}

// driftPayload returns a safe drift report payload.
func driftPayload(report domain.CounterDriftReport) map[string]any {
	return map[string]any{
		"mismatch_count":  len(report.Mismatches),
		"mismatch_sample": report.Mismatches,
		"repaired":        report.Repaired,
	}
}
