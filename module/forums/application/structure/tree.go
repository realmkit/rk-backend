package structure

import (
	"context"
	"sort"
	"strings"

	"github.com/google/uuid"
	"github.com/niflaot/gamehub-go/module/forums/domain"
	"github.com/niflaot/gamehub-go/module/forums/port"
)

// prepareForum validates and fills tree fields for a new forum.
func (service Service) prepareForum(ctx context.Context, forum domain.Forum) (domain.Forum, error) {
	if forum.ID == uuid.Nil {
		forum.ID = uuid.New()
	}
	if _, err := service.categories.FindByID(ctx, forum.CategoryID); err != nil {
		return domain.Forum{}, err
	}
	parentPath := "/"
	if forum.ParentForumID != nil {
		parent, err := service.forums.FindByID(ctx, *forum.ParentForumID)
		if err != nil {
			return domain.Forum{}, err
		}
		if parent.Kind == domain.ForumKindLink {
			return domain.Forum{}, port.ErrInvalidMove
		}
		forum.CategoryID = parent.CategoryID
		forum.Depth = parent.Depth + 1
		parentPath = parent.Path
	}
	forum.Path = parentPath + forum.ID.String() + "/"
	if err := forum.Validate(); err != nil {
		return domain.Forum{}, err
	}
	return forum, nil
}

// moveTarget returns the forum state after a move.
func (service Service) moveTarget(ctx context.Context, current domain.Forum, command port.MoveForumCommand) (domain.Forum, error) {
	target := current
	target.CategoryID = command.CategoryID
	target.ParentForumID = command.ParentForumID
	target.DisplayOrder = command.DisplayOrder
	parentPath := "/"
	if target.CategoryID == uuid.Nil {
		return domain.Forum{}, port.ErrInvalidMove
	}
	if _, err := service.categories.FindByID(ctx, target.CategoryID); err != nil {
		return domain.Forum{}, err
	}
	if command.ParentForumID != nil {
		if *command.ParentForumID == current.ID {
			return domain.Forum{}, port.ErrInvalidMove
		}
		parent, err := service.forums.FindByID(ctx, *command.ParentForumID)
		if err != nil {
			return domain.Forum{}, err
		}
		if strings.HasPrefix(parent.Path, current.Path) || parent.Kind == domain.ForumKindLink {
			return domain.Forum{}, port.ErrInvalidMove
		}
		target.CategoryID = parent.CategoryID
		target.Depth = parent.Depth + 1
		parentPath = parent.Path
	} else {
		target.Depth = 0
	}
	target.Path = parentPath + target.ID.String() + "/"
	if err := target.Validate(); err != nil {
		return domain.Forum{}, err
	}
	return target, nil
}

// loadTree builds the visible forum tree.
func (service Service) loadTree(ctx context.Context, actorUserID uuid.UUID) (domain.ForumTree, error) {
	categories, err := service.categories.List(ctx, port.CategoryFilter{Status: domain.CategoryStatusActive}, port.Page{Limit: 1000})
	if err != nil {
		return domain.ForumTree{}, err
	}
	forums, err := service.forums.ListTreeForums(ctx)
	if err != nil {
		return domain.ForumTree{}, err
	}
	forumIDs := make([]uuid.UUID, 0, len(forums))
	for _, forum := range forums {
		forumIDs = append(forumIDs, forum.ID)
	}
	if service.authorizer == nil {
		return domain.ForumTree{}, port.ErrForbidden
	}
	visible, err := service.authorizer.VisibleForums(ctx, actorUserID, forumIDs)
	if err != nil {
		return domain.ForumTree{}, err
	}
	stats, err := service.forums.ListStats(ctx, forumIDs)
	if err != nil {
		return domain.ForumTree{}, err
	}
	return treeFrom(categories.Items, forums, stats, visible), nil
}

// treeFrom nests visible forums under visible categories.
func treeFrom(
	categories []domain.ForumCategory,
	forums []domain.Forum,
	stats map[uuid.UUID]domain.ForumStats,
	visible map[uuid.UUID]bool,
) domain.ForumTree {
	nodes := map[uuid.UUID]domain.ForumNode{}
	byParent := map[uuid.UUID][]uuid.UUID{}
	rootsByCategory := map[uuid.UUID][]uuid.UUID{}
	for _, forum := range forums {
		if !visible[forum.ID] {
			continue
		}
		nodes[forum.ID] = domain.ForumNode{Forum: forum, Stats: stats[forum.ID], Children: []domain.ForumNode{}}
		if forum.ParentForumID != nil {
			byParent[*forum.ParentForumID] = append(byParent[*forum.ParentForumID], forum.ID)
			continue
		}
		rootsByCategory[forum.CategoryID] = append(rootsByCategory[forum.CategoryID], forum.ID)
	}
	categoryNodes := make([]domain.CategoryNode, 0, len(categories))
	for _, category := range categories {
		forums := buildNodes(rootsByCategory[category.ID], nodes, byParent)
		if len(forums) == 0 {
			continue
		}
		categoryNodes = append(categoryNodes, domain.CategoryNode{Category: category, Forums: forums})
	}
	sort.SliceStable(categoryNodes, func(left int, right int) bool {
		if categoryNodes[left].Category.DisplayOrder != categoryNodes[right].Category.DisplayOrder {
			return categoryNodes[left].Category.DisplayOrder < categoryNodes[right].Category.DisplayOrder
		}
		return categoryNodes[left].Category.ID.String() < categoryNodes[right].Category.ID.String()
	})
	return domain.ForumTree{Categories: categoryNodes}
}

// buildNodes recursively builds forum nodes.
func buildNodes(ids []uuid.UUID, nodes map[uuid.UUID]domain.ForumNode, byParent map[uuid.UUID][]uuid.UUID) []domain.ForumNode {
	result := make([]domain.ForumNode, 0, len(ids))
	for _, id := range ids {
		node, ok := nodes[id]
		if !ok {
			continue
		}
		node.Children = buildNodes(byParent[id], nodes, byParent)
		result = append(result, node)
	}
	return sortedNodes(result)
}

// sortedNodes returns nodes ordered for display.
func sortedNodes(nodes []domain.ForumNode) []domain.ForumNode {
	sort.SliceStable(nodes, func(left int, right int) bool {
		if nodes[left].Forum.DisplayOrder != nodes[right].Forum.DisplayOrder {
			return nodes[left].Forum.DisplayOrder < nodes[right].Forum.DisplayOrder
		}
		return nodes[left].Forum.ID.String() < nodes[right].Forum.ID.String()
	})
	return nodes
}

// treeCacheKey returns cache key for actor visibility.
func treeCacheKey(actorUserID uuid.UUID) string {
	if actorUserID == uuid.Nil {
		return "forums:tree:v1:anonymous"
	}
	return "forums:tree:v1:user:" + actorUserID.String()
}
