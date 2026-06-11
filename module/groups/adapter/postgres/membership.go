package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/module/groups/domain"
	"github.com/realmkit/rk-backend/module/groups/port"
	"github.com/realmkit/rk-backend/pkg/orm"
	"github.com/realmkit/rk-backend/pkg/pagination"
)

// MembershipRepository stores memberships in PostgreSQL.
type MembershipRepository struct {
	store orm.Store
}

// NewMembershipRepository creates a membership repository.
func NewMembershipRepository(store orm.Store) MembershipRepository {
	return MembershipRepository{store: store}
}

// Upsert stores or updates a membership.
func (repository MembershipRepository) Upsert(ctx context.Context, membership domain.Membership) (domain.Membership, bool, error) {
	current, err := repository.Find(ctx, membership.GroupID, membership.UserID)
	if err == nil {
		membership.ID = current.ID
		updated, err := repository.update(ctx, membership, current.Version)
		return updated, false, err
	}
	if !errors.Is(err, port.ErrNotFound) {
		return domain.Membership{}, false, err
	}
	model := membershipModelFromDomain(membership)
	if err := repository.store.DB(ctx).Create(&model).Error; err != nil {
		return domain.Membership{}, false, port.ErrConflict
	}
	return membershipFromModel(model), true, nil
}

// Find returns one membership.
func (repository MembershipRepository) Find(ctx context.Context, groupID uuid.UUID, userID uuid.UUID) (domain.Membership, error) {
	var model MembershipModel
	if err := repository.store.DB(ctx).First(&model, "group_id = ? AND user_id = ?", groupID, userID).Error; err != nil {
		return domain.Membership{}, mapError(err)
	}
	return membershipFromModel(model), nil
}

// ListByGroup returns group memberships.
func (repository MembershipRepository) ListByGroup(
	ctx context.Context,
	groupID uuid.UUID,
	page pagination.Page,
) (pagination.Result[domain.Membership], error) {
	query := repository.store.DB(ctx).Model(&MembershipModel{}).Where("group_id = ?", groupID).Order("created_at asc").Limit(page.Limit + 1)
	var models []MembershipModel
	if err := query.Find(&models).Error; err != nil {
		return pagination.Result[domain.Membership]{}, err
	}
	return membershipPage(models, page.Limit), nil
}

// ListByUser returns user memberships.
func (repository MembershipRepository) ListByUser(ctx context.Context, userID uuid.UUID) ([]domain.Membership, error) {
	var models []MembershipModel
	if err := repository.store.DB(ctx).Find(&models, "user_id = ?", userID).Error; err != nil {
		return nil, err
	}
	items := make([]domain.Membership, 0, len(models))
	for _, model := range models {
		items = append(items, membershipFromModel(model))
	}
	return items, nil
}

// Delete soft deletes a membership.
func (repository MembershipRepository) Delete(ctx context.Context, groupID uuid.UUID, userID uuid.UUID, expectedVersion *uint64) error {
	query := repository.store.DB(ctx).Where("group_id = ? AND user_id = ?", groupID, userID)
	if expectedVersion != nil {
		query = query.Where("version = ?", *expectedVersion)
	}
	result := query.Delete(&MembershipModel{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return port.ErrPreconditionFailed
	}
	return nil
}

// update stores mutable membership fields.
func (repository MembershipRepository) update(
	ctx context.Context,
	membership domain.Membership,
	expectedVersion uint64,
) (domain.Membership, error) {
	result := repository.store.DB(ctx).
		Model(&MembershipModel{}).
		Where("id = ? AND version = ?", membership.ID, expectedVersion).
		Updates(membershipUpdates(membership, expectedVersion))
	if result.Error != nil {
		return domain.Membership{}, result.Error
	}
	if result.RowsAffected == 0 {
		return domain.Membership{}, port.ErrPreconditionFailed
	}
	return repository.Find(ctx, membership.GroupID, membership.UserID)
}

// membershipUpdates returns update fields for a membership.
func membershipUpdates(membership domain.Membership, expectedVersion uint64) map[string]any {
	return map[string]any{
		"status":              string(membership.Status),
		"assigned_by_user_id": membership.AssignedByUserID,
		"assigned_reason":     membership.AssignedReason,
		"starts_at":           membership.StartsAt,
		"expires_at":          membership.ExpiresAt,
		"version":             expectedVersion + 1,
	}
}

// membershipPage maps membership models into a page.
func membershipPage(models []MembershipModel, limit int) pagination.Result[domain.Membership] {
	next := ""
	if len(models) > limit {
		next = models[limit-1].ID.ID.String()
		models = models[:limit]
	}
	items := make([]domain.Membership, 0, len(models))
	for _, model := range models {
		items = append(items, membershipFromModel(model))
	}
	return pagination.Result[domain.Membership]{Items: items, NextCursor: next}
}
