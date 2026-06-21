package application

import (
	"context"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/module/groups/domain"
	"github.com/realmkit/rk-backend/module/groups/port"
	"github.com/realmkit/rk-backend/pkg/events/emitter"
	"github.com/realmkit/rk-backend/pkg/pagination"
)

// Service manages groups, memberships, grants, and permission checks.
type Service struct {
	groups      port.GroupRepository      // groups stores the groups value.
	memberships port.MembershipRepository // memberships stores the memberships value.
	permissions port.PermissionRepository // permissions stores the permissions value.
	clock       func() time.Time          // clock stores the clock value.
	events      emitter.Publisher         // events stores the events value.
}

// NewService creates a groups service.
func NewService(
	groups port.GroupRepository,
	memberships port.MembershipRepository,
	permissions port.PermissionRepository,
) Service {
	return Service{
		groups:      groups,
		memberships: memberships,
		permissions: permissions,
		clock:       func() time.Time { return time.Now().UTC() },
	}
}

// WithEvents returns a copy of service that publishes group events.
func (service Service) WithEvents(events emitter.Publisher) Service {
	service.events = events
	return service
}

// Create creates a group.
func (service Service) Create(ctx context.Context, command port.CreateGroupCommand) (domain.Group, error) {
	group := command.Group
	if group.ID == uuid.Nil {
		group.ID = uuid.New()
	}
	if group.Status == "" {
		group.Status = domain.GroupStatusActive
	}
	if group.Version == 0 {
		group.Version = 1
	}
	if err := group.Validate(); err != nil {
		return domain.Group{}, err
	}
	created, err := service.groups.Create(ctx, group)
	if err != nil {
		return domain.Group{}, err
	}
	return created, service.publishGroupEvent(ctx, groupCreatedEvent, created)
}

// Update updates a group.
func (service Service) Update(ctx context.Context, command port.UpdateGroupCommand) (domain.Group, error) {
	current, err := service.groups.FindByID(ctx, command.Group.ID)
	if err != nil {
		return domain.Group{}, err
	}
	group := command.Group
	group.Key = current.Key
	group.Version = current.Version
	if err := group.Validate(); err != nil {
		return domain.Group{}, err
	}
	updated, err := service.groups.Update(ctx, group, command.ExpectedVersion)
	if err != nil {
		return domain.Group{}, err
	}
	return updated, service.publishGroupEvent(ctx, groupUpdatedEvent, updated)
}

// Get returns one group.
func (service Service) Get(ctx context.Context, id uuid.UUID) (domain.Group, error) {
	return service.groups.FindByID(ctx, id)
}

// List lists groups.
func (service Service) List(ctx context.Context, filter port.GroupFilter, page pagination.Page) (pagination.Result[domain.Group], error) {
	return service.groups.List(ctx, filter, page)
}

// Delete deletes a group.
func (service Service) Delete(ctx context.Context, command port.DeleteGroupCommand) error {
	group, err := service.groups.FindByID(ctx, command.ID)
	if err != nil {
		return err
	}
	if err := service.groups.Delete(ctx, command.ID, command.ExpectedVersion); err != nil {
		return err
	}
	return service.publishGroupEvent(ctx, groupDeletedEvent, group)
}

// Assign assigns a user to a group.
func (service Service) Assign(ctx context.Context, command port.AssignMembershipCommand) (domain.Membership, error) {
	membership := command.Membership
	if membership.ID == uuid.Nil {
		membership.ID = uuid.New()
	}
	if membership.Status == "" {
		membership.Status = domain.MembershipStatusActive
	}
	if membership.Version == 0 {
		membership.Version = 1
	}
	if err := membership.Validate(); err != nil {
		return domain.Membership{}, err
	}
	stored, _, err := service.memberships.Upsert(ctx, membership)
	if err != nil {
		return domain.Membership{}, err
	}
	return stored, service.publishMembershipEvent(ctx, membershipAddedEvent, stored)
}

// Remove removes a membership.
func (service Service) Remove(ctx context.Context, command port.RemoveMembershipCommand) error {
	membership, err := service.memberships.Find(ctx, command.GroupID, command.UserID)
	if err != nil {
		return err
	}
	if err := service.memberships.Delete(ctx, command.GroupID, command.UserID, command.ExpectedVersion); err != nil {
		return err
	}
	return service.publishMembershipEvent(ctx, membershipRemovedEvent, membership)
}

// ListGroupMembers lists memberships for a group.
func (service Service) ListGroupMembers(
	ctx context.Context,
	groupID uuid.UUID,
	page pagination.Page,
) (pagination.Result[domain.Membership], error) {
	return service.memberships.ListByGroup(ctx, groupID, page)
}

// ListUserGroups returns active groups for user.
func (service Service) ListUserGroups(ctx context.Context, userID uuid.UUID) (port.UserGroups, error) {
	memberships, err := service.memberships.ListByUser(ctx, userID)
	if err != nil {
		return port.UserGroups{}, err
	}
	groups := make([]domain.Group, 0, len(memberships))
	for _, membership := range memberships {
		group, err := service.groups.FindByID(ctx, membership.GroupID)
		if err != nil {
			return port.UserGroups{}, err
		}
		if membership.ActiveAt(service.clock()) && group.GrantsPermissions() {
			groups = append(groups, group)
		}
	}
	display, ok := domain.DisplayGroup(groups, memberships, service.clock())
	result := port.UserGroups{Groups: groups, EvaluatedAt: service.clock()}
	if ok {
		result.DisplayGroup = &display
	}
	return result, nil
}

// ListPermissionActions returns grantable permission actions.
func (service Service) ListPermissionActions(context.Context) ([]domain.PermissionAction, error) {
	actions := make([]domain.PermissionAction, 0, len(staticPermissionActions))
	for _, action := range staticPermissionActions {
		actions = append(actions, action)
	}
	sort.Slice(actions, func(left int, right int) bool {
		if actions[left].Area == actions[right].Area {
			return actions[left].Action < actions[right].Action
		}
		return actions[left].Area < actions[right].Area
	})
	return actions, nil
}

// ListPermissionGrants returns permission grants.
func (service Service) ListPermissionGrants(
	ctx context.Context,
	filter port.PermissionGrantFilter,
	page pagination.Page,
) (pagination.Result[domain.PermissionGrant], error) {
	return service.permissions.ListGrants(ctx, filter, page)
}

// CreatePermissionGrant creates a permission grant.
func (service Service) CreatePermissionGrant(
	ctx context.Context,
	command port.CreatePermissionGrantCommand,
) (domain.PermissionGrant, error) {
	grant := command.Grant
	if grant.ID == uuid.Nil {
		grant.ID = uuid.New()
	}
	if err := grant.Validate(); err != nil {
		return domain.PermissionGrant{}, err
	}
	action, err := permissionAction(grant.Action)
	if err != nil {
		return domain.PermissionGrant{}, err
	}
	if action.ScopeType != grant.ScopeType {
		return domain.PermissionGrant{}, domain.NewValidationError([]domain.Violation{{
			Field:   "scope_type",
			Message: "must match the permission scope",
		}})
	}
	if _, err := service.groups.FindByID(ctx, command.GroupID); err != nil {
		return domain.PermissionGrant{}, err
	}
	created, err := service.permissions.CreateGrant(ctx, command.GroupID, grant)
	if err != nil {
		return domain.PermissionGrant{}, err
	}
	return created, service.publishGrantEvent(ctx, permissionGrantCreatedEvent, created)
}

// DeletePermissionGrant deletes a permission grant.
func (service Service) DeletePermissionGrant(ctx context.Context, command port.DeletePermissionGrantCommand) error {
	if _, err := service.groups.FindByID(ctx, command.GroupID); err != nil {
		return err
	}
	if err := service.permissions.DeleteGrant(ctx, command.GroupID, command.ID); err != nil {
		return err
	}
	return service.publishGrantDeleted(ctx, command.ID)
}
