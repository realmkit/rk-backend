package structure

import (
	"strconv"
	"testing"

	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/module/forums/domain"
)

// benchmarkStructureTree stores the structure tree benchmark result.
var benchmarkStructureTree domain.ForumTree

// BenchmarkTreeFrom measures category/forum tree assembly and stable sibling ordering.
func BenchmarkTreeFrom(b *testing.B) {
	categories, forums, stats, visible := benchmarkTreeFixture()

	b.ReportAllocs()
	for index := 0; index < b.N; index++ {
		benchmarkStructureTree = treeFrom(categories, forums, stats, visible)
	}
}

// benchmarkTreeFixture builds a representative visible forum tree.
func benchmarkTreeFixture() (
	[]domain.ForumCategory,
	[]domain.Forum,
	map[uuid.UUID]domain.ForumStats,
	map[uuid.UUID]bool,
) {
	categories := make([]domain.ForumCategory, 0, 5)
	forums := make([]domain.Forum, 0, 100)
	stats := make(map[uuid.UUID]domain.ForumStats, 100)
	visible := make(map[uuid.UUID]bool, 100)
	for categoryIndex := 0; categoryIndex < 5; categoryIndex++ {
		categoryID := uuid.New()
		categories = append(categories, domain.ForumCategory{
			ID:           categoryID,
			Key:          domain.Key("category_" + strconv.Itoa(categoryIndex)),
			Name:         "Category " + strconv.Itoa(categoryIndex),
			DisplayOrder: categoryIndex,
			Status:       domain.CategoryStatusActive,
			Version:      1,
		})
		for rootIndex := 0; rootIndex < 5; rootIndex++ {
			root := benchmarkForum(categoryID, nil, rootIndex, "forum_"+strconv.Itoa(categoryIndex)+"_"+strconv.Itoa(rootIndex))
			forums = append(forums, root)
			stats[root.ID] = domain.ForumStats{ForumID: root.ID, ThreadCount: int64(rootIndex + 1)}
			visible[root.ID] = true
			for childIndex := 0; childIndex < 3; childIndex++ {
				parentID := root.ID
				child := benchmarkForum(
					categoryID,
					&parentID,
					childIndex,
					"forum_"+strconv.Itoa(categoryIndex)+"_"+strconv.Itoa(rootIndex)+"_"+strconv.Itoa(childIndex),
				)
				forums = append(forums, child)
				stats[child.ID] = domain.ForumStats{ForumID: child.ID, ThreadCount: int64(childIndex + 1)}
				visible[child.ID] = true
			}
		}
	}
	return categories, forums, stats, visible
}

// benchmarkForum returns one forum node for tree benchmarks.
func benchmarkForum(categoryID uuid.UUID, parentID *uuid.UUID, order int, key string) domain.Forum {
	id := uuid.New()
	return domain.Forum{
		ID:                   id,
		CategoryID:           categoryID,
		ParentForumID:        parentID,
		Kind:                 domain.ForumKindDiscussion,
		Key:                  domain.Key(key),
		Slug:                 domain.Slug(key),
		Name:                 key,
		DisplayOrder:         order,
		Path:                 "/" + id.String() + "/",
		ThreadVisibilityMode: domain.ThreadVisibilityAllThreads,
		DefaultThreadStatus:  domain.ThreadStatusOpen,
		Status:               domain.ForumStatusActive,
		Version:              1,
	}
}
