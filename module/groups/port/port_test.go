package port

import (
	"errors"
	"testing"

	"github.com/realmkit/rk-backend/pkg/search"
)

// TestGroupSortCatalogs verifies group sort defaults and allowed keys.
func TestGroupSortCatalogs(t *testing.T) {
	if got := DefaultGroupSort(); got.Key != "weight" || got.DefaultDirection != search.DirectionDesc {
		t.Fatalf("DefaultGroupSort() = %#v, want weight desc", got)
	}
	if len(AllowedGroupSorts()) < 5 {
		t.Fatalf("AllowedGroupSorts() returned too few options")
	}
}

// TestGroupErrorsAreStable verifies exported sentinel errors are usable.
func TestGroupErrorsAreStable(t *testing.T) {
	for _, err := range []error{ErrNotFound, ErrConflict, ErrPreconditionFailed, ErrForbidden, ErrUnknownPermission} {
		if !errors.Is(err, err) {
			t.Fatalf("errors.Is(%v, itself) = false", err)
		}
	}
}
