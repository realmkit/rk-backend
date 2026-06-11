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

// TestDefaultDefinitionsValidate verifies the code-owned job catalog.
func TestDefaultDefinitionsValidate(t *testing.T) {
	definitions := DefaultDefinitions(time.Unix(0, 0).UTC())
	if len(definitions) == 0 {
		t.Fatalf("DefaultDefinitions() returned no jobs")
	}
	for _, definition := range definitions {
		if err := definition.Validate(); err != nil {
			t.Fatalf("%s Validate() error = %v", definition.Key, err)
		}
	}
	required := map[string]bool{
		JobTicketsDetectSLABreaches: false,
		JobTicketsCloseStale:        false,
		JobTicketsVerifyStats:       false,
		JobTicketsRebuildStats:      false,
	}
	for _, definition := range definitions {
		if _, ok := required[definition.Key]; ok {
			required[definition.Key] = true
		}
	}
	for key, found := range required {
		if !found {
			t.Fatalf("DefaultDefinitions() missing %s", key)
		}
	}
}
