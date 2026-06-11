package port

import (
	"context"

	"github.com/google/uuid"
	"github.com/niflaot/gamehub-go/module/metadata/domain"
	"github.com/niflaot/gamehub-go/pkg/pagination"
)

// MetafieldDefinitionRepository stores metafield definitions.
type MetafieldDefinitionRepository interface {
	// Create stores definition.
	Create(ctx context.Context, definition domain.MetafieldDefinition) (domain.MetafieldDefinition, error)

	// Update stores mutable definition changes.
	Update(ctx context.Context, definition domain.MetafieldDefinition, expectedVersion uint64) (domain.MetafieldDefinition, error)

	// FindByID returns one definition by ID.
	FindByID(ctx context.Context, id uuid.UUID) (domain.MetafieldDefinition, error)

	// FindByKey returns one definition by owner type, namespace, and key.
	FindByKey(
		ctx context.Context,
		ownerType domain.OwnerType,
		namespace domain.Namespace,
		key domain.Key,
	) (domain.MetafieldDefinition, error)

	// List returns definitions matching filter.
	List(ctx context.Context, filter DefinitionFilter, page pagination.Page) (pagination.Result[domain.MetafieldDefinition], error)

	// Archive soft deletes definition.
	Archive(ctx context.Context, id uuid.UUID, expectedVersion uint64) error
}

// MetafieldValueRepository stores owner metafield values.
type MetafieldValueRepository interface {
	// Upsert creates or updates value.
	Upsert(ctx context.Context, value domain.MetafieldValue, expectedVersion *uint64) (domain.MetafieldValue, bool, error)

	// Find returns one owner value.
	Find(ctx context.Context, definitionID uuid.UUID, ownerType domain.OwnerType, ownerID uuid.UUID) (domain.MetafieldValue, error)

	// ListForOwner returns all values for owner.
	ListForOwner(ctx context.Context, ownerType domain.OwnerType, ownerID uuid.UUID) ([]domain.MetafieldValue, error)

	// Delete soft deletes one owner value.
	Delete(ctx context.Context, definitionID uuid.UUID, ownerType domain.OwnerType, ownerID uuid.UUID, expectedVersion uint64) error

	// CountByDefinition returns active value count for definition.
	CountByDefinition(ctx context.Context, definitionID uuid.UUID) (int64, error)
}

// MetaobjectDefinitionRepository stores metaobject definitions.
type MetaobjectDefinitionRepository interface {
	// Create stores definition.
	Create(ctx context.Context, definition domain.MetaobjectDefinition) (domain.MetaobjectDefinition, error)

	// Update stores mutable definition changes.
	Update(ctx context.Context, definition domain.MetaobjectDefinition, expectedVersion uint64) (domain.MetaobjectDefinition, error)

	// FindByID returns one definition by ID.
	FindByID(ctx context.Context, id uuid.UUID) (domain.MetaobjectDefinition, error)

	// FindByType returns one definition by type.
	FindByType(ctx context.Context, objectType domain.MetaobjectType) (domain.MetaobjectDefinition, error)

	// List returns definitions matching filter.
	List(
		ctx context.Context,
		filter MetaobjectDefinitionFilter,
		page pagination.Page,
	) (pagination.Result[domain.MetaobjectDefinition], error)

	// Archive soft deletes definition.
	Archive(ctx context.Context, id uuid.UUID, expectedVersion uint64) error
}

// MetaobjectEntryRepository stores metaobject entries.
type MetaobjectEntryRepository interface {
	// Create stores entry.
	Create(ctx context.Context, entry domain.MetaobjectEntry) (domain.MetaobjectEntry, error)

	// Update stores mutable entry changes.
	Update(ctx context.Context, entry domain.MetaobjectEntry, expectedVersion uint64) (domain.MetaobjectEntry, error)

	// FindByID returns one entry by ID.
	FindByID(ctx context.Context, id uuid.UUID) (domain.MetaobjectEntry, error)

	// FindByHandle returns one entry by definition and handle.
	FindByHandle(ctx context.Context, definitionID uuid.UUID, handle domain.Handle) (domain.MetaobjectEntry, error)

	// List returns entries for definition.
	List(ctx context.Context, definitionID uuid.UUID, page pagination.Page) (pagination.Result[domain.MetaobjectEntry], error)

	// Delete soft deletes entry.
	Delete(ctx context.Context, id uuid.UUID, expectedVersion uint64) error

	// CountByDefinition returns active entry count for definition.
	CountByDefinition(ctx context.Context, definitionID uuid.UUID) (int64, error)
}

// OwnerResolver checks owner existence without coupling metadata to owner modules.
type OwnerResolver interface {
	// Exists reports whether owner exists.
	Exists(ctx context.Context, ownerType domain.OwnerType, ownerID uuid.UUID) (bool, error)
}

// ReferenceResolver checks value references.
type ReferenceResolver interface {
	// OwnerExists reports whether owner reference exists.
	OwnerExists(ctx context.Context, reference domain.OwnerReference) (bool, error)

	// MetaobjectEntryExists reports whether metaobject entry reference exists.
	MetaobjectEntryExists(ctx context.Context, reference domain.MetaobjectReference) (bool, error)
}
