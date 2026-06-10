package application

import (
	"context"

	"github.com/google/uuid"
	"github.com/niflaot/gamehub-go/module/groups/domain"
	eventdomain "github.com/niflaot/gamehub-go/pkg/events/domain"
	"github.com/niflaot/gamehub-go/pkg/events/emitter"
)

const (
	// groupCreatedEvent is emitted when a group is created.
	groupCreatedEvent eventdomain.EventKey = "groups.group.created"

	// groupUpdatedEvent is emitted when a group is updated.
	groupUpdatedEvent eventdomain.EventKey = "groups.group.updated"

	// groupDeletedEvent is emitted when a group is deleted.
	groupDeletedEvent eventdomain.EventKey = "groups.group.deleted"

	// membershipAddedEvent is emitted when membership is assigned.
	membershipAddedEvent eventdomain.EventKey = "groups.membership.added"

	// membershipRemovedEvent is emitted when membership is removed.
	membershipRemovedEvent eventdomain.EventKey = "groups.membership.removed"

	// relationTupleCreatedEvent is emitted when a tuple is created.
	relationTupleCreatedEvent eventdomain.EventKey = "groups.relation_tuple.created"

	// relationTupleDeletedEvent is emitted when a tuple is deleted.
	relationTupleDeletedEvent eventdomain.EventKey = "groups.relation_tuple.deleted"
)

// publishGroupEvent publishes one group lifecycle event.
func (service Service) publishGroupEvent(
	ctx context.Context,
	key eventdomain.EventKey,
	group domain.Group,
) error {
	return emitter.Publish(ctx, service.events, eventdomain.Draft{
		Key:           key,
		SchemaVersion: 1,
		Producer:      eventdomain.ProducerGroups,
		AggregateType: "group",
		AggregateID:   emitter.UUID(group.ID),
		Payload:       groupPayload(group),
		Scopes:        groupScopes(group.ID),
	})
}

// publishMembershipEvent publishes one group membership event.
func (service Service) publishMembershipEvent(
	ctx context.Context,
	key eventdomain.EventKey,
	membership domain.Membership,
) error {
	return emitter.Publish(ctx, service.events, eventdomain.Draft{
		Key:           key,
		SchemaVersion: 1,
		Producer:      eventdomain.ProducerGroups,
		AggregateType: "group_membership",
		AggregateID:   emitter.UUID(membership.ID),
		Payload: map[string]any{
			"id":       membership.ID,
			"group_id": membership.GroupID,
			"user_id":  membership.UserID,
			"status":   membership.Status,
		},
		Scopes: []eventdomain.Scope{
			{Type: eventdomain.ScopeGroup, ID: membership.GroupID.String()},
			{Type: eventdomain.ScopeUser, ID: membership.UserID.String()},
		},
	})
}

// publishTupleEvent publishes one relation tuple event.
func (service Service) publishTupleEvent(
	ctx context.Context,
	key eventdomain.EventKey,
	tuple domain.RelationTuple,
) error {
	return emitter.Publish(ctx, service.events, eventdomain.Draft{
		Key:           key,
		SchemaVersion: 1,
		Producer:      eventdomain.ProducerGroups,
		AggregateType: "relation_tuple",
		AggregateID:   emitter.UUID(tuple.ID),
		Payload: map[string]any{
			"id":           tuple.ID,
			"object_type":  tuple.ObjectType,
			"object_id":    tuple.ObjectID,
			"relation":     tuple.Relation,
			"subject_type": tuple.SubjectType,
			"subject_id":   tuple.SubjectID,
		},
		Scopes: systemScopes(),
	})
}

// publishTupleDeleted publishes one relation tuple deletion event.
func (service Service) publishTupleDeleted(ctx context.Context, id uuid.UUID) error {
	return emitter.Publish(ctx, service.events, eventdomain.Draft{
		Key:           relationTupleDeletedEvent,
		SchemaVersion: 1,
		Producer:      eventdomain.ProducerGroups,
		AggregateType: "relation_tuple",
		AggregateID:   emitter.UUID(id),
		Payload:       map[string]any{"id": id},
		Scopes:        systemScopes(),
	})
}

// groupPayload returns a safe group payload.
func groupPayload(group domain.Group) map[string]any {
	return map[string]any{
		"id":      group.ID,
		"key":     group.Key,
		"name":    group.Name,
		"color":   group.Color,
		"weight":  group.Weight,
		"status":  group.Status,
		"version": group.Version,
	}
}

// groupScopes returns audience scopes for group events.
func groupScopes(groupID uuid.UUID) []eventdomain.Scope {
	return []eventdomain.Scope{{Type: eventdomain.ScopeGroup, ID: groupID.String()}}
}

// systemScopes returns private operational scopes.
func systemScopes() []eventdomain.Scope {
	return []eventdomain.Scope{{Type: eventdomain.ScopeSystem}}
}
