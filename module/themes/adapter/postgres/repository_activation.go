package postgres

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/module/themes/domain"
	"github.com/realmkit/rk-backend/module/themes/port"
	"github.com/realmkit/rk-backend/pkg/orm"
	"gorm.io/gorm"
)

// ActivationRepository stores active theme pointers in PostgreSQL.
type ActivationRepository struct {
	store orm.Store
}

// NewActivationRepository creates an activation repository.
func NewActivationRepository(store orm.Store) ActivationRepository {
	return ActivationRepository{store: store}
}

// Activate stores a new current activation for an environment.
func (repository ActivationRepository) Activate(
	ctx context.Context,
	activation domain.ThemeActivation,
) (domain.ThemeActivation, error) {
	model := activationModel(activation)
	if model.ActivatedAt.IsZero() {
		model.ActivatedAt = time.Now().UTC()
	}
	model.IsCurrent = true
	err := repository.store.DB(ctx).Transaction(func(tx *gorm.DB) error {
		err := tx.Model(&ActivationModel{}).
			Where("environment = ? AND is_current = ?", model.Environment, true).
			Update("is_current", false).
			Error
		if err != nil {
			return err
		}
		return tx.Create(&model).Error
	})
	if err != nil {
		return domain.ThemeActivation{}, err
	}
	return activationFromModel(model), nil
}

// Current returns the current activation for an environment.
func (repository ActivationRepository) Current(
	ctx context.Context,
	environment domain.ActivationEnvironment,
) (domain.ThemeActivation, error) {
	var model ActivationModel
	err := repository.store.DB(ctx).
		First(&model, "environment = ? AND is_current = ?", environment, true).
		Error
	if err != nil {
		return domain.ThemeActivation{}, mapError(err)
	}
	return activationFromModel(model), nil
}

// FindByID returns one activation history entry.
func (repository ActivationRepository) FindByID(ctx context.Context, id uuid.UUID) (domain.ThemeActivation, error) {
	var model ActivationModel
	if err := repository.store.DB(ctx).First(&model, "id = ?", id).Error; err != nil {
		return domain.ThemeActivation{}, mapError(err)
	}
	return activationFromModel(model), nil
}

// ListByTheme returns activation history for a theme.
func (repository ActivationRepository) ListByTheme(
	ctx context.Context,
	themeID uuid.UUID,
) ([]domain.ThemeActivation, error) {
	var models []ActivationModel
	err := repository.store.DB(ctx).
		Where("theme_id = ?", themeID).
		Order("activated_at DESC, id ASC").
		Find(&models).
		Error
	if err != nil {
		return nil, err
	}
	activations := make([]domain.ThemeActivation, 0, len(models))
	for _, model := range models {
		activations = append(activations, activationFromModel(model))
	}
	return activations, nil
}

// activationModel maps activation to persistence.
func activationModel(activation domain.ThemeActivation) ActivationModel {
	return ActivationModel{
		ID:                orm.ID{ID: activation.ID},
		ThemeID:           activation.ThemeID,
		VersionID:         activation.VersionID,
		Environment:       string(activation.Environment),
		IsCurrent:         activation.IsCurrent,
		Reason:            activation.Reason,
		SettingsDataJSON:  jsonString(activation.SettingsDataJSON),
		ActivatedByUserID: activation.ActivatedBy,
		ActivatedAt:       activation.ActivatedAt,
	}
}

// activationFromModel maps persistence to activation.
func activationFromModel(model ActivationModel) domain.ThemeActivation {
	return domain.ThemeActivation{
		ID:               model.ID.ID,
		ThemeID:          model.ThemeID,
		VersionID:        model.VersionID,
		Environment:      domain.ActivationEnvironment(model.Environment),
		IsCurrent:        model.IsCurrent,
		Reason:           model.Reason,
		SettingsDataJSON: []byte(model.SettingsDataJSON),
		ActivatedBy:      model.ActivatedByUserID,
		ActivatedAt:      model.ActivatedAt,
		CreatedAt:        model.CreatedAt,
	}
}

// Ensure activation repository implements its port.
var _ port.ActivationRepository = ActivationRepository{}
