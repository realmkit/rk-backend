package domain

import "strings"

// Scope identifies one event audience.
type Scope struct {
	// Type is the audience scope kind.
	Type ScopeType `json:"type"`

	// ID is the scoped object identifier when needed.
	ID string `json:"id,omitempty"`

	// Permission is the permission needed for permission scopes.
	Permission string `json:"permission,omitempty"`
}

// Validate validates the scope.
func (scope Scope) Validate() error {
	var violations []Violation
	violations = append(violations, ValidateScopeType("type", scope.Type)...)
	if requiresScopeID(scope.Type) && strings.TrimSpace(scope.ID) == "" {
		violations = AppendViolation(violations, "id", "is required")
	}
	if scope.Type == ScopePermission && strings.TrimSpace(scope.Permission) == "" {
		violations = AppendViolation(violations, "permission", "is required")
	}
	if scope.Type != ScopePermission && strings.TrimSpace(scope.Permission) != "" {
		violations = AppendViolation(violations, "permission", "is only valid for permission scope")
	}
	return ErrorIfInvalid(violations)
}

// requiresScopeID reports whether scope type needs an object id.
func requiresScopeID(value ScopeType) bool {
	switch value {
	case ScopeUser, ScopeGroup, ScopeForum, ScopeThread, ScopePost,
		ScopeAsset, ScopePunishment, ScopeTicket:
		return true
	default:
		return false
	}
}
