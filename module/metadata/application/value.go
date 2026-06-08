package application

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
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
	return service.values.Upsert(ctx, domain.MetafieldValue{
		DefinitionID: definition.ID,
		OwnerType:    command.Owner.Type,
		OwnerID:      command.Owner.ID,
		Value:        canonical,
	}, command.ExpectedVersion)
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
	return service.values.Delete(ctx, definition.ID, command.Owner.Type, command.Owner.ID, command.ExpectedVersion)
}

// ensureOwner verifies owner syntax and existence.
func (service Service) ensureOwner(ctx context.Context, owner port.OwnerRef) error {
	if violations := domain.ValidateOwnerType("owner_type", owner.Type); len(violations) > 0 {
		return domain.NewValidationError(violations)
	}
	if owner.ID == uuid.Nil {
		return domain.NewValidationError([]domain.Violation{{Field: "owner_id", Message: "is required"}})
	}
	ok, err := service.owners.Exists(ctx, owner.Type, owner.ID)
	if err != nil {
		return err
	}
	if !ok {
		return port.ErrNotFound
	}
	return nil
}

// ensureReferences verifies reference target existence.
func (service Service) ensureReferences(ctx context.Context, field domain.FieldDefinition, raw json.RawMessage) error {
	if field.List {
		return service.ensureReferenceList(ctx, field, raw)
	}
	return service.ensureReferenceValue(ctx, field, raw)
}

// ensureReferenceList verifies reference list target existence.
func (service Service) ensureReferenceList(ctx context.Context, field domain.FieldDefinition, raw json.RawMessage) error {
	if field.ValueType != domain.ValueOwnerReference && field.ValueType != domain.ValueMetaobjectReference {
		return nil
	}
	var items []json.RawMessage
	if err := json.Unmarshal(raw, &items); err != nil {
		return nil
	}
	for _, item := range items {
		if err := service.ensureReferenceValue(ctx, field, item); err != nil {
			return err
		}
	}
	return nil
}

// ensureReferenceValue verifies one reference target exists.
func (service Service) ensureReferenceValue(ctx context.Context, field domain.FieldDefinition, raw json.RawMessage) error {
	switch field.ValueType {
	case domain.ValueOwnerReference:
		var reference domain.OwnerReference
		if err := json.Unmarshal(raw, &reference); err != nil {
			return nil
		}
		ok, err := service.references.OwnerExists(ctx, reference)
		if err != nil {
			return err
		}
		if !ok {
			return port.ErrNotFound
		}
	case domain.ValueMetaobjectReference:
		var reference domain.MetaobjectReference
		if err := json.Unmarshal(raw, &reference); err != nil {
			return nil
		}
		ok, err := service.references.MetaobjectEntryExists(ctx, reference)
		if err != nil {
			return err
		}
		if !ok {
			return port.ErrNotFound
		}
	}
	return nil
}

// ownerMetadataView combines definitions and values.
func ownerMetadataView(owner port.OwnerRef, definitions []domain.MetafieldDefinition, values []domain.MetafieldValue, includeEmpty bool) port.OwnerMetadataView {
	byDefinition := make(map[uuid.UUID]domain.MetafieldValue, len(values))
	for _, value := range values {
		byDefinition[value.DefinitionID] = value
	}
	items := make([]port.OwnerMetafieldView, 0, len(definitions))
	for _, definition := range definitions {
		value, ok := byDefinition[definition.ID]
		if !ok && !includeEmpty {
			continue
		}
		item := port.OwnerMetafieldView{Definition: definition}
		if ok {
			item.Value = &value
		}
		items = append(items, item)
	}
	return port.OwnerMetadataView{Owner: owner, Metafields: items}
}
