// Package search contains reusable list-search primitives for admin and API surfaces.
package search

import (
	"errors"
	"strings"
	"unicode"
)

const (
	// DefaultMinQueryLength is the default minimum length for text searches.
	DefaultMinQueryLength = 2

	// DefaultMaxQueryLength is the default maximum length for text searches.
	DefaultMaxQueryLength = 120
)

var (
	// ErrQueryTooShort reports a search query below the configured minimum.
	ErrQueryTooShort = errors.New("search query too short")

	// ErrQueryTooLong reports a search query above the configured maximum.
	ErrQueryTooLong = errors.New("search query too long")

	// ErrQueryControlCharacter reports unsupported control characters.
	ErrQueryControlCharacter = errors.New("search query contains control characters")
)

// TextQuery is a normalized user-provided search query.
type TextQuery struct {
	value string // value stores the value value.
}

// QueryOptions controls text query validation.
type QueryOptions struct {
	// MinLength is the minimum accepted non-empty query length.
	MinLength int

	// MaxLength is the maximum accepted query length.
	MaxLength int
}

// NewTextQuery validates and normalizes a search query.
func NewTextQuery(raw string, options QueryOptions) (TextQuery, error) {
	value := normalizeWhitespace(raw)
	if value == "" {
		return TextQuery{}, nil
	}
	if containsControl(value) {
		return TextQuery{}, ErrQueryControlCharacter
	}
	minimum := options.MinLength
	if minimum == 0 {
		minimum = DefaultMinQueryLength
	}
	maximum := options.MaxLength
	if maximum == 0 {
		maximum = DefaultMaxQueryLength
	}
	if len([]rune(value)) < minimum {
		return TextQuery{}, ErrQueryTooShort
	}
	if len([]rune(value)) > maximum {
		return TextQuery{}, ErrQueryTooLong
	}
	return TextQuery{value: value}, nil
}

// String returns the normalized query text.
func (query TextQuery) String() string {
	return query.value
}

// LowerLike returns a lower-case SQL LIKE search value.
func (query TextQuery) LowerLike() string {
	return "%" + strings.ToLower(query.value) + "%"
}

// Empty reports whether the query is absent.
func (query TextQuery) Empty() bool {
	return query.value == ""
}

// normalizeWhitespace trims and collapses whitespace.
func normalizeWhitespace(value string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(value)), " ")
}

// containsControl reports unsupported control runes.
func containsControl(value string) bool {
	for _, current := range value {
		if unicode.IsControl(current) {
			return true
		}
	}
	return false
}
