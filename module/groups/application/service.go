package application

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/niflaot/gamehub-go/module/groups/domain"
	"github.com/niflaot/gamehub-go/module/groups/port"
	"github.com/niflaot/gamehub-go/pkg/events/emitter"
	"github.com/niflaot/gamehub-go/pkg/pagination"
)

// Service manages groups, memberships, tuples, and permission checks.
type Service struct {
	groups      port.GroupRepository
	memberships port.MembershipRepository
	tuples      port.TupleRepository
	policies    port.PermissionRepository
	clock       func() time.Time
	events      emitter.Publisher
}

// NewService creates a groups service.
func NewService(groups port.GroupRepository, memberships port.MembershipRepository, tuples port.TupleRepository, policies ...port.PermissionRepository) Service {
	var policyRepository port.PermissionRepository
	if len(policies) > 0 {
		policyRepository = policies[0]
	}
	return Service{
		groups:      groups,
		memberships: memberships,
		tuples:      tuples,
		policies:    policyRepository,
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
	if _, err := service.tuples.Create(ctx, membershipTuple(stored)); err != nil && !errors.Is(err, port.ErrConflict) {
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
	tuples, err := service.tuples.List(ctx, port.TupleFilter{ObjectType: domain.ObjectGroup, ObjectID: membership.GroupID, Relation: domain.RelationMember, SubjectType: domain.SubjectUser, SubjectID: membership.UserID}, pagination.Page{Limit: 100})
	if err != nil {
		return err
	}
	for _, tuple := range tuples.Items {
		if err := service.tuples.Delete(ctx, tuple.ID); err != nil {
			return err
		}
	}
	return service.publishMembershipEvent(ctx, membershipRemovedEvent, membership)
}

// ListGroupMembers lists memberships for a group.
func (service Service) ListGroupMembers(ctx context.Context, groupID uuid.UUID, page pagination.Page) (pagination.Result[domain.Membership], error) {
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

// Create creates a tuple.
func (service Service) CreateTuple(ctx context.Context, command port.CreateTupleCommand) (domain.RelationTuple, error) {
	tuple := command.Tuple
	if tuple.ID == uuid.Nil {
		tuple.ID = uuid.New()
	}
	if err := tuple.Validate(); err != nil {
		return domain.RelationTuple{}, err
	}
	created, err := service.tuples.Create(ctx, tuple)
	if err != nil {
		return domain.RelationTuple{}, err
	}
	return created, service.publishTupleEvent(ctx, relationTupleCreatedEvent, created)
}

// Delete deletes a tuple.
func (service Service) DeleteTuple(ctx context.Context, command port.DeleteTupleCommand) error {
	if err := service.tuples.Delete(ctx, command.ID); err != nil {
		return err
	}
	return service.publishTupleDeleted(ctx, command.ID)
}

// membershipTuple returns the canonical membership tuple.
func membershipTuple(membership domain.Membership) domain.RelationTuple {
	return domain.RelationTuple{
		ID:          uuid.New(),
		ObjectType:  domain.ObjectGroup,
		ObjectID:    membership.GroupID,
		Relation:    domain.RelationMember,
		SubjectType: domain.SubjectUser,
		SubjectID:   membership.UserID,
	}
}
