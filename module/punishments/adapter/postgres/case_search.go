package postgres

import (
	"time"

	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/module/punishments/port"
	"github.com/realmkit/rk-backend/pkg/search"
	"gorm.io/gorm"
)

// applyPunishmentFilter applies punishment case list filters.
func applyPunishmentFilter(query *gorm.DB, filter port.PunishmentFilter) *gorm.DB {
	if filter.TargetUserID != uuid.Nil {
		query = query.Where("target_user_id = ?", filter.TargetUserID)
	}
	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}
	if !filter.Query.Empty() {
		if query.Dialector.Name() == "postgres" {
			query = query.Where(punishmentPostgresSearchCondition(), filter.Query.String())
		} else {
			like := filter.Query.LowerLike()
			query = query.Where(punishmentSearchCondition(), like, like, like, like)
		}
	}
	return query
}

// punishmentSearchCondition returns safe text-search fields.
func punishmentSearchCondition() string {
	return "LOWER(reason) LIKE ? OR LOWER(source) LIKE ? OR LOWER(issuer_key) LIKE ? OR LOWER(target_ip_hash) LIKE ?"
}

// punishmentPostgresSearchCondition returns indexed PostgreSQL text search.
func punishmentPostgresSearchCondition() string {
	return "to_tsvector('simple', coalesce(reason, '') || ' ' || coalesce(source, '') || ' ' || coalesce(issuer_key, '') || ' ' || coalesce(target_ip_hash, '')) @@ plainto_tsquery('simple', ?)"
}

// applyPunishmentCursor applies keyset cursor filtering.
func applyPunishmentCursor(query *gorm.DB, cursor search.Cursor, ok bool, sort search.Sort) (*gorm.DB, error) {
	if !ok || len(cursor.Values) == 0 {
		return query, nil
	}
	id, err := uuid.Parse(cursor.ID)
	if err != nil {
		return nil, search.ErrInvalidCursor
	}
	column := punishmentSortColumn(sort.Key)
	value := punishmentCursorValue(cursor.Values[0], sort.Key)
	if sort.Desc() {
		return query.Where(column+" < ? OR ("+column+" = ? AND id > ?)", value, value, id), nil
	}
	return query.Where(column+" > ? OR ("+column+" = ? AND id > ?)", value, value, id), nil
}

// punishmentOrder returns deterministic ordering SQL.
func punishmentOrder(sort search.Sort) string {
	direction := "ASC"
	if sort.Desc() {
		direction = "DESC"
	}
	return punishmentSortColumn(sort.Key) + " " + direction + ", id ASC"
}

// punishmentSortColumn maps public sort keys to columns.
func punishmentSortColumn(key string) string {
	switch key {
	case "expires_at":
		return "expires_at"
	case "status":
		return "status"
	default:
		return "created_at"
	}
}

// punishmentCursor returns an encoded punishment cursor.
func punishmentCursor(model PunishmentModel, filterHash string, sort search.Sort) (string, error) {
	return search.EncodeCursor(search.Cursor{
		FilterHash: filterHash,
		Sort:       sort.Key,
		Direction:  sort.Direction,
		Values:     []string{punishmentModelSortValue(model, sort.Key)},
		ID:         model.ID.ID.String(),
	})
}

// punishmentModelSortValue returns the cursor value for a row.
func punishmentModelSortValue(model PunishmentModel, key string) string {
	switch key {
	case "expires_at":
		return optionalPunishmentTime(model.ExpiresAt)
	case "status":
		return model.Status
	default:
		return model.CreatedAt.Format(time.RFC3339Nano)
	}
}

// punishmentCursorValue converts a cursor value to the matching SQL type.
func punishmentCursorValue(value string, key string) any {
	if key == "created_at" || key == "expires_at" || key == "" {
		parsed, _ := time.Parse(time.RFC3339Nano, value)
		return parsed
	}
	return value
}

// optionalPunishmentTime formats a nullable punishment time.
func optionalPunishmentTime(value *time.Time) string {
	if value == nil {
		return time.Time{}.Format(time.RFC3339Nano)
	}
	return value.Format(time.RFC3339Nano)
}

// punishmentFilterHash binds cursors to active punishment filters.
func punishmentFilterHash(filter port.PunishmentFilter, sort search.Sort) string {
	return search.HashFilter(filter.TargetUserID, filter.Status, filter.Query.String(), sort)
}
