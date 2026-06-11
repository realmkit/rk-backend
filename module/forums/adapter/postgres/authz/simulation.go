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
	relations, err := simulationRelations(request.Permission)
	if err != nil {
		return forumsdomain.ForumPermissionSimulationResult{}, err
	}
	result := forumsdomain.ForumPermissionSimulationResult{
		Allowed:          false,
		Reason:           "no_matching_relation",
		Permission:       request.Permission,
		ObjectType:       request.ObjectType,
		ObjectID:         request.ObjectID,
		CheckedRelations: relationNames(relations),
	}
	var tuples []relationTupleRow
	err = authorizer.store.DB(ctx).
		Table("authorization_relation_tuples").
		Select("object_id, relation, subject_type, subject_id, subject_relation").
		Where(
			"object_type = ? AND object_id = ? AND relation IN ? AND deleted_at IS NULL",
			groupsdomain.ObjectForum,
			forumID,
			relations,
		).
		Find(&tuples).Error
	if err != nil {
		return forumsdomain.ForumPermissionSimulationResult{}, err
	}
	memberships, err := authorizer.activeMemberships(ctx, request.ActorUserID, groupSubjectIDs(tuples))
	if err != nil {
		return forumsdomain.ForumPermissionSimulationResult{}, err
	}
	for _, tuple := range tuples {
		if tupleMatchesActor(tuple, request.ActorUserID, memberships) {
			result.Allowed = true
			result.Reason = "matched_relation"
			result.MatchedRelation = tuple.Relation
			return result, nil
		}
	}
	return result, nil
}

// simulationRelations returns forum-level relations checked for permission.
func simulationRelations(permission string) ([]groupsdomain.Relation, error) {
	switch groupsdomain.Permission(permission) {
	case groupsdomain.PermissionForumsView,
		groupsdomain.PermissionThreadsView,
		groupsdomain.PermissionPostsView:
		return viewRelations(), nil
	case groupsdomain.PermissionForumsManageForum:
		return manageRelations(), nil
	case groupsdomain.PermissionForumsCreateThread:
		return creatorRelations(), nil
	case groupsdomain.PermissionForumsReply:
		return replyRelations(), nil
	case groupsdomain.PermissionForumsLikePosts,
		groupsdomain.PermissionPostsLike:
		return likeRelations(), nil
	case groupsdomain.PermissionForumsPinThreads,
		groupsdomain.PermissionForumsManageThreads,
		groupsdomain.PermissionForumsManagePosts,
		groupsdomain.PermissionThreadsUpdate,
		groupsdomain.PermissionThreadsClose,
		groupsdomain.PermissionThreadsOpen,
		groupsdomain.PermissionThreadsDelete,
		groupsdomain.PermissionThreadsPin,
		groupsdomain.PermissionPostsUpdate,
		groupsdomain.PermissionPostsDelete,
		groupsdomain.PermissionPostsViewHidden,
		groupsdomain.PermissionPostsViewRevisions:
		return moderateRelations(), nil
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

// relationNames maps relations to strings.
func relationNames(relations []groupsdomain.Relation) []string {
	names := make([]string, 0, len(relations))
	for _, relation := range relations {
		names = append(names, string(relation))
	}
	return names
}
