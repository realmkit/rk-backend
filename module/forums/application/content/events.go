package content

import (
	"context"

	"github.com/google/uuid"
	"github.com/niflaot/gamehub-go/module/forums/domain"
	eventdomain "github.com/niflaot/gamehub-go/pkg/events/domain"
	"github.com/niflaot/gamehub-go/pkg/events/emitter"
)

// publishThreadEvent publishes one thread event.
func (service Service) publishThreadEvent(
	ctx context.Context,
	key eventdomain.EventKey,
	thread domain.Thread,
	actorID uuid.UUID,
) error {
	return emitter.Publish(ctx, service.events, eventdomain.Draft{
		Key:           key,
		SchemaVersion: 1,
		Producer:      eventdomain.ProducerForums,
		AggregateType: "forum_thread",
		AggregateID:   emitter.UUID(thread.ID),
		ActorUserID:   emitter.UUID(actorID),
		Payload:       threadPayload(thread),
		Scopes:        threadScopes(thread),
	})
}

// publishPostEvent publishes one post event.
func (service Service) publishPostEvent(
	ctx context.Context,
	key eventdomain.EventKey,
	post domain.Post,
	actorID uuid.UUID,
) error {
	return emitter.Publish(ctx, service.events, eventdomain.Draft{
		Key:           key,
		SchemaVersion: 1,
		Producer:      eventdomain.ProducerForums,
		AggregateType: "forum_post",
		AggregateID:   emitter.UUID(post.ID),
		ActorUserID:   emitter.UUID(actorID),
		Payload:       postPayload(post),
		Scopes:        postScopes(post),
	})
}

// threadPayload returns a safe thread payload.
func threadPayload(thread domain.Thread) map[string]any {
	return map[string]any{
		"id":                 thread.ID,
		"forum_id":           thread.ForumID,
		"author_user_id":     thread.AuthorUserID,
		"title":              thread.Title,
		"slug":               thread.Slug,
		"status":             thread.Status,
		"sticky_state":       thread.StickyState,
		"post_count":         thread.PostCount,
		"visible_post_count": thread.VisiblePostCount,
		"version":            thread.Version,
	}
}

// postPayload returns a safe post payload.
func postPayload(post domain.Post) map[string]any {
	return map[string]any{
		"id":             post.ID,
		"thread_id":      post.ThreadID,
		"forum_id":       post.ForumID,
		"author_user_id": post.AuthorUserID,
		"sequence":       post.Sequence,
		"status":         post.Status,
		"like_count":     post.LikeCount,
		"edit_count":     post.EditCount,
		"version":        post.Version,
	}
}

// threadScopes returns event scopes for one thread.
func threadScopes(thread domain.Thread) []eventdomain.Scope {
	return []eventdomain.Scope{
		{Type: eventdomain.ScopeForum, ID: thread.ForumID.String()},
		{Type: eventdomain.ScopeThread, ID: thread.ID.String()},
	}
}

// postScopes returns event scopes for one post.
func postScopes(post domain.Post) []eventdomain.Scope {
	return []eventdomain.Scope{
		{Type: eventdomain.ScopeForum, ID: post.ForumID.String()},
		{Type: eventdomain.ScopeThread, ID: post.ThreadID.String()},
		{Type: eventdomain.ScopePost, ID: post.ID.String()},
	}
}
