package application

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/niflaot/gamehub-go/module/metadata/domain"
	"github.com/niflaot/gamehub-go/module/metadata/port"
	"github.com/niflaot/gamehub-go/pkg/pagination"
)

// ListMetaobjectEntries returns metaobject entries.
func (service Service) ListMetaobjectEntries(
	ctx context.Context,
	query port.ListMetaobjectEntriesQuery,
) (pagination.Result[port.MetaobjectEntryView], error) {
	if err := service.ensureDependencies(); err != nil {
		return pagination.Result[port.MetaobjectEntryView]{}, err
	}
	if err := service.policy.CanManageMetaobjects(ctx, query.Actor); err != nil {
		return pagination.Result[port.MetaobjectEntryView]{}, err
	}
	return service.metaobjectEntries.List(ctx, query.DefinitionID, query.Page)
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
		return service.ensureOwnerReference(ctx, reference)
	case domain.ValueMetaobjectReference:
		var reference domain.MetaobjectReference
		if err := json.Unmarshal(raw, &reference); err != nil {
			return nil
		}
		return service.ensureMetaobjectReference(ctx, reference)
	default:
		return nil
	}
}

// ensureOwnerReference verifies one owner reference.
func (service Service) ensureOwnerReference(ctx context.Context, reference domain.OwnerReference) error {
	ok, err := service.references.OwnerExists(ctx, reference)
	if err != nil {
		return err
	}
	if !ok {
		return port.ErrNotFound
	}
	return nil
}

// ensureMetaobjectReference verifies one metaobject reference.
func (service Service) ensureMetaobjectReference(ctx context.Context, reference domain.MetaobjectReference) error {
	ok, err := service.references.MetaobjectEntryExists(ctx, reference)
	if err != nil {
		return err
	}
	if !ok {
		return port.ErrNotFound
	}
	return nil
}

// ownerMetadataView combines definitions and values.
func ownerMetadataView(
	owner port.OwnerRef,
	definitions []domain.MetafieldDefinition,
	values []domain.MetafieldValue,
	includeEmpty bool,
) port.OwnerMetadataView {
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
