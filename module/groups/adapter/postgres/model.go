// Package postgres contains GORM repositories for groups and permissions.
package postgres

import (
	"time"

	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/pkg/orm"
)

// GroupModel is the GORM model for groups.
type GroupModel struct {
	orm.ID                    // ID embeds shared fields.
	Key            string     `gorm:"size:64;not null;index"`       // Key stores the key value.
	Name           string     `gorm:"size:120;not null"`            // Name stores the name value.
	Description    string     `gorm:"size:500;not null;default:''"` // Description stores the description value.
	Color          string     `gorm:"size:7;not null"`              // Color stores the color value.
	Weight         int        `gorm:"not null;default:0;index"`     // Weight stores the weight value.
	Status         string     `gorm:"size:32;not null;index"`       // Status stores the status value.
	IconAssetID    *uuid.UUID `gorm:"type:uuid;index"`              // IconAssetID stores the icon asset i d value.
	Version        uint64     `gorm:"not null;default:1"`           // Version stores the version value.
	orm.Timestamps            // Timestamps embeds shared fields.
	orm.SoftDelete            // SoftDelete embeds shared fields.
}

// TableName returns the database table name.
func (GroupModel) TableName() string {
	return "groups"
}

// MembershipModel is the GORM model for group memberships.
type MembershipModel struct {
	orm.ID                      // ID embeds shared fields.
	GroupID          uuid.UUID  `gorm:"type:uuid;not null;index"`     // GroupID stores the group i d value.
	UserID           uuid.UUID  `gorm:"type:uuid;not null;index"`     // UserID stores the user i d value.
	Status           string     `gorm:"size:32;not null;index"`       // Status stores the status value.
	AssignedByUserID *uuid.UUID `gorm:"type:uuid;index"`              // AssignedByUserID stores the assigned by user i d value.
	AssignedReason   string     `gorm:"size:500;not null;default:''"` // AssignedReason stores the assigned reason value.
	StartsAt         *time.Time `gorm:"index"`                        // StartsAt stores the starts at value.
	ExpiresAt        *time.Time `gorm:"index"`                        // ExpiresAt stores the expires at value.
	Version          uint64     `gorm:"not null;default:1"`           // Version stores the version value.
	orm.Timestamps              // Timestamps embeds shared fields.
	orm.SoftDelete              // SoftDelete embeds shared fields.
}

// TableName returns the database table name.
func (MembershipModel) TableName() string {
	return "group_memberships"
}

// PermissionGrantModel is the GORM model for permission grants.
type PermissionGrantModel struct {
	orm.ID                     // ID embeds shared fields.
	Action          string     `gorm:"size:120;not null;index"`            // Action stores the action value.
	ScopeType       string     `gorm:"size:64;not null;index"`             // ScopeType stores the scope type value.
	ScopeID         uuid.UUID  `gorm:"type:uuid;not null;index"`           // ScopeID stores the scope i d value.
	Inherit         bool       `gorm:"not null;default:false;index"`       // Inherit stores the inherit value.
	ConditionKey    string     `gorm:"size:120;not null;default:'';index"` // ConditionKey stores the condition key value.
	CreatedByUserID *uuid.UUID `gorm:"type:uuid;index"`                    // CreatedByUserID stores the created by user i d value.
	CreatedAt       time.Time  // CreatedAt stores the created at value.
	orm.SoftDelete             // SoftDelete embeds shared fields.
}

// TableName returns the database table name.
func (PermissionGrantModel) TableName() string {
	return "permission_grants"
}

// GroupPermissionGrantModel assigns a global grant to a group.
type GroupPermissionGrantModel struct {
	orm.ID                     // ID embeds shared fields.
	GroupID         uuid.UUID  `gorm:"type:uuid;not null;index"` // GroupID stores the group i d value.
	GrantID         uuid.UUID  `gorm:"type:uuid;not null;index"` // GrantID stores the grant i d value.
	CreatedByUserID *uuid.UUID `gorm:"type:uuid;index"`          // CreatedByUserID stores the created by user i d value.
	CreatedAt       time.Time  // CreatedAt stores the created at value.
	orm.SoftDelete             // SoftDelete embeds shared fields.
}

// TableName returns the database table name.
func (GroupPermissionGrantModel) TableName() string {
	return "group_permission_grants"
}
