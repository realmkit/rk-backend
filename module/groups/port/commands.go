package port

import (
	"time"

	"github.com/google/uuid"
	"github.com/niflaot/gamehub-go/module/groups/domain"
	"github.com/niflaot/gamehub-go/pkg/pagination"
)

// CreateGroupCommand creates a group.
type CreateGroupCommand struct {
	// Group is the group to create.
	Group domain.Group
}

// UpdateGroupCommand updates mutable group fields.
type UpdateGroupCommand struct {
	// Group is the replacement group state.
	Group domain.Group

	// ExpectedVersion is the required current version.
	ExpectedVersion uint64
}

// DeleteGroupCommand deletes a group.
type DeleteGroupCommand struct {
	// ID is the group identifier.
	ID uuid.UUID

	// ExpectedVersion is the required current version.
	ExpectedVersion uint64
}

// AssignMembershipCommand assigns a user to a group.
type AssignMembershipCommand struct {
	// Membership is the membership to assign.
	Membership domain.Membership
}

// RemoveMembershipCommand revokes a group membership.
type RemoveMembershipCommand struct {
	// GroupID is the group identifier.
	GroupID uuid.UUID

	// UserID is the user identifier.
	UserID uuid.UUID

	// ExpectedVersion is the required current version when known.
	ExpectedVersion *uint64
}

// CreateTupleCommand creates a relation tuple.
type CreateTupleCommand struct {
	// Tuple is the tuple to create.
	Tuple domain.RelationTuple
}

// DeleteTupleCommand deletes a relation tuple.
type DeleteTupleCommand struct {
	// ID is the tuple identifier.
	ID uuid.UUID
}

// GroupFilter filters groups.
type GroupFilter struct {
	// Status filters by group status.
	Status domain.GroupStatus
}

// TupleFilter filters relation tuples.
type TupleFilter struct {
	// ObjectType filters by object type.
	ObjectType domain.ObjectType

	// ObjectID filters by object identifier.
	ObjectID uuid.UUID

	// Relation filters by relation.
	Relation domain.Relation

	// SubjectType filters by subject type.
	SubjectType domain.SubjectType

	// SubjectID filters by subject identifier.
	SubjectID uuid.UUID
}

// CheckRequest requests a permission decision.
type CheckRequest struct {
	// ActorUserID is the authenticated user.
	ActorUserID uuid.UUID `json:"actor_user_id"`

	// Permission is the domain action.
	Permission domain.Permission `json:"permission"`

	// ObjectType is the target object type.
	ObjectType domain.ObjectType `json:"object_type"`

	// ObjectID is the target object identifier.
	ObjectID uuid.UUID `json:"object_id"`

	// Context contains module-provided fields used by policy conditions.
	Context map[string]any `json:"context,omitempty"`
}

// Decision contains an authorization result.
type Decision struct {
	// Allowed reports whether the action is allowed.
	Allowed bool `json:"allowed"`

	// Reason explains the decision.
	Reason string `json:"reason"`

	// MatchedRelation is the relation that allowed the action.
	MatchedRelation domain.Relation `json:"matched_relation,omitempty"`

	// MatchedConditions are the conditions that passed for the allowing rule.
	MatchedConditions []domain.PolicyCondition `json:"matched_conditions,omitempty"`

	// FailedConditions are conditions from matched relations that failed.
	FailedConditions []domain.PolicyCondition `json:"failed_conditions,omitempty"`
}

// UserGroups contains a user's groups and display group.
type UserGroups struct {
	// Groups contains active groups.
	Groups []domain.Group `json:"groups"`

	// DisplayGroup is the selected frontend display group when present.
	DisplayGroup *domain.Group `json:"display_group,omitempty"`

	// EvaluatedAt is the decision instant.
	EvaluatedAt time.Time `json:"evaluated_at"`
}

// Page aliases the shared pagination page.
type Page = pagination.Page

// Result aliases the shared pagination result.
type Result[T any] = pagination.Result[T]
