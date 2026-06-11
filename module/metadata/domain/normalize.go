package domain

import (
	"bytes"
	"encoding/json"
	"strconv"

	"github.com/google/uuid"
)

// OwnerReference points to another supported GameHub owner.
type OwnerReference struct {
	// Type is the referenced owner type.
	Type OwnerType `json:"type"`

	// ID is the referenced owner identifier.
	ID uuid.UUID `json:"id"`
}

// MetaobjectReference points to a metaobject entry.
type MetaobjectReference struct {
	// DefinitionID is the referenced metaobject definition.
	DefinitionID uuid.UUID `json:"definition_id"`

	// EntryID is the referenced metaobject entry.
	EntryID uuid.UUID `json:"entry_id"`
}

// NormalizeValue validates raw and returns canonical JSON for field.
func NormalizeValue(field FieldDefinition, raw json.RawMessage) (json.RawMessage, error) {
	if len(bytes.TrimSpace(raw)) == 0 {
		raw = json.RawMessage("null")
	}
	if isNull(raw) {
		if field.Required {
			return nil, NewValidationError([]Violation{{Field: "value", Message: "is required"}})
		}
		return json.RawMessage(`{"value":null}`), nil
	}
	if field.List {
		return normalizeList(field, raw)
	}
	item, err := normalizeItem(field, raw)
	if err != nil {
		return nil, err
	}
	if field.ValueType == ValueOwnerReference || field.ValueType == ValueMetaobjectReference {
		return item, nil
	}
	return marshalCanonical(map[string]json.RawMessage{"value": item})
}

// ValidateMetaobjectEntryFields validates and normalizes metaobject entry fields.
func ValidateMetaobjectEntryFields(definition MetaobjectDefinition, raw map[Key]json.RawMessage) (map[Key]json.RawMessage, error) {
	normalized := make(map[Key]json.RawMessage, len(definition.Fields))
	var violations []Violation
	fields := map[Key]FieldDefinition{}
	for _, field := range definition.Fields {
		fields[field.Key] = field
		value, ok := raw[field.Key]
		if !ok && field.Required {
			violations = AppendViolation(violations, "fields."+string(field.Key), "is required")
			continue
		}
		if !ok {
			continue
		}
		canonical, err := NormalizeValue(field, value)
		if validation, ok := err.(ValidationError); ok {
			for _, violation := range validation.Violations {
				violations = AppendViolation(violations, "fields."+string(field.Key)+"."+violation.Field, violation.Message)
			}
			continue
		}
		if err != nil {
			return nil, err
		}
		normalized[field.Key] = canonical
	}
	for key := range raw {
		if _, ok := fields[key]; !ok {
			violations = AppendViolation(violations, "fields."+string(key), "is not defined")
		}
	}
	return normalized, NewValidationError(violations)
}

// fmtInt formats an integer for validation paths.
func fmtInt(value int) string {
	return strconv.Itoa(value)
}

// isNull reports whether raw is JSON null.
func isNull(raw json.RawMessage) bool {
	return bytes.Equal(bytes.TrimSpace(raw), []byte("null"))
}

// normalizeList validates and normalizes list values.
func normalizeList(field FieldDefinition, raw json.RawMessage) (json.RawMessage, error) {
	var items []json.RawMessage
	if err := json.Unmarshal(raw, &items); err != nil {
		return nil, NewValidationError([]Violation{{Field: "value", Message: "must be an array"}})
	}
	var violations []Violation
	if field.Rules.MinItems != nil && len(items) < *field.Rules.MinItems {
		violations = AppendViolation(violations, "value", "has fewer items than min_items")
	}
	if field.Rules.MaxItems != nil && len(items) > *field.Rules.MaxItems {
		violations = AppendViolation(violations, "value", "has more items than max_items")
	}
	normalized := make([]json.RawMessage, 0, len(items))
	seen := map[string]struct{}{}
	for index, item := range items {
		canonical, err := normalizeItem(field, item)
		if validation, ok := err.(ValidationError); ok {
			for _, violation := range validation.Violations {
				violations = AppendViolation(violations, "value."+itoa(index)+"."+violation.Field, violation.Message)
			}
			continue
		}
		if err != nil {
			return nil, err
		}
		key := string(canonical)
		if field.Rules.UniqueItems {
			if _, ok := seen[key]; ok {
				violations = AppendViolation(violations, "value."+itoa(index), "must be unique")
			}
			seen[key] = struct{}{}
		}
		normalized = append(normalized, canonical)
	}
	if err := NewValidationError(violations); err != nil {
		return nil, err
	}
	return marshalCanonical(map[string][]json.RawMessage{"value": normalized})
}

// normalizeItem validates and normalizes one scalar item.
func normalizeItem(field FieldDefinition, raw json.RawMessage) (json.RawMessage, error) {
	switch field.ValueType {
	case ValueSingleLineText, ValueMultiLineText, ValueURL, ValueColor, ValueEnum:
		return normalizeString(field, raw)
	case ValueInteger:
		return normalizeInteger(field, raw)
	case ValueDecimal:
		return normalizeDecimal(field, raw)
	case ValueBoolean:
		return normalizeBoolean(raw)
	case ValueDate:
		return normalizeDate(field, raw)
	case ValueDatetime:
		return normalizeDatetime(field, raw)
	case ValueJSON:
		return normalizeJSON(field, raw)
	case ValueOwnerReference:
		return normalizeOwnerReference(field, raw)
	case ValueMetaobjectReference:
		return normalizeMetaobjectReference(field, raw)
	default:
		return nil, NewValidationError([]Violation{{Field: "value_type", Message: "is not supported"}})
	}
}
