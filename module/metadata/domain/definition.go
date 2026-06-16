// Package domain contains metadata entities, value objects, and validation.
package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// MetafieldDefinition describes one custom field assignable to an owner type.
type MetafieldDefinition struct {
	// ID is the definition identifier.
	ID uuid.UUID `json:"id"`

	// OwnerType is the allowlisted owner type.
	OwnerType OwnerType `json:"owner_type"`

	// Key is the stable field key.
	Key Key `json:"key"`

	// Name is the human-readable name.
	Name string `json:"name"`

	// Description is the optional human-readable description.
	Description string `json:"description,omitempty"`

	// ValueType is the accepted value type.
	ValueType ValueType `json:"value_type"`

	// List reports whether this definition accepts multiple values.
	List bool `json:"list"`

	// Required reports whether this definition is required for completeness.
	Required bool `json:"required"`

	// Rules contains type-specific validation rules.
	Rules Rules `json:"rules"`

	// SortOrder is the admin display order.
	SortOrder int `json:"sort_order"`

	// Active reports whether writes are accepted.
	Active bool `json:"active"`

	// Version is the optimistic concurrency version.
	Version uint64 `json:"version"`

	// CreatedAt is the creation timestamp.
	CreatedAt time.Time `json:"created_at"`

	// UpdatedAt is the update timestamp.
	UpdatedAt time.Time `json:"updated_at"`
}

// Field returns this definition as a reusable field definition.
func (definition MetafieldDefinition) Field() FieldDefinition {
	return FieldDefinition{
		Key:         definition.Key,
		Name:        definition.Name,
		Description: definition.Description,
		ValueType:   definition.ValueType,
		List:        definition.List,
		Required:    definition.Required,
		Rules:       definition.Rules,
	}
}

// Validate validates the metafield definition.
func (definition MetafieldDefinition) Validate() error {
	var violations []Violation
	violations = append(violations, ValidateOwnerType("owner_type", definition.OwnerType)...)
	violations = append(violations, definition.Field().Validate("definition")...)
	return NewValidationError(violations)
}

// MetafieldValue stores one canonical owner value.
type MetafieldValue struct {
	// ID is the value identifier.
	ID uuid.UUID `json:"id"`

	// DefinitionID is the owning definition identifier.
	DefinitionID uuid.UUID `json:"definition_id"`

	// OwnerType is the owner type receiving the value.
	OwnerType OwnerType `json:"owner_type"`

	// OwnerID is the owner identifier.
	OwnerID uuid.UUID `json:"owner_id"`

	// Value is the canonical JSON value.
	Value json.RawMessage `json:"value"`

	// Version is the optimistic concurrency version.
	Version uint64 `json:"version"`

	// CreatedAt is the creation timestamp.
	CreatedAt time.Time `json:"created_at"`

	// UpdatedAt is the update timestamp.
	UpdatedAt time.Time `json:"updated_at"`
}

// MetaobjectDefinition describes a reusable custom object schema.
type MetaobjectDefinition struct {
	// ID is the definition identifier.
	ID uuid.UUID `json:"id"`

	// Type is the stable metaobject type.
	Type MetaobjectType `json:"type"`

	// Name is the human-readable name.
	Name string `json:"name"`

	// Description is the optional human-readable description.
	Description string `json:"description,omitempty"`

	// Fields contains the embedded field definitions.
	Fields []FieldDefinition `json:"fields"`

	// Active reports whether entries can be written.
	Active bool `json:"active"`

	// Version is the optimistic concurrency version.
	Version uint64 `json:"version"`

	// CreatedAt is the creation timestamp.
	CreatedAt time.Time `json:"created_at"`

	// UpdatedAt is the update timestamp.
	UpdatedAt time.Time `json:"updated_at"`
}

// Validate validates the metaobject definition.
func (definition MetaobjectDefinition) Validate() error {
	var violations []Violation
	seen := map[Key]struct{}{}
	violations = append(violations, ValidateMetaobjectType("type", definition.Type)...)
	violations = append(violations, validateName("name", definition.Name)...)
	if definition.Description != "" && len(definition.Description) > 500 {
		violations = AppendViolation(violations, "description", "must be at most 500 characters")
	}
	for index, field := range definition.Fields {
		prefix := "fields." + itoa(index)
		violations = append(violations, field.Validate(prefix)...)
		if _, ok := seen[field.Key]; ok {
			violations = AppendViolation(violations, prefix+".key", "must be unique")
		}
		seen[field.Key] = struct{}{}
	}
	return NewValidationError(violations)
}

// MetaobjectEntry stores one concrete metaobject entry.
type MetaobjectEntry struct {
	// ID is the entry identifier.
	ID uuid.UUID `json:"id"`

	// DefinitionID is the metaobject definition identifier.
	DefinitionID uuid.UUID `json:"definition_id"`

	// Handle is the stable public handle.
	Handle Handle `json:"handle"`

	// DisplayName is the human-readable display name.
	DisplayName string `json:"display_name"`

	// Fields contains canonical values keyed by field key.
	Fields map[Key]json.RawMessage `json:"fields"`

	// Version is the optimistic concurrency version.
	Version uint64 `json:"version"`

	// CreatedAt is the creation timestamp.
	CreatedAt time.Time `json:"created_at"`

	// UpdatedAt is the update timestamp.
	UpdatedAt time.Time `json:"updated_at"`
}

// Validate validates the metaobject entry identity fields.
func (entry MetaobjectEntry) Validate() error {
	var violations []Violation
	violations = append(violations, ValidateHandle("handle", entry.Handle)...)
	violations = append(violations, validateName("display_name", entry.DisplayName)...)
	return NewValidationError(violations)
}
