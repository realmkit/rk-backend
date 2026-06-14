package application

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/module/groups/domain"
	"github.com/realmkit/rk-backend/module/groups/port"
	"github.com/realmkit/rk-backend/pkg/pagination"
)

// Check returns an authorization decision.
func (service Service) Check(ctx context.Context, request port.CheckRequest) (port.Decision, error) {
	if request.Action == "" {
		return port.Decision{Allowed: false, Reason: "missing_action"}, nil
	}
	if request.ScopeType == "" || request.ScopeID == uuid.Nil {
		return port.Decision{Allowed: false, Reason: "missing_scope"}, nil
	}
	action, err := service.permissionAction(ctx, request.Action)
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

// permissionAction returns a configured action or built-in fallback action.
func (service Service) permissionAction(ctx context.Context, action domain.Action) (domain.PermissionAction, error) {
	if service.permissions != nil {
		found, err := service.permissions.FindAction(ctx, action)
		if err == nil {
			return found, nil
		}
		if err != nil && !errors.Is(err, port.ErrNotFound) {
			return domain.PermissionAction{}, err
		}
	}
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
			Action:    request.Action,
			ScopeType: request.ScopeType,
			ScopeID:   request.ScopeID,
		},
		pagination.Page{Limit: 100},
	)
	if err != nil {
		return port.Decision{}, err
	}
	for _, grant := range grants.Items {
		ok, err := service.subjectMatches(ctx, request.ActorUserID, grant)
		if err != nil {
			return port.Decision{}, err
		}
		if ok {
			return port.Decision{
				Allowed:            true,
				Reason:             "matched_grant",
				MatchedGrantID:     grant.ID,
				MatchedSubjectType: grant.SubjectType,
				MatchedSubjectID:   grant.SubjectID,
				MatchedScopeType:   grant.ScopeType,
				MatchedScopeID:     grant.ScopeID,
			}, nil
		}
	}
	return port.Decision{Allowed: false, Reason: "no_matching_grant"}, nil
}

// subjectMatches reports whether actor matches grant subject.
func (service Service) subjectMatches(ctx context.Context, actorUserID uuid.UUID, grant domain.PermissionGrant) (bool, error) {
	switch grant.SubjectType {
	case domain.SubjectPublic:
		return grant.SubjectID == domain.PublicSubjectID(), nil
	case domain.SubjectAuthenticated:
		return actorUserID != uuid.Nil && grant.SubjectID == domain.AuthenticatedSubjectID(), nil
	case domain.SubjectUser:
		return actorUserID != uuid.Nil && grant.SubjectID == actorUserID, nil
	case domain.SubjectGroup:
		if actorUserID == uuid.Nil {
			return false, nil
		}
		return service.activeGroupMember(ctx, grant.SubjectID, actorUserID)
	default:
		return false, nil
	}
}

// activeGroupMember reports whether user is active in enabled group.
func (service Service) activeGroupMember(ctx context.Context, groupID uuid.UUID, userID uuid.UUID) (bool, error) {
	group, err := service.groups.FindByID(ctx, groupID)
	if errors.Is(err, port.ErrNotFound) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if !group.GrantsPermissions() {
		return false, nil
	}
	membership, err := service.memberships.Find(ctx, groupID, userID)
	if errors.Is(err, port.ErrNotFound) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return membership.ActiveAt(service.clock()), nil
}

// Ensure Service implements service contracts.
var _ port.GroupService = Service{}

// Ensure Service implements membership contracts.
var _ port.MembershipService = Service{}

// Ensure Service implements checker contracts.
var _ port.Checker = Service{}

// Ensure Service implements permission grant contracts.
var _ port.PermissionGrantService = Service{}
