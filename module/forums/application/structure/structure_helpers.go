package structure

import (
	"context"
	"strconv"

	"github.com/google/uuid"
	"github.com/niflaot/gamehub-go/module/forums/domain"
	"github.com/niflaot/gamehub-go/module/forums/port"
)

// Tree returns the visible forum tree.
func (service Service) Tree(
	ctx context.Context,
	actorUserID uuid.UUID,
) (domain.ForumTree, error) {
	key := treeCacheKey(actorUserID)
	if service.cache != nil {
		if cached, ok, err := service.cache.GetTree(ctx, key); err == nil && ok {
			return cached, nil
		}
	}
	tree, err := service.loadTree(ctx, actorUserID)
	if err != nil {
		return domain.ForumTree{}, err
	}
	if service.cache != nil {
		if err := service.cache.SetTree(ctx, key, tree, treeCacheTTL); err != nil {
			return tree, nil
		}
	}
	return tree, nil
}

// requireManage verifies structure-management permission.
func (service Service) requireManage(
	ctx context.Context,
	actorUserID uuid.UUID,
	forumID uuid.UUID,
) error {
	if service.authorizer == nil {
		return nil
	}
	allowed, err := service.authorizer.CanManageForum(ctx, actorUserID, forumID)
	if err != nil {
		return err
	}
	if !allowed {
		return port.ErrForbidden
	}
	return nil
}

// clearTree clears cached trees when a cache is configured.
func (service Service) clearTree(ctx context.Context) error {
	if service.cache == nil {
		return nil
	}
	return service.cache.ClearTree(ctx)
}

// validateReorder validates reorder items.
func (service Service) validateReorder(items []port.ReorderItem) error {
	var violations []domain.Violation
	if len(items) == 0 {
		violations = domain.AppendViolation(violations, "items", "must contain at least one item")
	}
	for index, item := range items {
		field := "items[" + strconv.Itoa(index) + "]"
		if item.ID == uuid.Nil {
			violations = domain.AppendViolation(violations, field+".id", "is required")
		}
		violations = append(
			violations,
			domain.ValidateDisplayOrder(field+".display_order", item.DisplayOrder)...,
		)
	}
	return domain.NewValidationError(violations)
}
