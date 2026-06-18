package authz

import (
	"time"

	"github.com/google/uuid"
	forumsdomain "github.com/realmkit/rk-backend/module/forums/domain"
	groupsdomain "github.com/realmkit/rk-backend/module/groups/domain"
)

// addGrantToSettings appends grant to the action bucket it belongs to.
func addGrantToSettings(
	settings *forumsdomain.ForumPermissionSettings,
	action groupsdomain.Action,
	grant forumsdomain.ForumPermissionGrant,
) {
	switch action {
	case groupsdomain.PermissionForumsView:
		settings.Viewers = append(settings.Viewers, grant)
	case groupsdomain.PermissionForumsCreateThread:
		settings.Creators = append(settings.Creators, grant)
	case groupsdomain.PermissionForumsReply:
		settings.Replyers = append(settings.Replyers, grant)
	case groupsdomain.PermissionForumsLikePosts:
		settings.Likers = append(settings.Likers, grant)
	case groupsdomain.PermissionForumsViewAllThreads:
		settings.AllThreadViewers = append(settings.AllThreadViewers, grant)
	case groupsdomain.PermissionForumsBypassThreadLimits:
		settings.LimitBypassers = append(settings.LimitBypassers, grant)
	case groupsdomain.PermissionForumsPinThreads:
		settings.ThreadPinners = append(settings.ThreadPinners, grant)
	case groupsdomain.PermissionForumsManageThreads:
		settings.ThreadManagers = append(settings.ThreadManagers, grant)
	case groupsdomain.PermissionForumsManagePosts:
		settings.PostManagers = append(settings.PostManagers, grant)
	case groupsdomain.PermissionForumsAdministrativeAccess:
		settings.Administrators = append(settings.Administrators, grant)
	}
}

// grantFromRow maps a grant projection to a forum grant.
func grantFromRow(row permissionGrantRow) forumsdomain.ForumPermissionGrant {
	return forumsdomain.ForumPermissionGrant{
		SubjectType: forumsdomain.PermissionSubjectType(row.SubjectType),
		SubjectID:   row.SubjectID,
	}
}

// rowsFromPermissionSettings maps settings to insert rows.
func rowsFromPermissionSettings(
	settings forumsdomain.ForumPermissionSettings,
	actorUserID uuid.UUID,
	now time.Time,
) []permissionGrantInsertRow {
	var actor *uuid.UUID
	if actorUserID != uuid.Nil {
		actor = &actorUserID
	}
	rows := []permissionGrantInsertRow{}
	rows = append(rows, rowsFromGrants(settings.ForumID, groupsdomain.PermissionForumsView, settings.Viewers, actor, now)...)
	rows = append(rows, rowsFromGrants(settings.ForumID, groupsdomain.PermissionForumsCreateThread, settings.Creators, actor, now)...)
	rows = append(rows, rowsFromGrants(settings.ForumID, groupsdomain.PermissionForumsReply, settings.Replyers, actor, now)...)
	rows = append(rows, rowsFromGrants(settings.ForumID, groupsdomain.PermissionForumsLikePosts, settings.Likers, actor, now)...)
	rows = append(rows, rowsFromGrants(settings.ForumID, groupsdomain.PermissionForumsViewAllThreads, settings.AllThreadViewers, actor, now)...)
	rows = append(rows, rowsFromGrants(settings.ForumID, groupsdomain.PermissionForumsBypassThreadLimits, settings.LimitBypassers, actor, now)...)
	rows = append(rows, rowsFromGrants(settings.ForumID, groupsdomain.PermissionForumsPinThreads, settings.ThreadPinners, actor, now)...)
	rows = append(rows, rowsFromGrants(settings.ForumID, groupsdomain.PermissionForumsManageThreads, settings.ThreadManagers, actor, now)...)
	rows = append(rows, rowsFromGrants(settings.ForumID, groupsdomain.PermissionForumsManagePosts, settings.PostManagers, actor, now)...)
	rows = append(rows, rowsFromGrants(settings.ForumID, groupsdomain.PermissionForumsAdministrativeAccess, settings.Administrators, actor, now)...)
	return rows
}

// rowsFromGrants maps one action and grant list to insert rows.
func rowsFromGrants(
	forumID uuid.UUID,
	action groupsdomain.Action,
	grants []forumsdomain.ForumPermissionGrant,
	actor *uuid.UUID,
	now time.Time,
) []permissionGrantInsertRow {
	rows := make([]permissionGrantInsertRow, 0, len(grants))
	for _, grant := range grants {
		grant = grant.Normalize()
		rows = append(rows, permissionGrantInsertRow{
			ID:              uuid.New(),
			SubjectType:     string(grant.SubjectType),
			SubjectID:       grant.SubjectID,
			Action:          string(action),
			ScopeType:       string(groupsdomain.ObjectForum),
			ScopeID:         forumID,
			ConditionKey:    "",
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
	grants = append(grants, settings.AllThreadViewers...)
	grants = append(grants, settings.LimitBypassers...)
	grants = append(grants, settings.ThreadPinners...)
	grants = append(grants, settings.ThreadManagers...)
	grants = append(grants, settings.PostManagers...)
	grants = append(grants, settings.Administrators...)
	return grants
}
