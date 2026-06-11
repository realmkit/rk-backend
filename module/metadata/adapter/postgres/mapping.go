package postgres

import (
	"encoding/json"

	"github.com/realmkit/rk-backend/module/metadata/domain"
)

// definitionModelFromDomain maps domain definition to persistence.
func definitionModelFromDomain(definition domain.MetafieldDefinition) MetafieldDefinitionModel {
	return MetafieldDefinitionModel{
		OwnerType:   string(definition.OwnerType),
		Namespace:   string(definition.Namespace),
		Key:         string(definition.Key),
		Name:        definition.Name,
		Description: definition.Description,
		ValueType:   string(definition.ValueType),
		List:        definition.List,
		Required:    definition.Required,
		Rules:       marshalJSON(definition.Rules),
		SortOrder:   definition.SortOrder,
		Active:      definition.Active,
		Version:     definition.Version,
	}
}

// definitionFromModel maps persistence definition to domain.
func definitionFromModel(model MetafieldDefinitionModel) (domain.MetafieldDefinition, error) {
	var rules domain.Rules
	if err := json.Unmarshal(model.Rules, &rules); err != nil {
		return domain.MetafieldDefinition{}, err
	}
	return domain.MetafieldDefinition{
		ID:          model.ID.ID,
		OwnerType:   domain.OwnerType(model.OwnerType),
		Namespace:   domain.Namespace(model.Namespace),
		Key:         domain.Key(model.Key),
		Name:        model.Name,
		Description: model.Description,
		ValueType:   domain.ValueType(model.ValueType),
		List:        model.List,
		Required:    model.Required,
		Rules:       rules,
		SortOrder:   model.SortOrder,
		Active:      model.Active,
		Version:     model.Version,
		CreatedAt:   model.CreatedAt,
		UpdatedAt:   model.UpdatedAt,
	}, nil
}

// valueModelFromDomain maps domain value to persistence.
func valueModelFromDomain(value domain.MetafieldValue) MetafieldValueModel {
	return MetafieldValueModel{
		DefinitionID: value.DefinitionID,
		OwnerType:    string(value.OwnerType),
		OwnerID:      value.OwnerID,
		Value:        JSON(value.Value),
		Version:      value.Version,
	}
}

// valueFromModel maps persistence value to domain.
func valueFromModel(model MetafieldValueModel) domain.MetafieldValue {
	return domain.MetafieldValue{
		ID:           model.ID.ID,
		DefinitionID: model.DefinitionID,
		OwnerType:    domain.OwnerType(model.OwnerType),
		OwnerID:      model.OwnerID,
		Value:        append([]byte(nil), model.Value...),
		Version:      model.Version,
		CreatedAt:    model.CreatedAt,
		UpdatedAt:    model.UpdatedAt,
	}
}

// metaobjectDefinitionModelFromDomain maps domain definition to persistence.
func metaobjectDefinitionModelFromDomain(definition domain.MetaobjectDefinition) MetaobjectDefinitionModel {
	return MetaobjectDefinitionModel{
		Type:        string(definition.Type),
		Name:        definition.Name,
		Description: definition.Description,
		Fields:      marshalJSON(definition.Fields),
		Active:      definition.Active,
		Version:     definition.Version,
	}
}

// metaobjectDefinitionFromModel maps persistence definition to domain.
func metaobjectDefinitionFromModel(model MetaobjectDefinitionModel) (domain.MetaobjectDefinition, error) {
	var fields []domain.FieldDefinition
	if err := json.Unmarshal(model.Fields, &fields); err != nil {
		return domain.MetaobjectDefinition{}, err
	}
	return domain.MetaobjectDefinition{
		ID:          model.ID.ID,
		Type:        domain.MetaobjectType(model.Type),
		Name:        model.Name,
		Description: model.Description,
		Fields:      fields,
		Active:      model.Active,
		Version:     model.Version,
		CreatedAt:   model.CreatedAt,
		UpdatedAt:   model.UpdatedAt,
	}, nil
}

// metaobjectEntryModelFromDomain maps domain entry to persistence.
func metaobjectEntryModelFromDomain(entry domain.MetaobjectEntry) MetaobjectEntryModel {
	return MetaobjectEntryModel{
		DefinitionID: entry.DefinitionID,
		Handle:       string(entry.Handle),
		DisplayName:  entry.DisplayName,
		Fields:       marshalJSON(entry.Fields),
		Version:      entry.Version,
	}
}

// metaobjectEntryFromModel maps persistence entry to domain.
func metaobjectEntryFromModel(model MetaobjectEntryModel) (domain.MetaobjectEntry, error) {
	var fields map[domain.Key]json.RawMessage
	if err := json.Unmarshal(model.Fields, &fields); err != nil {
		return domain.MetaobjectEntry{}, err
	}
	return domain.MetaobjectEntry{
		ID:           model.ID.ID,
		DefinitionID: model.DefinitionID,
		Handle:       domain.Handle(model.Handle),
		DisplayName:  model.DisplayName,
		Fields:       fields,
		Version:      model.Version,
		CreatedAt:    model.CreatedAt,
		UpdatedAt:    model.UpdatedAt,
	}, nil
}
