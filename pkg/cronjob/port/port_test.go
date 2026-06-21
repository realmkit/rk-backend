package port

import (
	"context"
	"errors"
	"testing"

	"github.com/realmkit/rk-backend/pkg/cronjob/domain"
)

// TestHandlerFuncRuns verifies HandlerFunc adapts a function.
func TestHandlerFuncRuns(t *testing.T) {
	expected := domain.Result{ProcessedCount: 3, ChangedCount: 2}
	handler := HandlerFunc(func(context.Context, RunContext) (domain.Result, error) {
		return expected, nil
	})
	got, err := handler.Run(context.Background(), RunContext{})
	if err != nil || got.ProcessedCount != expected.ProcessedCount || got.ChangedCount != expected.ChangedCount {
		t.Fatalf("Run() = %#v, %v; want %#v, nil", got, err, expected)
	}
}

// TestCronErrorsAreStable verifies exported sentinel errors are usable.
func TestCronErrorsAreStable(t *testing.T) {
	for _, err := range []error{ErrNotFound, ErrPreconditionFailed, ErrNoDueJob, ErrHandlerMissing} {
		if !errors.Is(err, err) {
			t.Fatalf("errors.Is(%v, itself) = false", err)
		}
	}
}
