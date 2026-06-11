package postgres

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/pkg/events/domain"
)

// eventFromModel maps persistence rows to a domain event.
func eventFromModel(model EventModel, scopes []ScopeModel) domain.Event {
	return domain.Event{
		ID:             model.ID,
		Key:            domain.EventKey(model.EventKey),
		SchemaVersion:  model.SchemaVersion,
		Producer:       domain.Producer(model.Producer),
		AggregateType:  domain.AggregateType(model.AggregateType),
		AggregateID:    model.AggregateID,
		Payload:        json.RawMessage(model.PayloadJSON),
		Metadata:       json.RawMessage(model.MetadataJSON),
		ActorUserID:    model.ActorUserID,
		RequestID:      model.RequestID,
		CorrelationID:  model.CorrelationID,
		IdempotencyKey: model.IdempotencyKey,
		DedupeKey:      valueFromPointer(model.DedupeKey),
		Scopes:         scopesFromModels(scopes),
		OccurredAt:     model.OccurredAt,
		AvailableAt:    model.AvailableAt,
		Status:         domain.Status(model.Status),
		AttemptCount:   model.AttemptCount,
		LockedBy:       model.LockedBy,
		LockedUntil:    model.LockedUntil,
		ProcessedAt:    model.ProcessedAt,
		DeadAt:         model.DeadAt,
		LastError:      model.LastError,
		CreatedAt:      model.CreatedAt,
		UpdatedAt:      model.UpdatedAt,
	}
}

// modelFromDraft maps a draft to persistence model.
func modelFromDraft(draft domain.Draft, now time.Time) (EventModel, []ScopeModel, error) {
	payload, err := json.Marshal(draft.Payload)
	if err != nil {
		return EventModel{}, nil, err
	}
	metadata, err := json.Marshal(draft.Metadata)
	if err != nil {
		return EventModel{}, nil, err
	}
	id := uuid.New()
	availableAt := draft.AvailableAt
	if availableAt.IsZero() {
		availableAt = now
	}
	model := EventModel{
		ID:             id,
		EventKey:       string(draft.Key),
		SchemaVersion:  draft.SchemaVersion,
		Producer:       string(draft.Producer),
		AggregateType:  string(draft.AggregateType),
		AggregateID:    draft.AggregateID,
		PayloadJSON:    string(payload),
		MetadataJSON:   string(metadata),
		ActorUserID:    draft.ActorUserID,
		RequestID:      draft.RequestID,
		CorrelationID:  draft.CorrelationID,
		IdempotencyKey: draft.IdempotencyKey,
		DedupeKey:      pointerFromValue(draft.DedupeKey),
		OccurredAt:     now,
		AvailableAt:    availableAt,
		Status:         string(domain.StatusPending),
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	return model, scopeModelsFromDraft(id, draft.Scopes, now), nil
}

// scopeModelsFromDraft maps draft scopes.
func scopeModelsFromDraft(eventID uuid.UUID, scopes []domain.Scope, now time.Time) []ScopeModel {
	models := make([]ScopeModel, 0, len(scopes))
	for _, scope := range scopes {
		models = append(models, ScopeModel{
			ID:         uuid.New(),
			EventID:    eventID,
			ScopeType:  string(scope.Type),
			ScopeID:    scope.ID,
			Permission: scope.Permission,
			CreatedAt:  now,
		})
	}
	return models
}

// scopesFromModels maps persisted scopes.
func scopesFromModels(models []ScopeModel) []domain.Scope {
	scopes := make([]domain.Scope, 0, len(models))
	for _, model := range models {
		scopes = append(scopes, domain.Scope{
			Type:       domain.ScopeType(model.ScopeType),
			ID:         model.ScopeID,
			Permission: model.Permission,
		})
	}
	return scopes
}

// pointerFromValue returns a pointer for non-empty values.
func pointerFromValue(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}

// valueFromPointer returns pointer value or empty string.
func valueFromPointer(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
