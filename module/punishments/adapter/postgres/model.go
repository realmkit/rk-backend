// Package postgres stores punishments in PostgreSQL through GORM.
package postgres

import (
	"time"

	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/pkg/orm"
)

// DefinitionModel is the GORM model for punishment definitions.
type DefinitionModel struct {
	orm.ID
	Key                    string `gorm:"size:64;not null;uniqueIndex"`
	Name                   string `gorm:"size:120;not null"`
	Description            string `gorm:"size:1000;not null;default:''"`
	Color                  string `gorm:"size:16;not null"`
	Severity               int    `gorm:"not null;default:0;index"`
	Status                 string `gorm:"size:64;not null;index"`
	DefaultDurationSeconds *int64
	MinDurationSeconds     *int64
	MaxDurationSeconds     *int64
	AllowPermanent         bool   `gorm:"not null;default:false"`
	RequiresReason         bool   `gorm:"not null;default:true"`
	RequiresTargetIP       bool   `gorm:"not null;default:false"`
	DisplayOrder           int    `gorm:"not null;default:0;index"`
	Version                uint64 `gorm:"not null;default:1"`
	orm.Timestamps
	orm.SoftDelete
}

// TableName returns the table name.
func (DefinitionModel) TableName() string { return "punishment_definitions" }

// ActionModel is the GORM model for action templates.
type ActionModel struct {
	orm.ID
	DefinitionID      uuid.UUID `gorm:"type:uuid;not null;index"`
	TargetSystem      string    `gorm:"size:64;not null;index"`
	ActionKey         string    `gorm:"size:160;not null;index"`
	Effect            string    `gorm:"size:64;not null;index"`
	ConfigurationJSON string    `gorm:"type:jsonb;not null"`
	DisplayOrder      int       `gorm:"not null;default:0;index"`
	Status            string    `gorm:"size:64;not null;index"`
	orm.Timestamps
	orm.SoftDelete
}

// TableName returns the table name.
func (ActionModel) TableName() string { return "punishment_definition_actions" }

// PunishmentModel is the GORM model for issued punishments.
type PunishmentModel struct {
	orm.ID
	DefinitionID       uuid.UUID  `gorm:"type:uuid;not null;index"`
	TargetUserID       uuid.UUID  `gorm:"type:uuid;not null;index"`
	TargetIPHash       string     `gorm:"size:128;not null;default:'';index"`
	TargetIPCiphertext string     `gorm:"type:text;not null;default:''"`
	IssuerType         string     `gorm:"size:64;not null;index"`
	IssuerUserID       *uuid.UUID `gorm:"type:uuid;index"`
	IssuerKey          string     `gorm:"size:160;not null;default:'';index"`
	Reason             string     `gorm:"size:1000;not null"`
	PrivateReason      string     `gorm:"size:2000;not null;default:''"`
	Status             string     `gorm:"size:64;not null;index"`
	StartsAt           time.Time  `gorm:"not null;index"`
	ExpiresAt          *time.Time `gorm:"index"`
	RevokedAt          *time.Time
	RevokedByUserID    *uuid.UUID `gorm:"type:uuid"`
	RevocationReason   string     `gorm:"size:1000;not null;default:''"`
	Source             string     `gorm:"size:160;not null;default:''"`
	IdempotencyKey     string     `gorm:"size:200;uniqueIndex"`
	Version            uint64     `gorm:"not null;default:1"`
	orm.Timestamps
	orm.SoftDelete
}

// TableName returns the table name.
func (PunishmentModel) TableName() string { return "punishments" }

// SnapshotModel is the GORM model for action snapshots.
type SnapshotModel struct {
	orm.ID
	PunishmentID       uuid.UUID `gorm:"type:uuid;not null;index"`
	DefinitionActionID uuid.UUID `gorm:"type:uuid;not null;index"`
	TargetSystem       string    `gorm:"size:64;not null;index"`
	ActionKey          string    `gorm:"size:160;not null;index"`
	Effect             string    `gorm:"size:64;not null;index"`
	ConfigurationJSON  string    `gorm:"type:jsonb;not null"`
	Status             string    `gorm:"size:64;not null;index"`
	CreatedAt          time.Time `gorm:"not null"`
	orm.SoftDelete
}

// TableName returns the table name.
func (SnapshotModel) TableName() string { return "punishment_action_snapshots" }

// RestrictionModel is the GORM model for active restrictions.
type RestrictionModel struct {
	orm.ID
	PunishmentID uuid.UUID  `gorm:"type:uuid;not null;index"`
	TargetUserID uuid.UUID  `gorm:"type:uuid;not null;index"`
	ActionKey    string     `gorm:"size:160;not null;index"`
	StartsAt     time.Time  `gorm:"not null;index"`
	ExpiresAt    *time.Time `gorm:"index"`
	CreatedAt    time.Time  `gorm:"not null"`
	orm.SoftDelete
}

// TableName returns the table name.
func (RestrictionModel) TableName() string { return "punishment_active_restrictions" }
