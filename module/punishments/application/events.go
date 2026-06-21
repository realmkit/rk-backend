package application

import (
	"context"

	"github.com/realmkit/rk-backend/module/punishments/domain"
	eventdomain "github.com/realmkit/rk-backend/pkg/events/domain"
	"github.com/realmkit/rk-backend/pkg/events/emitter"
)

// event producer names.
const producerPunishments eventdomain.Producer = "punishments"

// publishDefinitionEvent supports package behavior.
func (service Service) publishDefinitionEvent(
	ctx context.Context,
	key eventdomain.EventKey,
	definition domain.Definition,
) error {
	return emitter.Publish(ctx, service.events, eventdomain.Draft{
		Key:           key,
		SchemaVersion: 1,
		Producer:      producerPunishments,
		AggregateType: "punishment_definition",
		AggregateID:   emitter.UUID(definition.ID),
		Payload: map[string]any{
			"id":       definition.ID,
			"key":      definition.Key,
			"name":     definition.Name,
			"status":   definition.Status,
			"severity": definition.Severity,
		},
		Scopes: []eventdomain.Scope{
			{Type: eventdomain.ScopeStaff},
			{Type: eventdomain.ScopeSystem},
		},
	})
}

// publishPunishmentEvent supports package behavior.
func (service Service) publishPunishmentEvent(
	ctx context.Context,
	key eventdomain.EventKey,
	punishment domain.Punishment,
) error {
	payload := map[string]any{
		"id":             punishment.ID,
		"definition_id":  punishment.DefinitionID,
		"target_user_id": punishment.TargetUserID,
		"target_ip_hash": punishment.TargetIPHash,
		"issuer_type":    punishment.IssuerType,
		"issuer_user_id": punishment.IssuerUserID,
		"issuer_key":     punishment.IssuerKey,
		"reason":         punishment.Reason,
		"status":         punishment.Status,
		"starts_at":      punishment.StartsAt,
		"expires_at":     punishment.ExpiresAt,
	}
	return emitter.Publish(ctx, service.events, eventdomain.Draft{
		Key:           key,
		SchemaVersion: 1,
		Producer:      producerPunishments,
		AggregateType: "punishment",
		AggregateID:   emitter.UUID(punishment.ID),
		Payload:       payload,
		ActorUserID:   punishment.IssuerUserID,
		DedupeKey:     string(key) + ":" + punishment.ID.String(),
		Scopes: []eventdomain.Scope{
			{Type: eventdomain.ScopeStaff},
			{Type: eventdomain.ScopeUser, ID: punishment.TargetUserID.String()},
			{Type: eventdomain.ScopePunishment, ID: punishment.ID.String()},
			{Type: eventdomain.ScopeSystem},
		},
	})
}

// publishOperationsEvent supports package behavior.
func (service Service) publishOperationsEvent(
	ctx context.Context,
	key eventdomain.EventKey,
	count int64,
) error {
	return emitter.Publish(ctx, service.events, eventdomain.Draft{
		Key:           key,
		SchemaVersion: 1,
		Producer:      producerPunishments,
		AggregateType: "punishment_operation",
		Payload:       map[string]any{"count": count},
		Scopes: []eventdomain.Scope{
			{Type: eventdomain.ScopeStaff},
			{Type: eventdomain.ScopeSystem},
		},
	})
}
