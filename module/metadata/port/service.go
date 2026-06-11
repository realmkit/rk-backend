package port

import (
	"context"

	"github.com/niflaot/gamehub-go/module/metadata/domain"
	"github.com/niflaot/gamehub-go/pkg/pagination"
)

// DefinitionService manages metafield definitions.
type DefinitionService interface {
	// CreateDefinition creates a metafield definition.
	CreateDefinition(ctx context.Context, command CreateDefinitionCommand) (DefinitionView, error)

	// UpdateDefinition updates a metafield definition.
	UpdateDefinition(ctx context.Context, command UpdateDefinitionCommand) (DefinitionView, error)

	// ArchiveDefinition archives a metafield definition.
	ArchiveDefinition(ctx context.Context, command ArchiveDefinitionCommand) error

	// GetDefinition returns a metafield definition.
	GetDefinition(ctx context.Context, query GetDefinitionQuery) (DefinitionView, error)

	// ListDefinitions returns metafield definitions.
	ListDefinitions(ctx context.Context, query ListDefinitionsQuery) (pagination.Result[DefinitionView], error)
}

// ValueService manages owner metafield values.
type ValueService interface {
	// SetValue creates or updates an owner value.
	SetValue(ctx context.Context, command SetValueCommand) (ValueView, bool, error)

	// GetValue returns one owner value.
	GetValue(ctx context.Context, query GetValueQuery) (ValueView, error)

	// ListValuesForOwner returns owner metadata.
	ListValuesForOwner(ctx context.Context, query ListValuesForOwnerQuery) (OwnerMetadataView, error)

	// DeleteValue deletes one owner value.
	DeleteValue(ctx context.Context, command DeleteValueCommand) error
}

// MetaobjectService manages metaobject definitions and entries.
type MetaobjectService interface {
	// CreateMetaobjectDefinition creates a metaobject definition.
	CreateMetaobjectDefinition(ctx context.Context, command CreateMetaobjectDefinitionCommand) (MetaobjectDefinitionView, error)

	// UpdateMetaobjectDefinition updates a metaobject definition.
	UpdateMetaobjectDefinition(ctx context.Context, command UpdateMetaobjectDefinitionCommand) (MetaobjectDefinitionView, error)

	// ArchiveMetaobjectDefinition archives a metaobject definition.
	ArchiveMetaobjectDefinition(ctx context.Context, command ArchiveMetaobjectDefinitionCommand) error

	// ListMetaobjectDefinitions returns metaobject definitions.
	ListMetaobjectDefinitions(
		ctx context.Context,
		query ListMetaobjectDefinitionsQuery,
	) (pagination.Result[MetaobjectDefinitionView], error)

	// GetMetaobjectDefinition returns one metaobject definition.
	GetMetaobjectDefinition(ctx context.Context, query GetMetaobjectDefinitionQuery) (MetaobjectDefinitionView, error)

	// CreateMetaobjectEntry creates a metaobject entry.
	CreateMetaobjectEntry(ctx context.Context, command CreateMetaobjectEntryCommand) (MetaobjectEntryView, error)

	// UpdateMetaobjectEntry updates a metaobject entry.
	UpdateMetaobjectEntry(ctx context.Context, command UpdateMetaobjectEntryCommand) (MetaobjectEntryView, error)

	// GetMetaobjectEntry returns one metaobject entry.
	GetMetaobjectEntry(ctx context.Context, query GetMetaobjectEntryQuery) (MetaobjectEntryView, error)

	// ListMetaobjectEntries returns metaobject entries.
	ListMetaobjectEntries(ctx context.Context, query ListMetaobjectEntriesQuery) (pagination.Result[MetaobjectEntryView], error)

	// DeleteMetaobjectEntry deletes one metaobject entry.
	DeleteMetaobjectEntry(ctx context.Context, command DeleteMetaobjectEntryCommand) error
}

// DefinitionView is the application representation of a metafield definition.
type DefinitionView = domain.MetafieldDefinition

// ValueView is the application representation of a metafield value.
type ValueView = domain.MetafieldValue

// MetaobjectDefinitionView is the application representation of a metaobject definition.
type MetaobjectDefinitionView = domain.MetaobjectDefinition

// MetaobjectEntryView is the application representation of a metaobject entry.
type MetaobjectEntryView = domain.MetaobjectEntry
