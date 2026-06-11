package domain

import (
	"testing"
	"time"
)

// TestDefinitionValidateAcceptsInterval verifies interval schedules.
func TestDefinitionValidateAcceptsInterval(t *testing.T) {
	definition := Definition{
		Key:                JobEventsDispatchPending,
		Name:               "Dispatch events",
		ScheduleKind:       ScheduleInterval,
		ScheduleExpression: time.Minute.String(),
		Enabled:            true,
		ConcurrencyPolicy:  ConcurrencyForbid,
	}
	if err := definition.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	if definition.NextAfter(time.Unix(0, 0).UTC()) == nil {
		t.Fatalf("NextAfter() = nil, want next time")
	}
}

// TestManualAndDisabledDefinitionsDoNotSchedule verifies non-interval scheduling rules.
func TestManualAndDisabledDefinitionsDoNotSchedule(t *testing.T) {
	for _, kind := range []ScheduleKind{ScheduleManual, ScheduleDisabled} {
		definition := Definition{
			Key:               JobTicketsRebuildStats,
			Name:              "Ticket repair",
			ScheduleKind:      kind,
			ConcurrencyPolicy: ConcurrencyAllow,
			Enabled:           true,
		}
		if err := definition.Validate(); err != nil {
			t.Fatalf("Validate(%s) error = %v", kind, err)
		}
		if next := definition.NextAfter(time.Unix(0, 0).UTC()); next != nil {
			t.Fatalf("NextAfter(%s) = %v, want nil", kind, next)
		}
	}
}

// TestDisabledIntervalDoesNotSchedule verifies disabled definitions do not produce due times.
func TestDisabledIntervalDoesNotSchedule(t *testing.T) {
	definition := Definition{
		Key:                JobEventsDispatchPending,
		Name:               "Dispatch events",
		ScheduleKind:       ScheduleInterval,
		ScheduleExpression: time.Minute.String(),
		ConcurrencyPolicy:  ConcurrencyForbid,
	}
	if err := definition.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	if next := definition.NextAfter(time.Unix(0, 0).UTC()); next != nil {
		t.Fatalf("NextAfter() = %v, want nil when disabled", next)
	}
}

// TestDefinitionValidateRejectsBadSchedule verifies validation failures.
func TestDefinitionValidateRejectsBadSchedule(t *testing.T) {
	definition := Definition{
		Key:                "bad",
		Name:               "",
		ScheduleKind:       ScheduleInterval,
		ScheduleExpression: "not-duration",
		ConcurrencyPolicy:  "wild",
	}
	if err := definition.Validate(); err == nil {
		t.Fatalf("Validate() error = nil, want validation error")
	}
}

// TestValidateJobKeyRejectsUppercaseAndUndotted verifies job key shape.
func TestValidateJobKeyRejectsUppercaseAndUndotted(t *testing.T) {
	for _, key := range []string{"Bad.Key", "single"} {
		if violations := ValidateJobKey("key", key); len(violations) != 1 {
			t.Fatalf("ValidateJobKey(%q) = %+v, want one violation", key, violations)
		}
	}
}

// TestValidationErrorFormatting verifies cron validation helper behavior.
func TestValidationErrorFormatting(t *testing.T) {
	if err := ErrorIfInvalid(nil); err != nil {
		t.Fatalf("ErrorIfInvalid(nil) = %v, want nil", err)
	}
	err := ErrorIfInvalid(AppendViolation(nil, "key", "is required"))
	validation, ok := err.(ValidationError)
	if !ok {
		t.Fatalf("ErrorIfInvalid() = %T, want ValidationError", err)
	}
	if validation.Error() != "key: is required" {
		t.Fatalf("ValidationError.Error() = %q, want formatted message", validation.Error())
	}
}
