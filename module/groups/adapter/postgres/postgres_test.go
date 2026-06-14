package postgres

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/module/groups/domain"
	"github.com/realmkit/rk-backend/module/groups/port"
	"github.com/realmkit/rk-backend/pkg/orm"
	"github.com/realmkit/rk-backend/pkg/pagination"
	"github.com/realmkit/rk-backend/pkg/postgres/migrations"
	"github.com/realmkit/rk-backend/pkg/search"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// TestGroupRepositoryLifecycle verifies group CRUD behavior.
func TestGroupRepositoryLifecycle(t *testing.T) {
	groups, _, _, _ := newRepositories(t)
	group, err := groups.Create(context.Background(), testGroup())
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	group.Name = "Admins"
	updated, err := groups.Update(context.Background(), group, group.Version)
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	if updated.Version != 2 || updated.Name != "Admins" {
		t.Fatalf("updated = %+v, want version 2", updated)
	}
	found, err := groups.FindByKey(context.Background(), updated.Key)
	if err != nil {
		t.Fatalf("FindByKey() error = %v", err)
	}
	if found.ID != updated.ID {
		t.Fatalf("FindByKey() ID = %s, want %s", found.ID, updated.ID)
	}
	list, err := groups.List(context.Background(), port.GroupFilter{Status: domain.GroupStatusActive}, pagination.Page{Limit: 10})
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(list.Items) != 1 {
		t.Fatalf("List() items = %d, want 1", len(list.Items))
	}
	if err := groups.Delete(context.Background(), updated.ID, updated.Version); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	if _, err := groups.FindByID(context.Background(), updated.ID); !errors.Is(err, port.ErrNotFound) {
		t.Fatalf("FindByID() error = %v, want %v", err, port.ErrNotFound)
	}
}

// TestGroupRepositorySearchMatchesPartialTerms verifies admin finds Administrator.
func TestGroupRepositorySearchMatchesPartialTerms(t *testing.T) {
	groups, _, _, _ := newRepositories(t)
	administrator := testGroup()
	administrator.Key = "administrator"
	administrator.Name = "Administrator"
	if _, err := groups.Create(context.Background(), administrator); err != nil {
		t.Fatalf("Create() administrator error = %v", err)
	}
	moderator := testGroup()
	moderator.ID = uuid.New()
	moderator.Key = "moderator"
	moderator.Name = "Moderator"
	if _, err := groups.Create(context.Background(), moderator); err != nil {
		t.Fatalf("Create() moderator error = %v", err)
	}
	query, err := search.NewTextQuery("admin", search.QueryOptions{})
	if err != nil {
		t.Fatalf("NewTextQuery() error = %v", err)
	}
	list, err := groups.List(
		context.Background(),
		port.GroupFilter{Query: query},
		pagination.Page{Limit: 10},
	)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(list.Items) != 1 || list.Items[0].Key != "administrator" {
		t.Fatalf("items = %+v, want only administrator", list.Items)
	}
}

// TestGroupPostgresSearchConditionIncludesPartialFallback guards prefix search.
func TestGroupPostgresSearchConditionIncludesPartialFallback(t *testing.T) {
	condition := groupPostgresSearchCondition()
	if !strings.Contains(condition, "plainto_tsquery") {
		t.Fatalf("condition = %q, want full-text search", condition)
	}
	if strings.Count(condition, "LIKE ?") != 3 {
		t.Fatalf("condition = %q, want key, name, and description LIKE fallbacks", condition)
	}
}

// TestMembershipRepositoryUpsertListAndDelete verifies membership persistence.
func TestMembershipRepositoryUpsertListAndDelete(t *testing.T) {
	_, memberships, _, _ := newRepositories(t)
	membership := domain.Membership{
		ID:      uuid.New(),
		GroupID: uuid.New(),
		UserID:  uuid.New(),
		Status:  domain.MembershipStatusActive,
		Version: 1,
	}
	created, isCreated, err := memberships.Upsert(context.Background(), membership)
	if err != nil {
		t.Fatalf("Upsert() error = %v", err)
	}
	if !isCreated || created.Version != 1 {
		t.Fatalf("created=%v version=%d, want created version 1", isCreated, created.Version)
	}
	created.AssignedReason = "promoted"
	updated, isCreated, err := memberships.Upsert(context.Background(), created)
	if err != nil {
		t.Fatalf("Upsert() update error = %v", err)
	}
	if isCreated || updated.Version != 2 {
		t.Fatalf("updated created=%v version=%d, want update version 2", isCreated, updated.Version)
	}
	list, err := memberships.ListByUser(context.Background(), updated.UserID)
	if err != nil {
		t.Fatalf("ListByUser() error = %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("ListByUser() = %d, want 1", len(list))
	}
	if err := memberships.Delete(context.Background(), updated.GroupID, updated.UserID, &updated.Version); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	if _, err := memberships.Find(context.Background(), updated.GroupID, updated.UserID); !errors.Is(err, port.ErrNotFound) {
		t.Fatalf("Find() error = %v, want %v", err, port.ErrNotFound)
	}
}

// TestTupleRepositoryLifecycle verifies tuple persistence.
func TestTupleRepositoryLifecycle(t *testing.T) {
	_, _, tuples, _ := newRepositories(t)
	tuple := domain.RelationTuple{
		ID:          uuid.New(),
		ObjectType:  domain.ObjectGroup,
		ObjectID:    uuid.New(),
		Relation:    domain.RelationManager,
		SubjectType: domain.SubjectUser,
		SubjectID:   uuid.New(),
	}
	created, err := tuples.Create(context.Background(), tuple)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if _, err := tuples.Create(context.Background(), tuple); !errors.Is(err, port.ErrConflict) {
		t.Fatalf("Create() duplicate error = %v, want %v", err, port.ErrConflict)
	}
	list, err := tuples.List(
		context.Background(),
		port.TupleFilter{ObjectType: tuple.ObjectType, ObjectID: tuple.ObjectID},
		pagination.Page{Limit: 10},
	)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(list.Items) != 1 {
		t.Fatalf("List() items = %d, want 1", len(list.Items))
	}
	if err := tuples.Delete(context.Background(), created.ID); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	if _, err := tuples.FindByID(context.Background(), created.ID); !errors.Is(err, port.ErrNotFound) {
		t.Fatalf("FindByID() error = %v, want %v", err, port.ErrNotFound)
	}
}

// TestPermissionRepositoryLifecycle verifies policy definition and rule persistence.
func TestPermissionRepositoryLifecycle(t *testing.T) {
	_, _, _, policies := newRepositories(t)
	definition := domain.PermissionDefinition{
		ID:          uuid.New(),
		Permission:  "posts.update",
		ObjectType:  "post",
		Description: "Update posts",
		Enabled:     true,
		Version:     1,
	}
	stored, err := policies.UpsertDefinition(context.Background(), definition)
	if err != nil {
		t.Fatalf("UpsertDefinition() error = %v", err)
	}
	if stored.Permission != definition.Permission || stored.Version != 1 {
		t.Fatalf("definition = %+v, want stored", stored)
	}
	rule := domain.PermissionRule{
		ID:         uuid.New(),
		Permission: definition.Permission,
		ObjectType: definition.ObjectType,
		Relation:   "author",
		Conditions: []domain.PolicyCondition{{Type: domain.ConditionWithinDuration, Field: "post.created_at", Duration: "10m"}},
		Priority:   10,
		Enabled:    true,
	}
	if _, err := policies.UpsertRule(context.Background(), rule); err != nil {
		t.Fatalf("UpsertRule() error = %v", err)
	}
	rules, err := policies.ListRules(context.Background(), definition.Permission)
	if err != nil {
		t.Fatalf("ListRules() error = %v", err)
	}
	if len(rules) != 1 || len(rules[0].Conditions) != 1 {
		t.Fatalf("rules = %+v, want one condition rule", rules)
	}
}

// newRepositories creates migrated repositories.
func newRepositories(t *testing.T) (GroupRepository, MembershipRepository, TupleRepository, PermissionRepository) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("gorm.Open() error = %v", err)
	}
	if _, err := migrations.NewRunner(db, migrations.DefaultSource()).Up(context.Background()); err != nil {
		t.Fatalf("migrate Up() error = %v", err)
	}
	store := orm.NewStore(db)
	return NewGroupRepository(store), NewMembershipRepository(store), NewTupleRepository(store), NewPermissionRepository(store)
}

// testGroup returns a valid group.
func testGroup() domain.Group {
	return domain.Group{
		ID:      uuid.New(),
		Key:     "admin",
		Name:    "Admin",
		Color:   "#ff0000",
		Weight:  100,
		Status:  domain.GroupStatusActive,
		Version: 1,
	}
}
