// Package postgres stores punishments in PostgreSQL through GORM.
package postgres

import (
	"time"

	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/pkg/orm"
)

// DefinitionModel is the GORM model for punishment definitions.
type DefinitionModel struct {
	orm.ID                        // ID embeds shared fields.
	Key                    string `gorm:"size:64;not null;uniqueIndex"`  // Key stores the key value.
	Name                   string `gorm:"size:120;not null"`             // Name stores the name value.
	Description            string `gorm:"size:1000;not null;default:''"` // Description stores the description value.
	Color                  string `gorm:"size:16;not null"`              // Color stores the color value.
	Severity               int    `gorm:"not null;default:0;index"`      // Severity stores the severity value.
	Status                 string `gorm:"size:64;not null;index"`        // Status stores the status value.
	DefaultDurationSeconds *int64 // DefaultDurationSeconds stores the default duration seconds value.
	MinDurationSeconds     *int64 // MinDurationSeconds stores the min duration seconds value.
	MaxDurationSeconds     *int64 // MaxDurationSeconds stores the max duration seconds value.
	AllowPermanent         bool   `gorm:"not null;default:false"`   // AllowPermanent stores the allow permanent value.
	RequiresReason         bool   `gorm:"not null;default:true"`    // RequiresReason stores the requires reason value.
	RequiresTargetIP       bool   `gorm:"not null;default:false"`   // RequiresTargetIP stores the requires target i p value.
	DisplayOrder           int    `gorm:"not null;default:0;index"` // DisplayOrder stores the display order value.
	Version                uint64 `gorm:"not null;default:1"`       // Version stores the version value.
	orm.Timestamps                // Timestamps embeds shared fields.
	orm.SoftDelete                // SoftDelete embeds shared fields.
}

// TableName returns the table name.
func (DefinitionModel) TableName() string { return "punishment_definitions" }

// ActionModel is the GORM model for action templates.
type ActionModel struct {
	orm.ID                      // ID embeds shared fields.
	DefinitionID      uuid.UUID `gorm:"type:uuid;not null;index"` // DefinitionID stores the definition i d value.
	TargetSystem      string    `gorm:"size:64;not null;index"`   // TargetSystem stores the target system value.
	ActionType        string    `gorm:"size:160;not null;index"`  // ActionType stores the action type value.
	ConfigurationJSON string    `gorm:"type:jsonb;not null"`      // ConfigurationJSON stores the configuration j s o n value.
	DisplayOrder      int       `gorm:"not null;default:0;index"` // DisplayOrder stores the display order value.
	Status            string    `gorm:"size:64;not null;index"`   // Status stores the status value.
	orm.Timestamps              // Timestamps embeds shared fields.
	orm.SoftDelete              // SoftDelete embeds shared fields.
}

// TableName returns the table name.
func (ActionModel) TableName() string { return "punishment_definition_actions" }

// PunishmentModel is the GORM model for issued punishments.
type PunishmentModel struct {
	orm.ID                        // ID embeds shared fields.
	DefinitionID       uuid.UUID  `gorm:"type:uuid;not null;index"`           // DefinitionID stores the definition i d value.
	TargetUserID       uuid.UUID  `gorm:"type:uuid;not null;index"`           // TargetUserID stores the target user i d value.
	TargetIPHash       string     `gorm:"size:128;not null;default:'';index"` // TargetIPHash stores the target i p hash value.
	TargetIPCiphertext string     `gorm:"type:text;not null;default:''"`      // TargetIPCiphertext stores the target i p ciphertext value.
	IssuerType         string     `gorm:"size:64;not null;index"`             // IssuerType stores the issuer type value.
	IssuerUserID       *uuid.UUID `gorm:"type:uuid;index"`                    // IssuerUserID stores the issuer user i d value.
	IssuerKey          string     `gorm:"size:160;not null;default:'';index"` // IssuerKey stores the issuer key value.
	Reason             string     `gorm:"size:1000;not null"`                 // Reason stores the reason value.
	PrivateReason      string     `gorm:"size:2000;not null;default:''"`      // PrivateReason stores the private reason value.
	Status             string     `gorm:"size:64;not null;index"`             // Status stores the status value.
	StartsAt           time.Time  `gorm:"not null;index"`                     // StartsAt stores the starts at value.
	ExpiresAt          *time.Time `gorm:"index"`                              // ExpiresAt stores the expires at value.
	RevokedAt          *time.Time // RevokedAt stores the revoked at value.
	RevokedByUserID    *uuid.UUID `gorm:"type:uuid"`                     // RevokedByUserID stores the revoked by user i d value.
	RevocationReason   string     `gorm:"size:1000;not null;default:''"` // RevocationReason stores the revocation reason value.
	Source             string     `gorm:"size:160;not null;default:''"`  // Source stores the source value.
	IdempotencyKey     string     `gorm:"size:200;uniqueIndex"`          // IdempotencyKey stores the idempotency key value.
	Version            uint64     `gorm:"not null;default:1"`            // Version stores the version value.
	orm.Timestamps                // Timestamps embeds shared fields.
	orm.SoftDelete                // SoftDelete embeds shared fields.
}

// TableName returns the table name.
func (PunishmentModel) TableName() string { return "punishments" }

// SnapshotModel is the GORM model for action snapshots.
type SnapshotModel struct {
	orm.ID                       // ID embeds shared fields.
	PunishmentID       uuid.UUID `gorm:"type:uuid;not null;index"` // PunishmentID stores the punishment i d value.
	DefinitionActionID uuid.UUID `gorm:"type:uuid;not null;index"` // DefinitionActionID stores the definition action i d value.
	TargetSystem       string    `gorm:"size:64;not null;index"`   // TargetSystem stores the target system value.
	ActionType         string    `gorm:"size:160;not null;index"`  // ActionType stores the action type value.
	ConfigurationJSON  string    `gorm:"type:jsonb;not null"`      // ConfigurationJSON stores the configuration j s o n value.
	Status             string    `gorm:"size:64;not null;index"`   // Status stores the status value.
	CreatedAt          time.Time `gorm:"not null"`                 // CreatedAt stores the created at value.
	orm.SoftDelete               // SoftDelete embeds shared fields.
}

// TableName returns the table name.
func (SnapshotModel) TableName() string { return "punishment_action_snapshots" }

// RestrictionModel is the GORM model for active restrictions.
type RestrictionModel struct {
	orm.ID                    // ID embeds shared fields.
	PunishmentID   uuid.UUID  `gorm:"type:uuid;not null;index"` // PunishmentID stores the punishment i d value.
	TargetUserID   uuid.UUID  `gorm:"type:uuid;not null;index"` // TargetUserID stores the target user i d value.
	ActionKey      string     `gorm:"size:160;not null;index"`  // ActionKey stores the action key value.
	StartsAt       time.Time  `gorm:"not null;index"`           // StartsAt stores the starts at value.
	ExpiresAt      *time.Time `gorm:"index"`                    // ExpiresAt stores the expires at value.
	CreatedAt      time.Time  `gorm:"not null"`                 // CreatedAt stores the created at value.
	orm.SoftDelete            // SoftDelete embeds shared fields.
}

// TableName returns the table name.
func (RestrictionModel) TableName() string { return "punishment_active_restrictions" }
