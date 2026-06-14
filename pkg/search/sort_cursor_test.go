package search

import (
	"errors"
	"testing"
)

// TestNewSortDefaultsAndValidates verifies sort validation.
func TestNewSortDefaultsAndValidates(t *testing.T) {
	allowed := []SortOption{
		{Key: "name", DefaultDirection: DirectionAsc},
		{Key: "created_at", DefaultDirection: DirectionDesc},
	}
	sort, err := NewSort("", "", allowed[0], allowed)
	if err != nil {
		t.Fatalf("NewSort returned error: %v", err)
	}
	if sort.Key != "name" || sort.Direction != DirectionAsc {
		t.Fatalf("sort = %#v", sort)
	}

	sort, err = NewSort("created_at", "asc", allowed[0], allowed)
	if err != nil {
		t.Fatalf("NewSort returned error: %v", err)
	}
	if sort.Key != "created_at" || sort.Direction != DirectionAsc {
		t.Fatalf("sort = %#v", sort)
	}
}

// TestNewSortRejectsInvalidValues verifies bad sort input.
func TestNewSortRejectsInvalidValues(t *testing.T) {
	allowed := []SortOption{{Key: "name", DefaultDirection: DirectionAsc}}
	if _, err := NewSort("weight", "", allowed[0], allowed); !errors.Is(err, ErrInvalidSort) {
		t.Fatalf("invalid sort error = %v", err)
	}
	if _, err := NewSort("name", "sideways", allowed[0], allowed); !errors.Is(err, ErrInvalidDirection) {
		t.Fatalf("invalid direction error = %v", err)
	}
}

// TestCursorRoundTrip verifies cursor encoding and filter validation.
func TestCursorRoundTrip(t *testing.T) {
	sort := Sort{Key: "name", Direction: DirectionAsc}
	hash := HashFilter("active", "adm")
	token, err := EncodeCursor(Cursor{
		FilterHash: hash,
		Sort:       sort.Key,
		Direction:  sort.Direction,
		Values:     []string{"admin"},
		ID:         "00000000-0000-0000-0000-000000000001",
	})
	if err != nil {
		t.Fatalf("EncodeCursor returned error: %v", err)
	}
	cursor, ok, err := RequireCursor(token, hash, sort)
	if err != nil {
		t.Fatalf("RequireCursor returned error: %v", err)
	}
	if !ok || cursor.Values[0] != "admin" {
		t.Fatalf("cursor = %#v ok=%v", cursor, ok)
	}
}

// TestCursorRejectsMismatches verifies stale cursor protection.
func TestCursorRejectsMismatches(t *testing.T) {
	sort := Sort{Key: "name", Direction: DirectionAsc}
	token, err := EncodeCursor(Cursor{
		FilterHash: "one",
		Sort:       sort.Key,
		Direction:  sort.Direction,
		Values:     []string{"admin"},
		ID:         "00000000-0000-0000-0000-000000000001",
	})
	if err != nil {
		t.Fatalf("EncodeCursor returned error: %v", err)
	}
	_, _, err = RequireCursor(token, "two", sort)
	if !errors.Is(err, ErrInvalidCursor) {
		t.Fatalf("RequireCursor error = %v", err)
	}
}
