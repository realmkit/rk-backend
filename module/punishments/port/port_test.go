package port

import (
	"errors"
	"testing"

	"github.com/realmkit/rk-backend/pkg/search"
)

// TestPunishmentSortCatalogs verifies punishment sort defaults and allowed keys.
func TestPunishmentSortCatalogs(t *testing.T) {
	if got := DefaultDefinitionSort(); got.Key != "display_order" || got.DefaultDirection != search.DirectionAsc {
		t.Fatalf("DefaultDefinitionSort() = %#v, want display_order asc", got)
	}
	if got := DefaultPunishmentSort(); got.Key != "created_at" || got.DefaultDirection != search.DirectionDesc {
		t.Fatalf("DefaultPunishmentSort() = %#v, want created_at desc", got)
	}
	if len(AllowedDefinitionSorts()) < 4 || len(AllowedPunishmentSorts()) < 3 {
		t.Fatalf("allowed sort catalogs are incomplete")
	}
}

// TestPunishmentErrorsAreStable verifies exported sentinel errors are usable.
func TestPunishmentErrorsAreStable(t *testing.T) {
	for _, err := range []error{ErrNotFound, ErrConflict, ErrForbidden, ErrPreconditionFailed} {
		if !errors.Is(err, err) {
			t.Fatalf("errors.Is(%v, itself) = false", err)
		}
	}
}
