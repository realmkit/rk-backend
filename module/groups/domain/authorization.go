package domain

import (
	"time"

	"github.com/google/uuid"
)

// RelationTuple grants a relation on an object to a user or group subject.
type RelationTuple struct {
	// ID is the tuple identifier.
	ID uuid.UUID `json:"id"`

	// ObjectType is the object type.
	ObjectType ObjectType `json:"object_type"`

	// ObjectID is the object identifier.
	ObjectID uuid.UUID `json:"object_id"`

	// Relation is the granted object relation.
	Relation Relation `json:"relation"`

	// SubjectType is the subject type.
	SubjectType SubjectType `json:"subject_type"`

	// SubjectID is the subject identifier.
	SubjectID uuid.UUID `json:"subject_id"`

	// SubjectRelation is the subject relation required for group subjects.
	SubjectRelation Relation `json:"subject_relation,omitempty"`

	// CreatedByUserID is the creator when known.
	CreatedByUserID *uuid.UUID `json:"created_by_user_id,omitempty"`

	// CreatedAt is the creation timestamp.
	CreatedAt time.Time `json:"created_at"`
}

// Validate validates relation tuple fields.
func (tuple RelationTuple) Validate() error {
	var violations []Violation
	violations = append(violations, ValidateRelationTerm("object_type", string(tuple.ObjectType))...)
	if tuple.ObjectID == uuid.Nil {
		violations = AppendViolation(violations, "object_id", "is required")
	}
	violations = append(violations, ValidateRelationTerm("relation", string(tuple.Relation))...)
	violations = append(violations, ValidateRelationTerm("subject_type", string(tuple.SubjectType))...)
	violations = append(violations, validateTupleSubject(tuple)...)
	if tuple.SubjectRelation != "" {
		violations = append(violations, ValidateRelationTerm("subject_relation", string(tuple.SubjectRelation))...)
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

// validateTupleSubject validates subject-specific tuple invariants.
func validateTupleSubject(tuple RelationTuple) []Violation {
	switch tuple.SubjectType {
	case SubjectPublic:
		return validateSystemSubject(tuple, PublicSubjectID())
	case SubjectAuthenticated:
		return validateSystemSubject(tuple, AuthenticatedSubjectID())
	default:
		if tuple.SubjectID == uuid.Nil {
			return []Violation{{Field: "subject_id", Message: "is required"}}
		}
		if tuple.SubjectRelation != "" && tuple.SubjectType != SubjectGroup {
			return []Violation{{Field: "subject_relation", Message: "is only supported for group subjects"}}
		}
		return nil
	}
}

// validateSystemSubject validates public and authenticated subject tuples.
func validateSystemSubject(tuple RelationTuple, expectedID uuid.UUID) []Violation {
	var violations []Violation
	if tuple.SubjectID != expectedID {
		violations = AppendViolation(violations, "subject_id", "must use the reserved subject identifier")
	}
	if tuple.SubjectRelation != "" {
		violations = AppendViolation(violations, "subject_relation", "must be empty for this subject type")
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

// PermissionDefinition describes one customizable permission.
type PermissionDefinition struct {
	// ID is the definition identifier.
	ID uuid.UUID `json:"id"`

	// Permission is the domain action.
	Permission Permission `json:"permission"`

	// ObjectType is the target object type.
	ObjectType ObjectType `json:"object_type"`

	// Description explains the permission to administrators.
	Description string `json:"description"`

	// Enabled reports whether the permission can grant access.
	Enabled bool `json:"enabled"`

	// Version is the optimistic version.
	Version uint64 `json:"version"`

	// CreatedAt is the creation timestamp.
	CreatedAt time.Time `json:"created_at"`

	// UpdatedAt is the last update timestamp.
	UpdatedAt time.Time `json:"updated_at"`
}

// Validate validates permission definition fields.
func (definition PermissionDefinition) Validate() error {
	var violations []Violation
	violations = append(violations, ValidatePermission("permission", definition.Permission)...)
	violations = append(violations, ValidateRelationTerm("object_type", string(definition.ObjectType))...)
	if definition.ID == uuid.Nil {
		violations = AppendViolation(violations, "id", "is required")
	}
	if definition.Version == 0 {
		violations = AppendViolation(violations, "version", "is required")
	}
	return NewValidationError(violations)
}

// PermissionRule maps one relation to one permission with optional conditions.
type PermissionRule struct {
	// ID is the rule identifier.
	ID uuid.UUID `json:"id"`

	// Permission is the domain action.
	Permission Permission `json:"permission"`

	// ObjectType is the target object type.
	ObjectType ObjectType `json:"object_type"`

	// Relation is the object relation that can grant access.
	Relation Relation `json:"relation"`

	// Conditions must pass after the relation matches.
	Conditions []PolicyCondition `json:"conditions,omitempty"`

	// Priority orders rules from lowest number to highest.
	Priority int `json:"priority"`

	// Enabled reports whether the rule participates in checks.
	Enabled bool `json:"enabled"`

	// CreatedAt is the creation timestamp.
	CreatedAt time.Time `json:"created_at"`

	// UpdatedAt is the last update timestamp.
	UpdatedAt time.Time `json:"updated_at"`
}

// Validate validates permission rule fields.
func (rule PermissionRule) Validate() error {
	var violations []Violation
	violations = append(violations, ValidatePermission("permission", rule.Permission)...)
	violations = append(violations, ValidateRelationTerm("object_type", string(rule.ObjectType))...)
	violations = append(violations, ValidateRelationTerm("relation", string(rule.Relation))...)
	if rule.ID == uuid.Nil {
		violations = AppendViolation(violations, "id", "is required")
	}
	for index, condition := range rule.Conditions {
		violations = append(violations, condition.Validate("conditions["+itoa(index)+"]")...)
	}
	return NewValidationError(violations)
}

// requireField validates a condition field name.
func requireField(field string, value string) []Violation {
	if value == "" {
		return []Violation{{Field: field, Message: "is required"}}
	}
	return nil
}

// requireValue validates a condition scalar value.
func requireValue(field string, value string) []Violation {
	if value == "" {
		return []Violation{{Field: field, Message: "is required"}}
	}
	return nil
}

// itoa formats a small integer without importing strconv into callers.
func itoa(value int) string {
	if value == 0 {
		return "0"
	}
	var buf [20]byte
	index := len(buf)
	for value > 0 {
		index--
		buf[index] = byte('0' + value%10)
		value /= 10
	}
	return string(buf[index:])
}

// DisplayGroup returns the frontend display group from active memberships.
func DisplayGroup(groups []Group, memberships []Membership, instant time.Time) (Group, bool) {
	byID := map[uuid.UUID]Group{}
	for _, group := range groups {
		if group.GrantsPermissions() {
			byID[group.ID] = group
		}
	}
	var selected Group
	var selectedMembership Membership
	found := false
	for _, membership := range memberships {
		group, ok := byID[membership.GroupID]
		if !ok || !membership.ActiveAt(instant) {
			continue
		}
		if !found || betterDisplayGroup(group, membership, selected, selectedMembership) {
			selected = group
			selectedMembership = membership
			found = true
		}
	}
	return selected, found
}

// betterDisplayGroup reports whether candidate should be displayed first.
func betterDisplayGroup(candidate Group, candidateMembership Membership, current Group, currentMembership Membership) bool {
	if candidate.Weight != current.Weight {
		return candidate.Weight > current.Weight
	}
	if !candidateMembership.CreatedAt.Equal(currentMembership.CreatedAt) {
		return candidateMembership.CreatedAt.Before(currentMembership.CreatedAt)
	}
	return candidate.Key < current.Key
}
