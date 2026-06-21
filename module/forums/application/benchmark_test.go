package application

import (
	"context"
	"testing"

	"github.com/realmkit/rk-backend/module/forums/domain"
)

// benchmarkForumTree stores the forum facade tree benchmark result.
var benchmarkForumTree domain.ForumTree

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

// benchmarkUserID is the stable actor used by facade benchmarks.
var benchmarkUserID = testForum(testCategory().ID, nil, 1, "actor-scope").ID
