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
func groupSubjectIDs(tuples []relationTupleRow) []uuid.UUID {
	groupIDs := []uuid.UUID{}
	for _, tuple := range tuples {
		if tuple.SubjectType != string(groupsdomain.SubjectGroup) {
			continue
		}
		if !slices.Contains(groupIDs, tuple.SubjectID) {
			groupIDs = append(groupIDs, tuple.SubjectID)
		}
	}
	return groupIDs
}

// tupleMatchesActor reports whether tuple grants to actor.
func tupleMatchesActor(
	tuple relationTupleRow,
	actorUserID uuid.UUID,
	memberships map[uuid.UUID]bool,
) bool {
	switch groupsdomain.SubjectType(tuple.SubjectType) {
	case groupsdomain.SubjectPublic:
		return tuple.SubjectID == groupsdomain.PublicSubjectID()
	case groupsdomain.SubjectAuthenticated:
		return actorUserID != uuid.Nil && tuple.SubjectID == groupsdomain.AuthenticatedSubjectID()
	case groupsdomain.SubjectUser:
		return actorUserID != uuid.Nil && tuple.SubjectID == actorUserID
	case groupsdomain.SubjectGroup:
		return tuple.SubjectRelation == string(groupsdomain.RelationMember) &&
			memberships[tuple.SubjectID]
	default:
		return false
	}
}

// groupMembershipRow is a compact group membership projection.
type groupMembershipRow struct {
	GroupID uuid.UUID
}

// viewRelations returns relations that grant viewing.
func viewRelations() []groupsdomain.Relation {
	return []groupsdomain.Relation{
		groupsdomain.RelationViewer,
		groupsdomain.RelationManager,
		groupsdomain.RelationOwner,
	}
}

// manageRelations returns relations that grant structure management.
func manageRelations() []groupsdomain.Relation {
	return []groupsdomain.Relation{groupsdomain.RelationManager, groupsdomain.RelationOwner}
}

// creatorRelations returns relations that grant thread creation.
func creatorRelations() []groupsdomain.Relation {
	return []groupsdomain.Relation{
		groupsdomain.RelationCreator,
		groupsdomain.RelationManager,
		groupsdomain.RelationOwner,
	}
}

// replyRelations returns relations that grant replies.
func replyRelations() []groupsdomain.Relation {
	return []groupsdomain.Relation{
		groupsdomain.RelationReplyer,
		groupsdomain.RelationManager,
		groupsdomain.RelationOwner,
	}
}

// likeRelations returns relations that grant post likes.
func likeRelations() []groupsdomain.Relation {
	return []groupsdomain.Relation{
		groupsdomain.RelationLiker,
		groupsdomain.RelationManager,
		groupsdomain.RelationOwner,
	}
}

// moderateRelations returns relations that grant moderation.
func moderateRelations() []groupsdomain.Relation {
	return []groupsdomain.Relation{
		groupsdomain.RelationModerator,
		groupsdomain.RelationManager,
		groupsdomain.RelationOwner,
	}
}
