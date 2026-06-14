package postgres

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/module/user/domain"
	"github.com/realmkit/rk-backend/module/user/port"
	"github.com/realmkit/rk-backend/pkg/orm"
	"github.com/realmkit/rk-backend/pkg/pagination"
	"github.com/realmkit/rk-backend/pkg/search"
	"gorm.io/gorm"
)

// UserRepository stores local users.
type UserRepository struct {
	store orm.Store
}

// NewUserRepository creates a user repository.
func NewUserRepository(store orm.Store) UserRepository {
	return UserRepository{store: store}
}

// Create stores a user.
func (repository UserRepository) Create(ctx context.Context, user domain.User) (domain.User, error) {
	model := userModelFromDomain(user)
	if err := repository.store.DB(ctx).Create(&model).Error; err != nil {
		return domain.User{}, port.ErrConflict
	}
	return userFromModel(model), nil
}

// Update stores mutable user fields.
func (repository UserRepository) Update(ctx context.Context, user domain.User, expectedVersion uint64) (domain.User, error) {
	result := repository.store.DB(ctx).
		Model(&UserModel{}).
		Where("id = ? AND version = ?", user.ID, expectedVersion).
		Updates(map[string]any{"avatar_asset_id": user.AvatarAssetID, "version": expectedVersion + 1})
	if result.Error != nil {
		return domain.User{}, result.Error
	}
	if result.RowsAffected == 0 {
		return domain.User{}, port.ErrPreconditionFailed
	}
	return repository.FindByID(ctx, user.ID)
}

// FindByID returns one user.
func (repository UserRepository) FindByID(ctx context.Context, id uuid.UUID) (domain.User, error) {
	var model UserModel
	if err := repository.store.DB(ctx).First(&model, "id = ?", id).Error; err != nil {
		return domain.User{}, mapError(err)
	}
	return userFromModel(model), nil
}

// List returns matching users.
func (repository UserRepository) List(
	ctx context.Context,
	filter port.UserFilter,
	page pagination.Page,
) (pagination.Result[port.UserSummary], error) {
	sort := filter.Sort
	if sort.Key == "" {
		sort, _ = search.NewSort("", "", port.DefaultUserSort(), port.AllowedUserSorts())
	}
	filterHash := userFilterHash(filter, sort)
	cursor, hasCursor, err := search.RequireCursor(page.Cursor, filterHash, sort)
	if err != nil {
		return pagination.Result[port.UserSummary]{}, err
	}
	query := userListBase(repository.store.DB(ctx), filter)
	query, err = applyUserCursor(query, cursor, hasCursor, sort)
	if err != nil {
		return pagination.Result[port.UserSummary]{}, err
	}
	query = query.Order(userOrder(sort)).Limit(page.Limit + 1)
	var rows []userListRow
	if err := query.Scan(&rows).Error; err != nil {
		return pagination.Result[port.UserSummary]{}, err
	}
	return userPage(rows, page.Limit, filterHash, sort)
}

// FindSummariesByIDs returns display summaries keyed by local user ID.
func (repository UserRepository) FindSummariesByIDs(
	ctx context.Context,
	ids []uuid.UUID,
) (map[uuid.UUID]port.UserSummary, error) {
	if len(ids) == 0 {
		return map[uuid.UUID]port.UserSummary{}, nil
	}
	var rows []userListRow
	err := repository.store.DB(ctx).
		Table("users AS u").
		Select(userListSelect()).
		Joins("LEFT JOIN user_provider_claim_cache AS c ON c.user_id = u.id AND c.deleted_at IS NULL").
		Where("u.deleted_at IS NULL AND u.id IN ?", uniqueUserIDs(ids)).
		Scan(&rows).
		Error
	if err != nil {
		return nil, err
	}
	result := make(map[uuid.UUID]port.UserSummary, len(rows))
	for _, row := range rows {
		result[row.ID] = row.summary()
	}
	return result, nil
}

// TouchLastSeen stores the last-seen timestamp.
func (repository UserRepository) TouchLastSeen(ctx context.Context, id uuid.UUID) error {
	result := repository.store.DB(ctx).Model(&UserModel{}).Where("id = ?", id).Update("last_seen_at", time.Now().UTC())
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return port.ErrNotFound
	}
	return nil
}

// userListBase returns the base user list query.
func userListBase(db *gorm.DB, filter port.UserFilter) *gorm.DB {
	query := db.Table("users AS u").Select(userListSelect()).
		Joins("LEFT JOIN user_provider_claim_cache AS c ON c.user_id = u.id AND c.deleted_at IS NULL").
		Where("u.deleted_at IS NULL")
	if filter.Status != "" {
		query = query.Where("u.status = ?", filter.Status)
	}
	if !filter.Query.Empty() {
		if query.Dialector.Name() == "postgres" {
			query = query.Where(userPostgresSearchCondition(), filter.Query.String(), filter.Query.LowerLike())
		} else {
			like := filter.Query.LowerLike()
			query = query.Where(userSearchCondition(), like, like, like, like)
		}
	}
	return query
}

// userListSelect returns selected list columns.
func userListSelect() string {
	return strings.Join([]string{
		"u.id", "u.status", "u.avatar_asset_id", "u.first_seen_at", "u.last_seen_at",
		"u.version", "u.created_at", "u.updated_at", "c.id AS claim_id", "c.issuer",
		"c.subject", "c.username", "c.email", "c.email_verified", "c.display_name",
		"c.picture_url", "c.preferred_locale", "c.claims_hash", "c.synced_at",
	}, ", ")
}

// userSearchCondition returns the user text search predicate.
func userSearchCondition() string {
	return "LOWER(CAST(u.id AS text)) LIKE ? OR LOWER(c.username) LIKE ? OR LOWER(c.email) LIKE ? OR LOWER(c.display_name) LIKE ?"
}

// userPostgresSearchCondition returns indexed PostgreSQL text search.
func userPostgresSearchCondition() string {
	return "to_tsvector('simple', coalesce(c.username, '') || ' ' || coalesce(c.email, '') || ' ' || coalesce(c.display_name, '')) @@ plainto_tsquery('simple', ?) OR u.id::text ILIKE ?"
}

// applyUserCursor applies keyset cursor filtering.
func applyUserCursor(query *gorm.DB, cursor search.Cursor, ok bool, sort search.Sort) (*gorm.DB, error) {
	if !ok || len(cursor.Values) == 0 {
		return query, nil
	}
	id, err := uuid.Parse(cursor.ID)
	if err != nil {
		return nil, search.ErrInvalidCursor
	}
	column := userSortColumn(sort.Key)
	value := userCursorValue(cursor.Values[0], sort.Key)
	if sort.Desc() {
		return query.Where(column+" < ? OR ("+column+" = ? AND u.id > ?)", value, value, id), nil
	}
	return query.Where(column+" > ? OR ("+column+" = ? AND u.id > ?)", value, value, id), nil
}

// userOrder returns deterministic list ordering.
func userOrder(sort search.Sort) string {
	direction := "ASC"
	if sort.Desc() {
		direction = "DESC"
	}
	return userSortColumn(sort.Key) + " " + direction + ", u.id ASC"
}

// userSortColumn maps public sort keys to SQL columns.
func userSortColumn(key string) string {
	switch key {
	case "display_name":
		return "c.display_name"
	case "email":
		return "c.email"
	case "last_seen_at":
		return "u.last_seen_at"
	default:
		return "u.created_at"
	}
}

// userCursorValue converts a cursor value into the matching SQL type.
func userCursorValue(value string, key string) any {
	if key == "created_at" || key == "last_seen_at" || key == "" {
		parsed, _ := time.Parse(time.RFC3339Nano, value)
		return parsed
	}
	return value
}

// userPage maps rows into a paginated result.
func userPage(rows []userListRow, limit int, filterHash string, sort search.Sort) (pagination.Result[port.UserSummary], error) {
	next := ""
	if len(rows) > limit {
		cursor, err := userCursor(rows[limit-1], filterHash, sort)
		if err != nil {
			return pagination.Result[port.UserSummary]{}, err
		}
		next = cursor
		rows = rows[:limit]
	}
	items := make([]port.UserSummary, 0, len(rows))
	for _, row := range rows {
		items = append(items, row.summary())
	}
	return pagination.Result[port.UserSummary]{Items: items, NextCursor: next}, nil
}

// userCursor returns an encoded user cursor.
func userCursor(row userListRow, filterHash string, sort search.Sort) (string, error) {
	return search.EncodeCursor(search.Cursor{
		FilterHash: filterHash,
		Sort:       sort.Key,
		Direction:  sort.Direction,
		Values:     []string{row.sortValue(sort.Key)},
		ID:         row.ID.String(),
	})
}

// userFilterHash binds a user cursor to active filters.
func userFilterHash(filter port.UserFilter, sort search.Sort) string {
	return search.HashFilter(filter.Status, filter.Query.String(), sort)
}

// uniqueUserIDs removes duplicate IDs before a batch lookup.
func uniqueUserIDs(ids []uuid.UUID) []uuid.UUID {
	seen := make(map[uuid.UUID]struct{}, len(ids))
	result := make([]uuid.UUID, 0, len(ids))
	for _, id := range ids {
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		result = append(result, id)
	}
	return result
}

// mapError maps persistence errors to port errors.
func mapError(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return port.ErrNotFound
	}
	return err
}
