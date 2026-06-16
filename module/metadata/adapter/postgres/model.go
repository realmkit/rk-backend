// Package postgres contains GORM metadata repositories.
package postgres

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/pkg/orm"
)

// JSON stores JSON values in SQL.
type JSON []byte

// Value returns the driver representation.
func (value JSON) Value() (driver.Value, error) {
	if len(value) == 0 {
		return []byte("null"), nil
	}
	return []byte(value), nil
}

// Scan stores SQL JSON bytes.
func (value *JSON) Scan(input any) error {
	switch typed := input.(type) {
	case nil:
		*value = JSON("null")
	case []byte:
		*value = append((*value)[:0], typed...)
	case string:
		*value = append((*value)[:0], typed...)
	default:
		return fmt.Errorf("unsupported JSON value %T", input)
	}
	return nil
}

// MetafieldDefinitionModel is the GORM model for metafield definitions.
type MetafieldDefinitionModel struct {
	orm.ID
	OwnerType   string `gorm:"size:64;not null;index"`
	Key         string `gorm:"size:64;not null;index"`
	Name        string `gorm:"size:120;not null"`
	Description string `gorm:"size:500"`
	ValueType   string `gorm:"size:64;not null"`
	List        bool   `gorm:"not null;column:is_list"`
	Required    bool   `gorm:"not null;column:is_required"`
	Rules       JSON   `gorm:"not null;type:jsonb"`
	SortOrder   int    `gorm:"not null;default:0"`
	Active      bool   `gorm:"not null;default:true;index"`
	Version     uint64 `gorm:"not null;default:1"`
	orm.Timestamps
	orm.SoftDelete
}

// TableName returns the database table name.
func (MetafieldDefinitionModel) TableName() string {
	return "metadata_metafield_definitions"
}

// MetafieldValueModel is the GORM model for metafield values.
type MetafieldValueModel struct {
	orm.ID
	DefinitionID uuid.UUID `gorm:"type:uuid;not null;index"`
	OwnerType    string    `gorm:"size:64;not null;index"`
	OwnerID      uuid.UUID `gorm:"type:uuid;not null;index"`
	Value        JSON      `gorm:"not null;type:jsonb;column:value_json"`
	Version      uint64    `gorm:"not null;default:1"`
	orm.Timestamps
	orm.SoftDelete
}

// TableName returns the database table name.
func (MetafieldValueModel) TableName() string {
	return "metadata_metafield_values"
}

// MetaobjectDefinitionModel is the GORM model for metaobject definitions.
type MetaobjectDefinitionModel struct {
	orm.ID
	Type        string `gorm:"size:64;not null;index"`
	Name        string `gorm:"size:120;not null"`
	Description string `gorm:"size:500"`
	Fields      JSON   `gorm:"not null;type:jsonb;column:field_definitions"`
	Active      bool   `gorm:"not null;default:true;index"`
	Version     uint64 `gorm:"not null;default:1"`
	orm.Timestamps
	orm.SoftDelete
}

// TableName returns the database table name.
func (MetaobjectDefinitionModel) TableName() string {
	return "metadata_metaobject_definitions"
}

// MetaobjectEntryModel is the GORM model for metaobject entries.
type MetaobjectEntryModel struct {
	orm.ID
	DefinitionID uuid.UUID `gorm:"type:uuid;not null;index"`
	Handle       string    `gorm:"size:120;not null;index"`
	DisplayName  string    `gorm:"size:120;not null"`
	Fields       JSON      `gorm:"not null;type:jsonb;column:field_values"`
	Version      uint64    `gorm:"not null;default:1"`
	orm.Timestamps
	orm.SoftDelete
}

// TableName returns the database table name.
func (MetaobjectEntryModel) TableName() string {
	return "metadata_metaobject_entries"
}

// marshalJSON marshals value into JSON.
func marshalJSON(value any) JSON {
	data, _ := json.Marshal(value)
	return JSON(data)
}
