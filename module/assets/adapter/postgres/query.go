package postgres

import (
	"time"

	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/module/assets/domain"
	"github.com/realmkit/rk-backend/module/assets/port"
	"github.com/realmkit/rk-backend/pkg/pagination"
	"github.com/realmkit/rk-backend/pkg/search"
	"gorm.io/gorm"
)

// applyAssetFilter applies asset filters.
func applyAssetFilter(query *gorm.DB, filter port.AssetFilter) *gorm.DB {
	if filter.Namespace != "" {
		query = query.Where("namespace = ?", filter.Namespace)
	}
	if filter.PathExact {
		query = query.Where("path = ?", filter.Path)
	}
	if filter.PathPrefix != "" {
		query = query.Where("path = ? OR path LIKE ?", filter.PathPrefix, string(filter.PathPrefix)+"/%")
	}
	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}
	if filter.Visibility != "" {
		query = query.Where("visibility = ?", filter.Visibility)
	}
	if !filter.Query.Empty() {
		like := filter.Query.LowerLike()
		query = query.Where(assetSearchCondition(), like, like, like)
	}
	return query
}

// assetPage maps models into a page.
func assetPage(models []AssetModel, limit int, filterHash string, sort search.Sort) (pagination.Result[domain.Asset], error) {
	next := ""
	if len(models) > limit {
		cursor, err := assetCursor(models[limit-1], filterHash, sort)
		if err != nil {
			return pagination.Result[domain.Asset]{}, err
		}
		next = cursor
		models = models[:limit]
	}
	items := make([]domain.Asset, 0, len(models))
	for _, model := range models {
		items = append(items, assetFromModel(model))
	}
	return pagination.Result[domain.Asset]{Items: items, NextCursor: next}, nil
}

// applyAssetCursor applies keyset pagination to assets.
func applyAssetCursor(query *gorm.DB, cursor search.Cursor, ok bool, sort search.Sort) (*gorm.DB, error) {
	if !ok || len(cursor.Values) == 0 {
		return query, nil
	}
	id, err := uuid.Parse(cursor.ID)
	if err != nil {
		return nil, search.ErrInvalidCursor
	}
	column := assetSortColumn(sort.Key)
	value := assetCursorValue(cursor.Values[0], sort.Key)
	if sort.Desc() {
		return query.Where(column+" < ? OR ("+column+" = ? AND id > ?)", value, value, id), nil
	}
	return query.Where(column+" > ? OR ("+column+" = ? AND id > ?)", value, value, id), nil
}

// assetSearchCondition returns the asset text search predicate.
func assetSearchCondition() string {
	return "LOWER(path) LIKE ? OR LOWER(filename) LIKE ? OR LOWER(display_name) LIKE ?"
}

// assetOrder returns deterministic asset ordering SQL.
func assetOrder(sort search.Sort) string {
	direction := "ASC"
	if sort.Desc() {
		direction = "DESC"
	}
	return assetSortColumn(sort.Key) + " " + direction + ", id ASC"
}

// assetSortColumn maps public sort keys to SQL columns.
func assetSortColumn(key string) string {
	switch key {
	case "filename":
		return "filename"
	case "display_name":
		return "display_name"
	case "updated_at":
		return "updated_at"
	default:
		return "created_at"
	}
}

// assetCursor returns an encoded asset cursor.
func assetCursor(model AssetModel, filterHash string, sort search.Sort) (string, error) {
	return search.EncodeCursor(search.Cursor{
		FilterHash: filterHash,
		Sort:       sort.Key,
		Direction:  sort.Direction,
		Values:     []string{assetModelSortValue(model, sort.Key)},
		ID:         model.ID.ID.String(),
	})
}

// assetModelSortValue returns the cursor value for an asset row.
func assetModelSortValue(model AssetModel, key string) string {
	switch key {
	case "filename":
		return model.Filename
	case "display_name":
		return model.DisplayName
	case "updated_at":
		return model.UpdatedAt.Format(time.RFC3339Nano)
	default:
		return model.CreatedAt.Format(time.RFC3339Nano)
	}
}

// assetCursorValue converts a cursor value into the SQL type.
func assetCursorValue(value string, key string) any {
	if key == "created_at" || key == "updated_at" || key == "" {
		parsed, _ := time.Parse(time.RFC3339Nano, value)
		return parsed
	}
	return value
}

// assetFilterHash binds an asset cursor to current filters.
func assetFilterHash(filter port.AssetFilter, sort search.Sort) string {
	return search.HashFilter(filter.Namespace, filter.Path, filter.PathExact, filter.PathPrefix, filter.Status, filter.Visibility, filter.Query.String(), sort)
}
