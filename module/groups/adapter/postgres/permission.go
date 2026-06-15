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

// PermissionRepository stores permission records in PostgreSQL.
type PermissionRepository struct {
	store orm.Store
}

// NewPermissionRepository creates a permission repository.
func NewPermissionRepository(store orm.Store) PermissionRepository {
	return PermissionRepository{store: store}
}

// CreateGrant stores a permission grant.
func (repository PermissionRepository) CreateGrant(
	ctx context.Context,
	grant domain.PermissionGrant,
) (domain.PermissionGrant, error) {
	if err := grant.Validate(); err != nil {
		return domain.PermissionGrant{}, err
	}
	if existing, err := repository.findEquivalentGrant(ctx, grant); err == nil {
		return existing, port.ErrConflict
	} else if !errors.Is(err, port.ErrNotFound) {
		return domain.PermissionGrant{}, err
	}
	model := grantModelFromDomain(grant)
	if err := repository.store.DB(ctx).Create(&model).Error; err != nil {
		return domain.PermissionGrant{}, port.ErrConflict
	}
	return grantFromModel(model), nil
}

// ListGrants returns active permission grants.
func (repository PermissionRepository) ListGrants(
	ctx context.Context,
	filter port.PermissionGrantFilter,
	page pagination.Page,
) (pagination.Result[domain.PermissionGrant], error) {
	query := applyGrantFilter(repository.store.DB(ctx).Model(&PermissionGrantModel{}), filter).
		Order("created_at asc").
		Limit(page.Limit + 1)
	var models []PermissionGrantModel
	if err := query.Find(&models).Error; err != nil {
		return pagination.Result[domain.PermissionGrant]{}, err
	}
	return grantPage(models, page.Limit), nil
}

// DeleteGrant soft deletes one permission grant.
func (repository PermissionRepository) DeleteGrant(ctx context.Context, id uuid.UUID) error {
	result := repository.store.DB(ctx).Delete(&PermissionGrantModel{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return port.ErrNotFound
	}
	return nil
}

// findEquivalentGrant returns a matching active grant.
func (repository PermissionRepository) findEquivalentGrant(
	ctx context.Context,
	grant domain.PermissionGrant,
) (domain.PermissionGrant, error) {
	var model PermissionGrantModel
	err := repository.store.DB(ctx).
		Where("subject_type = ?", grant.SubjectType).
		Where("subject_id = ?", grant.SubjectID).
		Where("action = ?", grant.Action).
		Where("scope_type = ?", grant.ScopeType).
		Where("scope_id = ?", grant.ScopeID).
		Where("inherit = ?", grant.Inherit).
		Where("condition_key = ?", grant.ConditionKey).
		First(&model).
		Error
	if err != nil {
		return domain.PermissionGrant{}, mapError(err)
	}
	return grantFromModel(model), nil
}

// applyGrantFilter applies permission grant filters.
func applyGrantFilter(query *gorm.DB, filter port.PermissionGrantFilter) *gorm.DB {
	if filter.SubjectType != "" {
		query = query.Where("subject_type = ?", filter.SubjectType)
	}
	if filter.SubjectID != uuid.Nil {
		query = query.Where("subject_id = ?", filter.SubjectID)
	}
	if filter.Action != "" {
		query = query.Where("action = ?", filter.Action)
	}
	if filter.ScopeType != "" {
		query = query.Where("scope_type = ?", filter.ScopeType)
	}
	if filter.ScopeID != uuid.Nil {
		query = query.Where("scope_id = ?", filter.ScopeID)
	}
	return query
}

// grantPage maps grant models into a page.
func grantPage(models []PermissionGrantModel, limit int) pagination.Result[domain.PermissionGrant] {
	next := ""
	if len(models) > limit {
		next = models[limit-1].ID.ID.String()
		models = models[:limit]
	}
	items := make([]domain.PermissionGrant, 0, len(models))
	for _, model := range models {
		items = append(items, grantFromModel(model))
	}
	return pagination.Result[domain.PermissionGrant]{Items: items, NextCursor: next}
}
