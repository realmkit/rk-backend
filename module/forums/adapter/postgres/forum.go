package postgres

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	forumsauthz "github.com/niflaot/gamehub-go/module/forums/adapter/postgres/authz"
	forumsinteraction "github.com/niflaot/gamehub-go/module/forums/adapter/postgres/interaction"
	forumsoperations "github.com/niflaot/gamehub-go/module/forums/adapter/postgres/operations"
	"github.com/niflaot/gamehub-go/module/forums/domain"
	"github.com/niflaot/gamehub-go/module/forums/port"
	"github.com/niflaot/gamehub-go/pkg/orm"
	"github.com/niflaot/gamehub-go/pkg/pagination"
	"gorm.io/gorm"
)

// ForumRepository stores forums in PostgreSQL.
type ForumRepository struct {
	store orm.Store
}

// NewForumRepository creates a forum repository.
func NewForumRepository(store orm.Store) ForumRepository {
	return ForumRepository{store: store}
}

// Create stores a forum and creates its stats row.
func (repository ForumRepository) Create(ctx context.Context, forum domain.Forum) (domain.Forum, error) {
	model := forumModelFromDomain(forum)
	if err := repository.store.DB(ctx).Create(&model).Error; err != nil {
		return domain.Forum{}, port.ErrConflict
	}
	stats := StatsModel{ForumID: model.ID.ID, UpdatedAt: time.Now().UTC()}
	if err := repository.store.DB(ctx).Create(&stats).Error; err != nil {
		return domain.Forum{}, err
	}
	return forumFromModel(model), nil
}

// Update stores mutable forum fields.
func (repository ForumRepository) Update(ctx context.Context, forum domain.Forum, expectedVersion uint64) (domain.Forum, error) {
	result := repository.store.DB(ctx).Model(&ForumModel{}).Where("id = ? AND version = ?", forum.ID, expectedVersion).Updates(forumUpdates(forum, expectedVersion))
	if result.Error != nil {
		return domain.Forum{}, result.Error
	}
	if result.RowsAffected == 0 {
		return domain.Forum{}, port.ErrPreconditionFailed
	}
	return repository.FindByID(ctx, forum.ID)
}

// FindByID returns one forum.
func (repository ForumRepository) FindByID(ctx context.Context, id uuid.UUID) (domain.Forum, error) {
	var model ForumModel
	if err := repository.store.DB(ctx).First(&model, "id = ?", id).Error; err != nil {
		return domain.Forum{}, mapError(err)
	}
	return forumFromModel(model), nil
}

// List returns matching forums.
func (repository ForumRepository) List(ctx context.Context, filter port.ForumFilter, page pagination.Page) (pagination.Result[domain.Forum], error) {
	query := repository.store.DB(ctx).Model(&ForumModel{}).Order("path asc, display_order asc, id asc").Limit(page.Limit + 1)
	query = applyForumFilter(query, filter)
	var models []ForumModel
	if err := query.Find(&models).Error; err != nil {
		return pagination.Result[domain.Forum]{}, err
	}
	return forumPage(models, page.Limit), nil
}

// ListTreeForums returns forums used by tree reads.
func (repository ForumRepository) ListTreeForums(ctx context.Context) ([]domain.Forum, error) {
	var models []ForumModel
	if err := repository.store.DB(ctx).Model(&ForumModel{}).Where("status = ?", domain.ForumStatusActive).Order("path asc, display_order asc, id asc").Find(&models).Error; err != nil {
		return nil, err
	}
	items := make([]domain.Forum, 0, len(models))
	for _, model := range models {
		items = append(items, forumFromModel(model))
	}
	return items, nil
}

// ListStats returns stats for forum ids.
func (repository ForumRepository) ListStats(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]domain.ForumStats, error) {
	if len(ids) == 0 {
		return map[uuid.UUID]domain.ForumStats{}, nil
	}
	var models []StatsModel
	if err := repository.store.DB(ctx).Find(&models, "forum_id IN ?", ids).Error; err != nil {
		return nil, err
	}
	stats := map[uuid.UUID]domain.ForumStats{}
	for _, model := range models {
		stats[model.ForumID] = statsFromModel(model)
	}
	return stats, nil
}

// Move changes a forum path and descendant paths.
func (repository ForumRepository) Move(ctx context.Context, forum domain.Forum, oldPath string, expectedVersion uint64) (domain.Forum, error) {
	result := repository.store.DB(ctx).Model(&ForumModel{}).Where("id = ? AND version = ?", forum.ID, expectedVersion).Updates(forumUpdates(forum, expectedVersion))
	if result.Error != nil {
		return domain.Forum{}, result.Error
	}
	if result.RowsAffected == 0 {
		return domain.Forum{}, port.ErrPreconditionFailed
	}
	if oldPath != forum.Path {
		var descendants []ForumModel
		if err := repository.store.DB(ctx).Where("path LIKE ? AND id <> ?", oldPath+"%", forum.ID).Find(&descendants).Error; err != nil {
			return domain.Forum{}, err
		}
		for _, descendant := range descendants {
			nextPath := strings.Replace(descendant.Path, oldPath, forum.Path, 1)
			depthDelta := strings.Count(nextPath, "/") - strings.Count(descendant.Path, "/")
			if err := repository.store.DB(ctx).Model(&ForumModel{}).Where("id = ?", descendant.ID.ID).Updates(map[string]any{"path": nextPath, "category_id": forum.CategoryID, "depth": descendant.Depth + depthDelta}).Error; err != nil {
				return domain.Forum{}, err
			}
		}
	}
	return repository.FindByID(ctx, forum.ID)
}

// Delete soft deletes one forum.
func (repository ForumRepository) Delete(ctx context.Context, id uuid.UUID, expectedVersion uint64) error {
	result := repository.store.DB(ctx).Where("id = ? AND version = ?", id, expectedVersion).Delete(&ForumModel{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return port.ErrPreconditionFailed
	}
	return nil
}

// Reorder updates forum display order.
func (repository ForumRepository) Reorder(ctx context.Context, items []port.ReorderItem) error {
	for _, item := range items {
		if err := repository.store.DB(ctx).Model(&ForumModel{}).Where("id = ?", item.ID).Update("display_order", item.DisplayOrder).Error; err != nil {
			return err
		}
	}
	return nil
}

// applyForumFilter applies forum filters.
func applyForumFilter(query *gorm.DB, filter port.ForumFilter) *gorm.DB {
	if filter.CategoryID != uuid.Nil {
		query = query.Where("category_id = ?", filter.CategoryID)
	}
	if filter.ParentForumID != nil {
		query = query.Where("parent_forum_id = ?", *filter.ParentForumID)
	}
	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}
	return query
}

// forumUpdates returns update fields.
func forumUpdates(forum domain.Forum, expectedVersion uint64) map[string]any {
	return map[string]any{"category_id": forum.CategoryID, "parent_forum_id": forum.ParentForumID, "kind": string(forum.Kind), "key": string(forum.Key), "slug": string(forum.Slug), "name": forum.Name, "description": forum.Description, "display_order": forum.DisplayOrder, "path": forum.Path, "depth": forum.Depth, "external_url": forum.ExternalURL, "icon_asset_id": forum.IconAssetID, "thread_visibility_mode": string(forum.ThreadVisibilityMode), "max_sticky_threads": forum.MaxStickyThreads, "default_thread_status": string(forum.DefaultThreadStatus), "author_post_edit_window_seconds": forum.AuthorPostEditWindowSeconds, "author_post_delete_window_seconds": forum.AuthorPostDeleteWindowSeconds, "status": string(forum.Status), "version": expectedVersion + 1}
}

// forumPage maps models into a page.
func forumPage(models []ForumModel, limit int) pagination.Result[domain.Forum] {
	next := ""
	if len(models) > limit {
		next = models[limit-1].ID.ID.String()
		models = models[:limit]
	}
	items := make([]domain.Forum, 0, len(models))
	for _, model := range models {
		items = append(items, forumFromModel(model))
	}
	return pagination.Result[domain.Forum]{Items: items, NextCursor: next}
}

// statsFromModel maps persistence stats to domain.
func statsFromModel(model StatsModel) domain.ForumStats {
	return domain.ForumStats{
		ForumID:                model.ForumID,
		ThreadCount:            model.ThreadCount,
		VisibleThreadCount:     model.VisibleThreadCount,
		PostCount:              model.PostCount,
		VisiblePostCount:       model.VisiblePostCount,
		LatestThreadID:         model.LatestThreadID,
		LatestPostID:           model.LatestPostID,
		LatestPostAuthorUserID: model.LatestPostAuthorUserID,
		LatestPostAt:           model.LatestPostAt,
		UpdatedAt:              model.UpdatedAt,
	}
}

// Ensure ForumRepository implements port.ForumRepository.
var _ port.ForumRepository = ForumRepository{}

// VisibilityAuthorizer resolves forum permissions from authorization tuples.
type VisibilityAuthorizer = forumsauthz.VisibilityAuthorizer

// NewVisibilityAuthorizer creates a visibility authorizer.
func NewVisibilityAuthorizer(store orm.Store) VisibilityAuthorizer {
	return forumsauthz.NewVisibilityAuthorizer(store)
}

// InteractionRepository stores forum interactions in PostgreSQL.
type InteractionRepository = forumsinteraction.Repository

// NewInteractionRepository creates an interaction repository.
func NewInteractionRepository(store orm.Store) InteractionRepository {
	return forumsinteraction.NewRepository(store)
}

// OperationsRepository runs forum search, repair, and counter flushes in PostgreSQL.
type OperationsRepository = forumsoperations.Repository

// NewOperationsRepository creates an operations repository.
func NewOperationsRepository(store orm.Store) OperationsRepository {
	return forumsoperations.NewRepository(store)
}
