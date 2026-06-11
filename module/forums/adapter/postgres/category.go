package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/niflaot/gamehub-go/module/forums/domain"
	"github.com/niflaot/gamehub-go/module/forums/port"
	"github.com/niflaot/gamehub-go/pkg/orm"
	"github.com/niflaot/gamehub-go/pkg/pagination"
	"gorm.io/gorm"
)

// CategoryRepository stores forum categories in PostgreSQL.
type CategoryRepository struct {
	store orm.Store
}

// NewCategoryRepository creates a category repository.
func NewCategoryRepository(store orm.Store) CategoryRepository {
	return CategoryRepository{store: store}
}

// Create stores a category.
func (repository CategoryRepository) Create(ctx context.Context, category domain.ForumCategory) (domain.ForumCategory, error) {
	model := categoryModelFromDomain(category)
	if err := repository.store.DB(ctx).Create(&model).Error; err != nil {
		return domain.ForumCategory{}, port.ErrConflict
	}
	return categoryFromModel(model), nil
}

// Update stores mutable category fields.
func (repository CategoryRepository) Update(
	ctx context.Context,
	category domain.ForumCategory,
	expectedVersion uint64,
) (domain.ForumCategory, error) {
	result := repository.store.DB(ctx).
		Model(&CategoryModel{}).
		Where("id = ? AND version = ?", category.ID, expectedVersion).
		Updates(categoryUpdates(category, expectedVersion))
	if result.Error != nil {
		return domain.ForumCategory{}, result.Error
	}
	if result.RowsAffected == 0 {
		return domain.ForumCategory{}, port.ErrPreconditionFailed
	}
	return repository.FindByID(ctx, category.ID)
}

// FindByID returns one category.
func (repository CategoryRepository) FindByID(ctx context.Context, id uuid.UUID) (domain.ForumCategory, error) {
	var model CategoryModel
	if err := repository.store.DB(ctx).First(&model, "id = ?", id).Error; err != nil {
		return domain.ForumCategory{}, mapError(err)
	}
	return categoryFromModel(model), nil
}

// List returns matching categories.
func (repository CategoryRepository) List(
	ctx context.Context,
	filter port.CategoryFilter,
	page pagination.Page,
) (pagination.Result[domain.ForumCategory], error) {
	query := repository.store.DB(ctx).Model(&CategoryModel{}).Order("display_order asc, id asc").Limit(page.Limit + 1)
	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}
	var models []CategoryModel
	if err := query.Find(&models).Error; err != nil {
		return pagination.Result[domain.ForumCategory]{}, err
	}
	return categoryPage(models, page.Limit), nil
}

// Delete soft deletes one category.
func (repository CategoryRepository) Delete(ctx context.Context, id uuid.UUID, expectedVersion uint64) error {
	result := repository.store.DB(ctx).Where("id = ? AND version = ?", id, expectedVersion).Delete(&CategoryModel{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return port.ErrPreconditionFailed
	}
	return nil
}

// Reorder updates category display order.
func (repository CategoryRepository) Reorder(ctx context.Context, items []port.ReorderItem) error {
	for _, item := range items {
		err := repository.store.DB(ctx).
			Model(&CategoryModel{}).
			Where("id = ?", item.ID).
			Update("display_order", item.DisplayOrder).
			Error
		if err != nil {
			return err
		}
	}
	return nil
}

// categoryUpdates returns update fields.
func categoryUpdates(category domain.ForumCategory, expectedVersion uint64) map[string]any {
	return map[string]any{
		"key":           string(category.Key),
		"name":          category.Name,
		"description":   category.Description,
		"display_order": category.DisplayOrder,
		"status":        string(category.Status),
		"version":       expectedVersion + 1,
	}
}

// categoryPage maps models into a page.
func categoryPage(models []CategoryModel, limit int) pagination.Result[domain.ForumCategory] {
	next := ""
	if len(models) > limit {
		next = models[limit-1].ID.ID.String()
		models = models[:limit]
	}
	items := make([]domain.ForumCategory, 0, len(models))
	for _, model := range models {
		items = append(items, categoryFromModel(model))
	}
	return pagination.Result[domain.ForumCategory]{Items: items, NextCursor: next}
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

// categoryModelFromDomain maps category to persistence.
func categoryModelFromDomain(category domain.ForumCategory) CategoryModel {
	return CategoryModel{
		ID:           orm.ID{ID: category.ID},
		Key:          string(category.Key),
		Name:         category.Name,
		Description:  category.Description,
		DisplayOrder: category.DisplayOrder,
		Status:       string(category.Status),
		Version:      category.Version,
	}
}

// categoryFromModel maps persistence category to domain.
func categoryFromModel(model CategoryModel) domain.ForumCategory {
	return domain.ForumCategory{
		ID:           model.ID.ID,
		Key:          domain.Key(model.Key),
		Name:         model.Name,
		Description:  model.Description,
		DisplayOrder: model.DisplayOrder,
		Status:       domain.CategoryStatus(model.Status),
		Version:      model.Version,
		CreatedAt:    model.CreatedAt,
		UpdatedAt:    model.UpdatedAt,
	}
}

// mapError maps GORM errors into forum errors.
func mapError(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return port.ErrNotFound
	}
	return err
}

// Ensure CategoryRepository implements port.CategoryRepository.
var _ port.CategoryRepository = CategoryRepository{}
