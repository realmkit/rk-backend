package application

import (
	"context"

	"github.com/realmkit/rk-backend/module/metadata/domain"
	"github.com/realmkit/rk-backend/module/metadata/port"
	"github.com/realmkit/rk-backend/pkg/pagination"
)

// unlimitedPage returns a broad page for internal owner metadata composition.
func unlimitedPage() pagination.Page {
	return pagination.Page{Limit: pagination.MaxLimit}
}

// DeleteMetaobjectEntry deletes one metaobject entry.
func (service Service) DeleteMetaobjectEntry(ctx context.Context, command port.DeleteMetaobjectEntryCommand) error {
	if err := service.ensureDependencies(); err != nil {
		return err
	}
	if err := service.policy.CanManageMetaobjects(ctx, command.Actor); err != nil {
		return err
	}
	entry, err := service.metaobjectEntries.FindByID(ctx, command.ID)
	if err != nil {
		return err
	}
	if err := service.metaobjectEntries.Delete(ctx, command.ID, command.ExpectedVersion); err != nil {
		return err
	}
	return service.publishMetadataEvent(
		ctx,
		"metadata.entry.deleted",
		"metadata_entry",
		entry.ID,
		command.Actor,
		metaobjectEntryPayload(entry),
		nil,
	)
}

// incompatibleMetaobjectChange reports whether update changes existing field contracts.
func incompatibleMetaobjectChange(current domain.MetaobjectDefinition, next domain.MetaobjectDefinition) bool {
	fields := make(map[domain.Key]domain.FieldDefinition, len(current.Fields))
	for _, field := range current.Fields {
		fields[field.Key] = field
	}
	for _, field := range next.Fields {
		currentField, ok := fields[field.Key]
		if !ok {
			continue
		}
		if currentField.ValueType != field.ValueType || currentField.List != field.List {
			return true
		}
	}
	return false
}

// metaobjectDefinitionPayload returns a safe metaobject definition payload.
func metaobjectDefinitionPayload(definition domain.MetaobjectDefinition) map[string]any {
	return map[string]any{
		"id":      definition.ID,
		"type":    definition.Type,
		"active":  definition.Active,
		"version": definition.Version,
	}
}

// metaobjectEntryPayload returns a safe metaobject entry payload.
func metaobjectEntryPayload(entry domain.MetaobjectEntry) map[string]any {
	return map[string]any{
		"id":            entry.ID,
		"definition_id": entry.DefinitionID,
		"handle":        entry.Handle,
		"display_name":  entry.DisplayName,
		"version":       entry.Version,
	}
}
