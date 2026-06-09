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
	if tuple.SubjectID == uuid.Nil {
		violations = AppendViolation(violations, "subject_id", "is required")
	}
	if tuple.SubjectRelation != "" {
		violations = append(violations, ValidateRelationTerm("subject_relation", string(tuple.SubjectRelation))...)
	}
	return NewValidationError(violations)
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
