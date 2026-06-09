package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/niflaot/gamehub-go/module/forums/domain"
	"github.com/niflaot/gamehub-go/module/forums/port"
	"github.com/niflaot/gamehub-go/pkg/pagination"
)

// TestServiceCreateForumBuildsPathAndStats verifies forum creation prepares tree fields.
func TestServiceCreateForumBuildsPathAndStats(t *testing.T) {
	service, categories, forums, auth, _ := newTestService()
	actorID := uuid.New()
	auth.manage[domain.RootForumObjectID()] = true
	category := testCategory()
	categories.items[category.ID] = category

	created, err := service.CreateForum(context.Background(), port.CreateForumCommand{ActorUserID: actorID, Forum: domain.Forum{CategoryID: category.ID, Key: "news", Slug: "news", Name: "News"}})
	if err != nil {
		t.Fatalf("CreateForum() error = %v", err)
	}
	if created.Path != "/"+created.ID.String()+"/" || created.Depth != 0 {
		t.Fatalf("created path=%q depth=%d, want root path", created.Path, created.Depth)
	}
	if _, ok := forums.stats[created.ID]; !ok {
		t.Fatalf("stats missing for created forum")
	}
}

// TestServiceCreateCategoryRequiresManagePermission verifies structure admin checks.
func TestServiceCreateCategoryRequiresManagePermission(t *testing.T) {
	service, _, _, _, _ := newTestService()

	_, err := service.CreateCategory(context.Background(), port.CreateCategoryCommand{ActorUserID: uuid.New(), Category: testCategory()})
	if !errors.Is(err, port.ErrForbidden) {
		t.Fatalf("CreateCategory() error = %v, want %v", err, port.ErrForbidden)
	}
}

// TestServiceCreateForumUnderParentInheritsParentCategory verifies tree preparation.
func TestServiceCreateForumUnderParentInheritsParentCategory(t *testing.T) {
	service, categories, forums, auth, _ := newTestService()
	actorID := uuid.New()
	auth.manage[domain.RootForumObjectID()] = true
	rootCategory := testCategory()
	otherCategory := testCategory()
	otherCategory.ID = uuid.New()
	otherCategory.Key = "other"
	categories.items[rootCategory.ID] = rootCategory
	categories.items[otherCategory.ID] = otherCategory
	parent := testForum(rootCategory.ID, nil, 0, "support")
	forums.items[parent.ID] = parent

	created, err := service.CreateForum(context.Background(), port.CreateForumCommand{ActorUserID: actorID, Forum: domain.Forum{CategoryID: otherCategory.ID, ParentForumID: &parent.ID, Key: "help", Slug: "help", Name: "Help"}})
	if err != nil {
		t.Fatalf("CreateForum() error = %v", err)
	}
	if created.CategoryID != rootCategory.ID || created.Depth != 1 || created.Path != parent.Path+created.ID.String()+"/" {
		t.Fatalf("created = %+v, want child under parent category/path", created)
	}
}

// TestServiceCreateForumRejectsLinkParent verifies links cannot contain children.
func TestServiceCreateForumRejectsLinkParent(t *testing.T) {
	service, categories, forums, auth, _ := newTestService()
	actorID := uuid.New()
	auth.manage[domain.RootForumObjectID()] = true
	category := testCategory()
	categories.items[category.ID] = category
	parent := testForum(category.ID, nil, 0, "discord")
	parent.Kind = domain.ForumKindLink
	parent.ExternalURL = "https://discord.example"
	forums.items[parent.ID] = parent

	_, err := service.CreateForum(context.Background(), port.CreateForumCommand{ActorUserID: actorID, Forum: domain.Forum{CategoryID: category.ID, ParentForumID: &parent.ID, Key: "child", Slug: "child", Name: "Child"}})
	if !errors.Is(err, port.ErrInvalidMove) {
		t.Fatalf("CreateForum() error = %v, want %v", err, port.ErrInvalidMove)
	}
}

// TestServiceTreeFiltersInvisibleForumsAndCaches verifies tree visibility and cache.
func TestServiceTreeFiltersInvisibleForumsAndCaches(t *testing.T) {
	service, categories, forums, auth, cache := newTestService()
	category := testCategory()
	categories.items[category.ID] = category
	visible := testForum(category.ID, nil, 0, "visible")
	hidden := testForum(category.ID, nil, 1, "hidden")
	forums.items[visible.ID] = visible
	forums.items[hidden.ID] = hidden
	forums.stats[visible.ID] = domain.ForumStats{ForumID: visible.ID}
	forums.stats[hidden.ID] = domain.ForumStats{ForumID: hidden.ID}
	auth.visible[visible.ID] = true

	tree, err := service.Tree(context.Background(), uuid.Nil)
	if err != nil {
		t.Fatalf("Tree() error = %v", err)
	}
	if len(tree.Categories) != 1 || len(tree.Categories[0].Forums) != 1 || tree.Categories[0].Forums[0].Forum.ID != visible.ID {
		t.Fatalf("tree = %+v, want only visible forum", tree)
	}
	forums.items = map[uuid.UUID]domain.Forum{}
	cached, err := service.Tree(context.Background(), uuid.Nil)
	if err != nil {
		t.Fatalf("cached Tree() error = %v", err)
	}
	if len(cached.Categories) != 1 || cache.sets != 1 {
		t.Fatalf("cached tree=%+v sets=%d, want cache hit after one set", cached, cache.sets)
	}
}

// TestServiceReorderCategoriesRejectsInvalidItems verifies reorder validation.
func TestServiceReorderCategoriesRejectsInvalidItems(t *testing.T) {
	service, _, _, auth, _ := newTestService()
	auth.manage[domain.RootForumObjectID()] = true

	err := service.ReorderCategories(context.Background(), port.ReorderCategoriesCommand{ActorUserID: uuid.New()})
	var validation domain.ValidationError
	if !errors.As(err, &validation) {
		t.Fatalf("ReorderCategories() error = %v, want validation error", err)
	}
}

// TestServiceMoveForumRejectsDescendantParent verifies invalid tree moves.
func TestServiceMoveForumRejectsDescendantParent(t *testing.T) {
	service, categories, forums, auth, _ := newTestService()
	actorID := uuid.New()
	category := testCategory()
	categories.items[category.ID] = category
	parent := testForum(category.ID, nil, 0, "parent")
	child := testForum(category.ID, &parent.ID, 1, "child")
	child.Path = parent.Path + child.ID.String() + "/"
	forums.items[parent.ID] = parent
	forums.items[child.ID] = child
	auth.manage[parent.ID] = true

	_, err := service.MoveForum(context.Background(), port.MoveForumCommand{ActorUserID: actorID, ID: parent.ID, CategoryID: category.ID, ParentForumID: &child.ID, ExpectedVersion: parent.Version})
	if !errors.Is(err, port.ErrInvalidMove) {
		t.Fatalf("MoveForum() error = %v, want %v", err, port.ErrInvalidMove)
	}
}

// newTestService creates a forum service with in-memory fakes.
func newTestService() (Service, *memoryCategories, *memoryForums, *memoryAuthorizer, *memoryCache) {
	categories := &memoryCategories{items: map[uuid.UUID]domain.ForumCategory{}}
	forums := &memoryForums{items: map[uuid.UUID]domain.Forum{}, stats: map[uuid.UUID]domain.ForumStats{}}
	auth := &memoryAuthorizer{visible: map[uuid.UUID]bool{}, manage: map[uuid.UUID]bool{}}
	cache := &memoryCache{items: map[string]domain.ForumTree{}}
	service := NewService(Dependencies{Categories: categories, Forums: forums, Authorizer: auth, Cache: cache, Transactions: noopTx{}})
	return service, categories, forums, auth, cache
}

// memoryCategories stores categories in memory.
type memoryCategories struct {
	items map[uuid.UUID]domain.ForumCategory
}

// Create stores a category.
func (repository *memoryCategories) Create(_ context.Context, category domain.ForumCategory) (domain.ForumCategory, error) {
	repository.items[category.ID] = category
	return category, nil
}

// Update stores mutable category fields.
func (repository *memoryCategories) Update(_ context.Context, category domain.ForumCategory, expectedVersion uint64) (domain.ForumCategory, error) {
	current, ok := repository.items[category.ID]
	if !ok {
		return domain.ForumCategory{}, port.ErrNotFound
	}
	if current.Version != expectedVersion {
		return domain.ForumCategory{}, port.ErrPreconditionFailed
	}
	category.Version = expectedVersion + 1
	repository.items[category.ID] = category
	return category, nil
}

// FindByID returns one category.
func (repository *memoryCategories) FindByID(_ context.Context, id uuid.UUID) (domain.ForumCategory, error) {
	item, ok := repository.items[id]
	if !ok {
		return domain.ForumCategory{}, port.ErrNotFound
	}
	return item, nil
}

// List returns matching categories.
func (repository *memoryCategories) List(_ context.Context, filter port.CategoryFilter, _ pagination.Page) (pagination.Result[domain.ForumCategory], error) {
	items := []domain.ForumCategory{}
	for _, item := range repository.items {
		if filter.Status != "" && item.Status != filter.Status {
			continue
		}
		items = append(items, item)
	}
	return pagination.Result[domain.ForumCategory]{Items: items}, nil
}

// Delete removes a category.
func (repository *memoryCategories) Delete(_ context.Context, id uuid.UUID, _ uint64) error {
	delete(repository.items, id)
	return nil
}

// Reorder updates category order.
func (repository *memoryCategories) Reorder(_ context.Context, items []port.ReorderItem) error {
	for _, item := range items {
		category := repository.items[item.ID]
		category.DisplayOrder = item.DisplayOrder
		repository.items[item.ID] = category
	}
	return nil
}

// memoryForums stores forums in memory.
type memoryForums struct {
	items map[uuid.UUID]domain.Forum
	stats map[uuid.UUID]domain.ForumStats
}

// Create stores a forum.
func (repository *memoryForums) Create(_ context.Context, forum domain.Forum) (domain.Forum, error) {
	repository.items[forum.ID] = forum
	repository.stats[forum.ID] = domain.ForumStats{ForumID: forum.ID}
	return forum, nil
}

// Update stores mutable forum fields.
func (repository *memoryForums) Update(_ context.Context, forum domain.Forum, expectedVersion uint64) (domain.Forum, error) {
	forum.Version = expectedVersion + 1
	repository.items[forum.ID] = forum
	return forum, nil
}

// FindByID returns one forum.
func (repository *memoryForums) FindByID(_ context.Context, id uuid.UUID) (domain.Forum, error) {
	item, ok := repository.items[id]
	if !ok {
		return domain.Forum{}, port.ErrNotFound
	}
	return item, nil
}

// List returns matching forums.
func (repository *memoryForums) List(_ context.Context, _ port.ForumFilter, _ pagination.Page) (pagination.Result[domain.Forum], error) {
	items := make([]domain.Forum, 0, len(repository.items))
	for _, item := range repository.items {
		items = append(items, item)
	}
	return pagination.Result[domain.Forum]{Items: items}, nil
}

// ListTreeForums returns tree forums.
func (repository *memoryForums) ListTreeForums(context.Context) ([]domain.Forum, error) {
	items := make([]domain.Forum, 0, len(repository.items))
	for _, item := range repository.items {
		items = append(items, item)
	}
	return items, nil
}

// ListStats returns stats for ids.
func (repository *memoryForums) ListStats(_ context.Context, ids []uuid.UUID) (map[uuid.UUID]domain.ForumStats, error) {
	result := map[uuid.UUID]domain.ForumStats{}
	for _, id := range ids {
		result[id] = repository.stats[id]
	}
	return result, nil
}

// Move stores moved forum.
func (repository *memoryForums) Move(_ context.Context, forum domain.Forum, _ string, expectedVersion uint64) (domain.Forum, error) {
	forum.Version = expectedVersion + 1
	repository.items[forum.ID] = forum
	return forum, nil
}

// Delete removes a forum.
func (repository *memoryForums) Delete(_ context.Context, id uuid.UUID, _ uint64) error {
	delete(repository.items, id)
	return nil
}

// Reorder updates forum order.
func (repository *memoryForums) Reorder(_ context.Context, items []port.ReorderItem) error {
	for _, item := range items {
		forum := repository.items[item.ID]
		forum.DisplayOrder = item.DisplayOrder
		repository.items[item.ID] = forum
	}
	return nil
}

// memoryAuthorizer stores permission decisions in memory.
type memoryAuthorizer struct {
	visible map[uuid.UUID]bool
	manage  map[uuid.UUID]bool
}

// VisibleForums returns visible forum IDs.
func (authorizer *memoryAuthorizer) VisibleForums(_ context.Context, _ uuid.UUID, forumIDs []uuid.UUID) (map[uuid.UUID]bool, error) {
	result := map[uuid.UUID]bool{}
	for _, id := range forumIDs {
		result[id] = authorizer.visible[id]
	}
	return result, nil
}

// CanManageForum returns management decision.
func (authorizer *memoryAuthorizer) CanManageForum(_ context.Context, _ uuid.UUID, forumID uuid.UUID) (bool, error) {
	return authorizer.manage[forumID], nil
}

// memoryCache stores trees in memory.
type memoryCache struct {
	items map[string]domain.ForumTree
	sets  int
}

// GetTree returns a cached tree.
func (cache *memoryCache) GetTree(_ context.Context, key string) (domain.ForumTree, bool, error) {
	tree, ok := cache.items[key]
	return tree, ok, nil
}

// SetTree stores a tree.
func (cache *memoryCache) SetTree(_ context.Context, key string, tree domain.ForumTree, _ time.Duration) error {
	cache.items[key] = tree
	cache.sets++
	return nil
}

// ClearTree clears trees.
func (cache *memoryCache) ClearTree(context.Context) error {
	cache.items = map[string]domain.ForumTree{}
	return nil
}

// noopTx runs work without a real transaction.
type noopTx struct{}

// WithinTx runs fn.
func (noopTx) WithinTx(ctx context.Context, fn func(context.Context) error) error {
	return fn(ctx)
}

// testCategory returns a category.
func testCategory() domain.ForumCategory {
	return domain.ForumCategory{ID: uuid.New(), Key: "official", Name: "Official", Status: domain.CategoryStatusActive, Version: 1}
}

// testForum returns a forum.
func testForum(categoryID uuid.UUID, parentID *uuid.UUID, order int, key string) domain.Forum {
	id := uuid.New()
	path := "/" + id.String() + "/"
	return domain.Forum{ID: id, CategoryID: categoryID, ParentForumID: parentID, Kind: domain.ForumKindDiscussion, Key: domain.Key(key), Slug: domain.Slug(key), Name: key, DisplayOrder: order, Path: path, ThreadVisibilityMode: domain.ThreadVisibilityAllThreads, DefaultThreadStatus: domain.ThreadStatusOpen, Status: domain.ForumStatusActive, Version: 1}
}
