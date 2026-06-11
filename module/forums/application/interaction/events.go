package interaction

import (
	"context"

	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/module/forums/domain"
	eventdomain "github.com/realmkit/rk-backend/pkg/events/domain"
	"github.com/realmkit/rk-backend/pkg/events/emitter"
)

// publishPostInteraction publishes one post interaction event.
func (service Service) publishPostInteraction(
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
		Payload: map[string]any{
			"post_id":   post.ID,
			"thread_id": post.ThreadID,
			"forum_id":  post.ForumID,
			"user_id":   actorID,
		},
		Scopes: []eventdomain.Scope{
			{Type: eventdomain.ScopeForum, ID: post.ForumID.String()},
			{Type: eventdomain.ScopeThread, ID: post.ThreadID.String()},
			{Type: eventdomain.ScopePost, ID: post.ID.String()},
			{Type: eventdomain.ScopeUser, ID: actorID.String()},
		},
	})
}

// publishReadEvent publishes one read-state event.
func (service Service) publishReadEvent(
	ctx context.Context,
	key eventdomain.EventKey,
	actorID uuid.UUID,
	aggregateID uuid.UUID,
	payload map[string]any,
	scopes []eventdomain.Scope,
) error {
	return emitter.Publish(ctx, service.events, eventdomain.Draft{
		Key:           key,
		SchemaVersion: 1,
		Producer:      eventdomain.ProducerForums,
		AggregateType: readAggregateType(key),
		AggregateID:   emitter.UUID(aggregateID),
		ActorUserID:   emitter.UUID(actorID),
		Payload:       payload,
		Scopes:        scopes,
	})
}

// readAggregateType returns the aggregate type for a read event.
func readAggregateType(key eventdomain.EventKey) eventdomain.AggregateType {
	if key == "forums.forum.read" {
		return "forum"
	}
	return "forum_thread"
}
