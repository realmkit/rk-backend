package postgres

import (
	"time"

	"github.com/google/uuid"
)

// DefinitionModel stores one cron job definition.
type DefinitionModel struct {
	Key                string     `gorm:"column:key;primaryKey"`
	Name               string     `gorm:"column:name"`
	Description        string     `gorm:"column:description"`
	ScheduleKind       string     `gorm:"column:schedule_kind"`
	ScheduleExpression string     `gorm:"column:schedule_expression"`
	Enabled            bool       `gorm:"column:enabled"`
	ConcurrencyPolicy  string     `gorm:"column:concurrency_policy"`
	NextRunAt          *time.Time `gorm:"column:next_run_at"`
	LastRunAt          *time.Time `gorm:"column:last_run_at"`
	LastStatus         string     `gorm:"column:last_status"`
	LockedBy           string     `gorm:"column:locked_by"`
	LockedUntil        *time.Time `gorm:"column:locked_until"`
	Version            uint64     `gorm:"column:version"`
	CreatedAt          time.Time  `gorm:"column:created_at"`
	UpdatedAt          time.Time  `gorm:"column:updated_at"`
}

// TableName returns the database table.
func (DefinitionModel) TableName() string {
	return "cronjob_definitions"
}

// RunModel stores one cron job run.
type RunModel struct {
	ID                uuid.UUID  `gorm:"column:id;primaryKey"`
	JobKey            string     `gorm:"column:job_key"`
	Status            string     `gorm:"column:status"`
	ScheduledFor      *time.Time `gorm:"column:scheduled_for"`
	StartedAt         time.Time  `gorm:"column:started_at"`
	FinishedAt        *time.Time `gorm:"column:finished_at"`
	DurationMS        int64      `gorm:"column:duration_ms"`
	TriggerType       string     `gorm:"column:trigger_type"`
	TriggeredByUserID *uuid.UUID `gorm:"column:triggered_by_user_id"`
	WorkerID          string     `gorm:"column:worker_id"`
	ProcessedCount    int64      `gorm:"column:processed_count"`
	ChangedCount      int64      `gorm:"column:changed_count"`
	SkippedCount      int64      `gorm:"column:skipped_count"`
	MetadataJSON      string     `gorm:"column:metadata_json"`
	Error             string     `gorm:"column:error"`
	CreatedAt         time.Time  `gorm:"column:created_at"`
	UpdatedAt         time.Time  `gorm:"column:updated_at"`
}

// TableName returns the database table.
func (RunModel) TableName() string {
	return "cronjob_runs"
}
