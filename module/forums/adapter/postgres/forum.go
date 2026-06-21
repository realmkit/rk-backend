package postgres

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/module/forums/domain"
	"github.com/realmkit/rk-backend/module/forums/port"
	"github.com/realmkit/rk-backend/pkg/orm"
	"github.com/realmkit/rk-backend/pkg/pagination"
)

// ForumRepository stores forums in PostgreSQL.
type ForumRepository struct {
	store orm.Store // store stores the store value.
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
	result := repository.store.DB(ctx).
		Model(&ForumModel{}).
		Where("id = ? AND version = ?", forum.ID, expectedVersion).
		Updates(forumUpdates(forum, expectedVersion))
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
func (repository ForumRepository) List(
	ctx context.Context,
	filter port.ForumFilter,
	page pagination.Page,
) (pagination.Result[domain.Forum], error) {
	query := repository.store.DB(ctx).
		Model(&ForumModel{}).
		Order("path asc, display_order asc, id asc").
		Limit(page.Limit + 1)
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
	err := repository.store.DB(ctx).
		Model(&ForumModel{}).
		Where("status = ?", domain.ForumStatusActive).
		Order("path asc, display_order asc, id asc").
		Find(&models).
		Error
	if err != nil {
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
func (repository ForumRepository) Move(
	ctx context.Context,
	forum domain.Forum,
	oldPath string,
	expectedVersion uint64,
) (domain.Forum, error) {
	result := repository.store.DB(ctx).
		Model(&ForumModel{}).
		Where("id = ? AND version = ?", forum.ID, expectedVersion).
		Updates(forumUpdates(forum, expectedVersion))
	if result.Error != nil {
		return domain.Forum{}, result.Error
	}
	if result.RowsAffected == 0 {
		return domain.Forum{}, port.ErrPreconditionFailed
	}
	if oldPath != forum.Path {
		var descendants []ForumModel
		err := repository.store.DB(ctx).
			Where("path LIKE ? AND id <> ?", oldPath+"%", forum.ID).
			Find(&descendants).
			Error
		if err != nil {
			return domain.Forum{}, err
		}
		for _, descendant := range descendants {
			nextPath := strings.Replace(descendant.Path, oldPath, forum.Path, 1)
			depthDelta := strings.Count(nextPath, "/") - strings.Count(descendant.Path, "/")
			err := repository.store.DB(ctx).
				Model(&ForumModel{}).
				Where("id = ?", descendant.ID.ID).
				Updates(descendantMoveUpdates(forum, descendant, nextPath, depthDelta)).
				Error
			if err != nil {
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
		err := repository.store.DB(ctx).
			Model(&ForumModel{}).
			Where("id = ?", item.ID).
			Update("display_order", item.DisplayOrder).
			Error
		if err != nil {
			return err
		}
	}
	return nil
}

// Ensure ForumRepository implements port.ForumRepository.
var _ port.ForumRepository = ForumRepository{}
