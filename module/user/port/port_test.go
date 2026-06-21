package port

import (
	"errors"
	"testing"

	"github.com/realmkit/rk-backend/pkg/search"
)

// TestUserSortCatalogs verifies user sort defaults and allowed keys.
func TestUserSortCatalogs(t *testing.T) {
	if got := DefaultUserSort(); got.Key != "created_at" || got.DefaultDirection != search.DirectionDesc {
		t.Fatalf("DefaultUserSort() = %#v, want created_at desc", got)
	}
	if len(AllowedUserSorts()) < 4 {
		t.Fatalf("AllowedUserSorts() returned too few options")
	}
}

// TestUserErrorsAreStable verifies exported sentinel errors are usable.
func TestUserErrorsAreStable(t *testing.T) {
	for _, err := range []error{ErrNotFound, ErrConflict, ErrPreconditionFailed, ErrDisabled} {
		if !errors.Is(err, err) {
			t.Fatalf("errors.Is(%v, itself) = false", err)
		}
	}
}
