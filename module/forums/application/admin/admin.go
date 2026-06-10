package admin

import (
	"context"

	"github.com/google/uuid"
	"github.com/niflaot/gamehub-go/module/forums/domain"
	"github.com/niflaot/gamehub-go/module/forums/port"
)

// GetForumSettings returns admin forum settings.
func (service Service) GetForumSettings(
	ctx context.Context,
	actorUserID uuid.UUID,
	forumID uuid.UUID,
) (domain.ForumSettings, error) {
	if err := service.requireManage(ctx, actorUserID, forumID); err != nil {
		return domain.ForumSettings{}, err
	}
	forum, err := service.forums.FindByID(ctx, forumID)
	if err != nil {
		return domain.ForumSettings{}, err
	}
	return forum.Settings(), nil
}

// UpdateForumSettings updates admin forum settings.
func (service Service) UpdateForumSettings(
	ctx context.Context,
	command port.UpdateForumSettingsCommand,
) (domain.ForumSettings, error) {
	if err := service.requireManage(ctx, command.ActorUserID, command.Settings.ForumID); err != nil {
		return domain.ForumSettings{}, err
	}
	current, err := service.forums.FindByID(ctx, command.Settings.ForumID)
	if err != nil {
		return domain.ForumSettings{}, err
	}
	settings := command.Settings.Normalize()
	if err := settings.Validate(); err != nil {
		return domain.ForumSettings{}, err
	}
	updatedForum := forumWithSettings(current, settings)
	if err := updatedForum.Validate(); err != nil {
		return domain.ForumSettings{}, err
	}
	updated, err := service.forums.Update(ctx, updatedForum, command.ExpectedVersion)
	if err != nil {
		return domain.ForumSettings{}, err
	}
	settings = updated.Settings()
	if err := service.clearReadCache(ctx); err != nil {
		return domain.ForumSettings{}, err
	}
	return settings, service.publishForumAdminEvent(
		ctx,
		"forums.forum.settings_updated",
		settings.ForumID,
		command.ActorUserID,
		settingsPayload(settings),
	)
}

// GetForumPermissionSettings returns forum permission grants.
func (service Service) GetForumPermissionSettings(
	ctx context.Context,
	actorUserID uuid.UUID,
	forumID uuid.UUID,
) (domain.ForumPermissionSettings, error) {
	if err := service.requireManage(ctx, actorUserID, forumID); err != nil {
		return domain.ForumPermissionSettings{}, err
	}
	if _, err := service.forums.FindByID(ctx, forumID); err != nil {
		return domain.ForumPermissionSettings{}, err
	}
	if service.permissions == nil {
		return domain.ForumPermissionSettings{}, port.ErrForbidden
	}
	return service.permissions.ForumPermissionSettings(ctx, forumID)
}

// UpdateForumPermissionSettings replaces forum permission grants.
func (service Service) UpdateForumPermissionSettings(
	ctx context.Context,
	command port.UpdateForumPermissionSettingsCommand,
) error {
	settings := command.Settings.Normalize()
	if err := service.requireManage(ctx, command.ActorUserID, settings.ForumID); err != nil {
		return err
	}
	if _, err := service.forums.FindByID(ctx, settings.ForumID); err != nil {
		return err
	}
	if err := settings.Validate(); err != nil {
		return err
	}
	if service.permissions == nil {
		return port.ErrForbidden
	}
	return service.transactions.WithinTx(ctx, func(ctx context.Context) error {
		if err := service.permissions.UpdateForumPermissionSettings(
			ctx,
			command.ActorUserID,
			settings,
		); err != nil {
			return err
		}
		if err := service.clearReadCache(ctx); err != nil {
			return err
		}
		return service.publishForumAdminEvent(
			ctx,
			"forums.forum.permissions_updated",
			settings.ForumID,
			command.ActorUserID,
			map[string]any{
				"forum_id":    settings.ForumID,
				"grant_count": permissionGrantCount(settings),
			},
		)
	})
}

// SimulateForumPermission simulates one forum permission.
func (service Service) SimulateForumPermission(
	ctx context.Context,
	command port.SimulateForumPermissionCommand,
) (domain.ForumPermissionSimulationResult, error) {
	if err := service.requireManage(ctx, command.ActorUserID, command.ForumID); err != nil {
		return domain.ForumPermissionSimulationResult{}, err
	}
	if _, err := service.forums.FindByID(ctx, command.ForumID); err != nil {
		return domain.ForumPermissionSimulationResult{}, err
	}
	request := command.Request.Normalize(command.ForumID)
	if err := request.Validate(); err != nil {
		return domain.ForumPermissionSimulationResult{}, err
	}
	if service.permissions == nil {
		return domain.ForumPermissionSimulationResult{}, port.ErrForbidden
	}
	return service.permissions.SimulateForumPermission(ctx, command.ForumID, request)
}

func forumWithSettings(
	forum domain.Forum,
	settings domain.ForumSettings,
) domain.Forum {
	forum.Kind = settings.Kind
	forum.ExternalURL = settings.ExternalURL
	forum.ThreadVisibilityMode = settings.ThreadVisibilityMode
	forum.MaxStickyThreads = settings.MaxStickyThreads
	forum.DefaultThreadStatus = settings.DefaultThreadStatus
	forum.AuthorPostEditWindowSeconds = settings.AuthorPostEditWindowSeconds
	forum.AuthorPostDeleteWindowSeconds = settings.AuthorPostDeleteWindowSeconds
	return forum
}

// requireManage verifies forum-management permission.
func (service Service) requireManage(
	ctx context.Context,
	actorUserID uuid.UUID,
	forumID uuid.UUID,
) error {
	if service.authorizer == nil {
		return nil
	}
	allowed, err := service.authorizer.CanManageForum(ctx, actorUserID, forumID)
	if err != nil {
		return err
	}
	if !allowed {
		return port.ErrForbidden
	}
	return nil
}
