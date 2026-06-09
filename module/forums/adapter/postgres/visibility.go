package postgres

import (
	"context"
	"slices"

	"github.com/google/uuid"
	"github.com/niflaot/gamehub-go/module/forums/port"
	groupsdomain "github.com/niflaot/gamehub-go/module/groups/domain"
	"github.com/niflaot/gamehub-go/pkg/orm"
)

// VisibilityAuthorizer resolves forum permissions from authorization tuples.
type VisibilityAuthorizer struct {
	store orm.Store
}

// NewVisibilityAuthorizer creates a visibility authorizer.
func NewVisibilityAuthorizer(store orm.Store) VisibilityAuthorizer {
	return VisibilityAuthorizer{store: store}
}

// VisibleForums returns visible forum IDs for actor.
func (authorizer VisibilityAuthorizer) VisibleForums(ctx context.Context, actorUserID uuid.UUID, forumIDs []uuid.UUID) (map[uuid.UUID]bool, error) {
	return authorizer.allowedForums(ctx, actorUserID, forumIDs, []groupsdomain.Relation{groupsdomain.RelationViewer, groupsdomain.RelationManager, groupsdomain.RelationOwner})
}

// CanManageForum reports whether actor can manage target forum.
func (authorizer VisibilityAuthorizer) CanManageForum(ctx context.Context, actorUserID uuid.UUID, forumID uuid.UUID) (bool, error) {
	allowed, err := authorizer.allowedForums(ctx, actorUserID, []uuid.UUID{forumID}, []groupsdomain.Relation{groupsdomain.RelationManager, groupsdomain.RelationOwner})
	return allowed[forumID], err
}

// CanCreateThread reports whether actor can create a thread in forum.
func (authorizer VisibilityAuthorizer) CanCreateThread(ctx context.Context, actorUserID uuid.UUID, forumID uuid.UUID) (bool, error) {
	allowed, err := authorizer.allowedForums(ctx, actorUserID, []uuid.UUID{forumID}, []groupsdomain.Relation{groupsdomain.RelationCreator, groupsdomain.RelationManager, groupsdomain.RelationOwner})
	return allowed[forumID], err
}

// CanReply reports whether actor can reply in forum.
func (authorizer VisibilityAuthorizer) CanReply(ctx context.Context, actorUserID uuid.UUID, forumID uuid.UUID) (bool, error) {
	allowed, err := authorizer.allowedForums(ctx, actorUserID, []uuid.UUID{forumID}, []groupsdomain.Relation{groupsdomain.RelationReplyer, groupsdomain.RelationManager, groupsdomain.RelationOwner})
	return allowed[forumID], err
}

// CanManageThreads reports whether actor can manage threads in forum.
func (authorizer VisibilityAuthorizer) CanManageThreads(ctx context.Context, actorUserID uuid.UUID, forumID uuid.UUID) (bool, error) {
	allowed, err := authorizer.allowedForums(ctx, actorUserID, []uuid.UUID{forumID}, []groupsdomain.Relation{groupsdomain.RelationModerator, groupsdomain.RelationManager, groupsdomain.RelationOwner})
	return allowed[forumID], err
}

// CanManagePosts reports whether actor can manage posts in forum.
func (authorizer VisibilityAuthorizer) CanManagePosts(ctx context.Context, actorUserID uuid.UUID, forumID uuid.UUID) (bool, error) {
	allowed, err := authorizer.allowedForums(ctx, actorUserID, []uuid.UUID{forumID}, []groupsdomain.Relation{groupsdomain.RelationModerator, groupsdomain.RelationManager, groupsdomain.RelationOwner})
	return allowed[forumID], err
}

// allowedForums returns forum ids allowed by matching relations.
func (authorizer VisibilityAuthorizer) allowedForums(ctx context.Context, actorUserID uuid.UUID, forumIDs []uuid.UUID, relations []groupsdomain.Relation) (map[uuid.UUID]bool, error) {
	allowed := map[uuid.UUID]bool{}
	if len(forumIDs) == 0 {
		return allowed, nil
	}
	var tuples []relationTupleRow
	err := authorizer.store.DB(ctx).Table("authorization_relation_tuples").Select("object_id, relation, subject_type, subject_id, subject_relation").Where("object_type = ? AND object_id IN ? AND relation IN ? AND deleted_at IS NULL", groupsdomain.ObjectForum, forumIDs, relations).Find(&tuples).Error
	if err != nil {
		return nil, err
	}
	groupIDs := groupSubjectIDs(tuples)
	memberships, err := authorizer.activeMemberships(ctx, actorUserID, groupIDs)
	if err != nil {
		return nil, err
	}
	for _, tuple := range tuples {
		if tupleMatchesActor(tuple, actorUserID, memberships) {
			allowed[tuple.ObjectID] = true
		}
	}
	return allowed, nil
}

// activeMemberships returns actor group memberships that can grant permissions.
func (authorizer VisibilityAuthorizer) activeMemberships(ctx context.Context, actorUserID uuid.UUID, groupIDs []uuid.UUID) (map[uuid.UUID]bool, error) {
	result := map[uuid.UUID]bool{}
	if actorUserID == uuid.Nil || len(groupIDs) == 0 {
		return result, nil
	}
	var rows []groupMembershipRow
	err := authorizer.store.DB(ctx).Table("group_memberships").Joins("JOIN groups ON groups.id = group_memberships.group_id").Where("group_memberships.user_id = ? AND group_memberships.group_id IN ? AND group_memberships.status = ? AND group_memberships.deleted_at IS NULL AND groups.deleted_at IS NULL AND groups.status IN ?", actorUserID, groupIDs, groupsdomain.MembershipStatusActive, []groupsdomain.GroupStatus{groupsdomain.GroupStatusActive, groupsdomain.GroupStatusSystem}).Find(&rows).Error
	if err != nil {
		return nil, err
	}
	for _, row := range rows {
		result[row.GroupID] = true
	}
	return result, nil
}

// relationTupleRow is a compact authorization tuple projection.
type relationTupleRow struct {
	ObjectID        uuid.UUID
	Relation        string
	SubjectType     string
	SubjectID       uuid.UUID
	SubjectRelation string
}

// groupMembershipRow is a compact group membership projection.
type groupMembershipRow struct {
	GroupID uuid.UUID
}

// groupSubjectIDs extracts group subject IDs.
func groupSubjectIDs(tuples []relationTupleRow) []uuid.UUID {
	groupIDs := []uuid.UUID{}
	for _, tuple := range tuples {
		if tuple.SubjectType == string(groupsdomain.SubjectGroup) && !slices.Contains(groupIDs, tuple.SubjectID) {
			groupIDs = append(groupIDs, tuple.SubjectID)
		}
	}
	return groupIDs
}

// tupleMatchesActor reports whether tuple grants to actor.
func tupleMatchesActor(tuple relationTupleRow, actorUserID uuid.UUID, memberships map[uuid.UUID]bool) bool {
	switch groupsdomain.SubjectType(tuple.SubjectType) {
	case groupsdomain.SubjectPublic:
		return tuple.SubjectID == groupsdomain.PublicSubjectID()
	case groupsdomain.SubjectAuthenticated:
		return actorUserID != uuid.Nil && tuple.SubjectID == groupsdomain.AuthenticatedSubjectID()
	case groupsdomain.SubjectUser:
		return actorUserID != uuid.Nil && tuple.SubjectID == actorUserID
	case groupsdomain.SubjectGroup:
		return tuple.SubjectRelation == string(groupsdomain.RelationMember) && memberships[tuple.SubjectID]
	default:
		return false
	}
}

// Ensure VisibilityAuthorizer implements port.VisibilityAuthorizer.
var _ port.VisibilityAuthorizer = VisibilityAuthorizer{}
