package postgres

import (
	"time"

	"github.com/google/uuid"
)

// EventModel is the GORM model for event_outbox.
type EventModel struct {
	ID             uuid.UUID  `gorm:"column:id;primaryKey"`   // ID stores the i d value.
	EventKey       string     `gorm:"column:event_key"`       // EventKey stores the event key value.
	SchemaVersion  int        `gorm:"column:schema_version"`  // SchemaVersion stores the schema version value.
	Producer       string     `gorm:"column:producer"`        // Producer stores the producer value.
	AggregateType  string     `gorm:"column:aggregate_type"`  // AggregateType stores the aggregate type value.
	AggregateID    *uuid.UUID `gorm:"column:aggregate_id"`    // AggregateID stores the aggregate i d value.
	PayloadJSON    string     `gorm:"column:payload_json"`    // PayloadJSON stores the payload j s o n value.
	MetadataJSON   string     `gorm:"column:metadata_json"`   // MetadataJSON stores the metadata j s o n value.
	ActorUserID    *uuid.UUID `gorm:"column:actor_user_id"`   // ActorUserID stores the actor user i d value.
	RequestID      string     `gorm:"column:request_id"`      // RequestID stores the request i d value.
	CorrelationID  string     `gorm:"column:correlation_id"`  // CorrelationID stores the correlation i d value.
	IdempotencyKey string     `gorm:"column:idempotency_key"` // IdempotencyKey stores the idempotency key value.
	DedupeKey      *string    `gorm:"column:dedupe_key"`      // DedupeKey stores the dedupe key value.
	OccurredAt     time.Time  `gorm:"column:occurred_at"`     // OccurredAt stores the occurred at value.
	AvailableAt    time.Time  `gorm:"column:available_at"`    // AvailableAt stores the available at value.
	Status         string     `gorm:"column:status"`          // Status stores the status value.
	AttemptCount   int        `gorm:"column:attempt_count"`   // AttemptCount stores the attempt count value.
	LockedBy       string     `gorm:"column:locked_by"`       // LockedBy stores the locked by value.
	LockedUntil    *time.Time `gorm:"column:locked_until"`    // LockedUntil stores the locked until value.
	ProcessedAt    *time.Time `gorm:"column:processed_at"`    // ProcessedAt stores the processed at value.
	DeadAt         *time.Time `gorm:"column:dead_at"`         // DeadAt stores the dead at value.
	LastError      string     `gorm:"column:last_error"`      // LastError stores the last error value.
	CreatedAt      time.Time  `gorm:"column:created_at"`      // CreatedAt stores the created at value.
	UpdatedAt      time.Time  `gorm:"column:updated_at"`      // UpdatedAt stores the updated at value.
}

// TableName returns the database table name.
func (EventModel) TableName() string {
	return "event_outbox"
}

// ScopeModel is the GORM model for event_scopes.
type ScopeModel struct {
	ID         uuid.UUID `gorm:"column:id;primaryKey"` // ID stores the i d value.
	EventID    uuid.UUID `gorm:"column:event_id"`      // EventID stores the event i d value.
	ScopeType  string    `gorm:"column:scope_type"`    // ScopeType stores the scope type value.
	ScopeID    string    `gorm:"column:scope_id"`      // ScopeID stores the scope i d value.
	Permission string    `gorm:"column:permission"`    // Permission stores the permission value.
	CreatedAt  time.Time `gorm:"column:created_at"`    // CreatedAt stores the created at value.
}

// TableName returns the database table name.
func (ScopeModel) TableName() string {
	return "event_scopes"
}
