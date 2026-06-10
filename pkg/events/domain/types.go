package domain

import (
	"regexp"
	"strings"
)

// EventKey identifies one event fact type.
type EventKey string

// Producer identifies the package or module that published an event.
type Producer string

// AggregateType identifies the aggregate affected by an event.
type AggregateType string

// Status is the durable event dispatch status.
type Status string

// ScopeType identifies one event audience scope.
type ScopeType string

const (
	// StatusPending means the event is ready to be claimed.
	StatusPending Status = "pending"

	// StatusProcessing means a dispatcher currently owns the event.
	StatusProcessing Status = "processing"

	// StatusProcessed means dispatch completed.
	StatusProcessed Status = "processed"

	// StatusFailed means dispatch failed and can be retried.
	StatusFailed Status = "failed"

	// StatusDead means retry attempts are exhausted.
	StatusDead Status = "dead"

	// StatusCancelled means an operator cancelled the event.
	StatusCancelled Status = "cancelled"
)

const (
	// ScopeGlobal addresses public global subscribers.
	ScopeGlobal ScopeType = "global"

	// ScopeUser addresses one authenticated user.
	ScopeUser ScopeType = "user"

	// ScopeGroup addresses members of one group.
	ScopeGroup ScopeType = "group"

	// ScopePermission addresses actors with one permission.
	ScopePermission ScopeType = "permission"

	// ScopeForum addresses one forum.
	ScopeForum ScopeType = "forum"

	// ScopeThread addresses one forum thread.
	ScopeThread ScopeType = "thread"

	// ScopePost addresses one forum post.
	ScopePost ScopeType = "post"

	// ScopeAsset addresses one asset.
	ScopeAsset ScopeType = "asset"

	// ScopePunishment addresses one punishment.
	ScopePunishment ScopeType = "punishment"

	// ScopeStaff addresses staff-only subscribers.
	ScopeStaff ScopeType = "staff"

	// ScopeSystem addresses backend consumers only.
	ScopeSystem ScopeType = "system"
)

var eventKeyPattern = regexp.MustCompile(`^[a-z][a-z0-9_]*(\.[a-z][a-z0-9_]*)+$`)

// ValidateEventKey validates a dotted event key.
func ValidateEventKey(field string, value EventKey) []Violation {
	trimmed := strings.TrimSpace(string(value))
	if trimmed == "" {
		return []Violation{{Field: field, Message: "is required"}}
	}
	if !eventKeyPattern.MatchString(trimmed) {
		return []Violation{{Field: field, Message: "must be lower dotted words"}}
	}
	return nil
}

// ValidateStatus validates an event status.
func ValidateStatus(field string, value Status) []Violation {
	switch value {
	case StatusPending, StatusProcessing, StatusProcessed,
		StatusFailed, StatusDead, StatusCancelled:
		return nil
	default:
		return []Violation{{Field: field, Message: "is not supported"}}
	}
}

// ValidateScopeType validates a scope type.
func ValidateScopeType(field string, value ScopeType) []Violation {
	switch value {
	case ScopeGlobal, ScopeUser, ScopeGroup, ScopePermission, ScopeForum,
		ScopeThread, ScopePost, ScopeAsset, ScopePunishment, ScopeStaff,
		ScopeSystem:
		return nil
	default:
		return []Violation{{Field: field, Message: "is not supported"}}
	}
}
