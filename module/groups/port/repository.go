package port

import (
	"context"

	"github.com/google/uuid"
	"github.com/niflaot/gamehub-go/module/groups/domain"
	"github.com/niflaot/gamehub-go/pkg/pagination"
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

// TupleRepository stores relation tuples.
type TupleRepository interface {
	// Create stores a tuple.
	Create(ctx context.Context, tuple domain.RelationTuple) (domain.RelationTuple, error)

	// FindByID returns one tuple.
	FindByID(ctx context.Context, id uuid.UUID) (domain.RelationTuple, error)

	// List returns matching tuples.
	List(ctx context.Context, filter TupleFilter, page pagination.Page) (pagination.Result[domain.RelationTuple], error)

	// Delete soft deletes one tuple.
	Delete(ctx context.Context, id uuid.UUID) error
}

// PermissionRepository stores customizable permission definitions and rules.
type PermissionRepository interface {
	// UpsertDefinition stores or updates a permission definition.
	UpsertDefinition(ctx context.Context, definition domain.PermissionDefinition) (domain.PermissionDefinition, error)

	// FindDefinition returns one active permission definition.
	FindDefinition(ctx context.Context, permission domain.Permission) (domain.PermissionDefinition, error)

	// UpsertRule stores or updates a permission rule.
	UpsertRule(ctx context.Context, rule domain.PermissionRule) (domain.PermissionRule, error)

	// ListRules returns active rules for a permission.
	ListRules(ctx context.Context, permission domain.Permission) ([]domain.PermissionRule, error)
}
