package authz

import (
	"context"

	"github.com/google/uuid"
	forumsdomain "github.com/realmkit/rk-backend/module/forums/domain"
	groupsdomain "github.com/realmkit/rk-backend/module/groups/domain"
)

// SimulateForumPermission explains a forum permission decision.
func (authorizer VisibilityAuthorizer) SimulateForumPermission(
	ctx context.Context,
	forumID uuid.UUID,
	request forumsdomain.ForumPermissionSimulationRequest,
) (forumsdomain.ForumPermissionSimulationResult, error) {
	request = request.Normalize(forumID)
	if err := request.Validate(); err != nil {
		return forumsdomain.ForumPermissionSimulationResult{}, err
	}
	actions, err := simulationActions(request.Permission)
	if err != nil {
		return forumsdomain.ForumPermissionSimulationResult{}, err
	}
	result := forumsdomain.ForumPermissionSimulationResult{
		Allowed:        false,
		Reason:         "no_matching_grant",
		Permission:     request.Permission,
		ObjectType:     request.ObjectType,
		ObjectID:       request.ObjectID,
		CheckedActions: actionNames(actions),
	}
	var grants []permissionGrantRow
	err = authorizer.store.DB(ctx).
		Table("forum_permission_grants").
		Select("scope_id, action, subject_type, subject_id").
		Where(
			"scope_type = ? AND scope_id = ? AND action IN ? AND deleted_at IS NULL",
			groupsdomain.ObjectForum,
			forumID,
			actions,
		).
		Find(&grants).Error
	if err != nil {
		return forumsdomain.ForumPermissionSimulationResult{}, err
	}
	groupGrants, err := authorizer.groupPermissionGrants(ctx, []uuid.UUID{forumID}, actions)
	if err != nil {
		return forumsdomain.ForumPermissionSimulationResult{}, err
	}
	memberships, err := authorizer.activeMemberships(
		ctx,
		request.ActorUserID,
		grantGroupIDs(grants, groupGrants),
	)
	if err != nil {
		return forumsdomain.ForumPermissionSimulationResult{}, err
	}
	for _, grant := range grants {
		if grantMatchesActor(grant, request.ActorUserID, memberships) {
			result.Allowed = true
			result.Reason = "matched_grant"
			result.MatchedAction = grant.Action
			return result, nil
		}
	}
	for _, grant := range groupGrants {
		if groupGrantMatchesActor(grant, memberships) {
			result.Allowed = true
			result.Reason = "matched_grant"
			result.MatchedAction = grant.Action
			return result, nil
		}
	}
	return result, nil
}

// simulationActions returns forum-level actions checked for permission.
func simulationActions(permission string) ([]groupsdomain.Action, error) {
	switch groupsdomain.Permission(permission) {
	case groupsdomain.PermissionForumsView,
		groupsdomain.PermissionThreadsView,
		groupsdomain.PermissionPostsView:
		return viewActions(), nil
	case groupsdomain.PermissionForumsManageForum:
		return manageActions(), nil
	case groupsdomain.PermissionForumsCreateThread:
		return createThreadActions(), nil
	case groupsdomain.PermissionForumsReply:
		return replyActions(), nil
	case groupsdomain.PermissionForumsLikePosts,
		groupsdomain.PermissionPostsLike:
		return likeActions(), nil
	case groupsdomain.PermissionForumsPinThreads,
		groupsdomain.PermissionThreadsPin:
		return pinThreadActions(), nil
	case groupsdomain.PermissionForumsManageThreads,
		groupsdomain.PermissionThreadsClose,
		groupsdomain.PermissionThreadsOpen,
		groupsdomain.PermissionThreadsUpdate,
		groupsdomain.PermissionThreadsDelete:
		return threadManageActions(), nil
	case groupsdomain.PermissionForumsManagePosts,
		groupsdomain.PermissionPostsUpdate,
		groupsdomain.PermissionPostsDelete:
		return postManageActions(), nil
	case groupsdomain.PermissionForumsBypassThreadLimits:
		return limitBypassActions(), nil
	case groupsdomain.PermissionForumsViewAllThreads,
		groupsdomain.PermissionPostsViewHidden,
		groupsdomain.PermissionPostsViewRevisions:
		return viewAllThreadActions(), nil
	case groupsdomain.PermissionForumsAdministrativeAccess:
		return []groupsdomain.Action{groupsdomain.PermissionForumsAdministrativeAccess}, nil
	default:
		return nil, unsupportedSimulationPermission()
	}
}

// unsupportedSimulationPermission returns a validation error for unsupported permissions.
func unsupportedSimulationPermission() error {
	return forumsdomain.NewValidationError([]forumsdomain.Violation{{
		Field:   "permission",
		Message: "is not supported for forum simulation",
	}})
}

// actionNames maps actions to strings.
func actionNames(actions []groupsdomain.Action) []string {
	names := make([]string, 0, len(actions))
	for _, action := range actions {
		names = append(names, string(action))
	}
	return names
}
