package port

import (
	"errors"
	"testing"

	"github.com/realmkit/rk-backend/pkg/search"
)

// TestAssetSortCatalogs verifies asset sort defaults and allowed keys.
func TestAssetSortCatalogs(t *testing.T) {
	if got := DefaultAssetSort(); got.Key != "created_at" || got.DefaultDirection != search.DirectionDesc {
		t.Fatalf("DefaultAssetSort() = %#v, want created_at desc", got)
	}
	if len(AllowedAssetSorts()) < 4 {
		t.Fatalf("AllowedAssetSorts() returned too few options")
	}
}

// TestAssetErrorsAreStable verifies exported sentinel errors are usable.
func TestAssetErrorsAreStable(t *testing.T) {
	for _, err := range []error{ErrNotFound, ErrConflict, ErrPreconditionFailed, ErrInvalidState, ErrUploadMismatch} {
		if !errors.Is(err, err) {
			t.Fatalf("errors.Is(%v, itself) = false", err)
		}
	}
}
