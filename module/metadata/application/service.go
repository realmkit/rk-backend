package application

import (
	"context"

	"github.com/google/uuid"
	"github.com/niflaot/gamehub-go/module/metadata/domain"
	"github.com/niflaot/gamehub-go/module/metadata/port"
	eventdomain "github.com/niflaot/gamehub-go/pkg/events/domain"
	"github.com/niflaot/gamehub-go/pkg/events/emitter"
)

// Dependencies contains metadata service collaborators.
type Dependencies struct {
	// Definitions stores metafield definitions.
	Definitions port.MetafieldDefinitionRepository

	// Values stores metafield values.
	Values port.MetafieldValueRepository

	// MetaobjectDefinitions stores metaobject definitions.
	MetaobjectDefinitions port.MetaobjectDefinitionRepository

	// MetaobjectEntries stores metaobject entries.
	MetaobjectEntries port.MetaobjectEntryRepository

	// Owners resolves metadata owners.
	Owners port.OwnerResolver

	// References resolves value references.
	References port.ReferenceResolver

	// Policy authorizes metadata operations.
	Policy port.Policy

	// Events publishes metadata lifecycle events.
	Events emitter.Publisher
}

// Service implements metadata application use cases.
type Service struct {
	definitions           port.MetafieldDefinitionRepository
	values                port.MetafieldValueRepository
	metaobjectDefinitions port.MetaobjectDefinitionRepository
	metaobjectEntries     port.MetaobjectEntryRepository
	owners                port.OwnerResolver
	references            port.ReferenceResolver
	policy                port.Policy
	events                emitter.Publisher
}

// NewService creates a metadata service.
func NewService(dependencies Dependencies) Service {
	service := Service{
		definitions:           dependencies.Definitions,
		values:                dependencies.Values,
		metaobjectDefinitions: dependencies.MetaobjectDefinitions,
		metaobjectEntries:     dependencies.MetaobjectEntries,
		owners:                dependencies.Owners,
		references:            dependencies.References,
		policy:                dependencies.Policy,
		events:                dependencies.Events,
	}
	if service.owners == nil {
		service.owners = ExistingOwnerResolver{}
	}
	if service.references == nil {
		service.references = ExistingReferenceResolver{}
	}
	if service.policy == nil {
		service.policy = port.AllowAllPolicy{}
	}
	return service
}

// ExistingOwnerResolver treats every syntactically valid owner as existing.
type ExistingOwnerResolver struct{}

// Exists reports whether owner exists.
func (ExistingOwnerResolver) Exists(context.Context, domain.OwnerType, uuid.UUID) (bool, error) {
	return true, nil
}

// ExistingReferenceResolver treats every syntactically valid reference as existing.
type ExistingReferenceResolver struct{}

// OwnerExists reports whether owner reference exists.
func (ExistingReferenceResolver) OwnerExists(context.Context, domain.OwnerReference) (bool, error) {
	return true, nil
}

// MetaobjectEntryExists reports whether metaobject entry reference exists.
func (ExistingReferenceResolver) MetaobjectEntryExists(context.Context, domain.MetaobjectReference) (bool, error) {
	return true, nil
}

// ensureDependencies returns an error when required repositories are absent.
func (service Service) ensureDependencies() error {
	if service.definitions == nil || service.values == nil || service.metaobjectDefinitions == nil || service.metaobjectEntries == nil {
		return port.ErrNotFound
	}
	return nil
}

// publishMetadataEvent publishes one metadata event.
func (service Service) publishMetadataEvent(
	ctx context.Context,
	key eventdomain.EventKey,
	aggregateType eventdomain.AggregateType,
	aggregateID uuid.UUID,
	actor port.Actor,
	payload any,
	scopes []eventdomain.Scope,
) error {
	if len(scopes) == 0 {
		scopes = []eventdomain.Scope{{Type: eventdomain.ScopeSystem}}
	}
	return emitter.Publish(ctx, service.events, eventdomain.Draft{
		Key:           key,
		SchemaVersion: 1,
		Producer:      eventdomain.ProducerMetadata,
		AggregateType: aggregateType,
		AggregateID:   emitter.UUID(aggregateID),
		ActorUserID:   emitter.UUID(actor.ID),
		Payload:       payload,
		Scopes:        scopes,
	})
}

// metadataDefinitionPayload returns a safe definition payload.
func metadataDefinitionPayload(definition domain.MetafieldDefinition) map[string]any {
	return map[string]any{
		"id":         definition.ID,
		"owner_type": definition.OwnerType,
		"namespace":  definition.Namespace,
		"key":        definition.Key,
		"value_type": definition.ValueType,
		"active":     definition.Active,
		"version":    definition.Version,
	}
}

// metadataValuePayload returns a safe owner value payload.
func metadataValuePayload(value domain.MetafieldValue) map[string]any {
	return map[string]any{
		"id":            value.ID,
		"definition_id": value.DefinitionID,
		"owner_type":    value.OwnerType,
		"owner_id":      value.OwnerID,
		"version":       value.Version,
	}
}

// metadataOwnerScopes returns event scopes for one metadata owner.
func metadataOwnerScopes(ownerType domain.OwnerType, ownerID uuid.UUID) []eventdomain.Scope {
	return []eventdomain.Scope{{
		Type: ownerScopeType(ownerType),
		ID:   ownerID.String(),
	}}
}

// ownerScopeType maps metadata owner types to event scope types.
func ownerScopeType(ownerType domain.OwnerType) eventdomain.ScopeType {
	switch ownerType {
	case domain.OwnerAsset:
		return eventdomain.ScopeAsset
	case domain.OwnerForum:
		return eventdomain.ScopeForum
	case domain.OwnerForumThread:
		return eventdomain.ScopeThread
	case domain.OwnerGroup:
		return eventdomain.ScopeGroup
	case domain.OwnerUser:
		return eventdomain.ScopeUser
	default:
		return eventdomain.ScopeSystem
	}
}

// eventKey converts a stable event key string.
func eventKey(key string) eventdomain.EventKey {
	return eventdomain.EventKey(key)
}
