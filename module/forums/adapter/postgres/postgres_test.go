package postgres

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/niflaot/gamehub-go/module/forums/domain"
	"github.com/niflaot/gamehub-go/module/forums/port"
	groupsdomain "github.com/niflaot/gamehub-go/module/groups/domain"
	"github.com/niflaot/gamehub-go/pkg/orm"
	"github.com/niflaot/gamehub-go/pkg/pagination"
	"github.com/niflaot/gamehub-go/pkg/postgres/migrations"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// TestForumRepositoriesLifecycle verifies category and forum persistence.
func TestForumRepositoriesLifecycle(t *testing.T) {
	categories, forums, _ := newRepositories(t)
	category, err := categories.Create(context.Background(), testCategory())
	if err != nil {
		t.Fatalf("Create category error = %v", err)
	}
	forum := testForum(category.ID, nil, "news")
	created, err := forums.Create(context.Background(), forum)
	if err != nil {
		t.Fatalf("Create forum error = %v", err)
	}
	if created.Path != forum.Path {
		t.Fatalf("created path = %q, want %q", created.Path, forum.Path)
	}
	stats, err := forums.ListStats(context.Background(), []uuid.UUID{created.ID})
	if err != nil {
		t.Fatalf("ListStats() error = %v", err)
	}
	if stats[created.ID].ForumID != created.ID {
		t.Fatalf("stats = %+v, want created stats", stats)
	}
	created.Name = "News Updated"
	updated, err := forums.Update(context.Background(), created, created.Version)
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	if updated.Version != 2 || updated.Name != "News Updated" {
		t.Fatalf("updated = %+v, want version 2", updated)
	}
	list, err := forums.List(context.Background(), port.ForumFilter{CategoryID: category.ID}, pagination.Page{Limit: 10})
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(list.Items) != 1 {
		t.Fatalf("List() items = %d, want 1", len(list.Items))
	}
}

// TestCategoryRepositoryPreconditionAndSoftDelete verifies versioned deletes.
func TestCategoryRepositoryPreconditionAndSoftDelete(t *testing.T) {
	categories, _, _ := newRepositories(t)
	category, err := categories.Create(context.Background(), testCategory())
	if err != nil {
		t.Fatalf("Create category error = %v", err)
	}

	category.Name = "Changed"
	if _, err := categories.Update(context.Background(), category, category.Version+1); !errors.Is(err, port.ErrPreconditionFailed) {
		t.Fatalf("Update() error = %v, want %v", err, port.ErrPreconditionFailed)
	}
	if err := categories.Delete(context.Background(), category.ID, category.Version); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	if _, err := categories.FindByID(context.Background(), category.ID); !errors.Is(err, port.ErrNotFound) {
		t.Fatalf("FindByID() error = %v, want %v", err, port.ErrNotFound)
	}
}

// TestForumRepositoryMoveUpdatesDescendantPaths verifies materialized path moves.
func TestForumRepositoryMoveUpdatesDescendantPaths(t *testing.T) {
	categories, forums, _ := newRepositories(t)
	category, err := categories.Create(context.Background(), testCategory())
	if err != nil {
		t.Fatalf("Create category error = %v", err)
	}
	parent := testForum(category.ID, nil, "parent")
	parent, err = forums.Create(context.Background(), parent)
	if err != nil {
		t.Fatalf("Create parent error = %v", err)
	}
	child := testForum(category.ID, &parent.ID, "child")
	child.Path = parent.Path + child.ID.String() + "/"
	child.Depth = 1
	child, err = forums.Create(context.Background(), child)
	if err != nil {
		t.Fatalf("Create child error = %v", err)
	}
	oldPath := parent.Path
	parent.Path = "/" + uuid.NewString() + "/" + parent.ID.String() + "/"
	parent.Depth = 1
	moved, err := forums.Move(context.Background(), parent, oldPath, parent.Version)
	if err != nil {
		t.Fatalf("Move() error = %v", err)
	}
	foundChild, err := forums.FindByID(context.Background(), child.ID)
	if err != nil {
		t.Fatalf("Find child error = %v", err)
	}
	if foundChild.Path[:len(moved.Path)] != moved.Path || foundChild.Depth != 2 {
		t.Fatalf("child path=%q depth=%d, want under moved path %q depth 2", foundChild.Path, foundChild.Depth, moved.Path)
	}
}

// TestForumRepositoryReorderPersistsDisplayOrder verifies display order updates persist.
func TestForumRepositoryReorderPersistsDisplayOrder(t *testing.T) {
	categories, forums, _ := newRepositories(t)
	category, err := categories.Create(context.Background(), testCategory())
	if err != nil {
		t.Fatalf("Create category error = %v", err)
	}
	first, err := forums.Create(context.Background(), testForum(category.ID, nil, "first"))
	if err != nil {
		t.Fatalf("Create first forum error = %v", err)
	}
	second, err := forums.Create(context.Background(), testForum(category.ID, nil, "second"))
	if err != nil {
		t.Fatalf("Create second forum error = %v", err)
	}

	if err := forums.Reorder(context.Background(), []port.ReorderItem{{ID: first.ID, DisplayOrder: 20}, {ID: second.ID, DisplayOrder: 10}}); err != nil {
		t.Fatalf("Reorder() error = %v", err)
	}
	foundFirst, err := forums.FindByID(context.Background(), first.ID)
	if err != nil {
		t.Fatalf("Find first error = %v", err)
	}
	foundSecond, err := forums.FindByID(context.Background(), second.ID)
	if err != nil {
		t.Fatalf("Find second error = %v", err)
	}
	if foundFirst.DisplayOrder != 20 || foundSecond.DisplayOrder != 10 {
		t.Fatalf("orders = %d/%d, want 20/10", foundFirst.DisplayOrder, foundSecond.DisplayOrder)
	}
}

// TestVisibilityAuthorizerSupportsPublicAndAuthenticated verifies bulk visibility subjects.
func TestVisibilityAuthorizerSupportsPublicAndAuthenticated(t *testing.T) {
	_, _, db := newRepositories(t)
	store := orm.NewStore(db)
	authorizer := NewVisibilityAuthorizer(store)
	publicForumID := uuid.New()
	authForumID := uuid.New()
	createTuple(t, db, publicForumID, groupsdomain.SubjectPublic, groupsdomain.PublicSubjectID())
	createTuple(t, db, authForumID, groupsdomain.SubjectAuthenticated, groupsdomain.AuthenticatedSubjectID())

	anonymous, err := authorizer.VisibleForums(context.Background(), uuid.Nil, []uuid.UUID{publicForumID, authForumID})
	if err != nil {
		t.Fatalf("VisibleForums anonymous error = %v", err)
	}
	if !anonymous[publicForumID] || anonymous[authForumID] {
		t.Fatalf("anonymous = %+v, want only public", anonymous)
	}
	authenticated, err := authorizer.VisibleForums(context.Background(), uuid.New(), []uuid.UUID{publicForumID, authForumID})
	if err != nil {
		t.Fatalf("VisibleForums authenticated error = %v", err)
	}
	if !authenticated[publicForumID] || !authenticated[authForumID] {
		t.Fatalf("authenticated = %+v, want public and authenticated", authenticated)
	}
}

// TestVisibilityAuthorizerSupportsGroupMembership verifies group member grants.
func TestVisibilityAuthorizerSupportsGroupMembership(t *testing.T) {
	_, _, db := newRepositories(t)
	store := orm.NewStore(db)
	authorizer := NewVisibilityAuthorizer(store)
	forumID := uuid.New()
	groupID := uuid.New()
	memberID := uuid.New()
	otherID := uuid.New()
	createGroup(t, db, groupID)
	createMembership(t, db, groupID, memberID)
	createTuple(t, db, forumID, groupsdomain.SubjectGroup, groupID)

	memberVisible, err := authorizer.VisibleForums(context.Background(), memberID, []uuid.UUID{forumID})
	if err != nil {
		t.Fatalf("VisibleForums member error = %v", err)
	}
	if !memberVisible[forumID] {
		t.Fatalf("memberVisible = %+v, want forum visible", memberVisible)
	}
	otherVisible, err := authorizer.VisibleForums(context.Background(), otherID, []uuid.UUID{forumID})
	if err != nil {
		t.Fatalf("VisibleForums other error = %v", err)
	}
	if otherVisible[forumID] {
		t.Fatalf("otherVisible = %+v, want forum hidden", otherVisible)
	}
}

// newRepositories creates migrated forum repositories.
func newRepositories(t *testing.T) (CategoryRepository, ForumRepository, *gorm.DB) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("gorm.Open() error = %v", err)
	}
	if _, err := migrations.NewRunner(db, migrations.DefaultSource()).Up(context.Background()); err != nil {
		t.Fatalf("migrate Up() error = %v", err)
	}
	store := orm.NewStore(db)
	return NewCategoryRepository(store), NewForumRepository(store), db
}

// createTuple stores one visibility tuple.
func createTuple(t *testing.T, db *gorm.DB, forumID uuid.UUID, subjectType groupsdomain.SubjectType, subjectID uuid.UUID) {
	t.Helper()
	subjectRelation := ""
	if subjectType == groupsdomain.SubjectGroup {
		subjectRelation = string(groupsdomain.RelationMember)
	}
	err := db.Exec("INSERT INTO authorization_relation_tuples (id, object_type, object_id, relation, subject_type, subject_id, subject_relation, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)", uuid.New(), groupsdomain.ObjectForum, forumID, groupsdomain.RelationViewer, subjectType, subjectID, subjectRelation).Error
	if err != nil {
		t.Fatalf("insert tuple error = %v", err)
	}
}

// createGroup stores one active group.
func createGroup(t *testing.T, db *gorm.DB, groupID uuid.UUID) {
	t.Helper()
	err := db.Exec("INSERT INTO groups (id, key, name, description, color, weight, status, version, created_at, updated_at) VALUES (?, ?, ?, '', '#ffffff', 0, ?, 1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)", groupID, "mods"+groupID.String()[:8], "Moderators", groupsdomain.GroupStatusActive).Error
	if err != nil {
		t.Fatalf("insert group error = %v", err)
	}
}

// createMembership stores one active group membership.
func createMembership(t *testing.T, db *gorm.DB, groupID uuid.UUID, userID uuid.UUID) {
	t.Helper()
	err := db.Exec("INSERT INTO group_memberships (id, group_id, user_id, status, version, created_at, updated_at) VALUES (?, ?, ?, ?, 1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)", uuid.New(), groupID, userID, groupsdomain.MembershipStatusActive).Error
	if err != nil {
		t.Fatalf("insert membership error = %v", err)
	}
}

// testCategory returns a persisted category.
func testCategory() domain.ForumCategory {
	return domain.ForumCategory{ID: uuid.New(), Key: "official", Name: "Official", Status: domain.CategoryStatusActive, Version: 1}
}

// testForum returns a persisted forum.
func testForum(categoryID uuid.UUID, parentID *uuid.UUID, key string) domain.Forum {
	id := uuid.New()
	return domain.Forum{ID: id, CategoryID: categoryID, ParentForumID: parentID, Kind: domain.ForumKindDiscussion, Key: domain.Key(key), Slug: domain.Slug(key), Name: key, Path: "/" + id.String() + "/", ThreadVisibilityMode: domain.ThreadVisibilityAllThreads, DefaultThreadStatus: domain.ThreadStatusOpen, Status: domain.ForumStatusActive, Version: 1}
}
