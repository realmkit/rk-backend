package postgres

import (
	"time"

	"github.com/google/uuid"
)

// EventModel is the GORM model for event_outbox.
type EventModel struct {
	ID             uuid.UUID  `gorm:"column:id;primaryKey"`
	EventKey       string     `gorm:"column:event_key"`
	SchemaVersion  int        `gorm:"column:schema_version"`
	Producer       string     `gorm:"column:producer"`
	AggregateType  string     `gorm:"column:aggregate_type"`
	AggregateID    *uuid.UUID `gorm:"column:aggregate_id"`
	PayloadJSON    string     `gorm:"column:payload_json"`
	MetadataJSON   string     `gorm:"column:metadata_json"`
	ActorUserID    *uuid.UUID `gorm:"column:actor_user_id"`
	RequestID      string     `gorm:"column:request_id"`
	CorrelationID  string     `gorm:"column:correlation_id"`
	IdempotencyKey string     `gorm:"column:idempotency_key"`
	DedupeKey      *string    `gorm:"column:dedupe_key"`
	OccurredAt     time.Time  `gorm:"column:occurred_at"`
	AvailableAt    time.Time  `gorm:"column:available_at"`
	Status         string     `gorm:"column:status"`
	AttemptCount   int        `gorm:"column:attempt_count"`
	LockedBy       string     `gorm:"column:locked_by"`
	LockedUntil    *time.Time `gorm:"column:locked_until"`
	ProcessedAt    *time.Time `gorm:"column:processed_at"`
	DeadAt         *time.Time `gorm:"column:dead_at"`
	LastError      string     `gorm:"column:last_error"`
	CreatedAt      time.Time  `gorm:"column:created_at"`
	UpdatedAt      time.Time  `gorm:"column:updated_at"`
}

// TableName returns the database table name.
func (EventModel) TableName() string {
	return "event_outbox"
}

// ScopeModel is the GORM model for event_scopes.
type ScopeModel struct {
	ID         uuid.UUID `gorm:"column:id;primaryKey"`
	EventID    uuid.UUID `gorm:"column:event_id"`
	ScopeType  string    `gorm:"column:scope_type"`
	ScopeID    string    `gorm:"column:scope_id"`
	Permission string    `gorm:"column:permission"`
	CreatedAt  time.Time `gorm:"column:created_at"`
}

// TableName returns the database table name.
func (ScopeModel) TableName() string {
	return "event_scopes"
}
