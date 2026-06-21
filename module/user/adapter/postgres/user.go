package postgres

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/module/user/domain"
	"github.com/realmkit/rk-backend/module/user/port"
	"github.com/realmkit/rk-backend/pkg/orm"
	"github.com/realmkit/rk-backend/pkg/pagination"
	"github.com/realmkit/rk-backend/pkg/search"
)

// UserRepository stores local users.
type UserRepository struct {
	store orm.Store // store stores the store value.
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
