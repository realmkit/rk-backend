package domain

import "github.com/google/uuid"

// Rules defines validation rules for one metadata value.
type Rules struct {
	// MinLength is the minimum accepted string length.
	MinLength *int `json:"min_length,omitempty"`

	// MaxLength is the maximum accepted string length.
	MaxLength *int `json:"max_length,omitempty"`

	// Pattern is an optional string regular expression.
	Pattern string `json:"pattern,omitempty"`

	// AllowedValues contains accepted enum values.
	AllowedValues []string `json:"allowed_values,omitempty"`

	// Min is the minimum accepted number or decimal string.
	Min *float64 `json:"min,omitempty"`

	// Max is the maximum accepted number or decimal string.
	Max *float64 `json:"max,omitempty"`

	// Precision is the accepted decimal precision.
	Precision *int `json:"precision,omitempty"`

	// Scale is the accepted decimal scale.
	Scale *int `json:"scale,omitempty"`

	// MinDate is the minimum accepted date or datetime.
	MinDate string `json:"min_date,omitempty"`

	// MaxDate is the maximum accepted date or datetime.
	MaxDate string `json:"max_date,omitempty"`

	// FutureOnly requires a date or datetime after now.
	FutureOnly bool `json:"future_only,omitempty"`

	// PastOnly requires a date or datetime before now.
	PastOnly bool `json:"past_only,omitempty"`

	// MaxBytes limits encoded JSON value size.
	MaxBytes *int `json:"max_bytes,omitempty"`

	// AllowedOwnerTypes restricts owner reference types.
	AllowedOwnerTypes []OwnerType `json:"allowed_owner_types,omitempty"`

	// MetaobjectDefinitionIDs restricts metaobject reference definitions.
	MetaobjectDefinitionIDs []uuid.UUID `json:"metaobject_definition_ids,omitempty"`

	// MinItems is the minimum accepted list item count.
	MinItems *int `json:"min_items,omitempty"`

	// MaxItems is the maximum accepted list item count.
	MaxItems *int `json:"max_items,omitempty"`

	// UniqueItems requires normalized list items to be unique.
	UniqueItems bool `json:"unique_items,omitempty"`
}

// FieldDefinition defines validation for a metafield or metaobject field.
type FieldDefinition struct {
	// Key is the stable machine key.
	Key Key `json:"key"`

	// Name is the human-readable field name.
	Name string `json:"name"`

	// Description is the optional human-readable description.
	Description string `json:"description,omitempty"`

	// ValueType is the value type accepted by the field.
	ValueType ValueType `json:"value_type"`

	// List reports whether the field accepts multiple values.
	List bool `json:"list"`

	// Required reports whether the field is required.
	Required bool `json:"required"`

	// Rules contains type-specific validation rules.
	Rules Rules `json:"rules"`
}

// Validate validates the field definition.
func (definition FieldDefinition) Validate(prefix string) []Violation {
	var violations []Violation
	violations = append(violations, ValidateKey(prefix+".key", definition.Key)...)
	violations = append(violations, validateName(prefix+".name", definition.Name)...)
	if definition.Description != "" && len(definition.Description) > 500 {
		violations = AppendViolation(violations, prefix+".description", "must be at most 500 characters")
	}
	if !ValidValueType(definition.ValueType) {
		violations = AppendViolation(violations, prefix+".value_type", "is not supported")
	}
	violations = append(violations, definition.Rules.Validate(prefix+".rules", definition.ValueType, definition.List)...)
	return violations
}

// Validate validates rules for valueType.
func (rules Rules) Validate(prefix string, valueType ValueType, list bool) []Violation {
	var violations []Violation
	violations = append(violations, rules.validateListRules(prefix, list)...)
	violations = append(violations, rules.validateStringRules(prefix, valueType)...)
	violations = append(violations, rules.validateNumberRules(prefix, valueType)...)
	violations = append(violations, rules.validateTemporalRules(prefix, valueType)...)
	violations = append(violations, rules.validateJSONRules(prefix, valueType)...)
	violations = append(violations, rules.validateReferenceRules(prefix, valueType)...)
	return violations
}

// validateName validates a human-readable name.
func validateName(field string, value string) []Violation {
	if len(value) < 2 {
		return []Violation{{Field: field, Message: "must be at least 2 characters"}}
	}
	if len(value) > 120 {
		return []Violation{{Field: field, Message: "must be at most 120 characters"}}
	}
	return nil
}

// validateListRules validates list rule consistency.
func (rules Rules) validateListRules(prefix string, list bool) []Violation {
	var violations []Violation
	if rules.MinItems != nil && *rules.MinItems < 0 {
		violations = AppendViolation(violations, prefix+".min_items", "cannot be negative")
	}
	if rules.MaxItems != nil && *rules.MaxItems < 0 {
		violations = AppendViolation(violations, prefix+".max_items", "cannot be negative")
	}
	if rules.MinItems != nil && rules.MaxItems != nil && *rules.MaxItems < *rules.MinItems {
		violations = AppendViolation(violations, prefix+".max_items", "cannot be less than min_items")
	}
	if !list && (rules.MinItems != nil || rules.MaxItems != nil || rules.UniqueItems) {
		violations = AppendViolation(violations, prefix, "list rules require list=true")
	}
	return violations
}

// validateStringRules validates string rule consistency.
func (rules Rules) validateStringRules(prefix string, valueType ValueType) []Violation {
	var violations []Violation
	if rules.MinLength != nil && *rules.MinLength < 0 {
		violations = AppendViolation(violations, prefix+".min_length", "cannot be negative")
	}
	if rules.MaxLength != nil && *rules.MaxLength < 0 {
		violations = AppendViolation(violations, prefix+".max_length", "cannot be negative")
	}
	if rules.MinLength != nil && rules.MaxLength != nil && *rules.MaxLength < *rules.MinLength {
		violations = AppendViolation(violations, prefix+".max_length", "cannot be less than min_length")
	}
	if valueType == ValueEnum && len(rules.AllowedValues) == 0 {
		violations = AppendViolation(violations, prefix+".allowed_values", "is required for enum values")
	}
	return violations
}

// validateNumberRules validates number rule consistency.
func (rules Rules) validateNumberRules(prefix string, valueType ValueType) []Violation {
	var violations []Violation
	if rules.Min != nil && rules.Max != nil && *rules.Min > *rules.Max {
		violations = AppendViolation(violations, prefix+".min", "cannot be greater than max")
	}
	if rules.Precision != nil && (*rules.Precision < 1 || *rules.Precision > 38) {
		violations = AppendViolation(violations, prefix+".precision", "must be between 1 and 38")
	}
	if rules.Scale != nil && *rules.Scale < 0 {
		violations = AppendViolation(violations, prefix+".scale", "cannot be negative")
	}
	if rules.Precision != nil && rules.Scale != nil && *rules.Scale > *rules.Precision {
		violations = AppendViolation(violations, prefix+".scale", "cannot be greater than precision")
	}
	if valueType != ValueDecimal && (rules.Precision != nil || rules.Scale != nil) {
		violations = AppendViolation(violations, prefix, "precision and scale require decimal values")
	}
	return violations
}

// validateTemporalRules validates date rule consistency.
func (rules Rules) validateTemporalRules(prefix string, _ ValueType) []Violation {
	if rules.FutureOnly && rules.PastOnly {
		return []Violation{{Field: prefix, Message: "future_only and past_only cannot both be true"}}
	}
	return nil
}

// validateJSONRules validates JSON rule consistency.
func (rules Rules) validateJSONRules(prefix string, valueType ValueType) []Violation {
	if rules.MaxBytes != nil && *rules.MaxBytes < 1 {
		return []Violation{{Field: prefix + ".max_bytes", Message: "must be positive"}}
	}
	if valueType != ValueJSON && rules.MaxBytes != nil {
		return []Violation{{Field: prefix + ".max_bytes", Message: "requires json value type"}}
	}
	return nil
}

// validateReferenceRules validates reference rule consistency.
func (rules Rules) validateReferenceRules(prefix string, valueType ValueType) []Violation {
	var violations []Violation
	if valueType != ValueOwnerReference && len(rules.AllowedOwnerTypes) > 0 {
		violations = AppendViolation(violations, prefix+".allowed_owner_types", "requires owner_reference value type")
	}
	for index, ownerType := range rules.AllowedOwnerTypes {
		if !ValidOwnerType(ownerType) {
			violations = AppendViolation(violations, prefix+".allowed_owner_types", "contains unsupported owner type at index "+itoa(index))
		}
	}
	if valueType != ValueMetaobjectReference && len(rules.MetaobjectDefinitionIDs) > 0 {
		violations = AppendViolation(violations, prefix+".metaobject_definition_ids", "requires metaobject_reference value type")
	}
	return violations
}

// itoa formats an integer.
func itoa(value int) string {
	return fmtInt(value)
}
