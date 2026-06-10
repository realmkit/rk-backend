package domain

import "regexp"

// ScheduleKind identifies how a job is scheduled.
type ScheduleKind string

// ConcurrencyPolicy controls overlapping runs.
type ConcurrencyPolicy string

// RunStatus identifies one run state.
type RunStatus string

// TriggerType identifies why a run started.
type TriggerType string

const (
	// ScheduleInterval means schedule expression is a duration.
	ScheduleInterval ScheduleKind = "interval"

	// ScheduleManual means only manual triggers run the job.
	ScheduleManual ScheduleKind = "manual"

	// ScheduleDisabled means the schedule is disabled.
	ScheduleDisabled ScheduleKind = "disabled"
)

const (
	// ConcurrencyForbid prevents overlapping runs.
	ConcurrencyForbid ConcurrencyPolicy = "forbid"

	// ConcurrencyAllow permits overlapping runs.
	ConcurrencyAllow ConcurrencyPolicy = "allow"
)

const (
	// RunPending means the run has been created.
	RunPending RunStatus = "pending"

	// RunRunning means the handler is executing.
	RunRunning RunStatus = "running"

	// RunSucceeded means the handler completed.
	RunSucceeded RunStatus = "succeeded"

	// RunFailed means the handler failed.
	RunFailed RunStatus = "failed"

	// RunCancelled means the run was cancelled.
	RunCancelled RunStatus = "cancelled"

	// RunSkipped means the run was intentionally skipped.
	RunSkipped RunStatus = "skipped"
)

const (
	// TriggerSchedule means a schedule started the run.
	TriggerSchedule TriggerType = "schedule"

	// TriggerManual means an operator started the run.
	TriggerManual TriggerType = "manual"
)

var jobKeyPattern = regexp.MustCompile(`^[a-z][a-z0-9]*(\.[a-z][a-z0-9-]*)+$`)

// ValidateJobKey validates a dotted cron job key.
func ValidateJobKey(field string, value string) []Violation {
	if !jobKeyPattern.MatchString(value) {
		return []Violation{{Field: field, Message: "must be lower dotted words"}}
	}
	return nil
}
