package domain

import (
	"time"

	"github.com/google/uuid"
)

// PermissionAction describes one grantable action.
type PermissionAction struct {
	// ID is the action identifier.
	ID uuid.UUID `json:"id"`

	// Action is the stable dotted action key.
	Action Action `json:"action"`

	// Area groups actions for administration screens.
	Area string `json:"area"`

	// ScopeType is the resource type this action applies to.
	ScopeType ScopeType `json:"scope_type"`

	// Label is the human-readable action name.
	Label string `json:"label"`

	// Description explains the permission to administrators.
	Description string `json:"description"`

	// WarningLevel marks risky actions for UI confirmation.
	WarningLevel WarningLevel `json:"warning_level"`

	// Enabled reports whether grants can allow this action.
	Enabled bool `json:"enabled"`

	// Version is the optimistic version.
	Version uint64 `json:"version"`

	// CreatedAt is the creation timestamp.
	CreatedAt time.Time `json:"created_at"`

	// UpdatedAt is the last update timestamp.
	UpdatedAt time.Time `json:"updated_at"`
}

// Validate validates permission action fields.
func (action PermissionAction) Validate() error {
	var violations []Violation
	violations = append(violations, ValidatePermission("action", action.Action)...)
	violations = append(violations, ValidateRelationTerm("scope_type", string(action.ScopeType))...)
	violations = append(violations, requireValue("area", action.Area)...)
	violations = append(violations, requireValue("label", action.Label)...)
	if action.ID == uuid.Nil {
		violations = AppendViolation(violations, "id", "is required")
	}
	if action.WarningLevel == "" {
		violations = AppendViolation(violations, "warning_level", "is required")
	}
	if action.Version == 0 {
		violations = AppendViolation(violations, "version", "is required")
	}
	return NewValidationError(violations)
}

// PermissionGrant allows one subject to perform one action on one scope.
type PermissionGrant struct {
	// ID is the grant identifier.
	ID uuid.UUID `json:"id"`

	// SubjectType is the subject type.
	SubjectType SubjectType `json:"subject_type"`

	// SubjectID is the subject identifier.
	SubjectID uuid.UUID `json:"subject_id"`

	// Action is the granted dotted action key.
	Action Action `json:"action"`

	// ScopeType is the granted resource type.
	ScopeType ScopeType `json:"scope_type"`

	// ScopeID is the granted resource identifier.
	ScopeID uuid.UUID `json:"scope_id"`

	// Inherit reports whether descendant scopes inherit this grant.
	Inherit bool `json:"inherit"`

	// ConditionKey references an optional named runtime condition.
	ConditionKey string `json:"condition_key,omitempty"`

	// CreatedByUserID is the creator when known.
	CreatedByUserID *uuid.UUID `json:"created_by_user_id,omitempty"`

	// CreatedAt is the creation timestamp.
	CreatedAt time.Time `json:"created_at"`
}

// Validate validates permission grant fields.
func (grant PermissionGrant) Validate() error {
	var violations []Violation
	violations = append(violations, ValidateRelationTerm("subject_type", string(grant.SubjectType))...)
	violations = append(violations, validateGrantSubject(grant)...)
	violations = append(violations, ValidatePermission("action", grant.Action)...)
	violations = append(violations, ValidateRelationTerm("scope_type", string(grant.ScopeType))...)
	if grant.ID == uuid.Nil {
		violations = AppendViolation(violations, "id", "is required")
	}
	if grant.ScopeID == uuid.Nil {
		violations = AppendViolation(violations, "scope_id", "is required")
	}
	return NewValidationError(violations)
}

// PublicSubjectID returns the stable subject identifier used by public grants.
func PublicSubjectID() uuid.UUID {
	return uuid.MustParse("00000000-0000-0000-0000-000000000001")
}

// AuthenticatedSubjectID returns the stable subject identifier used by authenticated grants.
func AuthenticatedSubjectID() uuid.UUID {
	return uuid.MustParse("00000000-0000-0000-0000-000000000002")
}

// validateGrantSubject validates subject-specific grant invariants.
func validateGrantSubject(grant PermissionGrant) []Violation {
	switch grant.SubjectType {
	case SubjectPublic:
		return validateSystemSubject(grant, PublicSubjectID())
	case SubjectAuthenticated:
		return validateSystemSubject(grant, AuthenticatedSubjectID())
	default:
		if grant.SubjectID == uuid.Nil {
			return []Violation{{Field: "subject_id", Message: "is required"}}
		}
		return nil
	}
}

// validateSystemSubject validates public and authenticated subject grants.
func validateSystemSubject(grant PermissionGrant, expectedID uuid.UUID) []Violation {
	var violations []Violation
	if grant.SubjectID != expectedID {
		violations = AppendViolation(violations, "subject_id", "must use the reserved subject identifier")
	}
	return violations
}

// PolicyCondition describes one contextual condition for a permission rule.
type PolicyCondition struct {
	// Type is the condition evaluator.
	Type ConditionType `json:"type"`

	// Field is the context field read by the condition.
	Field string `json:"field,omitempty"`

	// Value is the expected scalar value.
	Value string `json:"value,omitempty"`

	// Values are the expected list values.
	Values []string `json:"values,omitempty"`

	// Duration is a Go duration string such as 10m or 24h.
	Duration string `json:"duration,omitempty"`
}

// Validate validates policy condition fields.
func (condition PolicyCondition) Validate(field string) []Violation {
	var violations []Violation
	violations = append(violations, ValidateConditionType(field+".type", condition.Type)...)
	switch condition.Type {
	case ConditionEquals:
		violations = append(violations, requireField(field+".field", condition.Field)...)
		violations = append(violations, requireValue(field+".value", condition.Value)...)
	case ConditionIn:
		violations = append(violations, requireField(field+".field", condition.Field)...)
		if len(condition.Values) == 0 {
			violations = AppendViolation(violations, field+".values", "must contain at least one value")
		}
	case ConditionFieldEqualsActor, ConditionFieldNotEqualsActor, ConditionIsUnset, ConditionAssignedToActor:
		violations = append(violations, requireField(field+".field", condition.Field)...)
	case ConditionWithinDuration, ConditionOlderThan:
		violations = append(violations, requireField(field+".field", condition.Field)...)
		violations = append(violations, requireValue(field+".duration", condition.Duration)...)
		if condition.Duration != "" {
			if _, err := time.ParseDuration(condition.Duration); err != nil {
				violations = AppendViolation(violations, field+".duration", "must be a valid duration")
			}
		}
	}
	return violations
}
