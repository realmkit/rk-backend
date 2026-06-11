package postgres

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/niflaot/gamehub-go/module/punishments/domain"
	"github.com/niflaot/gamehub-go/module/punishments/port"
	"github.com/niflaot/gamehub-go/pkg/orm"
)

// ActiveRestriction returns one matching active restriction.
func (repository CaseRepository) ActiveRestriction(ctx context.Context, userID uuid.UUID, actionKey string, now time.Time) (domain.ActiveRestriction, *domain.PunishmentSummary, error) {
	var model RestrictionModel
	err := repository.store.DB(ctx).
		Where("target_user_id = ? AND action_key = ? AND starts_at <= ?", userID, actionKey, now).
		Where("expires_at IS NULL OR expires_at > ?", now).
		Order("created_at desc, id desc").
		First(&model).Error
	if err != nil {
		return domain.ActiveRestriction{}, nil, mapError(err)
	}
	punishment, err := repository.FindByID(ctx, model.PunishmentID)
	if err != nil {
		return domain.ActiveRestriction{}, nil, err
	}
	summary := &domain.PunishmentSummary{
		ID:        punishment.ID,
		Reason:    punishment.Reason,
		StartsAt:  punishment.StartsAt,
		ExpiresAt: punishment.ExpiresAt,
	}
	return restrictionFromModel(model), summary, nil
}

// ListActiveRestrictions returns active restrictions for a user.
func (repository CaseRepository) ListActiveRestrictions(ctx context.Context, userID uuid.UUID, now time.Time) ([]domain.ActiveRestriction, error) {
	var models []RestrictionModel
	err := repository.store.DB(ctx).
		Where("target_user_id = ? AND starts_at <= ?", userID, now).
		Where("expires_at IS NULL OR expires_at > ?", now).
		Order("action_key, created_at desc").
		Find(&models).Error
	if err != nil {
		return nil, err
	}
	restrictions := make([]domain.ActiveRestriction, 0, len(models))
	for _, model := range models {
		restrictions = append(restrictions, restrictionFromModel(model))
	}
	return restrictions, nil
}

// VerifyRestrictions detects active projection drift.
func (repository CaseRepository) VerifyRestrictions(ctx context.Context, now time.Time) (domain.DriftReport, error) {
	expected, err := repository.expectedRestrictions(ctx, now)
	if err != nil {
		return domain.DriftReport{}, err
	}
	actual, err := repository.actualRestrictions(ctx)
	if err != nil {
		return domain.DriftReport{}, err
	}
	return driftReport(expected, actual, false), nil
}

// RebuildRestrictions rebuilds active restriction projection.
func (repository CaseRepository) RebuildRestrictions(ctx context.Context, now time.Time) (domain.DriftReport, error) {
	report, err := repository.VerifyRestrictions(ctx, now)
	if err != nil {
		return domain.DriftReport{}, err
	}
	if err := repository.store.DB(ctx).Where("1 = 1").Delete(&RestrictionModel{}).Error; err != nil {
		return domain.DriftReport{}, err
	}
	for _, restriction := range repository.expectedSlice(ctx, now) {
		model := restrictionModel(restriction)
		if err := repository.store.DB(ctx).Create(&model).Error; err != nil {
			return domain.DriftReport{}, err
		}
	}
	report.Repaired = true
	return report, nil
}

func (repository CaseRepository) expectedSlice(ctx context.Context, now time.Time) []domain.ActiveRestriction {
	expected, _ := repository.expectedRestrictions(ctx, now)
	out := make([]domain.ActiveRestriction, 0, len(expected))
	for _, restriction := range expected {
		out = append(out, restriction)
	}
	return out
}

func (repository CaseRepository) expectedRestrictions(ctx context.Context, now time.Time) (map[string]domain.ActiveRestriction, error) {
	var punishments []PunishmentModel
	err := repository.store.DB(ctx).
		Where("status = ? AND starts_at <= ?", domain.PunishmentActive, now).
		Where("expires_at IS NULL OR expires_at > ?", now).
		Find(&punishments).Error
	if err != nil {
		return nil, err
	}
	expected := map[string]domain.ActiveRestriction{}
	for _, model := range punishments {
		punishment := punishmentFromModel(model, repository.snapshots(ctx, model.ID.ID))
		for _, snapshot := range punishment.Snapshots {
			restriction, ok := domain.RestrictionFromSnapshot(punishment, snapshot)
			if ok {
				expected[restrictionKey(restriction.PunishmentID, restriction.ActionKey)] = restriction
			}
		}
	}
	return expected, nil
}

func (repository CaseRepository) actualRestrictions(ctx context.Context) (map[string]domain.ActiveRestriction, error) {
	var models []RestrictionModel
	if err := repository.store.DB(ctx).Find(&models).Error; err != nil {
		return nil, err
	}
	actual := map[string]domain.ActiveRestriction{}
	for _, model := range models {
		restriction := restrictionFromModel(model)
		actual[restrictionKey(restriction.PunishmentID, restriction.ActionKey)] = restriction
	}
	return actual, nil
}

// driftReport compares expected and actual active restriction projections.
func driftReport(expected map[string]domain.ActiveRestriction, actual map[string]domain.ActiveRestriction, repaired bool) domain.DriftReport {
	report := domain.DriftReport{Repaired: repaired}
	for key, restriction := range expected {
		if _, ok := actual[key]; !ok {
			report.Mismatches = append(report.Mismatches, domain.CounterDrift{PunishmentID: restriction.PunishmentID, ActionKey: restriction.ActionKey, Expected: true})
		}
	}
	for key, restriction := range actual {
		if _, ok := expected[key]; !ok {
			report.Mismatches = append(report.Mismatches, domain.CounterDrift{PunishmentID: restriction.PunishmentID, ActionKey: restriction.ActionKey, Actual: true})
		}
	}
	return report
}

// restrictionModel maps a domain restriction into persistence state.
func restrictionModel(restriction domain.ActiveRestriction) RestrictionModel {
	return RestrictionModel{
		ID:           orm.ID{ID: restriction.ID},
		PunishmentID: restriction.PunishmentID,
		TargetUserID: restriction.TargetUserID,
		ActionKey:    restriction.ActionKey,
		StartsAt:     restriction.StartsAt,
		ExpiresAt:    restriction.ExpiresAt,
		CreatedAt:    restriction.CreatedAt,
	}
}

// restrictionFromModel maps a persistence restriction row into domain state.
func restrictionFromModel(model RestrictionModel) domain.ActiveRestriction {
	return domain.ActiveRestriction{
		ID:           model.ID.ID,
		PunishmentID: model.PunishmentID,
		TargetUserID: model.TargetUserID,
		ActionKey:    model.ActionKey,
		StartsAt:     model.StartsAt,
		ExpiresAt:    model.ExpiresAt,
		CreatedAt:    model.CreatedAt,
	}
}

// restrictionKey returns a stable comparison key for one restriction projection.
func restrictionKey(punishmentID uuid.UUID, actionKey string) string {
	return punishmentID.String() + ":" + actionKey
}

var _ port.CaseRepository = CaseRepository{}
