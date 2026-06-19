package postgres

import (
	"encoding/json"
	"errors"

	"github.com/realmkit/rk-backend/module/punishments/domain"
	"github.com/realmkit/rk-backend/module/punishments/port"
	"github.com/realmkit/rk-backend/pkg/orm"
	"github.com/realmkit/rk-backend/pkg/search"
	"gorm.io/gorm"
)

// definitionModel maps a domain definition into persistence state.
func definitionModel(definition domain.Definition) DefinitionModel {
	return DefinitionModel{
		ID:                     orm.ID{ID: definition.ID},
		Key:                    string(definition.Key),
		Name:                   definition.Name,
		Description:            definition.Description,
		Color:                  string(definition.Color),
		Severity:               definition.Severity,
		Status:                 string(definition.Status),
		DefaultDurationSeconds: definition.DefaultDurationSeconds,
		MinDurationSeconds:     definition.MinDurationSeconds,
		MaxDurationSeconds:     definition.MaxDurationSeconds,
		AllowPermanent:         definition.AllowPermanent,
		RequiresReason:         definition.RequiresReason,
		RequiresTargetIP:       definition.RequiresTargetIP,
		DisplayOrder:           definition.DisplayOrder,
		Version:                definition.Version,
	}
}

// definitionFromModel maps persistence definition rows into domain state.
func definitionFromModel(model DefinitionModel, actions []ActionModel) domain.Definition {
	definition := domain.Definition{
		ID:                     model.ID.ID,
		Key:                    domain.Key(model.Key),
		Name:                   model.Name,
		Description:            model.Description,
		Color:                  domain.Color(model.Color),
		Severity:               model.Severity,
		Status:                 domain.DefinitionStatus(model.Status),
		DefaultDurationSeconds: model.DefaultDurationSeconds,
		MinDurationSeconds:     model.MinDurationSeconds,
		MaxDurationSeconds:     model.MaxDurationSeconds,
		AllowPermanent:         model.AllowPermanent,
		RequiresReason:         model.RequiresReason,
		RequiresTargetIP:       model.RequiresTargetIP,
		DisplayOrder:           model.DisplayOrder,
		Version:                model.Version,
		CreatedAt:              model.CreatedAt,
		UpdatedAt:              model.UpdatedAt,
	}
	for _, action := range actions {
		definition.Actions = append(definition.Actions, actionFromModel(action))
	}
	return definition
}

// definitionFilterHash binds cursors to definition filters.
func definitionFilterHash(filter port.DefinitionFilter, sort search.Sort) string {
	return search.HashFilter(filter.Status, filter.Query.String(), sort)
}

// mapError translates GORM and ORM errors to punishment ports.
func mapError(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return port.ErrNotFound
	}
	if errors.Is(orm.TranslateError(err), orm.ErrConflict) {
		return port.ErrConflict
	}
	return err
}

// actionModel maps a domain action template into persistence state.
func actionModel(action domain.ActionTemplate) ActionModel {
	return ActionModel{
		ID:                orm.ID{ID: action.ID},
		DefinitionID:      action.DefinitionID,
		TargetSystem:      string(action.TargetSystem),
		ActionType:        string(action.ActionType),
		ConfigurationJSON: string(action.ConfigurationJSON),
		DisplayOrder:      action.DisplayOrder,
		Status:            string(action.Status),
	}
}

// actionFromModel maps a persistence action row into domain state.
func actionFromModel(model ActionModel) domain.ActionTemplate {
	return domain.ActionTemplate{
		ID:                model.ID.ID,
		DefinitionID:      model.DefinitionID,
		TargetSystem:      domain.TargetSystem(model.TargetSystem),
		ActionType:        domain.ActionType(model.ActionType),
		ConfigurationJSON: json.RawMessage(model.ConfigurationJSON),
		DisplayOrder:      model.DisplayOrder,
		Status:            domain.DefinitionStatus(model.Status),
		CreatedAt:         model.CreatedAt,
		UpdatedAt:         model.UpdatedAt,
	}
}

// punishmentModel maps a domain punishment into persistence state.
func punishmentModel(punishment domain.Punishment) PunishmentModel {
	return PunishmentModel{
		ID:                 orm.ID{ID: punishment.ID},
		DefinitionID:       punishment.DefinitionID,
		TargetUserID:       punishment.TargetUserID,
		TargetIPHash:       punishment.TargetIPHash,
		TargetIPCiphertext: punishment.TargetIPCiphertext,
		IssuerType:         string(punishment.IssuerType),
		IssuerUserID:       punishment.IssuerUserID,
		IssuerKey:          punishment.IssuerKey,
		Reason:             punishment.Reason,
		PrivateReason:      punishment.PrivateReason,
		Status:             string(punishment.Status),
		StartsAt:           punishment.StartsAt,
		ExpiresAt:          punishment.ExpiresAt,
		RevokedAt:          punishment.RevokedAt,
		RevokedByUserID:    punishment.RevokedByUserID,
		RevocationReason:   punishment.RevocationReason,
		Source:             punishment.Source,
		IdempotencyKey:     punishment.IdempotencyKey,
		Version:            punishment.Version,
	}
}

// punishmentFromModel maps persistence punishment rows into domain state.
func punishmentFromModel(model PunishmentModel, snapshots []SnapshotModel) domain.Punishment {
	punishment := domain.Punishment{
		ID:                 model.ID.ID,
		DefinitionID:       model.DefinitionID,
		TargetUserID:       model.TargetUserID,
		TargetIPHash:       model.TargetIPHash,
		TargetIPCiphertext: model.TargetIPCiphertext,
		IssuerType:         domain.IssuerType(model.IssuerType),
		IssuerUserID:       model.IssuerUserID,
		IssuerKey:          model.IssuerKey,
		Reason:             model.Reason,
		PrivateReason:      model.PrivateReason,
		Status:             domain.PunishmentStatus(model.Status),
		StartsAt:           model.StartsAt,
		ExpiresAt:          model.ExpiresAt,
		RevokedAt:          model.RevokedAt,
		RevokedByUserID:    model.RevokedByUserID,
		RevocationReason:   model.RevocationReason,
		Source:             model.Source,
		IdempotencyKey:     model.IdempotencyKey,
		Version:            model.Version,
		CreatedAt:          model.CreatedAt,
		UpdatedAt:          model.UpdatedAt,
	}
	for _, snapshot := range snapshots {
		punishment.Snapshots = append(punishment.Snapshots, snapshotFromModel(snapshot))
	}
	return punishment
}
