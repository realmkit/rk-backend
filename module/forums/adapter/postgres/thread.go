package postgres

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/module/forums/domain"
	"github.com/realmkit/rk-backend/module/forums/port"
	"github.com/realmkit/rk-backend/pkg/orm"
	"github.com/realmkit/rk-backend/pkg/pagination"
	"gorm.io/gorm"
)

// ThreadRepository stores forum threads in PostgreSQL.
type ThreadRepository struct {
	store orm.Store // store stores the store value.
}

// NewThreadRepository creates a thread repository.
func NewThreadRepository(store orm.Store) ThreadRepository {
	return ThreadRepository{store: store}
}

// Create stores a thread.
func (repository ThreadRepository) Create(ctx context.Context, thread domain.Thread) (domain.Thread, error) {
	model := threadModelFromDomain(thread)
	if err := repository.store.DB(ctx).Create(&model).Error; err != nil {
		return domain.Thread{}, port.ErrConflict
	}
	if err := repository.incrementForumStats(ctx, thread); err != nil {
		return domain.Thread{}, err
	}
	return threadFromModel(model), nil
}

// FindByID returns one thread.
func (repository ThreadRepository) FindByID(ctx context.Context, id uuid.UUID) (domain.Thread, error) {
	var model ThreadModel
	if err := repository.store.DB(ctx).First(&model, "id = ?", id).Error; err != nil {
		return domain.Thread{}, mapError(err)
	}
	return threadFromModel(model), nil
}

// List returns matching threads.
func (repository ThreadRepository) List(
	ctx context.Context,
	filter port.ThreadFilter,
	page pagination.Page,
) (pagination.Result[domain.Thread], error) {
	query := repository.store.DB(ctx).
		Model(&ThreadModel{}).
		Where("forum_id = ?", filter.ForumID).
		Order("sticky_state desc, sticky_order asc, latest_post_at desc, id asc").
		Limit(page.Limit + 1)
	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}
	if filter.Section == "sticky" {
		query = query.Where("sticky_state <> ?", domain.StickyStateNormal)
	}
	if filter.Section == "normal" {
		query = query.Where("sticky_state = ?", domain.StickyStateNormal)
	}
	var models []ThreadModel
	if err := query.Find(&models).Error; err != nil {
		return pagination.Result[domain.Thread]{}, err
	}
	return threadPage(models, page.Limit), nil
}

// UpdateTitle updates thread title fields.
func (repository ThreadRepository) UpdateTitle(
	ctx context.Context,
	thread domain.Thread,
	expectedVersion uint64,
) (domain.Thread, error) {
	updates := map[string]any{
		"title":   thread.Title,
		"slug":    string(thread.Slug),
		"version": expectedVersion + 1,
	}
	result := repository.store.DB(ctx).
		Model(&ThreadModel{}).
		Where("id = ? AND version = ?", thread.ID, expectedVersion).
		Updates(updates)
	if result.Error != nil {
		return domain.Thread{}, result.Error
	}
	if result.RowsAffected == 0 {
		return domain.Thread{}, port.ErrPreconditionFailed
	}
	return repository.FindByID(ctx, thread.ID)
}

// Delete soft deletes a thread.
func (repository ThreadRepository) Delete(ctx context.Context, id uuid.UUID, expectedVersion uint64) error {
	result := repository.store.DB(ctx).
		Where("id = ? AND version = ?", id, expectedVersion).
		Delete(&ThreadModel{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return port.ErrPreconditionFailed
	}
	return nil
}

// incrementForumStats updates forum stats after thread creation.
func (repository ThreadRepository) incrementForumStats(ctx context.Context, thread domain.Thread) error {
	updates := map[string]any{
		"thread_count":               gorm.Expr("thread_count + ?", 1),
		"post_count":                 gorm.Expr("post_count + ?", thread.PostCount),
		"latest_thread_id":           thread.ID,
		"latest_post_id":             thread.LatestPostID,
		"latest_post_author_user_id": thread.LatestPostAuthorUserID,
		"latest_post_at":             thread.LatestPostAt,
		"updated_at":                 time.Now().UTC(),
	}
	if thread.Visible() {
		updates["visible_thread_count"] = gorm.Expr("visible_thread_count + ?", 1)
		updates["visible_post_count"] = gorm.Expr("visible_post_count + ?", thread.VisiblePostCount)
	}
	return repository.store.DB(ctx).
		Model(&StatsModel{}).
		Where("forum_id = ?", thread.ForumID).
		Updates(updates).Error
}

// threadPage maps models into a page.
func threadPage(models []ThreadModel, limit int) pagination.Result[domain.Thread] {
	next := ""
	if len(models) > limit {
		next = models[limit-1].ID.ID.String()
		models = models[:limit]
	}
	items := make([]domain.Thread, 0, len(models))
	for _, model := range models {
		items = append(items, threadFromModel(model))
	}
	return pagination.Result[domain.Thread]{Items: items, NextCursor: next}
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
	if !filter.Query.Empty() {
		like := filter.Query.LowerLike()
		if query.Dialector.Name() == "postgres" {
			query = query.Where(
				"key = ? OR slug = ? OR LOWER(name) LIKE ? OR LOWER(description) LIKE ?",
				filter.Query.String(),
				filter.Query.String(),
				like,
				like,
			)
		} else {
			query = query.Where(
				"LOWER(key) LIKE ? OR LOWER(slug) LIKE ? OR LOWER(name) LIKE ? OR LOWER(description) LIKE ?",
				like,
				like,
				like,
				like,
			)
		}
	}
	return query
}

// forumUpdates returns update fields.
func forumUpdates(forum domain.Forum, expectedVersion uint64) map[string]any {
	return map[string]any{
		"category_id":                       forum.CategoryID,
		"parent_forum_id":                   forum.ParentForumID,
		"kind":                              string(forum.Kind),
		"key":                               string(forum.Key),
		"slug":                              string(forum.Slug),
		"name":                              forum.Name,
		"description":                       forum.Description,
		"display_order":                     forum.DisplayOrder,
		"path":                              forum.Path,
		"depth":                             forum.Depth,
		"external_url":                      forum.ExternalURL,
		"icon_asset_id":                     forum.IconAssetID,
		"thread_visibility_mode":            string(forum.ThreadVisibilityMode),
		"max_sticky_threads":                forum.MaxStickyThreads,
		"default_thread_status":             string(forum.DefaultThreadStatus),
		"author_post_edit_window_seconds":   forum.AuthorPostEditWindowSeconds,
		"author_post_delete_window_seconds": forum.AuthorPostDeleteWindowSeconds,
		"status":                            string(forum.Status),
		"version":                           expectedVersion + 1,
	}
}

// descendantMoveUpdates returns update fields for one moved descendant.
func descendantMoveUpdates(
	forum domain.Forum,
	descendant ForumModel,
	nextPath string,
	depthDelta int,
) map[string]any {
	return map[string]any{
		"path":        nextPath,
		"category_id": forum.CategoryID,
		"depth":       descendant.Depth + depthDelta,
	}
}

// Ensure ThreadRepository implements port.ThreadRepository.
var _ port.ThreadRepository = ThreadRepository{}
