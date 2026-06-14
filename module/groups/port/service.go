package port

import (
	"context"

	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/module/groups/domain"
	"github.com/realmkit/rk-backend/pkg/pagination"
)

// GroupService manages groups.
type GroupService interface {
	// Create creates a group.
	Create(ctx context.Context, command CreateGroupCommand) (domain.Group, error)

	// Update updates a group.
	Update(ctx context.Context, command UpdateGroupCommand) (domain.Group, error)

	// Get returns one group.
	Get(ctx context.Context, id uuid.UUID) (domain.Group, error)

	// List lists groups.
	List(ctx context.Context, filter GroupFilter, page pagination.Page) (pagination.Result[domain.Group], error)

	// Delete deletes a group.
	Delete(ctx context.Context, command DeleteGroupCommand) error
}

// MembershipService manages memberships.
type MembershipService interface {
	// Assign assigns a user to a group.
	Assign(ctx context.Context, command AssignMembershipCommand) (domain.Membership, error)

	// Remove removes a membership.
	Remove(ctx context.Context, command RemoveMembershipCommand) error

	// ListGroupMembers lists memberships for a group.
	ListGroupMembers(ctx context.Context, groupID uuid.UUID, page pagination.Page) (pagination.Result[domain.Membership], error)

	// ListUserGroups returns active groups for user.
	ListUserGroups(ctx context.Context, userID uuid.UUID) (UserGroups, error)
}

// PermissionGrantService manages permission grants.
type PermissionGrantService interface {
	// CreatePermissionGrant creates a permission grant.
	CreatePermissionGrant(ctx context.Context, command CreatePermissionGrantCommand) (domain.PermissionGrant, error)

	// DeletePermissionGrant deletes a permission grant.
	DeletePermissionGrant(ctx context.Context, command DeletePermissionGrantCommand) error
}

// Checker checks permissions.
type Checker interface {
	// Check returns an authorization decision.
	Check(ctx context.Context, request CheckRequest) (Decision, error)
}
