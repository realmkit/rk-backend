package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/niflaot/gamehub-go/module/groups/domain"
	"github.com/niflaot/gamehub-go/pkg/orm"
	"gorm.io/gorm"
)

// PermissionRepository stores permission policy records in PostgreSQL.
type PermissionRepository struct {
	store orm.Store
}

// NewPermissionRepository creates a permission repository.
func NewPermissionRepository(store orm.Store) PermissionRepository {
	return PermissionRepository{store: store}
}

// UpsertDefinition stores or updates a permission definition.
func (repository PermissionRepository) UpsertDefinition(
	ctx context.Context,
	definition domain.PermissionDefinition,
) (domain.PermissionDefinition, error) {
	if err := definition.Validate(); err != nil {
		return domain.PermissionDefinition{}, err
	}
	model := definitionModelFromDomain(definition)
	var current PermissionDefinitionModel
	err := repository.store.DB(ctx).First(&current, "permission = ?", definition.Permission).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		if err := repository.store.DB(ctx).Create(&model).Error; err != nil {
			return domain.PermissionDefinition{}, err
		}
		return definitionFromModel(model), nil
	}
	if err != nil {
		return domain.PermissionDefinition{}, err
	}
	model.ID = current.ID
	model.Version = current.Version + 1
	err = repository.store.DB(ctx).
		Model(&PermissionDefinitionModel{}).
		Where("id = ?", current.ID.ID).
		Updates(definitionUpdates(model)).
		Error
	if err != nil {
		return domain.PermissionDefinition{}, err
	}
	return repository.FindDefinition(ctx, definition.Permission)
}

// FindDefinition returns one active permission definition.
func (repository PermissionRepository) FindDefinition(
	ctx context.Context,
	permission domain.Permission,
) (domain.PermissionDefinition, error) {
	var model PermissionDefinitionModel
	if err := repository.store.DB(ctx).First(&model, "permission = ?", permission).Error; err != nil {
		return domain.PermissionDefinition{}, mapError(err)
	}
	return definitionFromModel(model), nil
}

// UpsertRule stores or updates a permission rule.
func (repository PermissionRepository) UpsertRule(ctx context.Context, rule domain.PermissionRule) (domain.PermissionRule, error) {
	if err := rule.Validate(); err != nil {
		return domain.PermissionRule{}, err
	}
	model, err := ruleModelFromDomain(rule)
	if err != nil {
		return domain.PermissionRule{}, err
	}
	var current PermissionRuleModel
	err = repository.store.DB(ctx).First(&current, "id = ?", rule.ID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		if err := repository.store.DB(ctx).Create(&model).Error; err != nil {
			return domain.PermissionRule{}, err
		}
		return ruleFromModel(model)
	}
	if err != nil {
		return domain.PermissionRule{}, err
	}
	err = repository.store.DB(ctx).
		Model(&PermissionRuleModel{}).
		Where("id = ?", current.ID.ID).
		Updates(ruleUpdates(model)).
		Error
	if err != nil {
		return domain.PermissionRule{}, err
	}
	return repository.findRuleByID(ctx, rule.ID)
}

// ListRules returns active rules for a permission.
func (repository PermissionRepository) ListRules(ctx context.Context, permission domain.Permission) ([]domain.PermissionRule, error) {
	var models []PermissionRuleModel
	err := repository.store.DB(ctx).
		Where("permission = ? AND enabled = ?", permission, true).
		Order("priority asc").
		Find(&models).
		Error
	if err != nil {
		return nil, err
	}
	rules := make([]domain.PermissionRule, 0, len(models))
	for _, model := range models {
		rule, err := ruleFromModel(model)
		if err != nil {
			return nil, err
		}
		rules = append(rules, rule)
	}
	return rules, nil
}

// definitionUpdates returns update fields for a permission definition.
func definitionUpdates(model PermissionDefinitionModel) map[string]any {
	return map[string]any{
		"object_type": model.ObjectType,
		"description": model.Description,
		"enabled":     model.Enabled,
		"version":     model.Version,
	}
}

// ruleUpdates returns update fields for a permission rule.
func ruleUpdates(model PermissionRuleModel) map[string]any {
	return map[string]any{
		"permission":      model.Permission,
		"object_type":     model.ObjectType,
		"relation":        model.Relation,
		"conditions_json": model.ConditionsJSON,
		"priority":        model.Priority,
		"enabled":         model.Enabled,
	}
}

// findRuleByID returns one active rule by ID.
func (repository PermissionRepository) findRuleByID(ctx context.Context, id uuid.UUID) (domain.PermissionRule, error) {
	var model PermissionRuleModel
	if err := repository.store.DB(ctx).First(&model, "id = ?", id).Error; err != nil {
		return domain.PermissionRule{}, mapError(err)
	}
	return ruleFromModel(model)
}
