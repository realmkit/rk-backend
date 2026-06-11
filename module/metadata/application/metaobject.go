package application

import (
	"context"

	"github.com/realmkit/rk-backend/module/metadata/domain"
	"github.com/realmkit/rk-backend/module/metadata/port"
	"github.com/realmkit/rk-backend/pkg/pagination"
)

// CreateMetaobjectDefinition creates a metaobject definition.
func (service Service) CreateMetaobjectDefinition(
	ctx context.Context,
	command port.CreateMetaobjectDefinitionCommand,
) (port.MetaobjectDefinitionView, error) {
	if err := service.ensureDependencies(); err != nil {
		return port.MetaobjectDefinitionView{}, err
	}
	if err := service.policy.CanManageMetaobjects(ctx, command.Actor); err != nil {
		return port.MetaobjectDefinitionView{}, err
	}
	if err := command.Definition.Validate(); err != nil {
		return port.MetaobjectDefinitionView{}, err
	}
	if command.Definition.Version == 0 {
		command.Definition.Version = 1
	}
	created, err := service.metaobjectDefinitions.Create(ctx, command.Definition)
	if err != nil {
		return port.MetaobjectDefinitionView{}, err
	}
	return created, service.publishMetadataEvent(
		ctx,
		"metadata.definition.created",
		"metadata_definition",
		created.ID,
		command.Actor,
		metaobjectDefinitionPayload(created),
		nil,
	)
}

// UpdateMetaobjectDefinition updates a metaobject definition.
func (service Service) UpdateMetaobjectDefinition(
	ctx context.Context,
	command port.UpdateMetaobjectDefinitionCommand,
) (port.MetaobjectDefinitionView, error) {
	if err := service.ensureDependencies(); err != nil {
		return port.MetaobjectDefinitionView{}, err
	}
	if err := service.policy.CanManageMetaobjects(ctx, command.Actor); err != nil {
		return port.MetaobjectDefinitionView{}, err
	}
	current, err := service.metaobjectDefinitions.FindByID(ctx, command.Definition.ID)
	if err != nil {
		return port.MetaobjectDefinitionView{}, err
	}
	command.Definition.Type = current.Type
	if err := command.Definition.Validate(); err != nil {
		return port.MetaobjectDefinitionView{}, err
	}
	if incompatibleMetaobjectChange(current, command.Definition) {
		return port.MetaobjectDefinitionView{}, port.ErrConflict
	}
	updated, err := service.metaobjectDefinitions.Update(ctx, command.Definition, command.ExpectedVersion)
	if err != nil {
		return port.MetaobjectDefinitionView{}, err
	}
	return updated, service.publishMetadataEvent(
		ctx,
		"metadata.definition.updated",
		"metadata_definition",
		updated.ID,
		command.Actor,
		metaobjectDefinitionPayload(updated),
		nil,
	)
}

// ArchiveMetaobjectDefinition archives a metaobject definition.
func (service Service) ArchiveMetaobjectDefinition(ctx context.Context, command port.ArchiveMetaobjectDefinitionCommand) error {
	if err := service.ensureDependencies(); err != nil {
		return err
	}
	if err := service.policy.CanManageMetaobjects(ctx, command.Actor); err != nil {
		return err
	}
	count, err := service.metaobjectEntries.CountByDefinition(ctx, command.ID)
	if err != nil {
		return err
	}
	if count > 0 {
		return port.ErrReferenced
	}
	definition, err := service.metaobjectDefinitions.FindByID(ctx, command.ID)
	if err != nil {
		return err
	}
	if err := service.metaobjectDefinitions.Archive(ctx, command.ID, command.ExpectedVersion); err != nil {
		return err
	}
	return service.publishMetadataEvent(
		ctx,
		"metadata.definition.deleted",
		"metadata_definition",
		definition.ID,
		command.Actor,
		metaobjectDefinitionPayload(definition),
		nil,
	)
}

// ListMetaobjectDefinitions returns metaobject definitions.
func (service Service) ListMetaobjectDefinitions(
	ctx context.Context,
	query port.ListMetaobjectDefinitionsQuery,
) (pagination.Result[port.MetaobjectDefinitionView], error) {
	if err := service.ensureDependencies(); err != nil {
		return pagination.Result[port.MetaobjectDefinitionView]{}, err
	}
	if err := service.policy.CanManageMetaobjects(ctx, query.Actor); err != nil {
		return pagination.Result[port.MetaobjectDefinitionView]{}, err
	}
	return service.metaobjectDefinitions.List(ctx, query.Filter, query.Page)
}

// GetMetaobjectDefinition returns one metaobject definition.
func (service Service) GetMetaobjectDefinition(
	ctx context.Context,
	query port.GetMetaobjectDefinitionQuery,
) (port.MetaobjectDefinitionView, error) {
	if err := service.ensureDependencies(); err != nil {
		return port.MetaobjectDefinitionView{}, err
	}
	if err := service.policy.CanManageMetaobjects(ctx, query.Actor); err != nil {
		return port.MetaobjectDefinitionView{}, err
	}
	return service.metaobjectDefinitions.FindByID(ctx, query.ID)
}

// CreateMetaobjectEntry creates a metaobject entry.
func (service Service) CreateMetaobjectEntry(
	ctx context.Context,
	command port.CreateMetaobjectEntryCommand,
) (port.MetaobjectEntryView, error) {
	if err := service.ensureDependencies(); err != nil {
		return port.MetaobjectEntryView{}, err
	}
	if err := service.policy.CanManageMetaobjects(ctx, command.Actor); err != nil {
		return port.MetaobjectEntryView{}, err
	}
	definition, err := service.metaobjectDefinitions.FindByID(ctx, command.Entry.DefinitionID)
	if err != nil {
		return port.MetaobjectEntryView{}, err
	}
	if !definition.Active {
		return port.MetaobjectEntryView{}, port.ErrInactive
	}
	if err := command.Entry.Validate(); err != nil {
		return port.MetaobjectEntryView{}, err
	}
	fields, err := domain.ValidateMetaobjectEntryFields(definition, command.RawFields)
	if err != nil {
		return port.MetaobjectEntryView{}, err
	}
	command.Entry.Fields = fields
	if command.Entry.Version == 0 {
		command.Entry.Version = 1
	}
	created, err := service.metaobjectEntries.Create(ctx, command.Entry)
	if err != nil {
		return port.MetaobjectEntryView{}, err
	}
	return created, service.publishMetadataEvent(
		ctx,
		"metadata.entry.created",
		"metadata_entry",
		created.ID,
		command.Actor,
		metaobjectEntryPayload(created),
		nil,
	)
}

// UpdateMetaobjectEntry updates a metaobject entry.
func (service Service) UpdateMetaobjectEntry(
	ctx context.Context,
	command port.UpdateMetaobjectEntryCommand,
) (port.MetaobjectEntryView, error) {
	if err := service.ensureDependencies(); err != nil {
		return port.MetaobjectEntryView{}, err
	}
	if err := service.policy.CanManageMetaobjects(ctx, command.Actor); err != nil {
		return port.MetaobjectEntryView{}, err
	}
	current, err := service.metaobjectEntries.FindByID(ctx, command.Entry.ID)
	if err != nil {
		return port.MetaobjectEntryView{}, err
	}
	definition, err := service.metaobjectDefinitions.FindByID(ctx, current.DefinitionID)
	if err != nil {
		return port.MetaobjectEntryView{}, err
	}
	if err := command.Entry.Validate(); err != nil {
		return port.MetaobjectEntryView{}, err
	}
	fields, err := domain.ValidateMetaobjectEntryFields(definition, command.RawFields)
	if err != nil {
		return port.MetaobjectEntryView{}, err
	}
	command.Entry.DefinitionID = current.DefinitionID
	command.Entry.Handle = current.Handle
	command.Entry.Fields = fields
	updated, err := service.metaobjectEntries.Update(ctx, command.Entry, command.ExpectedVersion)
	if err != nil {
		return port.MetaobjectEntryView{}, err
	}
	return updated, service.publishMetadataEvent(
		ctx,
		"metadata.entry.updated",
		"metadata_entry",
		updated.ID,
		command.Actor,
		metaobjectEntryPayload(updated),
		nil,
	)
}

// GetMetaobjectEntry returns one metaobject entry.
func (service Service) GetMetaobjectEntry(ctx context.Context, query port.GetMetaobjectEntryQuery) (port.MetaobjectEntryView, error) {
	if err := service.ensureDependencies(); err != nil {
		return port.MetaobjectEntryView{}, err
	}
	if err := service.policy.CanManageMetaobjects(ctx, query.Actor); err != nil {
		return port.MetaobjectEntryView{}, err
	}
	return service.metaobjectEntries.FindByID(ctx, query.ID)
}
