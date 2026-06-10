package postgres

import (
	"context"
	"slices"
	"time"

	"github.com/google/uuid"
	forumsdomain "github.com/niflaot/gamehub-go/module/forums/domain"
	"github.com/niflaot/gamehub-go/module/forums/port"
	groupsdomain "github.com/niflaot/gamehub-go/module/groups/domain"
	"github.com/niflaot/gamehub-go/pkg/orm"
)

// managedForumRelations are relations replaced by forum permission settings.
var managedForumRelations = []groupsdomain.Relation{groupsdomain.RelationViewer, groupsdomain.RelationCreator, groupsdomain.RelationReplyer, groupsdomain.RelationLiker, groupsdomain.RelationModerator, groupsdomain.RelationManager}

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

// CanLikePosts reports whether actor can like posts in forum.
func (authorizer VisibilityAuthorizer) CanLikePosts(ctx context.Context, actorUserID uuid.UUID, forumID uuid.UUID) (bool, error) {
	allowed, err := authorizer.allowedForums(ctx, actorUserID, []uuid.UUID{forumID}, []groupsdomain.Relation{groupsdomain.RelationLiker, groupsdomain.RelationManager, groupsdomain.RelationOwner})
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

// ForumPermissionSettings returns permission grants for a forum.
func (authorizer VisibilityAuthorizer) ForumPermissionSettings(ctx context.Context, forumID uuid.UUID) (forumsdomain.ForumPermissionSettings, error) {
	settings := forumsdomain.ForumPermissionSettings{ForumID: forumID, Viewers: []forumsdomain.ForumPermissionGrant{}, Creators: []forumsdomain.ForumPermissionGrant{}, Replyers: []forumsdomain.ForumPermissionGrant{}, Likers: []forumsdomain.ForumPermissionGrant{}, Moderators: []forumsdomain.ForumPermissionGrant{}, Managers: []forumsdomain.ForumPermissionGrant{}}
	var tuples []relationTupleRow
	err := authorizer.store.DB(ctx).Table("authorization_relation_tuples").Select("object_id, relation, subject_type, subject_id, subject_relation").Where("object_type = ? AND object_id = ? AND relation IN ? AND deleted_at IS NULL", groupsdomain.ObjectForum, forumID, managedForumRelations).Order("relation asc, subject_type asc, subject_id asc").Find(&tuples).Error
	if err != nil {
		return forumsdomain.ForumPermissionSettings{}, err
	}
	for _, tuple := range tuples {
		grant := grantFromTuple(tuple)
		switch groupsdomain.Relation(tuple.Relation) {
		case groupsdomain.RelationViewer:
			settings.Viewers = append(settings.Viewers, grant)
		case groupsdomain.RelationCreator:
			settings.Creators = append(settings.Creators, grant)
		case groupsdomain.RelationReplyer:
			settings.Replyers = append(settings.Replyers, grant)
		case groupsdomain.RelationLiker:
			settings.Likers = append(settings.Likers, grant)
		case groupsdomain.RelationModerator:
			settings.Moderators = append(settings.Moderators, grant)
		case groupsdomain.RelationManager:
			settings.Managers = append(settings.Managers, grant)
		}
	}
	return settings, nil
}

// UpdateForumPermissionSettings replaces permission grants for a forum.
func (authorizer VisibilityAuthorizer) UpdateForumPermissionSettings(ctx context.Context, actorUserID uuid.UUID, settings forumsdomain.ForumPermissionSettings) error {
	settings = settings.Normalize()
	if err := settings.Validate(); err != nil {
		return err
	}
	if err := authorizer.validateGrantSubjects(ctx, settings); err != nil {
		return err
	}
	now := time.Now().UTC()
	if err := authorizer.store.DB(ctx).Table("authorization_relation_tuples").Where("object_type = ? AND object_id = ? AND relation IN ? AND deleted_at IS NULL", groupsdomain.ObjectForum, settings.ForumID, managedForumRelations).Update("deleted_at", now).Error; err != nil {
		return err
	}
	for _, tuple := range tuplesFromPermissionSettings(settings, actorUserID, now) {
		if err := authorizer.store.DB(ctx).Table("authorization_relation_tuples").Create(&tuple).Error; err != nil {
			return err
		}
	}
	return nil
}

// SimulateForumPermission explains a forum permission decision.
func (authorizer VisibilityAuthorizer) SimulateForumPermission(ctx context.Context, forumID uuid.UUID, request forumsdomain.ForumPermissionSimulationRequest) (forumsdomain.ForumPermissionSimulationResult, error) {
	request = request.Normalize(forumID)
	if err := request.Validate(); err != nil {
		return forumsdomain.ForumPermissionSimulationResult{}, err
	}
	relations, err := simulationRelations(request.Permission)
	if err != nil {
		return forumsdomain.ForumPermissionSimulationResult{}, err
	}
	checked := relationNames(relations)
	result := forumsdomain.ForumPermissionSimulationResult{Allowed: false, Reason: "no_matching_relation", Permission: request.Permission, ObjectType: request.ObjectType, ObjectID: request.ObjectID, CheckedRelations: checked}
	var tuples []relationTupleRow
	err = authorizer.store.DB(ctx).Table("authorization_relation_tuples").Select("object_id, relation, subject_type, subject_id, subject_relation").Where("object_type = ? AND object_id = ? AND relation IN ? AND deleted_at IS NULL", groupsdomain.ObjectForum, forumID, relations).Find(&tuples).Error
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

// relationTupleInsertRow is a write model for relation tuples.
type relationTupleInsertRow struct {
	ID              uuid.UUID  `gorm:"column:id"`
	ObjectType      string     `gorm:"column:object_type"`
	ObjectID        uuid.UUID  `gorm:"column:object_id"`
	Relation        string     `gorm:"column:relation"`
	SubjectType     string     `gorm:"column:subject_type"`
	SubjectID       uuid.UUID  `gorm:"column:subject_id"`
	SubjectRelation string     `gorm:"column:subject_relation"`
	CreatedByUserID *uuid.UUID `gorm:"column:created_by_user_id"`
	CreatedAt       time.Time  `gorm:"column:created_at"`
}

// grantFromTuple maps a tuple projection to a grant.
func grantFromTuple(tuple relationTupleRow) forumsdomain.ForumPermissionGrant {
	return forumsdomain.ForumPermissionGrant{SubjectType: forumsdomain.PermissionSubjectType(tuple.SubjectType), SubjectID: tuple.SubjectID, SubjectRelation: tuple.SubjectRelation}
}

// tuplesFromPermissionSettings maps settings to insert rows.
func tuplesFromPermissionSettings(settings forumsdomain.ForumPermissionSettings, actorUserID uuid.UUID, now time.Time) []relationTupleInsertRow {
	var actor *uuid.UUID
	if actorUserID != uuid.Nil {
		actor = &actorUserID
	}
	tuples := []relationTupleInsertRow{}
	tuples = append(tuples, tuplesFromGrants(settings.ForumID, groupsdomain.RelationViewer, settings.Viewers, actor, now)...)
	tuples = append(tuples, tuplesFromGrants(settings.ForumID, groupsdomain.RelationCreator, settings.Creators, actor, now)...)
	tuples = append(tuples, tuplesFromGrants(settings.ForumID, groupsdomain.RelationReplyer, settings.Replyers, actor, now)...)
	tuples = append(tuples, tuplesFromGrants(settings.ForumID, groupsdomain.RelationLiker, settings.Likers, actor, now)...)
	tuples = append(tuples, tuplesFromGrants(settings.ForumID, groupsdomain.RelationModerator, settings.Moderators, actor, now)...)
	tuples = append(tuples, tuplesFromGrants(settings.ForumID, groupsdomain.RelationManager, settings.Managers, actor, now)...)
	return tuples
}

// tuplesFromGrants maps one relation and grant list to insert rows.
func tuplesFromGrants(forumID uuid.UUID, relation groupsdomain.Relation, grants []forumsdomain.ForumPermissionGrant, actor *uuid.UUID, now time.Time) []relationTupleInsertRow {
	rows := make([]relationTupleInsertRow, 0, len(grants))
	for _, grant := range grants {
		grant = grant.Normalize()
		rows = append(rows, relationTupleInsertRow{ID: uuid.New(), ObjectType: string(groupsdomain.ObjectForum), ObjectID: forumID, Relation: string(relation), SubjectType: string(grant.SubjectType), SubjectID: grant.SubjectID, SubjectRelation: grant.SubjectRelation, CreatedByUserID: actor, CreatedAt: now})
	}
	return rows
}

// validateGrantSubjects verifies referenced users and groups exist.
func (authorizer VisibilityAuthorizer) validateGrantSubjects(ctx context.Context, settings forumsdomain.ForumPermissionSettings) error {
	for _, grant := range allPermissionGrants(settings) {
		switch grant.SubjectType {
		case forumsdomain.PermissionSubjectUser:
			if ok, err := authorizer.userExists(ctx, grant.SubjectID); err != nil || !ok {
				if err != nil {
					return err
				}
				return forumsdomain.NewValidationError([]forumsdomain.Violation{{Field: "subject_id", Message: "user does not exist"}})
			}
		case forumsdomain.PermissionSubjectGroup:
			if ok, err := authorizer.groupExists(ctx, grant.SubjectID); err != nil || !ok {
				if err != nil {
					return err
				}
				return forumsdomain.NewValidationError([]forumsdomain.Violation{{Field: "subject_id", Message: "group does not exist"}})
			}
		}
	}
	return nil
}

// allPermissionGrants flattens settings grants.
func allPermissionGrants(settings forumsdomain.ForumPermissionSettings) []forumsdomain.ForumPermissionGrant {
	grants := []forumsdomain.ForumPermissionGrant{}
	grants = append(grants, settings.Viewers...)
	grants = append(grants, settings.Creators...)
	grants = append(grants, settings.Replyers...)
	grants = append(grants, settings.Likers...)
	grants = append(grants, settings.Moderators...)
	grants = append(grants, settings.Managers...)
	return grants
}

// userExists reports whether a user exists.
func (authorizer VisibilityAuthorizer) userExists(ctx context.Context, userID uuid.UUID) (bool, error) {
	var count int64
	err := authorizer.store.DB(ctx).Table("users").Where("id = ? AND deleted_at IS NULL", userID).Count(&count).Error
	return count > 0, err
}

// groupExists reports whether an active/system group exists.
func (authorizer VisibilityAuthorizer) groupExists(ctx context.Context, groupID uuid.UUID) (bool, error) {
	var count int64
	err := authorizer.store.DB(ctx).Table("groups").Where("id = ? AND deleted_at IS NULL AND status IN ?", groupID, []groupsdomain.GroupStatus{groupsdomain.GroupStatusActive, groupsdomain.GroupStatusSystem}).Count(&count).Error
	return count > 0, err
}

// simulationRelations returns forum-level relations checked for permission.
func simulationRelations(permission string) ([]groupsdomain.Relation, error) {
	switch groupsdomain.Permission(permission) {
	case groupsdomain.PermissionForumsView, groupsdomain.PermissionThreadsView, groupsdomain.PermissionPostsView:
		return []groupsdomain.Relation{groupsdomain.RelationViewer, groupsdomain.RelationManager, groupsdomain.RelationOwner}, nil
	case groupsdomain.PermissionForumsManageForum:
		return []groupsdomain.Relation{groupsdomain.RelationManager, groupsdomain.RelationOwner}, nil
	case groupsdomain.PermissionForumsCreateThread:
		return []groupsdomain.Relation{groupsdomain.RelationCreator, groupsdomain.RelationManager, groupsdomain.RelationOwner}, nil
	case groupsdomain.PermissionForumsReply:
		return []groupsdomain.Relation{groupsdomain.RelationReplyer, groupsdomain.RelationManager, groupsdomain.RelationOwner}, nil
	case groupsdomain.PermissionForumsLikePosts, groupsdomain.PermissionPostsLike:
		return []groupsdomain.Relation{groupsdomain.RelationLiker, groupsdomain.RelationManager, groupsdomain.RelationOwner}, nil
	case groupsdomain.PermissionForumsPinThreads, groupsdomain.PermissionForumsManageThreads, groupsdomain.PermissionForumsManagePosts, groupsdomain.PermissionThreadsUpdate, groupsdomain.PermissionThreadsClose, groupsdomain.PermissionThreadsOpen, groupsdomain.PermissionThreadsDelete, groupsdomain.PermissionThreadsPin, groupsdomain.PermissionPostsUpdate, groupsdomain.PermissionPostsDelete, groupsdomain.PermissionPostsViewHidden, groupsdomain.PermissionPostsViewRevisions:
		return []groupsdomain.Relation{groupsdomain.RelationModerator, groupsdomain.RelationManager, groupsdomain.RelationOwner}, nil
	default:
		return nil, forumsdomain.NewValidationError([]forumsdomain.Violation{{Field: "permission", Message: "is not supported for forum simulation"}})
	}
}

// relationNames maps relations to strings.
func relationNames(relations []groupsdomain.Relation) []string {
	names := make([]string, 0, len(relations))
	for _, relation := range relations {
		names = append(names, string(relation))
	}
	return names
}

// Ensure VisibilityAuthorizer implements port.VisibilityAuthorizer.
var _ port.VisibilityAuthorizer = VisibilityAuthorizer{}

// Ensure VisibilityAuthorizer implements port.PermissionAdmin.
var _ port.PermissionAdmin = VisibilityAuthorizer{}
