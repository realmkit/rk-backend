package port

import (
	"context"

	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/module/groups/domain"
	"github.com/realmkit/rk-backend/pkg/pagination"
)

// GroupRepository stores groups.
type GroupRepository interface {
	// Create stores a group.
	Create(ctx context.Context, group domain.Group) (domain.Group, error)

	// Update stores mutable group fields.
	Update(ctx context.Context, group domain.Group, expectedVersion uint64) (domain.Group, error)

	// FindByID returns one group.
	FindByID(ctx context.Context, id uuid.UUID) (domain.Group, error)

	// FindByKey returns one group by key.
	FindByKey(ctx context.Context, key domain.Key) (domain.Group, error)

	// List returns matching groups.
	List(ctx context.Context, filter GroupFilter, page pagination.Page) (pagination.Result[domain.Group], error)

	// Delete soft deletes a group.
	Delete(ctx context.Context, id uuid.UUID, expectedVersion uint64) error
}

// MembershipRepository stores group memberships.
type MembershipRepository interface {
	// Upsert stores or updates a membership.
	Upsert(ctx context.Context, membership domain.Membership) (domain.Membership, bool, error)

	// Find returns one membership.
	Find(ctx context.Context, groupID uuid.UUID, userID uuid.UUID) (domain.Membership, error)

	// ListByGroup returns group memberships.
	ListByGroup(ctx context.Context, groupID uuid.UUID, page pagination.Page) (pagination.Result[domain.Membership], error)

	// ListByUser returns user memberships.
	ListByUser(ctx context.Context, userID uuid.UUID) ([]domain.Membership, error)

	// Delete soft deletes a membership.
	Delete(ctx context.Context, groupID uuid.UUID, userID uuid.UUID, expectedVersion *uint64) error
}

// PermissionRepository stores permission actions and grants.
type PermissionRepository interface {
	// UpsertAction stores or updates a permission action.
	UpsertAction(ctx context.Context, action domain.PermissionAction) (domain.PermissionAction, error)

	// FindAction returns one active permission action.
	FindAction(ctx context.Context, action domain.Action) (domain.PermissionAction, error)

	// CreateGrant stores a permission grant.
	CreateGrant(ctx context.Context, grant domain.PermissionGrant) (domain.PermissionGrant, error)

	// ListGrants returns active permission grants.
	ListGrants(ctx context.Context, filter PermissionGrantFilter, page pagination.Page) (pagination.Result[domain.PermissionGrant], error)

	// DeleteGrant soft deletes one permission grant.
	DeleteGrant(ctx context.Context, id uuid.UUID) error
}
