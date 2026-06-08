package application

import (
	"context"

	"github.com/google/uuid"
	"github.com/niflaot/gamehub-go/module/metadata/domain"
	"github.com/niflaot/gamehub-go/module/metadata/port"
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
