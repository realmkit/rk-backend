package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/module/groups/domain"
	"github.com/realmkit/rk-backend/module/groups/port"
	eventdomain "github.com/realmkit/rk-backend/pkg/events/domain"
	eventtesting "github.com/realmkit/rk-backend/pkg/events/testing"
	"github.com/realmkit/rk-backend/pkg/pagination"
)

// TestServiceAssignOnlyStoresMembership verifies assignment no longer creates permission state.
func TestServiceAssignOnlyStoresMembership(t *testing.T) {
	service, _, memberships, permissions := newTestService()
	group := testGroup("moderator", domain.GroupStatusActive)
	userID := uuid.New()

	membership, err := service.Assign(
		context.Background(),
		port.AssignMembershipCommand{Membership: domain.Membership{GroupID: group.ID, UserID: userID}},
	)
	if err != nil {
		t.Fatalf("Assign() error = %v", err)
	}
	if membership.Status != domain.MembershipStatusActive || len(memberships.items) != 1 {
		t.Fatalf("membership=%+v memberships=%d, want active membership", membership, len(memberships.items))
	}
	if len(permissions.grants) != 0 {
		t.Fatalf("grants = %d, want none from membership assignment", len(permissions.grants))
	}
}

// TestServiceGroupLifecycle verifies create, update, list, get, and delete paths.
func TestServiceGroupLifecycle(t *testing.T) {
	service, groups, _, _ := newTestService()
	created, err := service.Create(
		context.Background(),
		port.CreateGroupCommand{Group: domain.Group{Key: "admin", Name: "Admin", Color: "#ff0000", Weight: 100}},
	)
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

// TestServicePublishesGrantEvents verifies grants emit event facts.
func TestServicePublishesGrantEvents(t *testing.T) {
	events := &eventtesting.PublisherRecorder{}
	service, _, _, _ := newTestService()
	service = service.WithEvents(events)
	scopeID := uuid.New()
	grant, err := service.CreatePermissionGrant(context.Background(), port.CreatePermissionGrantCommand{
		Grant: domain.PermissionGrant{
			SubjectType: domain.SubjectUser,
			SubjectID:   uuid.New(),
			Action:      "groups.update",
			ScopeType:   domain.ObjectGroup,
			ScopeID:     scopeID,
		},
	})
	if err != nil {
		t.Fatalf("CreatePermissionGrant() error = %v", err)
	}
	if err := service.DeletePermissionGrant(context.Background(), port.DeletePermissionGrantCommand{ID: grant.ID}); err != nil {
		t.Fatalf("DeletePermissionGrant() error = %v", err)
	}
	assertEventKeys(t, events.Drafts(), []string{
		"groups.permission_grant.created",
		"groups.permission_grant.deleted",
	})
}

// TestServiceCheckAllowsDirectUserGrant verifies direct user grants allow actions.
func TestServiceCheckAllowsDirectUserGrant(t *testing.T) {
	service, _, _, permissions := newTestService()
	userID := uuid.New()
	groupID := uuid.New()
	permissions.grants[uuid.New()] = domain.PermissionGrant{
		ID:          uuid.New(),
		SubjectType: domain.SubjectUser,
		SubjectID:   userID,
		Action:      "groups.update",
		ScopeType:   domain.ObjectGroup,
		ScopeID:     groupID,
	}

	decision, err := service.Check(
		context.Background(),
		port.CheckRequest{ActorUserID: userID, Action: "groups.update", ScopeType: domain.ObjectGroup, ScopeID: groupID},
	)
	if err != nil {
		t.Fatalf("Check() error = %v", err)
	}
	if !decision.Allowed || decision.MatchedSubjectType != domain.SubjectUser {
		t.Fatalf("decision = %+v, want direct user grant", decision)
	}
}

// TestServiceCheckAllowsGroupGrant verifies active members inherit group grants.
func TestServiceCheckAllowsGroupGrant(t *testing.T) {
	service, groups, memberships, permissions := newTestService()
	group := testGroup("moderator", domain.GroupStatusActive)
	groups.items[group.ID] = group
	actorID := uuid.New()
	targetID := uuid.New()
	memberships.items[membershipKey(group.ID, actorID)] = domain.Membership{
		ID:      uuid.New(),
		GroupID: group.ID,
		UserID:  actorID,
		Status:  domain.MembershipStatusActive,
		Version: 1,
	}
	permissions.grants[uuid.New()] = domain.PermissionGrant{
		ID:          uuid.New(),
		SubjectType: domain.SubjectGroup,
		SubjectID:   group.ID,
		Action:      "assets.view",
		ScopeType:   domain.ObjectAsset,
		ScopeID:     targetID,
	}

	decision, err := service.Check(
		context.Background(),
		port.CheckRequest{ActorUserID: actorID, Action: "assets.view", ScopeType: domain.ObjectAsset, ScopeID: targetID},
	)
	if err != nil {
		t.Fatalf("Check() error = %v", err)
	}
	if !decision.Allowed || decision.MatchedSubjectType != domain.SubjectGroup {
		t.Fatalf("decision = %+v, want group grant", decision)
	}
}

// TestServiceCheckDeniesDisabledGroupGrant verifies disabled groups do not grant access.
func TestServiceCheckDeniesDisabledGroupGrant(t *testing.T) {
	service, groups, memberships, permissions := newTestService()
	group := testGroup("moderator", domain.GroupStatusDisabled)
	groups.items[group.ID] = group
	actorID := uuid.New()
	targetID := uuid.New()
	memberships.items[membershipKey(group.ID, actorID)] = domain.Membership{
		ID:      uuid.New(),
		GroupID: group.ID,
		UserID:  actorID,
		Status:  domain.MembershipStatusActive,
		Version: 1,
	}
	permissions.grants[uuid.New()] = domain.PermissionGrant{
		ID:          uuid.New(),
		SubjectType: domain.SubjectGroup,
		SubjectID:   group.ID,
		Action:      "assets.view",
		ScopeType:   domain.ObjectAsset,
		ScopeID:     targetID,
	}

	decision, err := service.Check(
		context.Background(),
		port.CheckRequest{ActorUserID: actorID, Action: "assets.view", ScopeType: domain.ObjectAsset, ScopeID: targetID},
	)
	if err != nil {
		t.Fatalf("Check() error = %v", err)
	}
	if decision.Allowed || decision.Reason != "no_matching_grant" {
		t.Fatalf("decision = %+v, want denied", decision)
	}
}

// TestServiceCheckAllowsPublicGrant verifies anonymous actors match public grants.
func TestServiceCheckAllowsPublicGrant(t *testing.T) {
	service, _, _, permissions := newTestService()
	forumID := uuid.New()
	permissions.grants[uuid.New()] = domain.PermissionGrant{
		ID:          uuid.New(),
		SubjectType: domain.SubjectPublic,
		SubjectID:   domain.PublicSubjectID(),
		Action:      domain.PermissionForumsView,
		ScopeType:   domain.ObjectForum,
		ScopeID:     forumID,
	}

	decision, err := service.Check(
		context.Background(),
		port.CheckRequest{Action: domain.PermissionForumsView, ScopeType: domain.ObjectForum, ScopeID: forumID},
	)
	if err != nil {
		t.Fatalf("Check() error = %v", err)
	}
	if !decision.Allowed || decision.MatchedSubjectType != domain.SubjectPublic {
		t.Fatalf("decision = %+v, want public grant", decision)
	}
}

// TestServiceCheckDeniesScopeMismatch verifies action scope type is enforced.
func TestServiceCheckDeniesScopeMismatch(t *testing.T) {
	service, _, _, _ := newTestService()
	decision, err := service.Check(
		context.Background(),
		port.CheckRequest{ActorUserID: uuid.New(), Action: "assets.view", ScopeType: domain.ObjectGroup, ScopeID: uuid.New()},
	)
	if err != nil {
		t.Fatalf("Check() error = %v", err)
	}
	if decision.Allowed || decision.Reason != "scope_type_mismatch" {
		t.Fatalf("decision = %+v, want scope mismatch", decision)
	}
}

// TestServiceCheckUnknownPermission verifies unknown permission errors.
func TestServiceCheckUnknownPermission(t *testing.T) {
	service, _, _, _ := newTestService()
	_, err := service.Check(
		context.Background(),
		port.CheckRequest{ActorUserID: uuid.New(), Action: "missing.permission", ScopeType: domain.ObjectGroup, ScopeID: uuid.New()},
	)
	if !errors.Is(err, port.ErrUnknownPermission) {
		t.Fatalf("Check() error = %v, want %v", err, port.ErrUnknownPermission)
	}
}

// assertEventKeys verifies event draft key order.
func assertEventKeys(t *testing.T, drafts []eventdomain.Draft, want []string) {
	t.Helper()
	if len(drafts) != len(want) {
		t.Fatalf("event count = %d, want %d", len(drafts), len(want))
	}
	for index, key := range want {
		if string(drafts[index].Key) != key {
			t.Fatalf("event[%d] = %s, want %s", index, drafts[index].Key, key)
		}
	}
}

// newTestService returns a service with in-memory repositories.
func newTestService() (Service, *memoryGroups, *memoryMemberships, *memoryPermissions) {
	groups := &memoryGroups{items: map[uuid.UUID]domain.Group{}}
	memberships := &memoryMemberships{items: map[string]domain.Membership{}}
	permissions := &memoryPermissions{
		actions: map[domain.Action]domain.PermissionAction{},
		grants:  map[uuid.UUID]domain.PermissionGrant{},
	}
	service := NewService(groups, memberships, permissions)
	return service, groups, memberships, permissions
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
func (repository *memoryMemberships) ListByGroup(
	_ context.Context,
	groupID uuid.UUID,
	_ pagination.Page,
) (pagination.Result[domain.Membership], error) {
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

// memoryPermissions stores permission actions and grants in memory.
type memoryPermissions struct {
	actions map[domain.Action]domain.PermissionAction
	grants  map[uuid.UUID]domain.PermissionGrant
}

// UpsertAction stores or updates a permission action.
func (repository *memoryPermissions) UpsertAction(_ context.Context, action domain.PermissionAction) (domain.PermissionAction, error) {
	repository.actions[action.Action] = action
	return action, nil
}

// FindAction returns one action.
func (repository *memoryPermissions) FindAction(_ context.Context, action domain.Action) (domain.PermissionAction, error) {
	found, ok := repository.actions[action]
	if !ok {
		return domain.PermissionAction{}, port.ErrNotFound
	}
	return found, nil
}

// CreateGrant stores a grant.
func (repository *memoryPermissions) CreateGrant(_ context.Context, grant domain.PermissionGrant) (domain.PermissionGrant, error) {
	for _, existing := range repository.grants {
		if equivalentGrant(existing, grant) {
			return existing, port.ErrConflict
		}
	}
	repository.grants[grant.ID] = grant
	return grant, nil
}

// ListGrants returns grants.
func (repository *memoryPermissions) ListGrants(
	_ context.Context,
	filter port.PermissionGrantFilter,
	_ pagination.Page,
) (pagination.Result[domain.PermissionGrant], error) {
	items := []domain.PermissionGrant{}
	for _, grant := range repository.grants {
		if filter.Action != "" && grant.Action != filter.Action {
			continue
		}
		if filter.ScopeType != "" && grant.ScopeType != filter.ScopeType {
			continue
		}
		if filter.ScopeID != uuid.Nil && grant.ScopeID != filter.ScopeID {
			continue
		}
		if filter.SubjectType != "" && grant.SubjectType != filter.SubjectType {
			continue
		}
		if filter.SubjectID != uuid.Nil && grant.SubjectID != filter.SubjectID {
			continue
		}
		items = append(items, grant)
	}
	return pagination.Result[domain.PermissionGrant]{Items: items}, nil
}

// DeleteGrant soft deletes one grant.
func (repository *memoryPermissions) DeleteGrant(_ context.Context, id uuid.UUID) error {
	if _, ok := repository.grants[id]; !ok {
		return port.ErrNotFound
	}
	delete(repository.grants, id)
	return nil
}

// equivalentGrant reports whether grants are equivalent.
func equivalentGrant(left domain.PermissionGrant, right domain.PermissionGrant) bool {
	return left.SubjectType == right.SubjectType &&
		left.SubjectID == right.SubjectID &&
		left.Action == right.Action &&
		left.ScopeType == right.ScopeType &&
		left.ScopeID == right.ScopeID &&
		left.Inherit == right.Inherit &&
		left.ConditionKey == right.ConditionKey
}

// fixedNow keeps the imported time package relevant for domain-like membership checks.
func fixedNow() time.Time {
	return time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
}
