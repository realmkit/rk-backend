package application

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/module/forums/domain"
	"github.com/realmkit/rk-backend/module/forums/port"
	"github.com/realmkit/rk-backend/pkg/pagination"
)

// benchmarkForumTree stores the forum facade tree benchmark result.
var benchmarkForumTree domain.ForumTree

// benchmarkCreatedThread stores the create-thread benchmark thread result.
var benchmarkCreatedThread domain.Thread

// benchmarkLatestPosts stores the latest-post benchmark result.
var benchmarkLatestPosts pagination.Result[domain.LatestPostSummary]

// BenchmarkServiceTreeCacheHit measures the facade tree read path after the read cache is warm.
func BenchmarkServiceTreeCacheHit(b *testing.B) {
	service, categories, forums, authorizer, _ := newTestService()
	category := testCategory()
	parent := testForum(category.ID, nil, 1, "general")
	child := testForum(category.ID, &parent.ID, 2, "support")
	categories.items[category.ID] = category
	forums.items[parent.ID] = parent
	forums.items[child.ID] = child
	forums.stats[parent.ID] = domain.ForumStats{ForumID: parent.ID, ThreadCount: 8, PostCount: 24}
	forums.stats[child.ID] = domain.ForumStats{ForumID: child.ID, ThreadCount: 3, PostCount: 7}
	authorizer.visible[parent.ID] = true
	authorizer.visible[child.ID] = true
	ctx := context.Background()

	if _, err := service.Tree(ctx, benchmarkUserID); err != nil {
		b.Fatalf("warm Tree() error = %v", err)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for index := 0; index < b.N; index++ {
		tree, err := service.Tree(ctx, benchmarkUserID)
		if err != nil {
			b.Fatalf("Tree() error = %v", err)
		}
		benchmarkForumTree = tree
	}
}

// BenchmarkCreateThread measures thread creation orchestration over the facade service.
func BenchmarkCreateThread(b *testing.B) {
	service, categories, forums, threads, posts, auth, _ := newContentTestService()
	actorID := uuid.New()
	category := testCategory()
	forum := testForum(category.ID, nil, 0, "benchmark")
	categories.items[category.ID] = category
	forums.items[forum.ID] = forum
	auth.create[forum.ID] = true
	command := port.CreateThreadCommand{
		ActorUserID:         actorID,
		ForumID:             forum.ID,
		Title:               "Benchmark thread",
		Slug:                "benchmark-thread",
		ContentDocumentJSON: []byte(`{"type":"doc","content":[{"type":"text","text":"Benchmark"}]}`),
		ContentText:         "Benchmark",
	}
	ctx := context.Background()

	b.ReportAllocs()
	for index := 0; index < b.N; index++ {
		thread, post, err := service.CreateThread(ctx, command)
		if err != nil {
			b.Fatalf("CreateThread() error = %v", err)
		}
		delete(threads.items, thread.ID)
		delete(posts.items, post.ID)
		benchmarkCreatedThread = thread
	}
}

// BenchmarkLatestPostsCacheHit measures the latest-post widget path after cache warming.
func BenchmarkLatestPostsCacheHit(b *testing.B) {
	service, _, forums, _, _, auth, interactions := newContentTestService()
	actorID := uuid.New()
	forum := testForum(uuid.New(), nil, 0, "latest")
	forums.items[forum.ID] = forum
	auth.visible[forum.ID] = true
	interactions.latest = []domain.LatestPostSummary{
		{
			ForumID:      forum.ID,
			ThreadID:     uuid.New(),
			PostID:       uuid.New(),
			AuthorUserID: uuid.New(),
			Sequence:     1,
			ThreadTitle:  "Cached",
		},
	}
	ctx := context.Background()
	page := pagination.Page{Limit: 10}
	if _, err := service.ListLatestPosts(ctx, actorID, uuid.Nil, page); err != nil {
		b.Fatalf("warm ListLatestPosts() error = %v", err)
	}
	interactions.latest = nil

	b.ReportAllocs()
	b.ResetTimer()
	for index := 0; index < b.N; index++ {
		result, err := service.ListLatestPosts(ctx, actorID, uuid.Nil, page)
		if err != nil {
			b.Fatalf("ListLatestPosts() error = %v", err)
		}
		benchmarkLatestPosts = result
	}
}

// benchmarkUserID is the stable actor used by facade benchmarks.
var benchmarkUserID = testForum(testCategory().ID, nil, 1, "actor-scope").ID
