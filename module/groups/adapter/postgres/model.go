// Package postgres contains GORM repositories for groups and permissions.
package postgres

import (
	"time"

	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/pkg/orm"
)

// GroupModel is the GORM model for groups.
type GroupModel struct {
	orm.ID
	Key         string     `gorm:"size:64;not null;index"`
	Name        string     `gorm:"size:120;not null"`
	Description string     `gorm:"size:500;not null;default:''"`
	Color       string     `gorm:"size:7;not null"`
	Weight      int        `gorm:"not null;default:0;index"`
	Status      string     `gorm:"size:32;not null;index"`
	IconAssetID *uuid.UUID `gorm:"type:uuid;index"`
	Version     uint64     `gorm:"not null;default:1"`
	orm.Timestamps
	orm.SoftDelete
}

// TableName returns the database table name.
func (GroupModel) TableName() string {
	return "groups"
}

// MembershipModel is the GORM model for group memberships.
type MembershipModel struct {
	orm.ID
	GroupID          uuid.UUID  `gorm:"type:uuid;not null;index"`
	UserID           uuid.UUID  `gorm:"type:uuid;not null;index"`
	Status           string     `gorm:"size:32;not null;index"`
	AssignedByUserID *uuid.UUID `gorm:"type:uuid;index"`
	AssignedReason   string     `gorm:"size:500;not null;default:''"`
	StartsAt         *time.Time `gorm:"index"`
	ExpiresAt        *time.Time `gorm:"index"`
	Version          uint64     `gorm:"not null;default:1"`
	orm.Timestamps
	orm.SoftDelete
}

// TableName returns the database table name.
func (MembershipModel) TableName() string {
	return "group_memberships"
}

// PermissionActionModel is the GORM model for permission actions.
type PermissionActionModel struct {
	orm.ID
	Action       string `gorm:"size:120;not null;index"`
	Area         string `gorm:"size:64;not null;index"`
	ScopeType    string `gorm:"size:64;not null;index"`
	Label        string `gorm:"size:120;not null"`
	Description  string `gorm:"size:500;not null;default:''"`
	WarningLevel string `gorm:"size:32;not null;default:'normal';index"`
	Enabled      bool   `gorm:"not null;default:true;index"`
	Version      uint64 `gorm:"not null;default:1"`
	orm.Timestamps
	orm.SoftDelete
}

// TableName returns the database table name.
func (PermissionActionModel) TableName() string {
	return "permission_actions"
}

// PermissionGrantModel is the GORM model for permission grants.
type PermissionGrantModel struct {
	orm.ID
	SubjectType     string     `gorm:"size:64;not null;index"`
	SubjectID       uuid.UUID  `gorm:"type:uuid;not null;index"`
	Action          string     `gorm:"size:120;not null;index"`
	ScopeType       string     `gorm:"size:64;not null;index"`
	ScopeID         uuid.UUID  `gorm:"type:uuid;not null;index"`
	Inherit         bool       `gorm:"not null;default:false;index"`
	ConditionKey    string     `gorm:"size:120;not null;default:'';index"`
	CreatedByUserID *uuid.UUID `gorm:"type:uuid;index"`
	CreatedAt       time.Time
	orm.SoftDelete
}

// TableName returns the database table name.
func (PermissionGrantModel) TableName() string {
	return "permission_grants"
}
