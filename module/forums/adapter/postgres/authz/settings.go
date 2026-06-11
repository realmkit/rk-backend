package authz

import (
	"context"
	"time"

	"github.com/google/uuid"
	forumsdomain "github.com/realmkit/rk-backend/module/forums/domain"
	groupsdomain "github.com/realmkit/rk-backend/module/groups/domain"
)

// ForumPermissionSettings returns permission grants for a forum.
func (authorizer VisibilityAuthorizer) ForumPermissionSettings(
	ctx context.Context,
	forumID uuid.UUID,
) (forumsdomain.ForumPermissionSettings, error) {
	settings := emptyPermissionSettings(forumID)
	var tuples []relationTupleRow
	err := authorizer.store.DB(ctx).
		Table("authorization_relation_tuples").
		Select("object_id, relation, subject_type, subject_id, subject_relation").
		Where(
			"object_type = ? AND object_id = ? AND relation IN ? AND deleted_at IS NULL",
			groupsdomain.ObjectForum,
			forumID,
			managedForumRelations,
		).
		Order("relation asc, subject_type asc, subject_id asc").
		Find(&tuples).Error
	if err != nil {
		return forumsdomain.ForumPermissionSettings{}, err
	}
	for _, tuple := range tuples {
		addGrantToSettings(&settings, groupsdomain.Relation(tuple.Relation), grantFromTuple(tuple))
	}
	return settings, nil
}

// UpdateForumPermissionSettings replaces permission grants for a forum.
func (authorizer VisibilityAuthorizer) UpdateForumPermissionSettings(
	ctx context.Context,
	actorUserID uuid.UUID,
	settings forumsdomain.ForumPermissionSettings,
) error {
	settings = settings.Normalize()
	if err := settings.Validate(); err != nil {
		return err
	}
	if err := authorizer.validateGrantSubjects(ctx, settings); err != nil {
		return err
	}
	now := time.Now().UTC()
	err := authorizer.store.DB(ctx).
		Table("authorization_relation_tuples").
		Where(
			"object_type = ? AND object_id = ? AND relation IN ? AND deleted_at IS NULL",
			groupsdomain.ObjectForum,
			settings.ForumID,
			managedForumRelations,
		).
		Update("deleted_at", now).Error
	if err != nil {
		return err
	}
	for _, tuple := range tuplesFromPermissionSettings(settings, actorUserID, now) {
		if err := authorizer.store.DB(ctx).Table("authorization_relation_tuples").Create(&tuple).Error; err != nil {
			return err
		}
	}
	return nil
}

// validateGrantSubjects verifies referenced users and groups exist.
func (authorizer VisibilityAuthorizer) validateGrantSubjects(
	ctx context.Context,
	settings forumsdomain.ForumPermissionSettings,
) error {
	for _, grant := range allPermissionGrants(settings) {
		switch grant.SubjectType {
		case forumsdomain.PermissionSubjectUser:
			ok, err := authorizer.userExists(ctx, grant.SubjectID)
			if err != nil {
				return err
			}
			if !ok {
				return subjectValidationError("user does not exist")
			}
		case forumsdomain.PermissionSubjectGroup:
			ok, err := authorizer.groupExists(ctx, grant.SubjectID)
			if err != nil {
				return err
			}
			if !ok {
				return subjectValidationError("group does not exist")
			}
		}
	}
	return nil
}

// userExists reports whether a user exists.
func (authorizer VisibilityAuthorizer) userExists(ctx context.Context, userID uuid.UUID) (bool, error) {
	var count int64
	err := authorizer.store.DB(ctx).
		Table("users").
		Where("id = ? AND deleted_at IS NULL", userID).
		Count(&count).Error
	return count > 0, err
}

// groupExists reports whether an active/system group exists.
func (authorizer VisibilityAuthorizer) groupExists(ctx context.Context, groupID uuid.UUID) (bool, error) {
	var count int64
	err := authorizer.store.DB(ctx).
		Table("groups").
		Where(
			"id = ? AND deleted_at IS NULL AND status IN ?",
			groupID,
			[]groupsdomain.GroupStatus{
				groupsdomain.GroupStatusActive,
				groupsdomain.GroupStatusSystem,
			},
		).
		Count(&count).Error
	return count > 0, err
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

// emptyPermissionSettings returns an initialized settings value.
func emptyPermissionSettings(forumID uuid.UUID) forumsdomain.ForumPermissionSettings {
	return forumsdomain.ForumPermissionSettings{
		ForumID:    forumID,
		Viewers:    []forumsdomain.ForumPermissionGrant{},
		Creators:   []forumsdomain.ForumPermissionGrant{},
		Replyers:   []forumsdomain.ForumPermissionGrant{},
		Likers:     []forumsdomain.ForumPermissionGrant{},
		Moderators: []forumsdomain.ForumPermissionGrant{},
		Managers:   []forumsdomain.ForumPermissionGrant{},
	}
}

// subjectValidationError returns a forum validation error for a bad subject.
func subjectValidationError(message string) error {
	return forumsdomain.NewValidationError([]forumsdomain.Violation{{
		Field:   "subject_id",
		Message: message,
	}})
}
