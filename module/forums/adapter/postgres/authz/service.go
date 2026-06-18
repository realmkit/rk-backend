// Package authz adapts forum authorization reads and writes to PostgreSQL.
package authz

import (
	"context"

	"github.com/google/uuid"
	forumsdomain "github.com/realmkit/rk-backend/module/forums/domain"
	"github.com/realmkit/rk-backend/module/forums/port"
	groupsdomain "github.com/realmkit/rk-backend/module/groups/domain"
	"github.com/realmkit/rk-backend/pkg/orm"
)

// managedForumActions are actions replaced by forum permission settings.
var managedForumActions = []groupsdomain.Action{
	groupsdomain.PermissionForumsView,
	groupsdomain.PermissionForumsCreateThread,
	groupsdomain.PermissionForumsReply,
	groupsdomain.PermissionForumsLikePosts,
	groupsdomain.PermissionForumsViewAllThreads,
	groupsdomain.PermissionForumsBypassThreadLimits,
	groupsdomain.PermissionForumsPinThreads,
	groupsdomain.PermissionForumsManageThreads,
	groupsdomain.PermissionForumsManagePosts,
	groupsdomain.PermissionForumsAdministrativeAccess,
}

// VisibilityAuthorizer resolves forum permissions from authorization tuples.
type VisibilityAuthorizer struct {
	store orm.Store
}

// NewVisibilityAuthorizer creates a visibility authorizer.
func NewVisibilityAuthorizer(store orm.Store) VisibilityAuthorizer {
	return VisibilityAuthorizer{store: store}
}

// VisibleForums returns visible forum IDs for actor.
func (authorizer VisibilityAuthorizer) VisibleForums(
	ctx context.Context,
	actorUserID uuid.UUID,
	forumIDs []uuid.UUID,
) (map[uuid.UUID]bool, error) {
	return authorizer.allowedForums(ctx, actorUserID, forumIDs, viewActions())
}

// CanManageForum reports whether actor can manage target forum.
func (authorizer VisibilityAuthorizer) CanManageForum(
	ctx context.Context,
	actorUserID uuid.UUID,
	forumID uuid.UUID,
) (bool, error) {
	return authorizer.allowed(ctx, actorUserID, forumID, manageActions())
}

// CanCreateThread reports whether actor can create a thread in forum.
func (authorizer VisibilityAuthorizer) CanCreateThread(
	ctx context.Context,
	actorUserID uuid.UUID,
	forumID uuid.UUID,
) (bool, error) {
	if actorUserID == uuid.Nil {
		return false, nil
	}
	return authorizer.allowed(ctx, actorUserID, forumID, createThreadActions())
}

// CanReply reports whether actor can reply in forum.
func (authorizer VisibilityAuthorizer) CanReply(
	ctx context.Context,
	actorUserID uuid.UUID,
	forumID uuid.UUID,
) (bool, error) {
	if actorUserID == uuid.Nil {
		return false, nil
	}
	return authorizer.allowed(ctx, actorUserID, forumID, replyActions())
}

// CanLikePosts reports whether actor can like posts in forum.
func (authorizer VisibilityAuthorizer) CanLikePosts(
	ctx context.Context,
	actorUserID uuid.UUID,
	forumID uuid.UUID,
) (bool, error) {
	if actorUserID == uuid.Nil {
		return false, nil
	}
	return authorizer.allowed(ctx, actorUserID, forumID, likeActions())
}

// CanManageThreads reports whether actor can manage threads in forum.
func (authorizer VisibilityAuthorizer) CanManageThreads(
	ctx context.Context,
	actorUserID uuid.UUID,
	forumID uuid.UUID,
) (bool, error) {
	if actorUserID == uuid.Nil {
		return false, nil
	}
	return authorizer.allowed(ctx, actorUserID, forumID, threadManageActions())
}

// CanManagePosts reports whether actor can manage posts in forum.
func (authorizer VisibilityAuthorizer) CanManagePosts(
	ctx context.Context,
	actorUserID uuid.UUID,
	forumID uuid.UUID,
) (bool, error) {
	if actorUserID == uuid.Nil {
		return false, nil
	}
	return authorizer.allowed(ctx, actorUserID, forumID, postManageActions())
}

// allowed reports whether actor matches any relation for one forum.
func (authorizer VisibilityAuthorizer) allowed(
	ctx context.Context,
	actorUserID uuid.UUID,
	forumID uuid.UUID,
	actions []groupsdomain.Action,
) (bool, error) {
	allowed, err := authorizer.allowedForums(ctx, actorUserID, []uuid.UUID{forumID}, actions)
	return allowed[forumID], err
}

// allowedForums returns forum ids allowed by matching actions.
func (authorizer VisibilityAuthorizer) allowedForums(
	ctx context.Context,
	actorUserID uuid.UUID,
	forumIDs []uuid.UUID,
	actions []groupsdomain.Action,
) (map[uuid.UUID]bool, error) {
	allowed := map[uuid.UUID]bool{}
	if len(forumIDs) == 0 {
		return allowed, nil
	}
	var grants []permissionGrantRow
	err := authorizer.store.DB(ctx).
		Table("forum_permission_grants").
		Select("scope_id, action, subject_type, subject_id").
		Where(
			"scope_type = ? AND scope_id IN ? AND action IN ? AND deleted_at IS NULL",
			groupsdomain.ObjectForum,
			forumIDs,
			actions,
		).
		Find(&grants).Error
	if err != nil {
		return nil, err
	}
	groupGrants, err := authorizer.groupPermissionGrants(ctx, forumIDs, actions)
	if err != nil {
		return nil, err
	}
	memberships, err := authorizer.activeMemberships(
		ctx,
		actorUserID,
		grantGroupIDs(grants, groupGrants),
	)
	if err != nil {
		return nil, err
	}
	for _, grant := range grants {
		if grantMatchesActor(grant, actorUserID, memberships) {
			allowed[grant.ScopeID] = true
		}
	}
	for _, grant := range groupGrants {
		if !groupGrantMatchesActor(grant, memberships) {
			continue
		}
		if grant.ScopeID == groupsdomain.AllScopeID() {
			for _, forumID := range forumIDs {
				allowed[forumID] = true
			}
			continue
		}
		allowed[grant.ScopeID] = true
	}
	return allowed, nil
}

// groupPermissionGrants returns group-owned forum grants.
func (authorizer VisibilityAuthorizer) groupPermissionGrants(
	ctx context.Context,
	forumIDs []uuid.UUID,
	actions []groupsdomain.Action,
) ([]groupPermissionGrantRow, error) {
	var grants []groupPermissionGrantRow
	err := authorizer.store.DB(ctx).
		Table("permission_grants").
		Select("permission_grants.scope_id, permission_grants.action, group_permission_grants.group_id").
		Joins("JOIN group_permission_grants ON group_permission_grants.grant_id = permission_grants.id").
		Where(
			"permission_grants.scope_type = ? AND (permission_grants.scope_id IN ? OR permission_grants.scope_id = ?) "+
				"AND permission_grants.action IN ? AND permission_grants.deleted_at IS NULL "+
				"AND group_permission_grants.deleted_at IS NULL",
			groupsdomain.ObjectForum,
			forumIDs,
			groupsdomain.AllScopeID(),
			actions,
		).
		Find(&grants).Error
	return grants, err
}

// permissionGrantRow is a compact permission grant projection.
type permissionGrantRow struct {
	ScopeID     uuid.UUID
	Action      string
	SubjectType string
	SubjectID   uuid.UUID
}

// groupPermissionGrantRow is a compact group permission grant projection.
type groupPermissionGrantRow struct {
	ScopeID uuid.UUID
	Action  string
	GroupID uuid.UUID
}

// Ensure VisibilityAuthorizer implements port.VisibilityAuthorizer.
var _ port.VisibilityAuthorizer = VisibilityAuthorizer{}

// Ensure VisibilityAuthorizer implements port.PermissionAdmin.
var _ port.PermissionAdmin = VisibilityAuthorizer{}

// keepForumDomainImport preserves godoc links for this package in generated docs.
var keepForumDomainImport forumsdomain.ForumPermissionSettings
