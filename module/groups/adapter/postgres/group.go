package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/module/groups/domain"
	"github.com/realmkit/rk-backend/module/groups/port"
	"github.com/realmkit/rk-backend/pkg/orm"
	"github.com/realmkit/rk-backend/pkg/pagination"
	"gorm.io/gorm"
)

// GroupRepository stores groups in PostgreSQL.
type GroupRepository struct {
	store orm.Store
}

// NewGroupRepository creates a group repository.
func NewGroupRepository(store orm.Store) GroupRepository {
	return GroupRepository{store: store}
}

// Create stores a group.
func (repository GroupRepository) Create(ctx context.Context, group domain.Group) (domain.Group, error) {
	model := groupModelFromDomain(group)
	if model.Version == 0 {
		model.Version = 1
	}
	if err := repository.store.DB(ctx).Create(&model).Error; err != nil {
		return domain.Group{}, port.ErrConflict
	}
	return groupFromModel(model), nil
}

// Update stores mutable group fields.
func (repository GroupRepository) Update(ctx context.Context, group domain.Group, expectedVersion uint64) (domain.Group, error) {
	result := repository.store.DB(ctx).
		Model(&GroupModel{}).
		Where("id = ? AND version = ?", group.ID, expectedVersion).
		Updates(groupUpdates(group, expectedVersion))
	if result.Error != nil {
		return domain.Group{}, result.Error
	}
	if result.RowsAffected == 0 {
		return domain.Group{}, port.ErrPreconditionFailed
	}
	return repository.FindByID(ctx, group.ID)
}

// groupUpdates returns update fields for a group.
func groupUpdates(group domain.Group, expectedVersion uint64) map[string]any {
	return map[string]any{
		"name":          group.Name,
		"description":   group.Description,
		"color":         string(group.Color),
		"weight":        group.Weight,
		"status":        string(group.Status),
		"icon_asset_id": group.IconAssetID,
		"version":       expectedVersion + 1,
	}
}

// FindByID returns one group.
func (repository GroupRepository) FindByID(ctx context.Context, id uuid.UUID) (domain.Group, error) {
	var model GroupModel
	if err := repository.store.DB(ctx).First(&model, "id = ?", id).Error; err != nil {
		return domain.Group{}, mapError(err)
	}
	return groupFromModel(model), nil
}

// FindByKey returns one group by key.
func (repository GroupRepository) FindByKey(ctx context.Context, key domain.Key) (domain.Group, error) {
	var model GroupModel
	if err := repository.store.DB(ctx).First(&model, "key = ?", key).Error; err != nil {
		return domain.Group{}, mapError(err)
	}
	return groupFromModel(model), nil
}

// List returns matching groups.
func (repository GroupRepository) List(
	ctx context.Context,
	filter port.GroupFilter,
	page pagination.Page,
) (pagination.Result[domain.Group], error) {
	query := repository.store.DB(ctx).Model(&GroupModel{}).Order("weight desc, key asc").Limit(page.Limit + 1)
	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}
	var models []GroupModel
	if err := query.Find(&models).Error; err != nil {
		return pagination.Result[domain.Group]{}, err
	}
	return groupPage(models, page.Limit), nil
}

// Delete soft deletes a group.
func (repository GroupRepository) Delete(ctx context.Context, id uuid.UUID, expectedVersion uint64) error {
	result := repository.store.DB(ctx).Where("id = ? AND version = ?", id, expectedVersion).Delete(&GroupModel{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return port.ErrPreconditionFailed
	}
	return nil
}

// groupPage maps models into a page.
func groupPage(models []GroupModel, limit int) pagination.Result[domain.Group] {
	next := ""
	if len(models) > limit {
		next = models[limit-1].ID.ID.String()
		models = models[:limit]
	}
	items := make([]domain.Group, 0, len(models))
	for _, model := range models {
		items = append(items, groupFromModel(model))
	}
	return pagination.Result[domain.Group]{Items: items, NextCursor: next}
}

// mapError maps GORM errors into groups errors.
func mapError(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return port.ErrNotFound
	}
	return err
}
