package search

import (
	"errors"
	"strings"
)

const (
	// DirectionAsc sorts values from low to high.
	DirectionAsc Direction = "asc"

	// DirectionDesc sorts values from high to low.
	DirectionDesc Direction = "desc"
)

var (
	// ErrInvalidSort reports an unsupported sort key.
	ErrInvalidSort = errors.New("invalid search sort")

	// ErrInvalidDirection reports an unsupported sort direction.
	ErrInvalidDirection = errors.New("invalid search direction")
)

// Direction is a normalized sort direction.
type Direction string

// Sort is a validated sort request.
type Sort struct {
	// Key is the public sort key.
	Key string

	// Direction is the public sort direction.
	Direction Direction
}

// SortOption is one allowed sort key.
type SortOption struct {
	// Key is the public sort key accepted from API clients.
	Key string

	// DefaultDirection is used when clients omit direction.
	DefaultDirection Direction
}

// NewSort validates a requested sort against allowed options.
func NewSort(rawKey string, rawDirection string, defaults SortOption, allowed []SortOption) (Sort, error) {
	key := strings.TrimSpace(rawKey)
	if key == "" {
		key = defaults.Key
	}
	option, ok := findSortOption(key, allowed)
	if !ok {
		return Sort{}, ErrInvalidSort
	}
	direction := Direction(strings.ToLower(strings.TrimSpace(rawDirection)))
	if direction == "" {
		direction = option.DefaultDirection
	}
	if direction != DirectionAsc && direction != DirectionDesc {
		return Sort{}, ErrInvalidDirection
	}
	return Sort{Key: option.Key, Direction: direction}, nil
}

// Desc reports whether this sort is descending.
func (sort Sort) Desc() bool {
	return sort.Direction == DirectionDesc
}

// findSortOption returns the matching sort option.
func findSortOption(key string, allowed []SortOption) (SortOption, bool) {
	for _, option := range allowed {
		if option.Key == key {
			return option, true
		}
	}
	return SortOption{}, false
}
