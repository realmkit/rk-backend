package port

import (
	"errors"
	"testing"
)

// TestThemeErrorsAreStable verifies exported sentinel errors are usable.
func TestThemeErrorsAreStable(t *testing.T) {
	for _, err := range []error{ErrNotFound, ErrConflict, ErrPreconditionFailed, ErrPermissionDenied, ErrInvalidState} {
		if !errors.Is(err, err) {
			t.Fatalf("errors.Is(%v, itself) = false", err)
		}
	}
}
