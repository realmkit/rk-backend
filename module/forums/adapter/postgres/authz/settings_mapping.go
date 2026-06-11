package authz

import (
	"time"

	"github.com/google/uuid"
	forumsdomain "github.com/realmkit/rk-backend/module/forums/domain"
	groupsdomain "github.com/realmkit/rk-backend/module/groups/domain"
)

// addGrantToSettings appends grant to the relation bucket it belongs to.
func addGrantToSettings(
	settings *forumsdomain.ForumPermissionSettings,
	relation groupsdomain.Relation,
	grant forumsdomain.ForumPermissionGrant,
) {
	switch relation {
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

// grantFromTuple maps a tuple projection to a grant.
func grantFromTuple(tuple relationTupleRow) forumsdomain.ForumPermissionGrant {
	return forumsdomain.ForumPermissionGrant{
		SubjectType:     forumsdomain.PermissionSubjectType(tuple.SubjectType),
		SubjectID:       tuple.SubjectID,
		SubjectRelation: tuple.SubjectRelation,
	}
}

// tuplesFromPermissionSettings maps settings to insert rows.
func tuplesFromPermissionSettings(
	settings forumsdomain.ForumPermissionSettings,
	actorUserID uuid.UUID,
	now time.Time,
) []relationTupleInsertRow {
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
func tuplesFromGrants(
	forumID uuid.UUID,
	relation groupsdomain.Relation,
	grants []forumsdomain.ForumPermissionGrant,
	actor *uuid.UUID,
	now time.Time,
) []relationTupleInsertRow {
	rows := make([]relationTupleInsertRow, 0, len(grants))
	for _, grant := range grants {
		grant = grant.Normalize()
		rows = append(rows, relationTupleInsertRow{
			ID:              uuid.New(),
			ObjectType:      string(groupsdomain.ObjectForum),
			ObjectID:        forumID,
			Relation:        string(relation),
			SubjectType:     string(grant.SubjectType),
			SubjectID:       grant.SubjectID,
			SubjectRelation: grant.SubjectRelation,
			CreatedByUserID: actor,
			CreatedAt:       now,
		})
	}
	return rows
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
