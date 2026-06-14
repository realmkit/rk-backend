package search

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

// ErrInvalidCursor reports an invalid or mismatched cursor.
var ErrInvalidCursor = errors.New("invalid search cursor")

// Cursor is an opaque keyset cursor payload.
type Cursor struct {
	// FilterHash binds the cursor to the filter that produced it.
	FilterHash string `json:"filter_hash"`

	// Sort is the public sort key that produced the cursor.
	Sort string `json:"sort"`

	// Direction is the sort direction that produced the cursor.
	Direction Direction `json:"direction"`

	// Values are repository-specific sortable values.
	Values []string `json:"values"`

	// ID is the stable row tie-breaker.
	ID string `json:"id"`
}

// EncodeCursor encodes a cursor into a URL-safe token.
func EncodeCursor(cursor Cursor) (string, error) {
	payload, err := json.Marshal(cursor)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(payload), nil
}

// DecodeCursor decodes a URL-safe cursor token.
func DecodeCursor(token string) (Cursor, error) {
	if strings.TrimSpace(token) == "" {
		return Cursor{}, nil
	}
	payload, err := base64.RawURLEncoding.DecodeString(token)
	if err != nil {
		return Cursor{}, ErrInvalidCursor
	}
	var cursor Cursor
	if err := json.Unmarshal(payload, &cursor); err != nil {
		return Cursor{}, ErrInvalidCursor
	}
	if cursor.FilterHash == "" || cursor.Sort == "" || cursor.Direction == "" || cursor.ID == "" {
		return Cursor{}, ErrInvalidCursor
	}
	return cursor, nil
}

// RequireCursor validates that a decoded cursor belongs to the current query.
func RequireCursor(token string, filterHash string, sort Sort) (Cursor, bool, error) {
	cursor, err := DecodeCursor(token)
	if err != nil {
		return Cursor{}, false, err
	}
	if cursor.ID == "" {
		return Cursor{}, false, nil
	}
	if cursor.FilterHash != filterHash || cursor.Sort != sort.Key || cursor.Direction != sort.Direction {
		return Cursor{}, false, ErrInvalidCursor
	}
	return cursor, true, nil
}

// HashFilter returns a stable hash for cursor-bound filters.
func HashFilter(parts ...any) string {
	payload, _ := json.Marshal(parts)
	sum := sha256.Sum256(payload)
	return fmt.Sprintf("%x", sum[:])
}
