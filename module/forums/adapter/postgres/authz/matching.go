package authz

import (
	"context"
	"slices"

	"github.com/google/uuid"
	groupsdomain "github.com/realmkit/rk-backend/module/groups/domain"
)

// activeMemberships returns actor group memberships that can grant permissions.
func (authorizer VisibilityAuthorizer) activeMemberships(
	ctx context.Context,
	actorUserID uuid.UUID,
	groupIDs []uuid.UUID,
) (map[uuid.UUID]bool, error) {
	result := map[uuid.UUID]bool{}
	if actorUserID == uuid.Nil || len(groupIDs) == 0 {
		return result, nil
	}
	var rows []groupMembershipRow
	err := authorizer.store.DB(ctx).
		Table("group_memberships").
		Joins("JOIN groups ON groups.id = group_memberships.group_id").
		Where(
			"group_memberships.user_id = ? AND group_memberships.group_id IN ? "+
				"AND group_memberships.status = ? AND group_memberships.deleted_at IS NULL "+
				"AND groups.deleted_at IS NULL AND groups.status IN ?",
			actorUserID,
			groupIDs,
			groupsdomain.MembershipStatusActive,
			[]groupsdomain.GroupStatus{
				groupsdomain.GroupStatusActive,
				groupsdomain.GroupStatusSystem,
			},
		).
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	for _, row := range rows {
		result[row.GroupID] = true
	}
	return result, nil
}

// groupSubjectIDs extracts group subject IDs.
func groupSubjectIDs(grants []permissionGrantRow) []uuid.UUID {
	groupIDs := []uuid.UUID{}
	for _, grant := range grants {
		if grant.SubjectType != string(groupsdomain.SubjectGroup) {
			continue
		}
		if !slices.Contains(groupIDs, grant.SubjectID) {
			groupIDs = append(groupIDs, grant.SubjectID)
		}
	}
	return groupIDs
}

// grantMatchesActor reports whether grant allows actor.
func grantMatchesActor(
	grant permissionGrantRow,
	actorUserID uuid.UUID,
	memberships map[uuid.UUID]bool,
) bool {
	switch groupsdomain.SubjectType(grant.SubjectType) {
	case groupsdomain.SubjectPublic:
		return grant.SubjectID == groupsdomain.PublicSubjectID()
	case groupsdomain.SubjectAuthenticated:
		return actorUserID != uuid.Nil && grant.SubjectID == groupsdomain.AuthenticatedSubjectID()
	case groupsdomain.SubjectUser:
		return actorUserID != uuid.Nil && grant.SubjectID == actorUserID
	case groupsdomain.SubjectGroup:
		return memberships[grant.SubjectID]
	default:
		return false
	}
}

// groupMembershipRow is a compact group membership projection.
type groupMembershipRow struct {
	GroupID uuid.UUID
}

// viewActions returns actions that grant viewing.
func viewActions() []groupsdomain.Action {
	return []groupsdomain.Action{
		groupsdomain.PermissionForumsView,
		groupsdomain.PermissionForumsManageForum,
	}
}

// manageActions returns actions that grant structure management.
func manageActions() []groupsdomain.Action {
	return []groupsdomain.Action{groupsdomain.PermissionForumsManageForum}
}

// createThreadActions returns actions that grant thread creation.
func createThreadActions() []groupsdomain.Action {
	return []groupsdomain.Action{
		groupsdomain.PermissionForumsCreateThread,
		groupsdomain.PermissionForumsManageForum,
	}
}

// replyActions returns actions that grant replies.
func replyActions() []groupsdomain.Action {
	return []groupsdomain.Action{
		groupsdomain.PermissionForumsReply,
		groupsdomain.PermissionForumsManageForum,
	}
}

// likeActions returns actions that grant post likes.
func likeActions() []groupsdomain.Action {
	return []groupsdomain.Action{
		groupsdomain.PermissionForumsLikePosts,
		groupsdomain.PermissionForumsManageForum,
	}
}

// moderateActions returns actions that grant moderation.
func moderateActions() []groupsdomain.Action {
	return []groupsdomain.Action{
		groupsdomain.PermissionForumsManageThreads,
		groupsdomain.PermissionForumsManageForum,
	}
}
