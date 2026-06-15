package port

import (
	"time"

	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/module/groups/domain"
	"github.com/realmkit/rk-backend/pkg/pagination"
	"github.com/realmkit/rk-backend/pkg/search"
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

// CreatePermissionGrantCommand creates a permission grant.
type CreatePermissionGrantCommand struct {
	// GroupID is the group receiving the global permission.
	GroupID uuid.UUID

	// Grant is the grant to create.
	Grant domain.PermissionGrant
}

// DeletePermissionGrantCommand deletes a permission grant.
type DeletePermissionGrantCommand struct {
	// GroupID is the group losing the global permission.
	GroupID uuid.UUID

	// ID is the grant identifier.
	ID uuid.UUID
}

// GroupFilter filters groups.
type GroupFilter struct {
	// Status filters by group status.
	Status domain.GroupStatus

	// Query filters by key, name, or description.
	Query search.TextQuery

	// HasIcon filters by icon presence when set.
	HasIcon *bool

	// MinWeight filters by minimum display weight when set.
	MinWeight *int

	// MaxWeight filters by maximum display weight when set.
	MaxWeight *int

	// Sort controls deterministic result ordering.
	Sort search.Sort
}

// DefaultGroupSort returns the default group list sort.
func DefaultGroupSort() search.SortOption {
	return search.SortOption{Key: "weight", DefaultDirection: search.DirectionDesc}
}

// AllowedGroupSorts returns public group list sort keys.
func AllowedGroupSorts() []search.SortOption {
	return []search.SortOption{
		DefaultGroupSort(),
		{Key: "key", DefaultDirection: search.DirectionAsc},
		{Key: "name", DefaultDirection: search.DirectionAsc},
		{Key: "created_at", DefaultDirection: search.DirectionDesc},
		{Key: "updated_at", DefaultDirection: search.DirectionDesc},
	}
}

// PermissionGrantFilter filters permission grants.
type PermissionGrantFilter struct {
	// GroupID filters by assigned group.
	GroupID uuid.UUID

	// ActorUserID filters by active memberships for permission checks.
	ActorUserID uuid.UUID

	// Action filters by permission action.
	Action domain.Action

	// ScopeType filters by resource type.
	ScopeType domain.ScopeType

	// ScopeID filters by resource identifier.
	ScopeID uuid.UUID

	// IncludeAllScopes includes grants that apply to all resources of the scope type.
	IncludeAllScopes bool

	// AllScopeOnly filters to grants that apply to every resource of the scope type.
	AllScopeOnly bool
}

// CheckRequest requests a permission decision.
type CheckRequest struct {
	// ActorUserID is the authenticated user.
	ActorUserID uuid.UUID `json:"actor_user_id"`

	// Action is the domain action.
	Action domain.Action `json:"action"`

	// ScopeType is the target resource type.
	ScopeType domain.ScopeType `json:"scope_type"`

	// ScopeID is the target resource identifier.
	ScopeID uuid.UUID `json:"scope_id"`

	// Context contains module-provided fields used by policy conditions.
	Context map[string]any `json:"context,omitempty"`
}

// Decision contains an authorization result.
type Decision struct {
	// Allowed reports whether the action is allowed.
	Allowed bool `json:"allowed"`

	// Reason explains the decision.
	Reason string `json:"reason"`

	// MatchedGrantID is the grant that allowed the action.
	MatchedGrantID uuid.UUID `json:"matched_grant_id,omitempty"`

	// MatchedScopeType is the resource type that allowed the action.
	MatchedScopeType domain.ScopeType `json:"matched_scope_type,omitempty"`

	// MatchedScopeID is the resource identifier that allowed the action.
	MatchedScopeID uuid.UUID `json:"matched_scope_id,omitempty"`

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
