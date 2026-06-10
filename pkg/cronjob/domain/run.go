package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// Run is one cron job execution attempt.
type Run struct {
	// ID is the run identifier.
	ID uuid.UUID `json:"id"`

	// JobKey is the job definition key.
	JobKey string `json:"job_key"`

	// Status is the run status.
	Status RunStatus `json:"status"`

	// ScheduledFor is the due time.
	ScheduledFor *time.Time `json:"scheduled_for,omitempty"`

	// StartedAt is when the run started.
	StartedAt time.Time `json:"started_at"`

	// FinishedAt is when the run finished.
	FinishedAt *time.Time `json:"finished_at,omitempty"`

	// DurationMS is the duration in milliseconds.
	DurationMS int64 `json:"duration_ms"`

	// TriggerType is why the run started.
	TriggerType TriggerType `json:"trigger_type"`

	// TriggeredByUserID is the manual trigger actor.
	TriggeredByUserID *uuid.UUID `json:"triggered_by_user_id,omitempty"`

	// WorkerID is the worker executing the run.
	WorkerID string `json:"worker_id"`

	// ProcessedCount is the number of processed items.
	ProcessedCount int64 `json:"processed_count"`

	// ChangedCount is the number of changed items.
	ChangedCount int64 `json:"changed_count"`

	// SkippedCount is the number of skipped items.
	SkippedCount int64 `json:"skipped_count"`

	// Metadata contains safe run metadata.
	Metadata json.RawMessage `json:"metadata_json"`

	// Error contains safe error text.
	Error string `json:"error,omitempty"`

	// CreatedAt is when the run row was created.
	CreatedAt time.Time `json:"created_at"`

	// UpdatedAt is when the run row changed.
	UpdatedAt time.Time `json:"updated_at"`
}

// Result is returned by job handlers.
type Result struct {
	// ProcessedCount is the number of processed items.
	ProcessedCount int64

	// ChangedCount is the number of changed items.
	ChangedCount int64

	// SkippedCount is the number of skipped items.
	SkippedCount int64

	// Metadata is safe run metadata.
	Metadata any
}
