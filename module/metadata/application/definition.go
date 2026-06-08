package application

import (
	"context"

	"github.com/niflaot/gamehub-go/module/metadata/port"
	"github.com/niflaot/gamehub-go/pkg/pagination"
)

// CreateDefinition creates a metafield definition.
func (service Service) CreateDefinition(ctx context.Context, command port.CreateDefinitionCommand) (port.DefinitionView, error) {
	if err := service.ensureDependencies(); err != nil {
		return port.DefinitionView{}, err
	}
	if err := service.policy.CanManageDefinitions(ctx, command.Actor); err != nil {
		return port.DefinitionView{}, err
	}
	if err := command.Definition.Validate(); err != nil {
		return port.DefinitionView{}, err
	}
	if command.Definition.Version == 0 {
		command.Definition.Version = 1
	}
	return service.definitions.Create(ctx, command.Definition)
}

// UpdateDefinition updates a metafield definition.
func (service Service) UpdateDefinition(ctx context.Context, command port.UpdateDefinitionCommand) (port.DefinitionView, error) {
	if err := service.ensureDependencies(); err != nil {
		return port.DefinitionView{}, err
	}
	if err := service.policy.CanManageDefinitions(ctx, command.Actor); err != nil {
		return port.DefinitionView{}, err
	}
	current, err := service.definitions.FindByID(ctx, command.Definition.ID)
	if err != nil {
		return port.DefinitionView{}, err
	}
	command.Definition.OwnerType = current.OwnerType
	command.Definition.Namespace = current.Namespace
	command.Definition.Key = current.Key
	command.Definition.ValueType = current.ValueType
	command.Definition.List = current.List
	if err := command.Definition.Validate(); err != nil {
		return port.DefinitionView{}, err
	}
	return service.definitions.Update(ctx, command.Definition, command.ExpectedVersion)
}

// ArchiveDefinition archives a metafield definition.
func (service Service) ArchiveDefinition(ctx context.Context, command port.ArchiveDefinitionCommand) error {
	if err := service.ensureDependencies(); err != nil {
		return err
	}
	if err := service.policy.CanManageDefinitions(ctx, command.Actor); err != nil {
		return err
	}
	count, err := service.values.CountByDefinition(ctx, command.ID)
	if err != nil {
		return err
	}
	if count > 0 {
		return port.ErrReferenced
	}
	return service.definitions.Archive(ctx, command.ID, command.ExpectedVersion)
}

// GetDefinition returns a metafield definition.
func (service Service) GetDefinition(ctx context.Context, query port.GetDefinitionQuery) (port.DefinitionView, error) {
	if err := service.ensureDependencies(); err != nil {
		return port.DefinitionView{}, err
	}
	if err := service.policy.CanManageDefinitions(ctx, query.Actor); err != nil {
		return port.DefinitionView{}, err
	}
	return service.definitions.FindByID(ctx, query.ID)
}

// ListDefinitions returns metafield definitions.
func (service Service) ListDefinitions(ctx context.Context, query port.ListDefinitionsQuery) (pagination.Result[port.DefinitionView], error) {
	if err := service.ensureDependencies(); err != nil {
		return pagination.Result[port.DefinitionView]{}, err
	}
	if err := service.policy.CanManageDefinitions(ctx, query.Actor); err != nil {
		return pagination.Result[port.DefinitionView]{}, err
	}
	return service.definitions.List(ctx, query.Filter, query.Page)
}
