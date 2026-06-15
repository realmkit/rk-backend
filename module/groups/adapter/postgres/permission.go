package postgres

import (
	"context"
	"errors"
	"time"

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

// CreateGrant stores a global permission grant and assigns it to a group.
func (repository PermissionRepository) CreateGrant(
	ctx context.Context,
	groupID uuid.UUID,
	grant domain.PermissionGrant,
) (domain.PermissionGrant, error) {
	if err := grant.Validate(); err != nil {
		return domain.PermissionGrant{}, err
	}
	model, err := repository.findOrCreateGrant(ctx, grant)
	if err != nil {
		return domain.PermissionGrant{}, err
	}
	if err := repository.assignGrant(ctx, groupID, model.ID.ID, grant.CreatedByUserID); err != nil {
		return domain.PermissionGrant{}, err
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
		Order("permission_grants.created_at asc").
		Limit(page.Limit + 1)
	var models []PermissionGrantModel
	if err := query.Find(&models).Error; err != nil {
		return pagination.Result[domain.PermissionGrant]{}, err
	}
	return grantPage(models, page.Limit), nil
}

// DeleteGrant removes one global grant assignment from a group.
func (repository PermissionRepository) DeleteGrant(
	ctx context.Context,
	groupID uuid.UUID,
	id uuid.UUID,
) error {
	result := repository.store.DB(ctx).
		Where("group_id = ? AND grant_id = ?", groupID, id).
		Delete(&GroupPermissionGrantModel{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return port.ErrNotFound
	}
	return nil
}

// findOrCreateGrant returns an existing global grant or creates it.
func (repository PermissionRepository) findOrCreateGrant(
	ctx context.Context,
	grant domain.PermissionGrant,
) (PermissionGrantModel, error) {
	existing, err := repository.findEquivalentGrantModel(ctx, grant)
	if err == nil {
		return existing, nil
	}
	if !errors.Is(err, port.ErrNotFound) {
		return PermissionGrantModel{}, err
	}
	model := grantModelFromDomain(grant)
	if err := repository.store.DB(ctx).Create(&model).Error; err != nil {
		return PermissionGrantModel{}, port.ErrConflict
	}
	return model, nil
}

// assignGrant creates or restores one group grant assignment.
func (repository PermissionRepository) assignGrant(
	ctx context.Context,
	groupID uuid.UUID,
	grantID uuid.UUID,
	createdByUserID *uuid.UUID,
) error {
	if repository.activeAssignmentExists(ctx, groupID, grantID) {
		return port.ErrConflict
	}
	restored, err := repository.restoreAssignment(ctx, groupID, grantID, createdByUserID)
	if err != nil || restored {
		return err
	}
	return repository.createAssignment(ctx, groupID, grantID, createdByUserID)
}

// activeAssignmentExists reports whether the assignment is already active.
func (repository PermissionRepository) activeAssignmentExists(
	ctx context.Context,
	groupID uuid.UUID,
	grantID uuid.UUID,
) bool {
	var model GroupPermissionGrantModel
	err := repository.store.DB(ctx).
		Where("group_id = ? AND grant_id = ?", groupID, grantID).
		First(&model).
		Error
	return err == nil
}

// restoreAssignment restores a previously deleted group grant assignment.
func (repository PermissionRepository) restoreAssignment(
	ctx context.Context,
	groupID uuid.UUID,
	grantID uuid.UUID,
	createdByUserID *uuid.UUID,
) (bool, error) {
	var model GroupPermissionGrantModel
	err := repository.store.DB(ctx).
		Unscoped().
		Where("group_id = ? AND grant_id = ? AND deleted_at IS NOT NULL", groupID, grantID).
		Order("created_at desc").
		First(&model).
		Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	err = repository.store.DB(ctx).
		Unscoped().
		Model(&GroupPermissionGrantModel{}).
		Where("id = ?", model.ID.ID).
		Updates(map[string]any{
			"created_at":         time.Now().UTC(),
			"created_by_user_id": createdByUserID,
			"deleted_at":         nil,
		}).
		Error
	return true, err
}

// createAssignment creates a new group grant assignment.
func (repository PermissionRepository) createAssignment(
	ctx context.Context,
	groupID uuid.UUID,
	grantID uuid.UUID,
	createdByUserID *uuid.UUID,
) error {
	link := GroupPermissionGrantModel{
		ID:              orm.ID{ID: uuid.New()},
		GroupID:         groupID,
		GrantID:         grantID,
		CreatedByUserID: createdByUserID,
	}
	if err := repository.store.DB(ctx).Create(&link).Error; err != nil {
		return port.ErrConflict
	}
	return nil
}

// findEquivalentGrantModel returns a matching active global grant.
func (repository PermissionRepository) findEquivalentGrantModel(
	ctx context.Context,
	grant domain.PermissionGrant,
) (PermissionGrantModel, error) {
	var model PermissionGrantModel
	err := repository.store.DB(ctx).
		Where("action = ?", grant.Action).
		Where("scope_type = ?", grant.ScopeType).
		Where("scope_id = ?", grant.ScopeID).
		Where("inherit = ?", grant.Inherit).
		Where("condition_key = ?", grant.ConditionKey).
		First(&model).
		Error
	if err != nil {
		return PermissionGrantModel{}, mapError(err)
	}
	return model, nil
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
