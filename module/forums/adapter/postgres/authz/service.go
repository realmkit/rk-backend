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

// managedForumRelations are relations replaced by forum permission settings.
var managedForumRelations = []groupsdomain.Relation{
	groupsdomain.RelationViewer,
	groupsdomain.RelationCreator,
	groupsdomain.RelationReplyer,
	groupsdomain.RelationLiker,
	groupsdomain.RelationModerator,
	groupsdomain.RelationManager,
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
	return authorizer.allowedForums(ctx, actorUserID, forumIDs, viewRelations())
}

// CanManageForum reports whether actor can manage target forum.
func (authorizer VisibilityAuthorizer) CanManageForum(
	ctx context.Context,
	actorUserID uuid.UUID,
	forumID uuid.UUID,
) (bool, error) {
	return authorizer.allowed(ctx, actorUserID, forumID, manageRelations())
}

// CanCreateThread reports whether actor can create a thread in forum.
func (authorizer VisibilityAuthorizer) CanCreateThread(
	ctx context.Context,
	actorUserID uuid.UUID,
	forumID uuid.UUID,
) (bool, error) {
	return authorizer.allowed(ctx, actorUserID, forumID, creatorRelations())
}

// CanReply reports whether actor can reply in forum.
func (authorizer VisibilityAuthorizer) CanReply(
	ctx context.Context,
	actorUserID uuid.UUID,
	forumID uuid.UUID,
) (bool, error) {
	return authorizer.allowed(ctx, actorUserID, forumID, replyRelations())
}

// CanLikePosts reports whether actor can like posts in forum.
func (authorizer VisibilityAuthorizer) CanLikePosts(
	ctx context.Context,
	actorUserID uuid.UUID,
	forumID uuid.UUID,
) (bool, error) {
	return authorizer.allowed(ctx, actorUserID, forumID, likeRelations())
}

// CanManageThreads reports whether actor can manage threads in forum.
func (authorizer VisibilityAuthorizer) CanManageThreads(
	ctx context.Context,
	actorUserID uuid.UUID,
	forumID uuid.UUID,
) (bool, error) {
	return authorizer.allowed(ctx, actorUserID, forumID, moderateRelations())
}

// CanManagePosts reports whether actor can manage posts in forum.
func (authorizer VisibilityAuthorizer) CanManagePosts(
	ctx context.Context,
	actorUserID uuid.UUID,
	forumID uuid.UUID,
) (bool, error) {
	return authorizer.allowed(ctx, actorUserID, forumID, moderateRelations())
}

// allowed reports whether actor matches any relation for one forum.
func (authorizer VisibilityAuthorizer) allowed(
	ctx context.Context,
	actorUserID uuid.UUID,
	forumID uuid.UUID,
	relations []groupsdomain.Relation,
) (bool, error) {
	allowed, err := authorizer.allowedForums(ctx, actorUserID, []uuid.UUID{forumID}, relations)
	return allowed[forumID], err
}

// allowedForums returns forum ids allowed by matching relations.
func (authorizer VisibilityAuthorizer) allowedForums(
	ctx context.Context,
	actorUserID uuid.UUID,
	forumIDs []uuid.UUID,
	relations []groupsdomain.Relation,
) (map[uuid.UUID]bool, error) {
	allowed := map[uuid.UUID]bool{}
	if len(forumIDs) == 0 {
		return allowed, nil
	}
	var tuples []relationTupleRow
	err := authorizer.store.DB(ctx).
		Table("authorization_relation_tuples").
		Select("object_id, relation, subject_type, subject_id, subject_relation").
		Where(
			"object_type = ? AND object_id IN ? AND relation IN ? AND deleted_at IS NULL",
			groupsdomain.ObjectForum,
			forumIDs,
			relations,
		).
		Find(&tuples).Error
	if err != nil {
		return nil, err
	}
	memberships, err := authorizer.activeMemberships(ctx, actorUserID, groupSubjectIDs(tuples))
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

// relationTupleRow is a compact authorization tuple projection.
type relationTupleRow struct {
	ObjectID        uuid.UUID
	Relation        string
	SubjectType     string
	SubjectID       uuid.UUID
	SubjectRelation string
}

// Ensure VisibilityAuthorizer implements port.VisibilityAuthorizer.
var _ port.VisibilityAuthorizer = VisibilityAuthorizer{}

// Ensure VisibilityAuthorizer implements port.PermissionAdmin.
var _ port.PermissionAdmin = VisibilityAuthorizer{}

// keepForumDomainImport preserves godoc links for this package in generated docs.
var keepForumDomainImport forumsdomain.ForumPermissionSettings
