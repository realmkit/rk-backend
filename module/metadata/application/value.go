package application

import (
	"context"

	"github.com/niflaot/gamehub-go/module/metadata/domain"
	"github.com/niflaot/gamehub-go/module/metadata/port"
)

// SetValue creates or updates an owner value.
func (service Service) SetValue(ctx context.Context, command port.SetValueCommand) (port.ValueView, bool, error) {
	if err := service.ensureDependencies(); err != nil {
		return port.ValueView{}, false, err
	}
	if err := service.policy.CanWriteOwnerMetadata(ctx, command.Actor, command.Owner); err != nil {
		return port.ValueView{}, false, err
	}
	if err := service.ensureOwner(ctx, command.Owner); err != nil {
		return port.ValueView{}, false, err
	}
	definition, err := service.definitions.FindByKey(ctx, command.Owner.Type, command.Namespace, command.Key)
	if err != nil {
		return port.ValueView{}, false, err
	}
	if !definition.Active {
		return port.ValueView{}, false, port.ErrInactive
	}
	if err := service.ensureReferences(ctx, definition.Field(), command.RawValue); err != nil {
		return port.ValueView{}, false, err
	}
	canonical, err := domain.NormalizeValue(definition.Field(), command.RawValue)
	if err != nil {
		return port.ValueView{}, false, err
	}
	value, created, err := service.values.Upsert(ctx, domain.MetafieldValue{
		DefinitionID: definition.ID,
		OwnerType:    command.Owner.Type,
		OwnerID:      command.Owner.ID,
		Value:        canonical,
	}, command.ExpectedVersion)
	if err != nil {
		return port.ValueView{}, false, err
	}
	key := "metadata.entry.updated"
	if created {
		key = "metadata.entry.created"
	}
	if err := service.publishMetadataEvent(
		ctx,
		"metadata.metafield.set",
		"metafield",
		value.ID,
		command.Actor,
		metadataValuePayload(value),
		metadataOwnerScopes(value.OwnerType, value.OwnerID),
	); err != nil {
		return port.ValueView{}, false, err
	}
	return value, created, service.publishMetadataEvent(
		ctx,
		eventKey(key),
		"metadata_entry",
		value.ID,
		command.Actor,
		metadataValuePayload(value),
		metadataOwnerScopes(value.OwnerType, value.OwnerID),
	)
}

// GetValue returns one owner value.
func (service Service) GetValue(ctx context.Context, query port.GetValueQuery) (port.ValueView, error) {
	if err := service.ensureDependencies(); err != nil {
		return port.ValueView{}, err
	}
	if err := service.policy.CanReadOwnerMetadata(ctx, query.Actor, query.Owner); err != nil {
		return port.ValueView{}, err
	}
	definition, err := service.definitions.FindByKey(ctx, query.Owner.Type, query.Namespace, query.Key)
	if err != nil {
		return port.ValueView{}, err
	}
	return service.values.Find(ctx, definition.ID, query.Owner.Type, query.Owner.ID)
}

// ListValuesForOwner returns owner metadata.
func (service Service) ListValuesForOwner(ctx context.Context, query port.ListValuesForOwnerQuery) (port.OwnerMetadataView, error) {
	if err := service.ensureDependencies(); err != nil {
		return port.OwnerMetadataView{}, err
	}
	if err := service.policy.CanReadOwnerMetadata(ctx, query.Actor, query.Owner); err != nil {
		return port.OwnerMetadataView{}, err
	}
	if err := service.ensureOwner(ctx, query.Owner); err != nil {
		return port.OwnerMetadataView{}, err
	}
	active := true
	definitions, err := service.definitions.List(ctx, port.DefinitionFilter{
		OwnerType: query.Owner.Type,
		Namespace: query.Namespace,
		Active:    &active,
	}, unlimitedPage())
	if err != nil {
		return port.OwnerMetadataView{}, err
	}
	values, err := service.values.ListForOwner(ctx, query.Owner.Type, query.Owner.ID)
	if err != nil {
		return port.OwnerMetadataView{}, err
	}
	return ownerMetadataView(query.Owner, definitions.Items, values, query.IncludeEmpty), nil
}

// DeleteValue deletes one owner value.
func (service Service) DeleteValue(ctx context.Context, command port.DeleteValueCommand) error {
	if err := service.ensureDependencies(); err != nil {
		return err
	}
	if err := service.policy.CanWriteOwnerMetadata(ctx, command.Actor, command.Owner); err != nil {
		return err
	}
	definition, err := service.definitions.FindByKey(ctx, command.Owner.Type, command.Namespace, command.Key)
	if err != nil {
		return err
	}
	value, err := service.values.Find(ctx, definition.ID, command.Owner.Type, command.Owner.ID)
	if err != nil {
		return err
	}
	if err := service.values.Delete(ctx, definition.ID, command.Owner.Type, command.Owner.ID, command.ExpectedVersion); err != nil {
		return err
	}
	if err := service.publishMetadataEvent(
		ctx,
		"metadata.metafield.deleted",
		"metafield",
		value.ID,
		command.Actor,
		metadataValuePayload(value),
		metadataOwnerScopes(value.OwnerType, value.OwnerID),
	); err != nil {
		return err
	}
	return service.publishMetadataEvent(
		ctx,
		"metadata.entry.deleted",
		"metadata_entry",
		value.ID,
		command.Actor,
		metadataValuePayload(value),
		metadataOwnerScopes(value.OwnerType, value.OwnerID),
	)
}
