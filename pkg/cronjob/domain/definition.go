package domain

import (
	"strings"
	"time"
)

// Definition is a durable cron job schedule.
type Definition struct {
	// Key is the stable job key.
	Key string `json:"key"`

	// Name is the display name.
	Name string `json:"name"`

	// Description explains the job.
	Description string `json:"description"`

	// ScheduleKind identifies schedule behavior.
	ScheduleKind ScheduleKind `json:"schedule_kind"`

	// ScheduleExpression stores duration text for interval jobs.
	ScheduleExpression string `json:"schedule_expression"`

	// Enabled reports whether scheduled runs are enabled.
	Enabled bool `json:"enabled"`

	// ConcurrencyPolicy controls overlapping runs.
	ConcurrencyPolicy ConcurrencyPolicy `json:"concurrency_policy"`

	// NextRunAt is the next due time.
	NextRunAt *time.Time `json:"next_run_at,omitempty"`

	// LastRunAt is the latest run time.
	LastRunAt *time.Time `json:"last_run_at,omitempty"`

	// LastStatus is the latest run status.
	LastStatus RunStatus `json:"last_status,omitempty"`

	// LockedBy is the current worker lock.
	LockedBy string `json:"locked_by,omitempty"`

	// LockedUntil is when the lock expires.
	LockedUntil *time.Time `json:"locked_until,omitempty"`

	// Version is the optimistic version.
	Version uint64 `json:"version"`

	// CreatedAt is when the definition was created.
	CreatedAt time.Time `json:"created_at"`

	// UpdatedAt is when the definition changed.
	UpdatedAt time.Time `json:"updated_at"`
}

// Validate validates the definition.
func (definition Definition) Validate() error {
	var violations []Violation
	violations = append(violations, ValidateJobKey("key", definition.Key)...)
	if strings.TrimSpace(definition.Name) == "" {
		violations = AppendViolation(violations, "name", "is required")
	}
	switch definition.ScheduleKind {
	case ScheduleInterval:
		if _, err := time.ParseDuration(definition.ScheduleExpression); err != nil {
			violations = AppendViolation(violations, "schedule_expression", "must be a duration")
		}
	case ScheduleManual, ScheduleDisabled:
	default:
		violations = AppendViolation(violations, "schedule_kind", "is not supported")
	}
	switch definition.ConcurrencyPolicy {
	case ConcurrencyForbid, ConcurrencyAllow:
	default:
		violations = AppendViolation(violations, "concurrency_policy", "is not supported")
	}
	return ErrorIfInvalid(violations)
}

// NextAfter returns the next due time after now.
func (definition Definition) NextAfter(now time.Time) *time.Time {
	if !definition.Enabled || definition.ScheduleKind != ScheduleInterval {
		return nil
	}
	duration, err := time.ParseDuration(definition.ScheduleExpression)
	if err != nil {
		return nil
	}
	next := now.Add(duration)
	return &next
}
