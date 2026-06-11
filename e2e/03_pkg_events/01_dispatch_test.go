// Package events_e2e verifies durable event infrastructure.
package events_e2e

import (
	"context"
	"testing"

	"github.com/realmkit/rk-backend/e2e/harness"
	eventshttp "github.com/realmkit/rk-backend/pkg/events/adapter/http"
	eventspostgres "github.com/realmkit/rk-backend/pkg/events/adapter/postgres"
	eventsapplication "github.com/realmkit/rk-backend/pkg/events/application"
	"github.com/realmkit/rk-backend/pkg/events/domain"
)

// TestEventsPublishAndDispatch verifies outbox persistence and dispatch.
func TestEventsPublishAndDispatch(t *testing.T) {
	steps := harness.NewSteps(t)
	steps.Log("start ecosystem with migrated event tables")
	ecosystem := harness.New(t)
	repository := eventspostgres.NewRepository(ecosystem.Database.Store)
	service := eventsapplication.NewService(eventsapplication.Dependencies{
		Repository: repository,
		Broker:     eventshttp.NewHub(),
	})

	steps.Log("publish global event")
	event, err := service.Publish(context.Background(), domain.Draft{
		Key:           domain.EventKey("e2e.event.created"),
		SchemaVersion: 1,
		Producer:      domain.Producer("e2e"),
		AggregateType: domain.AggregateType("e2e_event"),
		Payload:       map[string]string{"ok": "true"},
		Scopes:        []domain.Scope{{Type: domain.ScopeGlobal}},
	})
	if err != nil {
		t.Fatalf("Publish() error = %v", err)
	}
	if event.Status != domain.StatusPending {
		t.Fatalf("event status = %s, want pending", event.Status)
	}

	steps.Log("dispatch pending event")
	result, err := service.DispatchOnce(context.Background(), "e2e-worker")
	if err != nil {
		t.Fatalf("DispatchOnce() error = %v", err)
	}
	if result.Claimed != 1 || result.Processed != 1 || result.Failed != 0 {
		t.Fatalf("dispatch result = %+v, want one processed", result)
	}

	steps.Log("verify event is processed")
	stored, err := service.Get(context.Background(), event.ID)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if stored.Status != domain.StatusProcessed || stored.ProcessedAt == nil {
		t.Fatalf("stored event = %+v, want processed", stored)
	}
}
