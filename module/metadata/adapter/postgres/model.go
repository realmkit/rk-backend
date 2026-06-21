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
	orm.ID                // ID embeds shared fields.
	OwnerType      string `gorm:"size:64;not null;index"`      // OwnerType stores the owner type value.
	Key            string `gorm:"size:64;not null;index"`      // Key stores the key value.
	Name           string `gorm:"size:120;not null"`           // Name stores the name value.
	Description    string `gorm:"size:500"`                    // Description stores the description value.
	ValueType      string `gorm:"size:64;not null"`            // ValueType stores the value type value.
	List           bool   `gorm:"not null;column:is_list"`     // List stores the list value.
	Required       bool   `gorm:"not null;column:is_required"` // Required stores the required value.
	Rules          JSON   `gorm:"not null;type:jsonb"`         // Rules stores the rules value.
	SortOrder      int    `gorm:"not null;default:0"`          // SortOrder stores the sort order value.
	Active         bool   `gorm:"not null;default:true;index"` // Active stores the active value.
	Version        uint64 `gorm:"not null;default:1"`          // Version stores the version value.
	orm.Timestamps        // Timestamps embeds shared fields.
	orm.SoftDelete        // SoftDelete embeds shared fields.
}

// TableName returns the database table name.
func (MetafieldDefinitionModel) TableName() string {
	return "metadata_metafield_definitions"
}

// MetafieldValueModel is the GORM model for metafield values.
type MetafieldValueModel struct {
	orm.ID                   // ID embeds shared fields.
	DefinitionID   uuid.UUID `gorm:"type:uuid;not null;index"`              // DefinitionID stores the definition i d value.
	OwnerType      string    `gorm:"size:64;not null;index"`                // OwnerType stores the owner type value.
	OwnerID        uuid.UUID `gorm:"type:uuid;not null;index"`              // OwnerID stores the owner i d value.
	Value          JSON      `gorm:"not null;type:jsonb;column:value_json"` // Value stores the value value.
	Version        uint64    `gorm:"not null;default:1"`                    // Version stores the version value.
	orm.Timestamps           // Timestamps embeds shared fields.
	orm.SoftDelete           // SoftDelete embeds shared fields.
}

// TableName returns the database table name.
func (MetafieldValueModel) TableName() string {
	return "metadata_metafield_values"
}

// MetaobjectDefinitionModel is the GORM model for metaobject definitions.
type MetaobjectDefinitionModel struct {
	orm.ID                // ID embeds shared fields.
	Type           string `gorm:"size:64;not null;index"`                       // Type stores the type value.
	Name           string `gorm:"size:120;not null"`                            // Name stores the name value.
	Description    string `gorm:"size:500"`                                     // Description stores the description value.
	Fields         JSON   `gorm:"not null;type:jsonb;column:field_definitions"` // Fields stores the fields value.
	Active         bool   `gorm:"not null;default:true;index"`                  // Active stores the active value.
	Version        uint64 `gorm:"not null;default:1"`                           // Version stores the version value.
	orm.Timestamps        // Timestamps embeds shared fields.
	orm.SoftDelete        // SoftDelete embeds shared fields.
}

// TableName returns the database table name.
func (MetaobjectDefinitionModel) TableName() string {
	return "metadata_metaobject_definitions"
}

// MetaobjectEntryModel is the GORM model for metaobject entries.
type MetaobjectEntryModel struct {
	orm.ID                   // ID embeds shared fields.
	DefinitionID   uuid.UUID `gorm:"type:uuid;not null;index"`                // DefinitionID stores the definition i d value.
	Handle         string    `gorm:"size:120;not null;index"`                 // Handle stores the handle value.
	DisplayName    string    `gorm:"size:120;not null"`                       // DisplayName stores the display name value.
	Fields         JSON      `gorm:"not null;type:jsonb;column:field_values"` // Fields stores the fields value.
	Version        uint64    `gorm:"not null;default:1"`                      // Version stores the version value.
	orm.Timestamps           // Timestamps embeds shared fields.
	orm.SoftDelete           // SoftDelete embeds shared fields.
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
