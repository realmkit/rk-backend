package postgres

import (
	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/module/groups/domain"
	"github.com/realmkit/rk-backend/module/groups/port"
	"gorm.io/gorm"
)

// applyGrantFilter applies permission grant filters.
func applyGrantFilter(query *gorm.DB, filter port.PermissionGrantFilter) *gorm.DB {
	query = filterGrantGroup(query, filter)
	query = filterGrantActor(query, filter)
	query = filterGrantAction(query, filter)
	return filterGrantScope(query, filter)
}

// filterGrantGroup limits grants to one assigned group.
func filterGrantGroup(query *gorm.DB, filter port.PermissionGrantFilter) *gorm.DB {
	if filter.GroupID == uuid.Nil {
		return query
	}
	return query.Joins(
		"JOIN group_permission_grants ON group_permission_grants.grant_id = permission_grants.id",
	).Where(
		"group_permission_grants.group_id = ? AND group_permission_grants.deleted_at IS NULL",
		filter.GroupID,
	)
}

// filterGrantActor limits grants to active memberships for one actor.
func filterGrantActor(query *gorm.DB, filter port.PermissionGrantFilter) *gorm.DB {
	if filter.ActorUserID == uuid.Nil {
		return query
	}
	return query.Joins(
		"JOIN group_permission_grants actor_grants ON actor_grants.grant_id = permission_grants.id",
	).Joins(
		"JOIN group_memberships ON group_memberships.group_id = actor_grants.group_id",
	).Joins(
		"JOIN groups ON groups.id = actor_grants.group_id",
	).Where(
		"group_memberships.user_id = ? AND group_memberships.status = ? "+
			"AND group_memberships.deleted_at IS NULL AND actor_grants.deleted_at IS NULL "+
			"AND groups.deleted_at IS NULL AND groups.status IN ?",
		filter.ActorUserID,
		domain.MembershipStatusActive,
		[]domain.GroupStatus{domain.GroupStatusActive, domain.GroupStatusSystem},
	).Distinct("permission_grants.*")
}

// filterGrantAction limits grants by action and scope type.
func filterGrantAction(query *gorm.DB, filter port.PermissionGrantFilter) *gorm.DB {
	if filter.Action != "" {
		query = query.Where("permission_grants.action = ?", filter.Action)
	}
	if filter.ScopeType != "" {
		query = query.Where("permission_grants.scope_type = ?", filter.ScopeType)
	}
	return query
}

// filterGrantScope limits grants by concrete or all-resource scope.
func filterGrantScope(query *gorm.DB, filter port.PermissionGrantFilter) *gorm.DB {
	if filter.AllScopeOnly {
		return query.Where("permission_grants.scope_id = ?", domain.AllScopeID())
	}
	if filter.ScopeID == uuid.Nil {
		return query
	}
	if filter.IncludeAllScopes {
		return query.Where(
			"(permission_grants.scope_id = ? OR permission_grants.scope_id = ?)",
			filter.ScopeID,
			domain.AllScopeID(),
		)
	}
	return query.Where("permission_grants.scope_id = ?", filter.ScopeID)
}
