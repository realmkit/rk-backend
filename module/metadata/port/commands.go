package port

import (
	"encoding/json"

	"github.com/google/uuid"
	"github.com/niflaot/gamehub-go/module/metadata/domain"
	"github.com/niflaot/gamehub-go/pkg/pagination"
)

// CreateDefinitionCommand creates a metafield definition.
type CreateDefinitionCommand struct {
	// Actor is the caller.
	Actor Actor

	// Definition is the definition to create.
	Definition domain.MetafieldDefinition
}

// UpdateDefinitionCommand updates a metafield definition.
type UpdateDefinitionCommand struct {
	// Actor is the caller.
	Actor Actor

	// Definition is the updated definition state.
	Definition domain.MetafieldDefinition

	// ExpectedVersion is the required current version.
	ExpectedVersion uint64
}

// ArchiveDefinitionCommand archives a metafield definition.
type ArchiveDefinitionCommand struct {
	// Actor is the caller.
	Actor Actor

	// ID is the definition identifier.
	ID uuid.UUID

	// ExpectedVersion is the required current version.
	ExpectedVersion uint64
}

// GetDefinitionQuery fetches a metafield definition.
type GetDefinitionQuery struct {
	// Actor is the caller.
	Actor Actor

	// ID is the definition identifier.
	ID uuid.UUID
}

// ListDefinitionsQuery lists metafield definitions.
type ListDefinitionsQuery struct {
	// Actor is the caller.
	Actor Actor

	// Filter contains definition filters.
	Filter DefinitionFilter

	// Page contains pagination options.
	Page pagination.Page
}

// SetValueCommand writes one owner value.
type SetValueCommand struct {
	// Actor is the caller.
	Actor Actor

	// Owner is the metadata owner.
	Owner OwnerRef

	// Namespace is the definition namespace.
	Namespace domain.Namespace

	// Key is the definition key.
	Key domain.Key

	// RawValue is the incoming JSON value.
	RawValue json.RawMessage

	// ExpectedVersion is the optional required current value version.
	ExpectedVersion *uint64
}

// GetValueQuery fetches one owner value.
type GetValueQuery struct {
	// Actor is the caller.
	Actor Actor

	// Owner is the metadata owner.
	Owner OwnerRef

	// Namespace is the definition namespace.
	Namespace domain.Namespace

	// Key is the definition key.
	Key domain.Key
}

// ListValuesForOwnerQuery lists owner metadata.
type ListValuesForOwnerQuery struct {
	// Actor is the caller.
	Actor Actor

	// Owner is the metadata owner.
	Owner OwnerRef

	// Namespace filters definitions when present.
	Namespace domain.Namespace

	// IncludeEmpty reports whether absent values should be included.
	IncludeEmpty bool
}

// DeleteValueCommand deletes one owner value.
type DeleteValueCommand struct {
	// Actor is the caller.
	Actor Actor

	// Owner is the metadata owner.
	Owner OwnerRef

	// Namespace is the definition namespace.
	Namespace domain.Namespace

	// Key is the definition key.
	Key domain.Key

	// ExpectedVersion is the required current value version.
	ExpectedVersion uint64
}

// CreateMetaobjectDefinitionCommand creates a metaobject definition.
type CreateMetaobjectDefinitionCommand struct {
	// Actor is the caller.
	Actor Actor

	// Definition is the definition to create.
	Definition domain.MetaobjectDefinition
}

// UpdateMetaobjectDefinitionCommand updates a metaobject definition.
type UpdateMetaobjectDefinitionCommand struct {
	// Actor is the caller.
	Actor Actor

	// Definition is the updated definition state.
	Definition domain.MetaobjectDefinition

	// ExpectedVersion is the required current version.
	ExpectedVersion uint64
}

// ArchiveMetaobjectDefinitionCommand archives a metaobject definition.
type ArchiveMetaobjectDefinitionCommand struct {
	// Actor is the caller.
	Actor Actor

	// ID is the definition identifier.
	ID uuid.UUID

	// ExpectedVersion is the required current version.
	ExpectedVersion uint64
}

// ListMetaobjectDefinitionsQuery lists metaobject definitions.
type ListMetaobjectDefinitionsQuery struct {
	// Actor is the caller.
	Actor Actor

	// Filter contains definition filters.
	Filter MetaobjectDefinitionFilter

	// Page contains pagination options.
	Page pagination.Page
}

// GetMetaobjectDefinitionQuery fetches one metaobject definition.
type GetMetaobjectDefinitionQuery struct {
	// Actor is the caller.
	Actor Actor

	// ID is the definition identifier.
	ID uuid.UUID
}

// CreateMetaobjectEntryCommand creates a metaobject entry.
type CreateMetaobjectEntryCommand struct {
	// Actor is the caller.
	Actor Actor

	// Entry is the entry to create.
	Entry domain.MetaobjectEntry

	// RawFields contains incoming field values.
	RawFields map[domain.Key]json.RawMessage
}

// UpdateMetaobjectEntryCommand updates a metaobject entry.
type UpdateMetaobjectEntryCommand struct {
	// Actor is the caller.
	Actor Actor

	// Entry is the updated entry state.
	Entry domain.MetaobjectEntry

	// RawFields contains incoming field values.
	RawFields map[domain.Key]json.RawMessage

	// ExpectedVersion is the required current version.
	ExpectedVersion uint64
}

// GetMetaobjectEntryQuery fetches one metaobject entry.
type GetMetaobjectEntryQuery struct {
	// Actor is the caller.
	Actor Actor

	// ID is the entry identifier.
	ID uuid.UUID
}

// ListMetaobjectEntriesQuery lists metaobject entries.
type ListMetaobjectEntriesQuery struct {
	// Actor is the caller.
	Actor Actor

	// DefinitionID is the definition identifier.
	DefinitionID uuid.UUID

	// Page contains pagination options.
	Page pagination.Page
}

// DeleteMetaobjectEntryCommand deletes a metaobject entry.
type DeleteMetaobjectEntryCommand struct {
	// Actor is the caller.
	Actor Actor

	// ID is the entry identifier.
	ID uuid.UUID

	// ExpectedVersion is the required current version.
	ExpectedVersion uint64
}
