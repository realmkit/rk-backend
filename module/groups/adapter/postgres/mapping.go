package postgres

import (
	"errors"

	"github.com/realmkit/rk-backend/module/groups/domain"
	"github.com/realmkit/rk-backend/module/groups/port"
	"github.com/realmkit/rk-backend/pkg/orm"
	"gorm.io/gorm"
)

// groupModelFromDomain maps domain group to persistence.
func groupModelFromDomain(group domain.Group) GroupModel {
	return GroupModel{
		ID:          orm.ID{ID: group.ID},
		Key:         string(group.Key),
		Name:        group.Name,
		Description: group.Description,
		Color:       string(group.Color),
		Weight:      group.Weight,
		Status:      string(group.Status),
		IconAssetID: group.IconAssetID,
		Version:     group.Version,
	}
}

// groupFromModel maps persistence group to domain.
func groupFromModel(model GroupModel) domain.Group {
	return domain.Group{
		ID:          model.ID.ID,
		Key:         domain.Key(model.Key),
		Name:        model.Name,
		Description: model.Description,
		Color:       domain.Color(model.Color),
		Weight:      model.Weight,
		Status:      domain.GroupStatus(model.Status),
		IconAssetID: model.IconAssetID,
		Version:     model.Version,
		CreatedAt:   model.CreatedAt,
		UpdatedAt:   model.UpdatedAt,
	}
}

// groupUpdates returns update fields for a group.
func groupUpdates(group domain.Group, expectedVersion uint64) map[string]any {
	return map[string]any{
		"name":          group.Name,
		"description":   group.Description,
		"color":         string(group.Color),
		"weight":        group.Weight,
		"status":        string(group.Status),
		"icon_asset_id": group.IconAssetID,
		"version":       expectedVersion + 1,
	}
}

// groupPostgresSearchCondition returns exact full-text plus prefix fallback search.
func groupPostgresSearchCondition() string {
	return `
		to_tsvector(
			'simple',
			coalesce(key, '') || ' ' || coalesce(name, '') || ' ' || coalesce(description, '')
		) @@ plainto_tsquery('simple', ?)
		OR LOWER(key) LIKE ?
		OR LOWER(name) LIKE ?
		OR LOWER(description) LIKE ?
	`
}

// mapError maps GORM errors into groups errors.
func mapError(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return port.ErrNotFound
	}
	return err
}

// membershipModelFromDomain maps domain membership to persistence.
func membershipModelFromDomain(membership domain.Membership) MembershipModel {
	return MembershipModel{
		ID:               orm.ID{ID: membership.ID},
		GroupID:          membership.GroupID,
		UserID:           membership.UserID,
		Status:           string(membership.Status),
		AssignedByUserID: membership.AssignedByUserID,
		AssignedReason:   membership.AssignedReason,
		StartsAt:         membership.StartsAt,
		ExpiresAt:        membership.ExpiresAt,
		Version:          membership.Version,
	}
}

// membershipFromModel maps persistence membership to domain.
func membershipFromModel(model MembershipModel) domain.Membership {
	return domain.Membership{
		ID:               model.ID.ID,
		GroupID:          model.GroupID,
		UserID:           model.UserID,
		Status:           domain.MembershipStatus(model.Status),
		AssignedByUserID: model.AssignedByUserID,
		AssignedReason:   model.AssignedReason,
		StartsAt:         model.StartsAt,
		ExpiresAt:        model.ExpiresAt,
		Version:          model.Version,
		CreatedAt:        model.CreatedAt,
		UpdatedAt:        model.UpdatedAt,
	}
}

// actionModelFromDomain maps domain permission action to persistence.
func actionModelFromDomain(action domain.PermissionAction) PermissionActionModel {
	return PermissionActionModel{
		ID:           orm.ID{ID: action.ID},
		Action:       string(action.Action),
		Area:         action.Area,
		ScopeType:    string(action.ScopeType),
		Label:        action.Label,
		Description:  action.Description,
		WarningLevel: string(action.WarningLevel),
		Enabled:      action.Enabled,
		Version:      action.Version,
	}
}

// actionFromModel maps persistence permission action to domain.
func actionFromModel(model PermissionActionModel) domain.PermissionAction {
	return domain.PermissionAction{
		ID:           model.ID.ID,
		Action:       domain.Action(model.Action),
		Area:         model.Area,
		ScopeType:    domain.ScopeType(model.ScopeType),
		Label:        model.Label,
		Description:  model.Description,
		WarningLevel: domain.WarningLevel(model.WarningLevel),
		Enabled:      model.Enabled,
		Version:      model.Version,
		CreatedAt:    model.CreatedAt,
		UpdatedAt:    model.UpdatedAt,
	}
}

// grantModelFromDomain maps domain permission grant to persistence.
func grantModelFromDomain(grant domain.PermissionGrant) PermissionGrantModel {
	return PermissionGrantModel{
		ID:              orm.ID{ID: grant.ID},
		SubjectType:     string(grant.SubjectType),
		SubjectID:       grant.SubjectID,
		Action:          string(grant.Action),
		ScopeType:       string(grant.ScopeType),
		ScopeID:         grant.ScopeID,
		Inherit:         grant.Inherit,
		ConditionKey:    grant.ConditionKey,
		CreatedByUserID: grant.CreatedByUserID,
	}
}

// grantFromModel maps persistence permission grant to domain.
func grantFromModel(model PermissionGrantModel) domain.PermissionGrant {
	return domain.PermissionGrant{
		ID:              model.ID.ID,
		SubjectType:     domain.SubjectType(model.SubjectType),
		SubjectID:       model.SubjectID,
		Action:          domain.Action(model.Action),
		ScopeType:       domain.ScopeType(model.ScopeType),
		ScopeID:         model.ScopeID,
		Inherit:         model.Inherit,
		ConditionKey:    model.ConditionKey,
		CreatedByUserID: model.CreatedByUserID,
		CreatedAt:       model.CreatedAt,
	}
}
