package application

import (
	"context"

	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/module/user/domain"
	eventdomain "github.com/realmkit/rk-backend/pkg/events/domain"
	"github.com/realmkit/rk-backend/pkg/events/emitter"
)

const (
	// userProvisionedEvent is emitted for first local user creation.
	userProvisionedEvent eventdomain.EventKey = "users.user.provisioned"

	// userUpdatedEvent is emitted when local user settings change.
	userUpdatedEvent eventdomain.EventKey = "users.user.updated"

	// identityLinkedEvent is emitted when an identity is linked.
	identityLinkedEvent eventdomain.EventKey = "users.identity.linked"

	// identityClaimRefreshedEvent is emitted when identity claims are refreshed.
	identityClaimRefreshedEvent eventdomain.EventKey = "users.identity.claim_refreshed"
)

// publishUserEvent publishes one user event.
func (service Service) publishUserEvent(
	ctx context.Context,
	key eventdomain.EventKey,
	user domain.User,
	actorID uuid.UUID,
) error {
	return emitter.Publish(ctx, service.events, eventdomain.Draft{
		Key:           key,
		SchemaVersion: 1,
		Producer:      eventdomain.ProducerUsers,
		AggregateType: "user",
		AggregateID:   emitter.UUID(user.ID),
		ActorUserID:   emitter.UUID(actorID),
		Payload: map[string]any{
			"id":      user.ID,
			"status":  user.Status,
			"version": user.Version,
		},
		Scopes: []eventdomain.Scope{{Type: eventdomain.ScopeUser, ID: user.ID.String()}},
	})
}

// publishIdentityEvent publishes one identity-link event.
func (service Service) publishIdentityEvent(
	ctx context.Context,
	key eventdomain.EventKey,
	link domain.IdentityLink,
) error {
	return emitter.Publish(ctx, service.events, eventdomain.Draft{
		Key:           key,
		SchemaVersion: 1,
		Producer:      eventdomain.ProducerUsers,
		AggregateType: "identity",
		AggregateID:   emitter.UUID(link.ID),
		ActorUserID:   emitter.UUID(link.UserID),
		Payload: map[string]any{
			"id":           link.ID,
			"user_id":      link.UserID,
			"provider":     link.Provider,
			"issuer":       link.Issuer,
			"subject_hash": link.SubjectHash,
		},
		Scopes: []eventdomain.Scope{{Type: eventdomain.ScopeUser, ID: link.UserID.String()}},
	})
}
