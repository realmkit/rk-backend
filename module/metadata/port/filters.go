package port

import (
	"context"

	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/module/metadata/domain"
	"github.com/realmkit/rk-backend/pkg/pagination"
)

// DefinitionFilter contains metafield definition list filters.
type DefinitionFilter struct {
	// OwnerType filters by owner type when present.
	OwnerType domain.OwnerType

	// Active filters by active state when present.
	Active *bool
}

// MetaobjectDefinitionFilter contains metaobject definition list filters.
type MetaobjectDefinitionFilter struct {
	// Type filters by metaobject type when present.
	Type domain.MetaobjectType

	// Active filters by active state when present.
	Active *bool
}

// Actor identifies the caller for policy checks.
type Actor struct {
	// ID is the actor identifier.
	ID uuid.UUID
}

// OwnerRef identifies one metadata owner.
type OwnerRef struct {
	// Type is the owner type.
	Type domain.OwnerType `json:"type"`

	// ID is the owner identifier.
	ID uuid.UUID `json:"id"`
}

// OwnerMetadataView represents all metadata for one owner.
type OwnerMetadataView struct {
	// Owner is the metadata owner.
	Owner OwnerRef `json:"owner"`

	// Metafields contains metadata fields with optional values.
	Metafields []OwnerMetafieldView `json:"metafields"`
}

// OwnerMetafieldView represents one definition and optional value.
type OwnerMetafieldView struct {
	// Definition is the metafield definition.
	Definition domain.MetafieldDefinition `json:"definition"`

	// Value is the owner value when present.
	Value *domain.MetafieldValue `json:"value,omitempty"`
}

// Policy authorizes metadata operations.
type Policy interface {
	// CanManageDefinitions authorizes metafield definition management.
	CanManageDefinitions(ctx context.Context, actor Actor) error

	// CanReadOwnerMetadata authorizes owner metadata reads.
	CanReadOwnerMetadata(ctx context.Context, actor Actor, owner OwnerRef) error

	// CanWriteOwnerMetadata authorizes owner metadata writes.
	CanWriteOwnerMetadata(ctx context.Context, actor Actor, owner OwnerRef) error

	// CanManageMetaobjects authorizes metaobject management.
	CanManageMetaobjects(ctx context.Context, actor Actor) error
}

// AllowAllPolicy allows every metadata operation.
type AllowAllPolicy struct{}

// CanManageDefinitions authorizes metafield definition management.
func (AllowAllPolicy) CanManageDefinitions(context.Context, Actor) error {
	return nil
}

// CanReadOwnerMetadata authorizes owner metadata reads.
func (AllowAllPolicy) CanReadOwnerMetadata(context.Context, Actor, OwnerRef) error {
	return nil
}

// CanWriteOwnerMetadata authorizes owner metadata writes.
func (AllowAllPolicy) CanWriteOwnerMetadata(context.Context, Actor, OwnerRef) error {
	return nil
}

// CanManageMetaobjects authorizes metaobject management.
func (AllowAllPolicy) CanManageMetaobjects(context.Context, Actor) error {
	return nil
}

// DefinitionPage is a paginated metafield definition result.
type DefinitionPage = pagination.Result[domain.MetafieldDefinition]

// MetaobjectDefinitionPage is a paginated metaobject definition result.
type MetaobjectDefinitionPage = pagination.Result[domain.MetaobjectDefinition]

// MetaobjectEntryPage is a paginated metaobject entry result.
type MetaobjectEntryPage = pagination.Result[domain.MetaobjectEntry]
