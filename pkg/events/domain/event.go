package domain

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Event is one durable outbox event.
type Event struct {
	// ID is the event identifier.
	ID uuid.UUID `json:"id"`

	// Key is the stable event key.
	Key EventKey `json:"key"`

	// SchemaVersion is the payload schema version.
	SchemaVersion int `json:"schema_version"`

	// Producer is the publishing module or package.
	Producer Producer `json:"producer"`

	// AggregateType is the affected aggregate type.
	AggregateType AggregateType `json:"aggregate_type"`

	// AggregateID is the affected aggregate identifier.
	AggregateID *uuid.UUID `json:"aggregate_id,omitempty"`

	// Payload contains the sanitized event payload.
	Payload json.RawMessage `json:"payload_json"`

	// Metadata contains safe operational metadata.
	Metadata json.RawMessage `json:"metadata_json"`

	// ActorUserID is the user that caused the event when present.
	ActorUserID *uuid.UUID `json:"actor_user_id,omitempty"`

	// RequestID is the request identifier.
	RequestID string `json:"request_id,omitempty"`

	// CorrelationID is the correlation identifier.
	CorrelationID string `json:"correlation_id,omitempty"`

	// IdempotencyKey is the request idempotency key when present.
	IdempotencyKey string `json:"idempotency_key,omitempty"`

	// DedupeKey prevents duplicate durable events.
	DedupeKey string `json:"dedupe_key,omitempty"`

	// Scopes contains event audience scopes.
	Scopes []Scope `json:"scopes"`

	// OccurredAt is when the fact happened.
	OccurredAt time.Time `json:"occurred_at"`

	// AvailableAt is when dispatch can claim the event.
	AvailableAt time.Time `json:"available_at"`

	// Status is the dispatch status.
	Status Status `json:"status"`

	// AttemptCount is the dispatch attempt count.
	AttemptCount int `json:"attempt_count"`

	// LockedBy is the current worker lock owner.
	LockedBy string `json:"locked_by,omitempty"`

	// LockedUntil is when the current worker lock expires.
	LockedUntil *time.Time `json:"locked_until,omitempty"`

	// ProcessedAt is when dispatch succeeded.
	ProcessedAt *time.Time `json:"processed_at,omitempty"`

	// DeadAt is when retries were exhausted.
	DeadAt *time.Time `json:"dead_at,omitempty"`

	// LastError is the latest dispatch failure.
	LastError string `json:"last_error,omitempty"`

	// CreatedAt is when the row was created.
	CreatedAt time.Time `json:"created_at"`

	// UpdatedAt is when the row last changed.
	UpdatedAt time.Time `json:"updated_at"`
}

// Draft is a publishable event before persistence.
type Draft struct {
	// Key is the stable event key.
	Key EventKey

	// SchemaVersion is the payload schema version.
	SchemaVersion int

	// Producer is the publishing module or package.
	Producer Producer

	// AggregateType is the affected aggregate type.
	AggregateType AggregateType

	// AggregateID is the affected aggregate identifier.
	AggregateID *uuid.UUID

	// Payload is marshalled into the event payload JSON.
	Payload any

	// Metadata is marshalled into metadata JSON.
	Metadata any

	// ActorUserID is the user that caused the event when present.
	ActorUserID *uuid.UUID

	// RequestID is the request identifier.
	RequestID string

	// CorrelationID is the correlation identifier.
	CorrelationID string

	// IdempotencyKey is the request idempotency key.
	IdempotencyKey string

	// DedupeKey prevents duplicate durable events.
	DedupeKey string

	// Scopes contains event audience scopes.
	Scopes []Scope

	// AvailableAt is when dispatch can claim the event.
	AvailableAt time.Time
}

// Validate validates an event draft.
func (draft Draft) Validate() error {
	var violations []Violation
	violations = append(violations, ValidateEventKey("key", draft.Key)...)
	if draft.SchemaVersion <= 0 {
		violations = AppendViolation(violations, "schema_version", "must be positive")
	}
	if strings.TrimSpace(string(draft.Producer)) == "" {
		violations = AppendViolation(violations, "producer", "is required")
	}
	if strings.TrimSpace(string(draft.AggregateType)) == "" {
		violations = AppendViolation(violations, "aggregate_type", "is required")
	}
	if len(draft.Scopes) == 0 {
		violations = AppendViolation(violations, "scopes", "must contain at least one scope")
	}
	for _, scope := range draft.Scopes {
		if err := scope.Validate(); err != nil {
			if validation, ok := err.(ValidationError); ok {
				violations = append(violations, validation.Violations...)
			}
		}
	}
	return ErrorIfInvalid(violations)
}
