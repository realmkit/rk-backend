package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/niflaot/gamehub-go/module/groups/domain"
	"github.com/niflaot/gamehub-go/module/groups/port"
	"github.com/niflaot/gamehub-go/pkg/pagination"
)

// TestServiceAssignCreatesMembershipTuple verifies membership assignment creates tuple.
func TestServiceAssignCreatesMembershipTuple(t *testing.T) {
	service, _, memberships, tuples := newTestService()
	group := testGroup("moderator", domain.GroupStatusActive)
	userID := uuid.New()

	membership, err := service.Assign(context.Background(), port.AssignMembershipCommand{Membership: domain.Membership{GroupID: group.ID, UserID: userID}})
	if err != nil {
		t.Fatalf("Assign() error = %v", err)
	}
	if membership.Status != domain.MembershipStatusActive || len(memberships.items) != 1 || len(tuples.items) != 1 {
		t.Fatalf("membership=%+v memberships=%d tuples=%d, want active and tuple", membership, len(memberships.items), len(tuples.items))
	}
}

// TestServiceGroupLifecycle verifies create, update, list, get, and delete paths.
func TestServiceGroupLifecycle(t *testing.T) {
	service, groups, _, _ := newTestService()
	created, err := service.Create(context.Background(), port.CreateGroupCommand{Group: domain.Group{Key: "admin", Name: "Admin", Color: "#ff0000", Weight: 100}})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	created.Name = "Admins"
	updated, err := service.Update(context.Background(), port.UpdateGroupCommand{Group: created, ExpectedVersion: created.Version})
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	if updated.Name != "Admins" || updated.Version != 2 {
		t.Fatalf("updated = %+v, want name and version 2", updated)
	}
	found, err := service.Get(context.Background(), updated.ID)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if found.ID != updated.ID {
		t.Fatalf("Get() ID = %s, want %s", found.ID, updated.ID)
	}
	list, err := service.List(context.Background(), port.GroupFilter{}, pagination.Page{Limit: 10})
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(list.Items) != 1 {
		t.Fatalf("List() items = %d, want 1", len(list.Items))
	}
	if err := service.Delete(context.Background(), port.DeleteGroupCommand{ID: updated.ID, ExpectedVersion: updated.Version}); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	if len(groups.items) != 0 {
		t.Fatalf("groups = %d, want deleted", len(groups.items))
	}
}

// TestServiceCreateRejectsInvalidGroup verifies validation bubbles from create.
func TestServiceCreateRejectsInvalidGroup(t *testing.T) {
	service, _, _, _ := newTestService()
	_, err := service.Create(context.Background(), port.CreateGroupCommand{Group: domain.Group{Key: "bad"}})
	if !errors.Is(err, domain.ErrInvalid) {
		t.Fatalf("Create() error = %v, want %v", err, domain.ErrInvalid)
	}
}

// TestServiceCheckAllowsDirectUserTuple verifies direct tuple decisions.
func TestServiceCheckAllowsDirectUserTuple(t *testing.T) {
	service, _, _, tuples := newTestService()
	userID := uuid.New()
	groupID := uuid.New()
	tuples.items[uuid.New()] = domain.RelationTuple{ID: uuid.New(), ObjectType: domain.ObjectGroup, ObjectID: groupID, Relation: domain.RelationManager, SubjectType: domain.SubjectUser, SubjectID: userID}

	decision, err := service.Check(context.Background(), port.CheckRequest{ActorUserID: userID, Permission: "groups.update", ObjectType: domain.ObjectGroup, ObjectID: groupID})
	if err != nil {
		t.Fatalf("Check() error = %v", err)
	}
	if !decision.Allowed || decision.MatchedRelation != domain.RelationManager {
		t.Fatalf("Decision = %+v, want allowed manager", decision)
	}
}

// TestServiceCheckAllowsGroupSubjectMember verifies group subject relation decisions.
func TestServiceCheckAllowsGroupSubjectMember(t *testing.T) {
	service, groups, memberships, tuples := newTestService()
	actorID := uuid.New()
	staffGroup := testGroup("staff", domain.GroupStatusActive)
	targetID := uuid.New()
	groups.items[staffGroup.ID] = staffGroup
	memberships.items[membershipKey(staffGroup.ID, actorID)] = domain.Membership{ID: uuid.New(), GroupID: staffGroup.ID, UserID: actorID, Status: domain.MembershipStatusActive}
	tuples.items[uuid.New()] = domain.RelationTuple{ID: uuid.New(), ObjectType: domain.ObjectAsset, ObjectID: targetID, Relation: domain.RelationViewer, SubjectType: domain.SubjectGroup, SubjectID: staffGroup.ID, SubjectRelation: domain.RelationMember}

	decision, err := service.Check(context.Background(), port.CheckRequest{ActorUserID: actorID, Permission: "assets.view", ObjectType: domain.ObjectAsset, ObjectID: targetID})
	if err != nil {
		t.Fatalf("Check() error = %v", err)
	}
	if !decision.Allowed {
		t.Fatalf("Decision = %+v, want allowed", decision)
	}
}

// TestServiceCheckDeniesDisabledGroupAndExpiredMembership verifies inactive subjects deny.
func TestServiceCheckDeniesDisabledGroupAndExpiredMembership(t *testing.T) {
	service, groups, memberships, tuples := newTestService()
	actorID := uuid.New()
	staffGroup := testGroup("staff", domain.GroupStatusDisabled)
	targetID := uuid.New()
	expired := time.Now().UTC().Add(-time.Minute)
	groups.items[staffGroup.ID] = staffGroup
	memberships.items[membershipKey(staffGroup.ID, actorID)] = domain.Membership{ID: uuid.New(), GroupID: staffGroup.ID, UserID: actorID, Status: domain.MembershipStatusActive, ExpiresAt: &expired}
	tuples.items[uuid.New()] = domain.RelationTuple{ID: uuid.New(), ObjectType: domain.ObjectAsset, ObjectID: targetID, Relation: domain.RelationViewer, SubjectType: domain.SubjectGroup, SubjectID: staffGroup.ID, SubjectRelation: domain.RelationMember}

	decision, err := service.Check(context.Background(), port.CheckRequest{ActorUserID: actorID, Permission: "assets.view", ObjectType: domain.ObjectAsset, ObjectID: targetID})
	if err != nil {
		t.Fatalf("Check() error = %v", err)
	}
	if decision.Allowed {
		t.Fatalf("Decision = %+v, want denied", decision)
	}
}

// TestServiceCheckDeniesObjectTypeMismatch verifies wrong object type denies without error.
func TestServiceCheckDeniesObjectTypeMismatch(t *testing.T) {
	service, _, _, _ := newTestService()
	decision, err := service.Check(context.Background(), port.CheckRequest{ActorUserID: uuid.New(), Permission: "assets.view", ObjectType: domain.ObjectGroup, ObjectID: uuid.New()})
	if err != nil {
		t.Fatalf("Check() error = %v", err)
	}
	if decision.Allowed || decision.Reason != "object_type_mismatch" {
		t.Fatalf("Decision = %+v, want mismatch deny", decision)
	}
}

// TestServiceCheckEvaluatesPolicyConditions verifies configured policy conditions.
func TestServiceCheckEvaluatesPolicyConditions(t *testing.T) {
	service, _, _, tuples, policies := newPolicyTestService()
	now := time.Date(2026, 6, 9, 12, 0, 0, 0, time.UTC)
	service.clock = func() time.Time { return now }
	actorID := uuid.New()
	postID := uuid.New()
	policies.definitions["posts.update"] = domain.PermissionDefinition{ID: uuid.New(), Permission: "posts.update", ObjectType: "post", Enabled: true, Version: 1}
	policies.rules["posts.update"] = []domain.PermissionRule{{
		ID:         uuid.New(),
		Permission: "posts.update",
		ObjectType: "post",
		Relation:   "author",
		Conditions: []domain.PolicyCondition{{Type: domain.ConditionWithinDuration, Field: "post.created_at", Duration: "10m"}},
		Enabled:    true,
	}}
	tuples.items[uuid.New()] = domain.RelationTuple{ID: uuid.New(), ObjectType: "post", ObjectID: postID, Relation: "author", SubjectType: domain.SubjectUser, SubjectID: actorID}

	allowed, err := service.Check(context.Background(), port.CheckRequest{ActorUserID: actorID, Permission: "posts.update", ObjectType: "post", ObjectID: postID, Context: map[string]any{"post.created_at": now.Add(-5 * time.Minute)}})
	if err != nil {
		t.Fatalf("Check() error = %v", err)
	}
	if !allowed.Allowed || len(allowed.MatchedConditions) != 1 {
		t.Fatalf("Decision = %+v, want allowed with matched condition", allowed)
	}
	denied, err := service.Check(context.Background(), port.CheckRequest{ActorUserID: actorID, Permission: "posts.update", ObjectType: "post", ObjectID: postID, Context: map[string]any{"post.created_at": now.Add(-15 * time.Minute)}})
	if err != nil {
		t.Fatalf("Check() stale error = %v", err)
	}
	if denied.Allowed || denied.Reason != "conditions_failed" || len(denied.FailedConditions) != 1 {
		t.Fatalf("Decision = %+v, want condition failure", denied)
	}
}

// TestServiceCheckDisabledPolicyDoesNotFallback verifies configured disabled policies deny.
func TestServiceCheckDisabledPolicyDoesNotFallback(t *testing.T) {
	service, _, _, tuples, policies := newPolicyTestService()
	actorID := uuid.New()
	groupID := uuid.New()
	policies.definitions["groups.update"] = domain.PermissionDefinition{ID: uuid.New(), Permission: "groups.update", ObjectType: domain.ObjectGroup, Enabled: false, Version: 1}
	tuples.items[uuid.New()] = domain.RelationTuple{ID: uuid.New(), ObjectType: domain.ObjectGroup, ObjectID: groupID, Relation: domain.RelationManager, SubjectType: domain.SubjectUser, SubjectID: actorID}

	decision, err := service.Check(context.Background(), port.CheckRequest{ActorUserID: actorID, Permission: "groups.update", ObjectType: domain.ObjectGroup, ObjectID: groupID})
	if !errors.Is(err, port.ErrUnknownPermission) {
		t.Fatalf("Check() error = %v, want %v", err, port.ErrUnknownPermission)
	}
	if decision.Allowed {
		t.Fatalf("Decision = %+v, want denied", decision)
	}
}

// TestServiceRemoveDeletesMembershipAndTuple verifies membership removal cleans tuples.
func TestServiceRemoveDeletesMembershipAndTuple(t *testing.T) {
	service, _, memberships, tuples := newTestService()
	groupID := uuid.New()
	userID := uuid.New()
	memberships.items[membershipKey(groupID, userID)] = domain.Membership{ID: uuid.New(), GroupID: groupID, UserID: userID, Status: domain.MembershipStatusActive, Version: 1}
	tupleID := uuid.New()
	tuples.items[tupleID] = domain.RelationTuple{ID: tupleID, ObjectType: domain.ObjectGroup, ObjectID: groupID, Relation: domain.RelationMember, SubjectType: domain.SubjectUser, SubjectID: userID}

	if err := service.Remove(context.Background(), port.RemoveMembershipCommand{GroupID: groupID, UserID: userID}); err != nil {
		t.Fatalf("Remove() error = %v", err)
	}
	if len(memberships.items) != 0 || len(tuples.items) != 0 {
		t.Fatalf("memberships=%d tuples=%d, want cleanup", len(memberships.items), len(tuples.items))
	}
}

// TestServiceTupleLifecycle verifies create and delete tuple paths.
func TestServiceTupleLifecycle(t *testing.T) {
	service, _, _, tuples := newTestService()
	tuple, err := service.CreateTuple(context.Background(), port.CreateTupleCommand{Tuple: domain.RelationTuple{ObjectType: domain.ObjectGroup, ObjectID: uuid.New(), Relation: domain.RelationViewer, SubjectType: domain.SubjectUser, SubjectID: uuid.New()}})
	if err != nil {
		t.Fatalf("CreateTuple() error = %v", err)
	}
	if len(tuples.items) != 1 {
		t.Fatalf("tuples = %d, want 1", len(tuples.items))
	}
	if err := service.DeleteTuple(context.Background(), port.DeleteTupleCommand{ID: tuple.ID}); err != nil {
		t.Fatalf("DeleteTuple() error = %v", err)
	}
	if len(tuples.items) != 0 {
		t.Fatalf("tuples = %d, want 0", len(tuples.items))
	}
}

// TestServiceListUserGroupsSelectsDisplayGroup verifies display group selection.
func TestServiceListUserGroupsSelectsDisplayGroup(t *testing.T) {
	service, groups, memberships, _ := newTestService()
	userID := uuid.New()
	member := testGroup("member", domain.GroupStatusActive)
	member.Weight = 1
	vip := testGroup("vip", domain.GroupStatusActive)
	vip.Weight = 20
	groups.items[member.ID] = member
	groups.items[vip.ID] = vip
	memberships.items[membershipKey(member.ID, userID)] = domain.Membership{ID: uuid.New(), GroupID: member.ID, UserID: userID, Status: domain.MembershipStatusActive}
	memberships.items[membershipKey(vip.ID, userID)] = domain.Membership{ID: uuid.New(), GroupID: vip.ID, UserID: userID, Status: domain.MembershipStatusActive}

	result, err := service.ListUserGroups(context.Background(), userID)
	if err != nil {
		t.Fatalf("ListUserGroups() error = %v", err)
	}
	if result.DisplayGroup == nil || result.DisplayGroup.ID != vip.ID {
		t.Fatalf("DisplayGroup = %+v, want vip", result.DisplayGroup)
	}
}

// TestServiceListGroupMembersDelegatesRepository verifies member listing.
func TestServiceListGroupMembersDelegatesRepository(t *testing.T) {
	service, _, memberships, _ := newTestService()
	groupID := uuid.New()
	userID := uuid.New()
	memberships.items[membershipKey(groupID, userID)] = domain.Membership{ID: uuid.New(), GroupID: groupID, UserID: userID, Status: domain.MembershipStatusActive}

	result, err := service.ListGroupMembers(context.Background(), groupID, pagination.Page{Limit: 10})
	if err != nil {
		t.Fatalf("ListGroupMembers() error = %v", err)
	}
	if len(result.Items) != 1 {
		t.Fatalf("ListGroupMembers() items = %d, want 1", len(result.Items))
	}
}

// newTestService returns a service with in-memory repositories.
func newTestService() (Service, *memoryGroups, *memoryMemberships, *memoryTuples) {
	groups := &memoryGroups{items: map[uuid.UUID]domain.Group{}}
	memberships := &memoryMemberships{items: map[string]domain.Membership{}}
	tuples := &memoryTuples{items: map[uuid.UUID]domain.RelationTuple{}}
	service := NewService(groups, memberships, tuples)
	return service, groups, memberships, tuples
}

// newPolicyTestService returns a service with policy repositories.
func newPolicyTestService() (Service, *memoryGroups, *memoryMemberships, *memoryTuples, *memoryPolicies) {
	groups := &memoryGroups{items: map[uuid.UUID]domain.Group{}}
	memberships := &memoryMemberships{items: map[string]domain.Membership{}}
	tuples := &memoryTuples{items: map[uuid.UUID]domain.RelationTuple{}}
	policies := &memoryPolicies{definitions: map[domain.Permission]domain.PermissionDefinition{}, rules: map[domain.Permission][]domain.PermissionRule{}}
	service := NewService(groups, memberships, tuples, policies)
	return service, groups, memberships, tuples, policies
}

// testGroup returns a test group.
func testGroup(key domain.Key, status domain.GroupStatus) domain.Group {
	return domain.Group{ID: uuid.New(), Key: key, Name: string(key), Color: "#00aaff", Weight: 1, Status: status, Version: 1}
}

// membershipKey returns a memory key.
func membershipKey(groupID uuid.UUID, userID uuid.UUID) string {
	return groupID.String() + ":" + userID.String()
}

// memoryGroups stores groups in memory.
type memoryGroups struct {
	items map[uuid.UUID]domain.Group
}

// Create stores a group.
func (repository *memoryGroups) Create(_ context.Context, group domain.Group) (domain.Group, error) {
	repository.items[group.ID] = group
	return group, nil
}

// Update stores mutable group fields.
func (repository *memoryGroups) Update(_ context.Context, group domain.Group, expectedVersion uint64) (domain.Group, error) {
	current := repository.items[group.ID]
	if current.Version != expectedVersion {
		return domain.Group{}, port.ErrPreconditionFailed
	}
	group.Version = expectedVersion + 1
	repository.items[group.ID] = group
	return group, nil
}

// FindByID returns one group.
func (repository *memoryGroups) FindByID(_ context.Context, id uuid.UUID) (domain.Group, error) {
	group, ok := repository.items[id]
	if !ok {
		return domain.Group{}, port.ErrNotFound
	}
	return group, nil
}

// FindByKey returns one group by key.
func (repository *memoryGroups) FindByKey(_ context.Context, key domain.Key) (domain.Group, error) {
	for _, group := range repository.items {
		if group.Key == key {
			return group, nil
		}
	}
	return domain.Group{}, port.ErrNotFound
}

// List returns matching groups.
func (repository *memoryGroups) List(context.Context, port.GroupFilter, pagination.Page) (pagination.Result[domain.Group], error) {
	items := make([]domain.Group, 0, len(repository.items))
	for _, group := range repository.items {
		items = append(items, group)
	}
	return pagination.Result[domain.Group]{Items: items}, nil
}

// Delete soft deletes a group.
func (repository *memoryGroups) Delete(_ context.Context, id uuid.UUID, expectedVersion uint64) error {
	current := repository.items[id]
	if current.Version != expectedVersion {
		return port.ErrPreconditionFailed
	}
	delete(repository.items, id)
	return nil
}

// memoryMemberships stores memberships in memory.
type memoryMemberships struct {
	items map[string]domain.Membership
}

// Upsert stores or updates a membership.
func (repository *memoryMemberships) Upsert(_ context.Context, membership domain.Membership) (domain.Membership, bool, error) {
	key := membershipKey(membership.GroupID, membership.UserID)
	if current, ok := repository.items[key]; ok {
		membership.ID = current.ID
		membership.Version = current.Version + 1
		repository.items[key] = membership
		return membership, false, nil
	}
	repository.items[key] = membership
	return membership, true, nil
}

// Find returns one membership.
func (repository *memoryMemberships) Find(_ context.Context, groupID uuid.UUID, userID uuid.UUID) (domain.Membership, error) {
	membership, ok := repository.items[membershipKey(groupID, userID)]
	if !ok {
		return domain.Membership{}, port.ErrNotFound
	}
	return membership, nil
}

// ListByGroup returns group memberships.
func (repository *memoryMemberships) ListByGroup(_ context.Context, groupID uuid.UUID, _ pagination.Page) (pagination.Result[domain.Membership], error) {
	items := []domain.Membership{}
	for _, membership := range repository.items {
		if membership.GroupID == groupID {
			items = append(items, membership)
		}
	}
	return pagination.Result[domain.Membership]{Items: items}, nil
}

// ListByUser returns user memberships.
func (repository *memoryMemberships) ListByUser(_ context.Context, userID uuid.UUID) ([]domain.Membership, error) {
	items := []domain.Membership{}
	for _, membership := range repository.items {
		if membership.UserID == userID {
			items = append(items, membership)
		}
	}
	return items, nil
}

// Delete soft deletes a membership.
func (repository *memoryMemberships) Delete(_ context.Context, groupID uuid.UUID, userID uuid.UUID, _ *uint64) error {
	delete(repository.items, membershipKey(groupID, userID))
	return nil
}

// memoryTuples stores tuples in memory.
type memoryTuples struct {
	items map[uuid.UUID]domain.RelationTuple
}

// Create stores a tuple.
func (repository *memoryTuples) Create(_ context.Context, tuple domain.RelationTuple) (domain.RelationTuple, error) {
	for _, existing := range repository.items {
		if equivalentTuple(existing, tuple) {
			return existing, port.ErrConflict
		}
	}
	repository.items[tuple.ID] = tuple
	return tuple, nil
}

// FindByID returns one tuple.
func (repository *memoryTuples) FindByID(_ context.Context, id uuid.UUID) (domain.RelationTuple, error) {
	tuple, ok := repository.items[id]
	if !ok {
		return domain.RelationTuple{}, port.ErrNotFound
	}
	return tuple, nil
}

// List returns matching tuples.
func (repository *memoryTuples) List(_ context.Context, filter port.TupleFilter, _ pagination.Page) (pagination.Result[domain.RelationTuple], error) {
	items := []domain.RelationTuple{}
	for _, tuple := range repository.items {
		if filter.ObjectType != "" && tuple.ObjectType != filter.ObjectType {
			continue
		}
		if filter.ObjectID != uuid.Nil && tuple.ObjectID != filter.ObjectID {
			continue
		}
		if filter.Relation != "" && tuple.Relation != filter.Relation {
			continue
		}
		if filter.SubjectType != "" && tuple.SubjectType != filter.SubjectType {
			continue
		}
		if filter.SubjectID != uuid.Nil && tuple.SubjectID != filter.SubjectID {
			continue
		}
		items = append(items, tuple)
	}
	return pagination.Result[domain.RelationTuple]{Items: items}, nil
}

// Delete soft deletes one tuple.
func (repository *memoryTuples) Delete(_ context.Context, id uuid.UUID) error {
	if _, ok := repository.items[id]; !ok {
		return port.ErrNotFound
	}
	delete(repository.items, id)
	return nil
}

// equivalentTuple reports whether tuples are equivalent.
func equivalentTuple(left domain.RelationTuple, right domain.RelationTuple) bool {
	return left.ObjectType == right.ObjectType && left.ObjectID == right.ObjectID && left.Relation == right.Relation && left.SubjectType == right.SubjectType && left.SubjectID == right.SubjectID && left.SubjectRelation == right.SubjectRelation
}

// memoryPolicies stores policy records in memory.
type memoryPolicies struct {
	definitions map[domain.Permission]domain.PermissionDefinition
	rules       map[domain.Permission][]domain.PermissionRule
}

// UpsertDefinition stores a permission definition.
func (repository *memoryPolicies) UpsertDefinition(_ context.Context, definition domain.PermissionDefinition) (domain.PermissionDefinition, error) {
	repository.definitions[definition.Permission] = definition
	return definition, nil
}

// FindDefinition returns a permission definition.
func (repository *memoryPolicies) FindDefinition(_ context.Context, permission domain.Permission) (domain.PermissionDefinition, error) {
	definition, ok := repository.definitions[permission]
	if !ok {
		return domain.PermissionDefinition{}, port.ErrNotFound
	}
	return definition, nil
}

// UpsertRule stores a permission rule.
func (repository *memoryPolicies) UpsertRule(_ context.Context, rule domain.PermissionRule) (domain.PermissionRule, error) {
	repository.rules[rule.Permission] = append(repository.rules[rule.Permission], rule)
	return rule, nil
}

// ListRules returns permission rules.
func (repository *memoryPolicies) ListRules(_ context.Context, permission domain.Permission) ([]domain.PermissionRule, error) {
	return repository.rules[permission], nil
}

// TestServiceCheckUnknownPermission verifies unknown permission errors.
func TestServiceCheckUnknownPermission(t *testing.T) {
	service, _, _, _ := newTestService()
	_, err := service.Check(context.Background(), port.CheckRequest{ActorUserID: uuid.New(), Permission: "missing.permission", ObjectType: domain.ObjectGroup, ObjectID: uuid.New()})
	if !errors.Is(err, port.ErrUnknownPermission) {
		t.Fatalf("Check() error = %v, want %v", err, port.ErrUnknownPermission)
	}
}
