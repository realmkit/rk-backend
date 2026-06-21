package structure

import (
	"testing"

	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/module/forums/domain"
	"github.com/realmkit/rk-backend/module/forums/port"
)

// TestValidateReorderRejectsInvalidItems verifies reorder command validation.
func TestValidateReorderRejectsInvalidItems(t *testing.T) {
	service := Service{}
	if err := service.validateReorder(nil); err == nil {
		t.Fatalf("validateReorder(nil) error = nil, want validation error")
	}
	items := []port.ReorderItem{{ID: uuid.New(), DisplayOrder: 1}}
	if err := service.validateReorder(items); err != nil {
		t.Fatalf("validateReorder(valid) error = %v", err)
	}
}

// TestTreeCacheKeyPartitionsAnonymousAndUsers verifies tree cache key scope.
func TestTreeCacheKeyPartitionsAnonymousAndUsers(t *testing.T) {
	userID := uuid.New()
	if got := treeCacheKey(uuid.Nil); got != "forums:tree:v1:anonymous" {
		t.Fatalf("treeCacheKey(nil) = %q", got)
	}
	if got := treeCacheKey(userID); got != "forums:tree:v1:user:"+userID.String() {
		t.Fatalf("treeCacheKey(user) = %q", got)
	}
}

// TestSortedNodesOrdersByDisplayOrder verifies stable tree ordering.
func TestSortedNodesOrdersByDisplayOrder(t *testing.T) {
	later := domain.ForumNode{Forum: domain.Forum{ID: uuid.New(), DisplayOrder: 20}}
	earlier := domain.ForumNode{Forum: domain.Forum{ID: uuid.New(), DisplayOrder: 10}}
	nodes := sortedNodes([]domain.ForumNode{later, earlier})
	if nodes[0].Forum.ID != earlier.Forum.ID {
		t.Fatalf("sortedNodes()[0] = %#v, want earlier", nodes[0])
	}
}
