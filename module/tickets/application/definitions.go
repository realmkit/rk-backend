package application

import (
	"context"

	"github.com/google/uuid"
	"github.com/niflaot/gamehub-go/module/tickets/domain"
	"github.com/niflaot/gamehub-go/module/tickets/port"
	"github.com/niflaot/gamehub-go/pkg/pagination"
)

// CreateDefinition stores a ticket definition.
func (service Service) CreateDefinition(ctx context.Context, definition domain.Definition) (domain.Definition, error) {
	definition = definition.Normalize()
	if err := definition.Validate(); err != nil {
		return domain.Definition{}, err
	}
	created, err := service.definitions.Create(ctx, definition)
	if err != nil {
		return domain.Definition{}, err
	}
	return created, service.publishDefinition(ctx, "tickets.definition.created", created)
}

// UpdateDefinition updates a ticket definition.
func (service Service) UpdateDefinition(ctx context.Context, definition domain.Definition, expectedVersion uint64) (domain.Definition, error) {
	definition = definition.Normalize()
	if err := definition.Validate(); err != nil {
		return domain.Definition{}, err
	}
	updated, err := service.definitions.Update(ctx, definition, expectedVersion)
	if err != nil {
		return domain.Definition{}, err
	}
	return updated, service.publishDefinition(ctx, "tickets.definition.updated", updated)
}

// DeleteDefinition soft deletes a ticket definition.
func (service Service) DeleteDefinition(ctx context.Context, id uuid.UUID, expectedVersion uint64) error {
	current, err := service.definitions.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if err := service.definitions.Delete(ctx, id, expectedVersion); err != nil {
		return err
	}
	return service.publishDefinition(ctx, "tickets.definition.deleted", current)
}

// GetDefinition returns one ticket definition.
func (service Service) GetDefinition(ctx context.Context, id uuid.UUID) (domain.Definition, error) {
	return service.definitions.FindByID(ctx, id)
}

// ListDefinitions returns ticket definitions.
func (service Service) ListDefinitions(ctx context.Context, filter port.DefinitionFilter, page pagination.Page) (pagination.Result[domain.Definition], error) {
	return service.definitions.List(ctx, filter, page)
}
