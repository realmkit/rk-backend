package harness

import (
	"fmt"
	"testing"
)

// Steps writes readable progress logs for e2e scenarios.
type Steps struct {
	t     testing.TB
	index int
}

// NewSteps creates a readable scenario step logger.
func NewSteps(t testing.TB) *Steps {
	t.Helper()
	return &Steps{t: t}
}

// Log writes one numbered step to the test log.
func (steps *Steps) Log(format string, args ...any) {
	steps.t.Helper()
	steps.index++
	steps.t.Logf("[e2e step %02d] %s", steps.index, fmt.Sprintf(format, args...))
}

// Do logs a step and runs fn.
func (steps *Steps) Do(name string, fn func()) {
	steps.t.Helper()
	steps.Log("%s", name)
	fn()
}
