package port

import (
	"errors"
	"testing"

	"github.com/realmkit/rk-backend/pkg/search"
)

// TestTicketSortCatalogs verifies ticket sort defaults and allowed keys.
func TestTicketSortCatalogs(t *testing.T) {
	if got := DefaultDefinitionSort(); got.Key != "display_order" || got.DefaultDirection != search.DirectionAsc {
		t.Fatalf("DefaultDefinitionSort() = %#v, want display_order asc", got)
	}
	if got := DefaultTicketSort(); got.Key != "updated_at" || got.DefaultDirection != search.DirectionDesc {
		t.Fatalf("DefaultTicketSort() = %#v, want updated_at desc", got)
	}
	if len(AllowedDefinitionSorts()) < 3 || len(AllowedTicketSorts()) < 4 {
		t.Fatalf("allowed sort catalogs are incomplete")
	}
}

// TestTicketErrorsAreStable verifies exported sentinel errors are usable.
func TestTicketErrorsAreStable(t *testing.T) {
	for _, err := range []error{ErrNotFound, ErrConflict, ErrForbidden, ErrPreconditionFailed} {
		if !errors.Is(err, err) {
			t.Fatalf("errors.Is(%v, itself) = false", err)
		}
	}
}
