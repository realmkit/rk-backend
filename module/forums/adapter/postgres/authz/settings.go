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
	var grants []permissionGrantRow
	err := authorizer.store.DB(ctx).
		Table("forum_permission_grants").
		Select("scope_id, action, subject_type, subject_id").
		Where(
			"scope_type = ? AND scope_id = ? AND action IN ? AND deleted_at IS NULL",
			groupsdomain.ObjectForum,
			forumID,
			managedForumActions,
		).
		Order("action asc, subject_type asc, subject_id asc").
		Find(&grants).Error
	if err != nil {
		return forumsdomain.ForumPermissionSettings{}, err
	}
	for _, grant := range grants {
		addGrantToSettings(&settings, groupsdomain.Action(grant.Action), grantFromRow(grant))
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
		Table("forum_permission_grants").
		Where(
			"scope_type = ? AND scope_id = ? AND action IN ? AND deleted_at IS NULL",
			groupsdomain.ObjectForum,
			settings.ForumID,
			managedForumActions,
		).
		Update("deleted_at", now).Error
	if err != nil {
		return err
	}
	for _, grant := range rowsFromPermissionSettings(settings, actorUserID, now) {
		if err := authorizer.store.DB(ctx).Table("forum_permission_grants").Create(&grant).Error; err != nil {
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

// permissionGrantInsertRow is a write model for permission grants.
type permissionGrantInsertRow struct {
	ID              uuid.UUID  `gorm:"column:id"`                 // ID stores the i d value.
	SubjectType     string     `gorm:"column:subject_type"`       // SubjectType stores the subject type value.
	SubjectID       uuid.UUID  `gorm:"column:subject_id"`         // SubjectID stores the subject i d value.
	Action          string     `gorm:"column:action"`             // Action stores the action value.
	ScopeType       string     `gorm:"column:scope_type"`         // ScopeType stores the scope type value.
	ScopeID         uuid.UUID  `gorm:"column:scope_id"`           // ScopeID stores the scope i d value.
	Inherit         bool       `gorm:"column:inherit"`            // Inherit stores the inherit value.
	ConditionKey    string     `gorm:"column:condition_key"`      // ConditionKey stores the condition key value.
	CreatedByUserID *uuid.UUID `gorm:"column:created_by_user_id"` // CreatedByUserID stores the created by user i d value.
	CreatedAt       time.Time  `gorm:"column:created_at"`         // CreatedAt stores the created at value.
}

// emptyPermissionSettings returns an initialized settings value.
func emptyPermissionSettings(forumID uuid.UUID) forumsdomain.ForumPermissionSettings {
	return forumsdomain.ForumPermissionSettings{
		ForumID:          forumID,
		Viewers:          []forumsdomain.ForumPermissionGrant{},
		Creators:         []forumsdomain.ForumPermissionGrant{},
		Replyers:         []forumsdomain.ForumPermissionGrant{},
		Likers:           []forumsdomain.ForumPermissionGrant{},
		ThreadPinners:    []forumsdomain.ForumPermissionGrant{},
		ThreadManagers:   []forumsdomain.ForumPermissionGrant{},
		PostManagers:     []forumsdomain.ForumPermissionGrant{},
		LimitBypassers:   []forumsdomain.ForumPermissionGrant{},
		AllThreadViewers: []forumsdomain.ForumPermissionGrant{},
		Administrators:   []forumsdomain.ForumPermissionGrant{},
	}
}

// subjectValidationError returns a forum validation error for a bad subject.
func subjectValidationError(message string) error {
	return forumsdomain.NewValidationError([]forumsdomain.Violation{{
		Field:   "subject_id",
		Message: message,
	}})
}
