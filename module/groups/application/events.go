package application

import (
	"context"

	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/module/groups/domain"
	eventdomain "github.com/realmkit/rk-backend/pkg/events/domain"
	"github.com/realmkit/rk-backend/pkg/events/emitter"
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

	// permissionGrantCreatedEvent is emitted when a grant is created.
	permissionGrantCreatedEvent eventdomain.EventKey = "groups.permission_grant.created"

	// permissionGrantDeletedEvent is emitted when a grant is deleted.
	permissionGrantDeletedEvent eventdomain.EventKey = "groups.permission_grant.deleted"
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

// publishGrantEvent publishes one permission grant event.
func (service Service) publishGrantEvent(
	ctx context.Context,
	key eventdomain.EventKey,
	grant domain.PermissionGrant,
) error {
	return emitter.Publish(ctx, service.events, eventdomain.Draft{
		Key:           key,
		SchemaVersion: 1,
		Producer:      eventdomain.ProducerGroups,
		AggregateType: "permission_grant",
		AggregateID:   emitter.UUID(grant.ID),
		Payload: map[string]any{
			"id":           grant.ID,
			"action":       grant.Action,
			"scope_type":   grant.ScopeType,
			"scope_id":     grant.ScopeID,
			"subject_type": grant.SubjectType,
			"subject_id":   grant.SubjectID,
		},
		Scopes: systemScopes(),
	})
}

// publishGrantDeleted publishes one permission grant deletion event.
func (service Service) publishGrantDeleted(ctx context.Context, id uuid.UUID) error {
	return emitter.Publish(ctx, service.events, eventdomain.Draft{
		Key:           permissionGrantDeletedEvent,
		SchemaVersion: 1,
		Producer:      eventdomain.ProducerGroups,
		AggregateType: "permission_grant",
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
