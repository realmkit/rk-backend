package search

import (
	"errors"
	"strings"
	"testing"
)

// TestNewTextQueryNormalizesWhitespace verifies query normalization.
func TestNewTextQueryNormalizesWhitespace(t *testing.T) {
	query, err := NewTextQuery("  alpha \n beta\t", QueryOptions{})
	if err != nil {
		t.Fatalf("NewTextQuery returned error: %v", err)
	}
	if query.String() != "alpha beta" {
		t.Fatalf("query = %q", query.String())
	}
	if query.LowerLike() != "%alpha beta%" {
		t.Fatalf("like = %q", query.LowerLike())
	}
}

// TestNewTextQueryAllowsEmpty verifies empty search is not an error.
func TestNewTextQueryAllowsEmpty(t *testing.T) {
	query, err := NewTextQuery("   ", QueryOptions{})
	if err != nil {
		t.Fatalf("NewTextQuery returned error: %v", err)
	}
	if !query.Empty() {
		t.Fatal("expected empty query")
	}
}

// TestNewTextQueryRejectsInvalidText verifies query validation failures.
func TestNewTextQueryRejectsInvalidText(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  error
	}{
		{name: "short", value: "a", want: ErrQueryTooShort},
		{name: "long", value: strings.Repeat("a", DefaultMaxQueryLength+1), want: ErrQueryTooLong},
		{name: "control", value: "ab\u0000", want: ErrQueryControlCharacter},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := NewTextQuery(test.value, QueryOptions{})
			if !errors.Is(err, test.want) {
				t.Fatalf("error = %v, want %v", err, test.want)
			}
		})
	}
}
