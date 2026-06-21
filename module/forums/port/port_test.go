package port

import (
	"errors"
	"testing"
)

// TestForumErrorsAreStable verifies exported sentinel errors are usable.
func TestForumErrorsAreStable(t *testing.T) {
	for _, err := range []error{ErrNotFound, ErrConflict, ErrPreconditionFailed, ErrForbidden, ErrInvalidMove} {
		if !errors.Is(err, err) {
			t.Fatalf("errors.Is(%v, itself) = false", err)
		}
	}
}

// TestPaginationAliasesCompile verifies public aliases remain available.
func TestPaginationAliasesCompile(t *testing.T) {
	page := Page{Limit: 10, Cursor: "cursor"}
	result := Result[string]{Items: []string{"one"}, NextCursor: "next"}
	if page.Limit != 10 || result.Items[0] != "one" || result.NextCursor != "next" {
		t.Fatalf("aliases produced unexpected values: %#v %#v", page, result)
	}
}
