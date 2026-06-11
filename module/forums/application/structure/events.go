package structure

import (
	"context"

	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/module/forums/domain"
	eventdomain "github.com/realmkit/rk-backend/pkg/events/domain"
	"github.com/realmkit/rk-backend/pkg/events/emitter"
)

// publishCategoryEvent publishes one forum category event.
func (service Service) publishCategoryEvent(
	ctx context.Context,
	key eventdomain.EventKey,
	category domain.ForumCategory,
	actorID uuid.UUID,
) error {
	return emitter.Publish(ctx, service.events, eventdomain.Draft{
		Key:           key,
		SchemaVersion: 1,
		Producer:      eventdomain.ProducerForums,
		AggregateType: "forum_category",
		AggregateID:   emitter.UUID(category.ID),
		ActorUserID:   emitter.UUID(actorID),
		Payload:       categoryPayload(category),
		Scopes:        []eventdomain.Scope{{Type: eventdomain.ScopeGlobal}},
	})
}

// publishForumEvent publishes one forum structure event.
func (service Service) publishForumEvent(
	ctx context.Context,
	key eventdomain.EventKey,
	forum domain.Forum,
	actorID uuid.UUID,
) error {
	return emitter.Publish(ctx, service.events, eventdomain.Draft{
		Key:           key,
		SchemaVersion: 1,
		Producer:      eventdomain.ProducerForums,
		AggregateType: "forum",
		AggregateID:   emitter.UUID(forum.ID),
		ActorUserID:   emitter.UUID(actorID),
		Payload:       forumPayload(forum),
		Scopes:        []eventdomain.Scope{{Type: eventdomain.ScopeForum, ID: forum.ID.String()}},
	})
}

// categoryPayload returns a safe category payload.
func categoryPayload(category domain.ForumCategory) map[string]any {
	return map[string]any{
		"id":            category.ID,
		"key":           category.Key,
		"name":          category.Name,
		"display_order": category.DisplayOrder,
		"status":        category.Status,
		"version":       category.Version,
	}
}

// forumPayload returns a safe forum payload.
func forumPayload(forum domain.Forum) map[string]any {
	return map[string]any{
		"id":                     forum.ID,
		"category_id":            forum.CategoryID,
		"parent_forum_id":        forum.ParentForumID,
		"kind":                   forum.Kind,
		"key":                    forum.Key,
		"slug":                   forum.Slug,
		"name":                   forum.Name,
		"display_order":          forum.DisplayOrder,
		"path":                   forum.Path,
		"depth":                  forum.Depth,
		"thread_visibility_mode": forum.ThreadVisibilityMode,
		"status":                 forum.Status,
		"version":                forum.Version,
	}
}
