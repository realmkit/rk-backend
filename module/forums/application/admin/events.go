package admin

import (
	"context"

	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/module/forums/domain"
	eventdomain "github.com/realmkit/rk-backend/pkg/events/domain"
	"github.com/realmkit/rk-backend/pkg/events/emitter"
)

// publishForumAdminEvent publishes one forum admin event.
func (service Service) publishForumAdminEvent(
	ctx context.Context,
	key eventdomain.EventKey,
	forumID uuid.UUID,
	actorID uuid.UUID,
	payload any,
) error {
	return emitter.Publish(ctx, service.events, eventdomain.Draft{
		Key:           key,
		SchemaVersion: 1,
		Producer:      eventdomain.ProducerForums,
		AggregateType: "forum",
		AggregateID:   emitter.UUID(forumID),
		ActorUserID:   emitter.UUID(actorID),
		Payload:       payload,
		Scopes: []eventdomain.Scope{
			{Type: eventdomain.ScopeForum, ID: forumID.String()},
			{Type: eventdomain.ScopeSystem},
		},
	})
}

// settingsPayload returns a safe settings payload.
func settingsPayload(settings domain.ForumSettings) map[string]any {
	return map[string]any{
		"forum_id":                          settings.ForumID,
		"kind":                              settings.Kind,
		"thread_visibility_mode":            settings.ThreadVisibilityMode,
		"max_sticky_threads":                settings.MaxStickyThreads,
		"default_thread_status":             settings.DefaultThreadStatus,
		"author_post_edit_window_seconds":   settings.AuthorPostEditWindowSeconds,
		"author_post_delete_window_seconds": settings.AuthorPostDeleteWindowSeconds,
		"external_url_configured":           settings.ExternalURL != "",
	}
}

// permissionGrantCount returns the total configured grant count.
func permissionGrantCount(settings domain.ForumPermissionSettings) int {
	return len(settings.Viewers) +
		len(settings.Creators) +
		len(settings.Replyers) +
		len(settings.Likers) +
		len(settings.Moderators) +
		len(settings.Managers)
}
