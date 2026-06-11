package operations

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/niflaot/gamehub-go/module/forums/domain"
	"github.com/niflaot/gamehub-go/module/forums/port"
	eventtesting "github.com/niflaot/gamehub-go/pkg/events/testing"
	"github.com/niflaot/gamehub-go/pkg/pagination"
)

// TestSearchValidatesAndScopesVisibleForums covers search validation and visibility.
func TestSearchValidatesAndScopesVisibleForums(t *testing.T) {
	forumID := uuid.New()
	operations := &operationsRepoFake{}
	service := NewService(Dependencies{
		Forums:     forumRepoFake{forums: []domain.Forum{{ID: forumID}}},
		Operations: operations,
		Authorizer: visibilityFake{visible: map[uuid.UUID]bool{forumID: true}},
	})

	if _, err := service.Search(context.Background(), port.SearchCommand{Query: "x"}, pagination.Page{}); err == nil {
		t.Fatalf("expected short query to fail validation")
	}
	result, err := service.Search(
		context.Background(),
		port.SearchCommand{Query: "  castle  "},
		pagination.Page{Limit: 10},
	)
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}
	if len(result.Items) != 1 {
		t.Fatalf("search result count = %d, want 1", len(result.Items))
	}
	if operations.searchFilter.Query != "castle" || operations.searchFilter.ForumIDs[0] != forumID {
		t.Fatalf("search filter = %#v", operations.searchFilter)
	}

	service.authorizer = visibilityFake{}
	empty, err := service.Search(
		context.Background(),
		port.SearchCommand{Query: "castle", ForumID: forumID},
		pagination.Page{Limit: 10},
	)
	if err != nil {
		t.Fatalf("empty Search() error = %v", err)
	}
	if len(empty.Items) != 0 {
		t.Fatalf("empty search count = %d, want 0", len(empty.Items))
	}
}

// TestRepairAndCacheOperationsPublishEvents covers operational commands.
func TestRepairAndCacheOperationsPublishEvents(t *testing.T) {
	threadID := uuid.New()
	events := &eventtesting.PublisherRecorder{}
	operations := &operationsRepoFake{
		report: domain.CounterDriftReport{
			Mismatches: []domain.CounterDrift{{ObjectType: "thread", ObjectID: threadID, Field: "view_count"}},
			Repaired:   true,
		},
	}
	cache := &readCacheFake{
		views: map[string]int64{
			threadID.String(): 2,
			uuid.NewString():  3,
			"bad-id":          10,
		},
	}
	service := NewService(Dependencies{
		Operations: operations,
		Cache:      cache,
		Events:     events,
	})

	if _, err := service.VerifyStats(context.Background()); err != nil {
		t.Fatalf("VerifyStats() error = %v", err)
	}
	if _, err := service.VerifyLikes(context.Background()); err != nil {
		t.Fatalf("VerifyLikes() error = %v", err)
	}
	if _, err := service.RebuildStats(context.Background()); err != nil {
		t.Fatalf("RebuildStats() error = %v", err)
	}
	if _, err := service.RebuildLikes(context.Background()); err != nil {
		t.Fatalf("RebuildLikes() error = %v", err)
	}

	total, err := service.FlushThreadViews(context.Background())
	if err != nil {
		t.Fatalf("FlushThreadViews() error = %v", err)
	}
	if total != 5 {
		t.Fatalf("flushed views = %d, want 5", total)
	}
	if len(operations.appliedViews) != 2 {
		t.Fatalf("applied thread count = %d, want 2", len(operations.appliedViews))
	}
	if err := service.ClearReadCache(context.Background()); err != nil {
		t.Fatalf("ClearReadCache() error = %v", err)
	}
	if !cache.cleared {
		t.Fatalf("expected cache clear")
	}
	if len(events.Drafts()) != 3 {
		t.Fatalf("operation event count = %d, want 3", len(events.Drafts()))
	}
}

type forumRepoFake struct {
	forums []domain.Forum
}

func (fake forumRepoFake) Create(context.Context, domain.Forum) (domain.Forum, error) {
	return domain.Forum{}, nil
}

func (fake forumRepoFake) Update(context.Context, domain.Forum, uint64) (domain.Forum, error) {
	return domain.Forum{}, nil
}

func (fake forumRepoFake) FindByID(context.Context, uuid.UUID) (domain.Forum, error) {
	return domain.Forum{}, nil
}

func (fake forumRepoFake) List(
	context.Context,
	port.ForumFilter,
	pagination.Page,
) (pagination.Result[domain.Forum], error) {
	return pagination.Result[domain.Forum]{Items: fake.forums}, nil
}

func (fake forumRepoFake) ListTreeForums(context.Context) ([]domain.Forum, error) {
	return nil, nil
}

func (fake forumRepoFake) ListStats(context.Context, []uuid.UUID) (map[uuid.UUID]domain.ForumStats, error) {
	return nil, nil
}

func (fake forumRepoFake) Move(context.Context, domain.Forum, string, uint64) (domain.Forum, error) {
	return domain.Forum{}, nil
}

func (fake forumRepoFake) Delete(context.Context, uuid.UUID, uint64) error {
	return nil
}

func (fake forumRepoFake) Reorder(context.Context, []port.ReorderItem) error {
	return nil
}

type visibilityFake struct {
	visible map[uuid.UUID]bool
}

func (fake visibilityFake) VisibleForums(
	context.Context,
	uuid.UUID,
	[]uuid.UUID,
) (map[uuid.UUID]bool, error) {
	return fake.visible, nil
}

func (fake visibilityFake) CanManageForum(context.Context, uuid.UUID, uuid.UUID) (bool, error) {
	return false, nil
}

func (fake visibilityFake) CanCreateThread(context.Context, uuid.UUID, uuid.UUID) (bool, error) {
	return false, nil
}

func (fake visibilityFake) CanReply(context.Context, uuid.UUID, uuid.UUID) (bool, error) {
	return false, nil
}

func (fake visibilityFake) CanLikePosts(context.Context, uuid.UUID, uuid.UUID) (bool, error) {
	return false, nil
}

func (fake visibilityFake) CanManageThreads(context.Context, uuid.UUID, uuid.UUID) (bool, error) {
	return false, nil
}

func (fake visibilityFake) CanManagePosts(context.Context, uuid.UUID, uuid.UUID) (bool, error) {
	return false, nil
}

type operationsRepoFake struct {
	searchFilter port.SearchFilter
	report       domain.CounterDriftReport
	appliedViews map[uuid.UUID]int64
}

func (fake *operationsRepoFake) Search(
	_ context.Context,
	filter port.SearchFilter,
	_ pagination.Page,
) (pagination.Result[domain.SearchResult], error) {
	fake.searchFilter = filter
	return pagination.Result[domain.SearchResult]{Items: []domain.SearchResult{{Type: "thread"}}}, nil
}

func (fake *operationsRepoFake) VerifyStats(context.Context) (domain.CounterDriftReport, error) {
	return fake.report, nil
}

func (fake *operationsRepoFake) RebuildStats(context.Context) (domain.CounterDriftReport, error) {
	return fake.report, nil
}

func (fake *operationsRepoFake) VerifyLikes(context.Context) (domain.CounterDriftReport, error) {
	return fake.report, nil
}

func (fake *operationsRepoFake) RebuildLikes(context.Context) (domain.CounterDriftReport, error) {
	return fake.report, nil
}

func (fake *operationsRepoFake) ApplyThreadViews(_ context.Context, increments map[uuid.UUID]int64) error {
	fake.appliedViews = increments
	return nil
}

type readCacheFake struct {
	views   map[string]int64
	cleared bool
}

func (fake *readCacheFake) GetTree(context.Context, string) (domain.ForumTree, bool, error) {
	return domain.ForumTree{}, false, nil
}

func (fake *readCacheFake) SetTree(context.Context, string, domain.ForumTree, time.Duration) error {
	return nil
}

func (fake *readCacheFake) ClearTree(context.Context) error {
	return nil
}

func (fake *readCacheFake) GetLatestPosts(
	context.Context,
	string,
) (pagination.Result[domain.LatestPostSummary], bool, error) {
	return pagination.Result[domain.LatestPostSummary]{}, false, nil
}

func (fake *readCacheFake) SetLatestPosts(
	context.Context,
	string,
	pagination.Result[domain.LatestPostSummary],
	time.Duration,
) error {
	return nil
}

func (fake *readCacheFake) ClearLatestPosts(context.Context) error {
	return nil
}

func (fake *readCacheFake) GetMostLikedPosts(
	context.Context,
	string,
) (pagination.Result[domain.MostLikedPost], bool, error) {
	return pagination.Result[domain.MostLikedPost]{}, false, nil
}

func (fake *readCacheFake) SetMostLikedPosts(
	context.Context,
	string,
	pagination.Result[domain.MostLikedPost],
	time.Duration,
) error {
	return nil
}

func (fake *readCacheFake) ClearMostLikedPosts(context.Context) error {
	return nil
}

func (fake *readCacheFake) IncrementThreadView(context.Context, string) error {
	return nil
}

func (fake *readCacheFake) DrainThreadViews(context.Context) (map[string]int64, error) {
	return fake.views, nil
}

func (fake *readCacheFake) ClearAll(context.Context) error {
	fake.cleared = true
	return nil
}
