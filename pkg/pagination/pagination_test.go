package pagination

import (
	"errors"
	"testing"
)

// TestNewUsesDefaultLimit verifies omitted limits receive the default limit.
func TestNewUsesDefaultLimit(t *testing.T) {
	page, err := New(Request{})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	if page.Limit != DefaultLimit {
		t.Fatalf("Limit = %d, want %d", page.Limit, DefaultLimit)
	}
}

// TestNewCapsMaxLimit verifies oversized limits are capped.
func TestNewCapsMaxLimit(t *testing.T) {
	page, err := New(Request{Limit: MaxLimit + 1})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	if page.Limit != MaxLimit {
		t.Fatalf("Limit = %d, want %d", page.Limit, MaxLimit)
	}
}

// TestNewRejectsNegativeLimit verifies negative limits fail validation.
func TestNewRejectsNegativeLimit(t *testing.T) {
	if _, err := New(Request{Limit: -1}); !errors.Is(err, ErrInvalidLimit) {
		t.Fatalf("New() error = %v, want %v", err, ErrInvalidLimit)
	}
}

// TestNewTrimsCursor verifies cursors are normalized.
func TestNewTrimsCursor(t *testing.T) {
	page, err := New(Request{Limit: 10, Cursor: " abc "})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	if page.Cursor != "abc" {
		t.Fatalf("Cursor = %q, want %q", page.Cursor, "abc")
	}
}
