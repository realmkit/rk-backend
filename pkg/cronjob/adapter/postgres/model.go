package postgres

import (
	"time"

	"github.com/google/uuid"
)

// DefinitionModel stores one cron job definition.
type DefinitionModel struct {
	Key                string     `gorm:"column:key;primaryKey"`      // Key stores the key value.
	Name               string     `gorm:"column:name"`                // Name stores the name value.
	Description        string     `gorm:"column:description"`         // Description stores the description value.
	ScheduleKind       string     `gorm:"column:schedule_kind"`       // ScheduleKind stores the schedule kind value.
	ScheduleExpression string     `gorm:"column:schedule_expression"` // ScheduleExpression stores the schedule expression value.
	Enabled            bool       `gorm:"column:enabled"`             // Enabled stores the enabled value.
	ConcurrencyPolicy  string     `gorm:"column:concurrency_policy"`  // ConcurrencyPolicy stores the concurrency policy value.
	NextRunAt          *time.Time `gorm:"column:next_run_at"`         // NextRunAt stores the next run at value.
	LastRunAt          *time.Time `gorm:"column:last_run_at"`         // LastRunAt stores the last run at value.
	LastStatus         string     `gorm:"column:last_status"`         // LastStatus stores the last status value.
	LockedBy           string     `gorm:"column:locked_by"`           // LockedBy stores the locked by value.
	LockedUntil        *time.Time `gorm:"column:locked_until"`        // LockedUntil stores the locked until value.
	Version            uint64     `gorm:"column:version"`             // Version stores the version value.
	CreatedAt          time.Time  `gorm:"column:created_at"`          // CreatedAt stores the created at value.
	UpdatedAt          time.Time  `gorm:"column:updated_at"`          // UpdatedAt stores the updated at value.
}

// TableName returns the database table.
func (DefinitionModel) TableName() string {
	return "cronjob_definitions"
}

// RunModel stores one cron job run.
type RunModel struct {
	ID                uuid.UUID  `gorm:"column:id;primaryKey"`        // ID stores the i d value.
	JobKey            string     `gorm:"column:job_key"`              // JobKey stores the job key value.
	Status            string     `gorm:"column:status"`               // Status stores the status value.
	ScheduledFor      *time.Time `gorm:"column:scheduled_for"`        // ScheduledFor stores the scheduled for value.
	StartedAt         time.Time  `gorm:"column:started_at"`           // StartedAt stores the started at value.
	FinishedAt        *time.Time `gorm:"column:finished_at"`          // FinishedAt stores the finished at value.
	DurationMS        int64      `gorm:"column:duration_ms"`          // DurationMS stores the duration m s value.
	TriggerType       string     `gorm:"column:trigger_type"`         // TriggerType stores the trigger type value.
	TriggeredByUserID *uuid.UUID `gorm:"column:triggered_by_user_id"` // TriggeredByUserID stores the triggered by user i d value.
	WorkerID          string     `gorm:"column:worker_id"`            // WorkerID stores the worker i d value.
	ProcessedCount    int64      `gorm:"column:processed_count"`      // ProcessedCount stores the processed count value.
	ChangedCount      int64      `gorm:"column:changed_count"`        // ChangedCount stores the changed count value.
	SkippedCount      int64      `gorm:"column:skipped_count"`        // SkippedCount stores the skipped count value.
	MetadataJSON      string     `gorm:"column:metadata_json"`        // MetadataJSON stores the metadata j s o n value.
	Error             string     `gorm:"column:error"`                // Error stores the error value.
	CreatedAt         time.Time  `gorm:"column:created_at"`           // CreatedAt stores the created at value.
	UpdatedAt         time.Time  `gorm:"column:updated_at"`           // UpdatedAt stores the updated at value.
}

// TableName returns the database table.
func (RunModel) TableName() string {
	return "cronjob_runs"
}
