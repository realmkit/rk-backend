package postgres

import (
	"context"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/module/groups/domain"
	"github.com/realmkit/rk-backend/module/groups/port"
	"github.com/realmkit/rk-backend/pkg/orm"
	"github.com/realmkit/rk-backend/pkg/pagination"
	"github.com/realmkit/rk-backend/pkg/search"
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
	sort := filter.Sort
	if sort.Key == "" {
		sort, _ = search.NewSort("", "", port.DefaultGroupSort(), port.AllowedGroupSorts())
	}
	filterHash := groupFilterHash(filter, sort)
	cursor, hasCursor, err := search.RequireCursor(page.Cursor, filterHash, sort)
	if err != nil {
		return pagination.Result[domain.Group]{}, err
	}
	query := applyGroupFilter(repository.store.DB(ctx).Model(&GroupModel{}), filter)
	query, err = applyGroupCursor(query, cursor, hasCursor, sort)
	if err != nil {
		return pagination.Result[domain.Group]{}, err
	}
	query = query.Order(groupOrder(sort)).Limit(page.Limit + 1)
	var models []GroupModel
	if err := query.Find(&models).Error; err != nil {
		return pagination.Result[domain.Group]{}, err
	}
	return groupPage(models, page.Limit, filterHash, sort)
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
func groupPage(models []GroupModel, limit int, filterHash string, sort search.Sort) (pagination.Result[domain.Group], error) {
	next := ""
	if len(models) > limit {
		cursor, err := groupCursor(models[limit-1], filterHash, sort)
		if err != nil {
			return pagination.Result[domain.Group]{}, err
		}
		next = cursor
		models = models[:limit]
	}
	items := make([]domain.Group, 0, len(models))
	for _, model := range models {
		items = append(items, groupFromModel(model))
	}
	return pagination.Result[domain.Group]{Items: items, NextCursor: next}, nil
}

// applyGroupFilter applies group list filters.
func applyGroupFilter(query *gorm.DB, filter port.GroupFilter) *gorm.DB {
	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}
	if !filter.Query.Empty() {
		like := filter.Query.LowerLike()
		if query.Dialector.Name() == "postgres" {
			query = query.Where(groupPostgresSearchCondition(), filter.Query.String(), like, like, like)
		} else {
			query = query.Where("LOWER(key) LIKE ? OR LOWER(name) LIKE ? OR LOWER(description) LIKE ?", like, like, like)
		}
	}
	if filter.HasIcon != nil && *filter.HasIcon {
		query = query.Where("icon_asset_id IS NOT NULL")
	}
	if filter.HasIcon != nil && !*filter.HasIcon {
		query = query.Where("icon_asset_id IS NULL")
	}
	if filter.MinWeight != nil {
		query = query.Where("weight >= ?", *filter.MinWeight)
	}
	if filter.MaxWeight != nil {
		query = query.Where("weight <= ?", *filter.MaxWeight)
	}
	return query
}

// applyGroupCursor applies keyset cursor filtering.
func applyGroupCursor(query *gorm.DB, cursor search.Cursor, ok bool, sort search.Sort) (*gorm.DB, error) {
	if !ok || len(cursor.Values) == 0 {
		return query, nil
	}
	column := groupSortColumn(sort.Key)
	id, err := uuid.Parse(cursor.ID)
	if err != nil {
		return nil, search.ErrInvalidCursor
	}
	value := groupCursorValue(cursor.Values[0], sort.Key)
	if sort.Desc() {
		return query.Where(column+" < ? OR ("+column+" = ? AND id > ?)", value, value, id), nil
	}
	return query.Where(column+" > ? OR ("+column+" = ? AND id > ?)", value, value, id), nil
}

// groupOrder returns deterministic ordering SQL.
func groupOrder(sort search.Sort) string {
	direction := "ASC"
	if sort.Desc() {
		direction = "DESC"
	}
	return groupSortColumn(sort.Key) + " " + direction + ", id ASC"
}

// groupSortColumn maps public sort keys to columns.
func groupSortColumn(key string) string {
	switch key {
	case "key":
		return "key"
	case "name":
		return "name"
	case "created_at":
		return "created_at"
	case "updated_at":
		return "updated_at"
	default:
		return "weight"
	}
}

// groupCursor returns an encoded cursor for a group row.
func groupCursor(model GroupModel, filterHash string, sort search.Sort) (string, error) {
	value := groupModelSortValue(model, sort.Key)
	return search.EncodeCursor(search.Cursor{
		FilterHash: filterHash,
		Sort:       sort.Key,
		Direction:  sort.Direction,
		Values:     []string{value},
		ID:         model.ID.ID.String(),
	})
}

// groupModelSortValue returns the cursor value for the current sort.
func groupModelSortValue(model GroupModel, key string) string {
	switch key {
	case "key":
		return model.Key
	case "name":
		return model.Name
	case "created_at":
		return model.CreatedAt.Format(time.RFC3339Nano)
	case "updated_at":
		return model.UpdatedAt.Format(time.RFC3339Nano)
	default:
		return strconv.Itoa(model.Weight)
	}
}

// groupCursorValue converts a cursor value to the matching SQL type.
func groupCursorValue(value string, key string) any {
	if key == "weight" || key == "" {
		parsed, _ := strconv.Atoi(value)
		return parsed
	}
	if key == "created_at" || key == "updated_at" {
		parsed, _ := time.Parse(time.RFC3339Nano, value)
		return parsed
	}
	return value
}

// groupFilterHash binds cursors to the current filter.
func groupFilterHash(filter port.GroupFilter, sort search.Sort) string {
	return search.HashFilter(filter.Status, filter.Query.String(), filter.HasIcon, filter.MinWeight, filter.MaxWeight, sort)
}
