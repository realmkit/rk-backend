package postgres

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/module/punishments/domain"
	"github.com/realmkit/rk-backend/module/punishments/port"
	"github.com/realmkit/rk-backend/pkg/orm"
	"github.com/realmkit/rk-backend/pkg/pagination"
	"github.com/realmkit/rk-backend/pkg/search"
)

// CaseRepository stores issued punishments and restriction projections.
type CaseRepository struct {
	store orm.Store // store stores the store value.
}

// NewCaseRepository creates a punishment case repository.
func NewCaseRepository(store orm.Store) CaseRepository {
	return CaseRepository{store: store}
}

// Issue stores a punishment with snapshots and active restrictions.
func (repository CaseRepository) Issue(
	ctx context.Context,
	punishment domain.Punishment,
	restrictions []domain.ActiveRestriction,
) (domain.Punishment, error) {
	model := punishmentModel(punishment)
	if err := repository.store.DB(ctx).Create(&model).Error; err != nil {
		return domain.Punishment{}, port.ErrConflict
	}
	for _, snapshot := range punishment.Snapshots {
		model := snapshotModel(snapshot)
		if err := repository.store.DB(ctx).Create(&model).Error; err != nil {
			return domain.Punishment{}, err
		}
	}
	for _, restriction := range restrictions {
		model := restrictionModel(restriction)
		if err := repository.store.DB(ctx).Create(&model).Error; err != nil {
			return domain.Punishment{}, err
		}
	}
	return repository.FindByID(ctx, punishment.ID)
}

// Update updates punishment note fields.
func (repository CaseRepository) Update(
	ctx context.Context,
	punishment domain.Punishment,
	expectedVersion uint64,
) (domain.Punishment, error) {
	result := repository.store.DB(ctx).Model(&PunishmentModel{}).
		Where("id = ? AND version = ?", punishment.ID, expectedVersion).
		Updates(map[string]any{
			"reason":         punishment.Reason,
			"private_reason": punishment.PrivateReason,
			"version":        expectedVersion + 1,
		})
	if result.Error != nil {
		return domain.Punishment{}, result.Error
	}
	if result.RowsAffected == 0 {
		return domain.Punishment{}, port.ErrPreconditionFailed
	}
	return repository.FindByID(ctx, punishment.ID)
}

// Revoke revokes a punishment and deletes restrictions.
func (repository CaseRepository) Revoke(ctx context.Context, punishment domain.Punishment, expectedVersion uint64) error {
	result := repository.store.DB(ctx).Model(&PunishmentModel{}).
		Where("id = ? AND version = ?", punishment.ID, expectedVersion).
		Updates(map[string]any{
			"status":             string(punishment.Status),
			"revoked_at":         punishment.RevokedAt,
			"revoked_by_user_id": punishment.RevokedByUserID,
			"revocation_reason":  punishment.RevocationReason,
			"version":            expectedVersion + 1,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return port.ErrPreconditionFailed
	}
	return repository.store.DB(ctx).Where("punishment_id = ?", punishment.ID).Delete(&RestrictionModel{}).Error
}

// ExpireDue expires due punishments and removes restrictions.
func (repository CaseRepository) ExpireDue(ctx context.Context, now time.Time) (int64, error) {
	var models []PunishmentModel
	if err := repository.store.DB(ctx).
		Where("status = ? AND expires_at IS NOT NULL AND expires_at <= ?", domain.PunishmentActive, now).
		Find(&models).Error; err != nil {
		return 0, err
	}
	for _, model := range models {
		if err := repository.store.DB(ctx).Model(&PunishmentModel{}).
			Where("id = ?", model.ID.ID).
			Updates(map[string]any{"status": string(domain.PunishmentExpired), "version": model.Version + 1}).Error; err != nil {
			return 0, err
		}
		if err := repository.store.DB(ctx).Where("punishment_id = ?", model.ID.ID).Delete(&RestrictionModel{}).Error; err != nil {
			return 0, err
		}
	}
	return int64(len(models)), nil
}

// FindByID returns one punishment.
func (repository CaseRepository) FindByID(ctx context.Context, id uuid.UUID) (domain.Punishment, error) {
	var model PunishmentModel
	if err := repository.store.DB(ctx).First(&model, "id = ?", id).Error; err != nil {
		return domain.Punishment{}, mapError(err)
	}
	return punishmentFromModel(model, repository.snapshots(ctx, id)), nil
}

// FindByIdempotencyKey returns one punishment by idempotency key.
func (repository CaseRepository) FindByIdempotencyKey(ctx context.Context, key string) (domain.Punishment, error) {
	var model PunishmentModel
	if err := repository.store.DB(ctx).First(&model, "idempotency_key = ?", key).Error; err != nil {
		return domain.Punishment{}, mapError(err)
	}
	return punishmentFromModel(model, repository.snapshots(ctx, model.ID.ID)), nil
}

// List returns punishments.
func (repository CaseRepository) List(
	ctx context.Context,
	filter port.PunishmentFilter,
	page pagination.Page,
) (pagination.Result[domain.Punishment], error) {
	sort := filter.Sort
	if sort.Key == "" {
		sort, _ = search.NewSort("", "", port.DefaultPunishmentSort(), port.AllowedPunishmentSorts())
	}
	filterHash := punishmentFilterHash(filter, sort)
	cursor, hasCursor, err := search.RequireCursor(page.Cursor, filterHash, sort)
	if err != nil {
		return pagination.Result[domain.Punishment]{}, err
	}
	query := applyPunishmentFilter(repository.store.DB(ctx).Model(&PunishmentModel{}), filter)
	query, err = applyPunishmentCursor(query, cursor, hasCursor, sort)
	if err != nil {
		return pagination.Result[domain.Punishment]{}, err
	}
	query = query.Order(punishmentOrder(sort)).Limit(page.Limit + 1)
	var models []PunishmentModel
	if err := query.Find(&models).Error; err != nil {
		return pagination.Result[domain.Punishment]{}, err
	}
	next := ""
	if len(models) > page.Limit {
		next, err = punishmentCursor(models[page.Limit-1], filterHash, sort)
		if err != nil {
			return pagination.Result[domain.Punishment]{}, err
		}
		models = models[:page.Limit]
	}
	items := make([]domain.Punishment, 0, len(models))
	for _, model := range models {
		items = append(items, punishmentFromModel(model, repository.snapshots(ctx, model.ID.ID)))
	}
	return pagination.Result[domain.Punishment]{Items: items, NextCursor: next}, nil
}

// snapshots supports package behavior.
func (repository CaseRepository) snapshots(ctx context.Context, punishmentID uuid.UUID) []SnapshotModel {
	var snapshots []SnapshotModel
	_ = repository.store.DB(ctx).Where("punishment_id = ?", punishmentID).Order("created_at, id").Find(&snapshots).Error
	return snapshots
}

// snapshotModel maps a domain action snapshot into persistence state.
func snapshotModel(snapshot domain.ActionSnapshot) SnapshotModel {
	return SnapshotModel{
		ID:                 orm.ID{ID: snapshot.ID},
		PunishmentID:       snapshot.PunishmentID,
		DefinitionActionID: snapshot.DefinitionActionID,
		TargetSystem:       string(snapshot.TargetSystem),
		ActionType:         string(snapshot.ActionType),
		ConfigurationJSON:  string(snapshot.ConfigurationJSON),
		Status:             string(snapshot.Status),
		CreatedAt:          snapshot.CreatedAt,
	}
}

// snapshotFromModel maps a persistence snapshot row into domain state.
func snapshotFromModel(model SnapshotModel) domain.ActionSnapshot {
	return domain.ActionSnapshot{
		ID:                 model.ID.ID,
		PunishmentID:       model.PunishmentID,
		DefinitionActionID: model.DefinitionActionID,
		TargetSystem:       domain.TargetSystem(model.TargetSystem),
		ActionType:         domain.ActionType(model.ActionType),
		ConfigurationJSON:  json.RawMessage(model.ConfigurationJSON),
		Status:             domain.DefinitionStatus(model.Status),
		CreatedAt:          model.CreatedAt,
	}
}
