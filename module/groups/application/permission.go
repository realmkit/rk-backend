package application

import (
	"context"

	"github.com/realmkit/rk-backend/module/groups/domain"
	"github.com/realmkit/rk-backend/module/groups/port"
	"github.com/realmkit/rk-backend/pkg/pagination"
)

// Check returns an authorization decision.
func (service Service) Check(ctx context.Context, request port.CheckRequest) (port.Decision, error) {
	if request.Action == "" {
		return port.Decision{Allowed: false, Reason: "missing_action"}, nil
	}
	if request.ScopeType == "" {
		return port.Decision{Allowed: false, Reason: "missing_scope"}, nil
	}
	action, err := permissionAction(request.Action)
	if err != nil {
		return port.Decision{Allowed: false, Reason: "unknown_permission"}, err
	}
	if !action.Enabled {
		return port.Decision{Allowed: false, Reason: "permission_disabled"}, nil
	}
	if action.ScopeType != request.ScopeType {
		return port.Decision{Allowed: false, Reason: "scope_type_mismatch"}, nil
	}
	return service.checkGrants(ctx, request)
}

// permissionAction returns one app-owned action definition.
func permissionAction(action domain.Action) (domain.PermissionAction, error) {
	found, ok := staticPermissionActions[action]
	if !ok {
		return domain.PermissionAction{}, port.ErrUnknownPermission
	}
	return found, nil
}

// checkGrants checks action grants for the requested scope.
func (service Service) checkGrants(ctx context.Context, request port.CheckRequest) (port.Decision, error) {
	grants, err := service.permissions.ListGrants(
		ctx,
		port.PermissionGrantFilter{
			ActorUserID:      request.ActorUserID,
			Action:           request.Action,
			ScopeType:        request.ScopeType,
			ScopeID:          request.ScopeID,
			IncludeAllScopes: true,
			AllScopeOnly:     request.ScopeID == domain.AllScopeID(),
		},
		pagination.Page{Limit: 100},
	)
	if err != nil {
		return port.Decision{}, err
	}
	for _, grant := range grants.Items {
		return port.Decision{
			Allowed:          true,
			Reason:           "matched_grant",
			MatchedGrantID:   grant.ID,
			MatchedScopeType: grant.ScopeType,
			MatchedScopeID:   grant.ScopeID,
		}, nil
	}
	return port.Decision{Allowed: false, Reason: "no_matching_grant"}, nil
}

// Ensure Service implements service contracts.
var _ port.GroupService = Service{}

// Ensure Service implements membership contracts.
var _ port.MembershipService = Service{}

// Ensure Service implements checker contracts.
var _ port.Checker = Service{}

// Ensure Service implements permission grant contracts.
var _ port.PermissionGrantService = Service{}
