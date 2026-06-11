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

// RelationTupleModel is the GORM model for authorization relation tuples.
type RelationTupleModel struct {
	orm.ID
	ObjectType      string     `gorm:"size:64;not null;index"`
	ObjectID        uuid.UUID  `gorm:"type:uuid;not null;index"`
	Relation        string     `gorm:"size:64;not null;index"`
	SubjectType     string     `gorm:"size:64;not null;index"`
	SubjectID       uuid.UUID  `gorm:"type:uuid;not null;index"`
	SubjectRelation string     `gorm:"size:64;not null;default:'';index"`
	CreatedByUserID *uuid.UUID `gorm:"type:uuid;index"`
	CreatedAt       time.Time
	orm.SoftDelete
}

// TableName returns the database table name.
func (RelationTupleModel) TableName() string {
	return "authorization_relation_tuples"
}

// PermissionDefinitionModel is the GORM model for permission definitions.
type PermissionDefinitionModel struct {
	orm.ID
	Permission  string `gorm:"size:120;not null;index"`
	ObjectType  string `gorm:"size:64;not null;index"`
	Description string `gorm:"size:500;not null;default:''"`
	Enabled     bool   `gorm:"not null;default:true;index"`
	Version     uint64 `gorm:"not null;default:1"`
	orm.Timestamps
	orm.SoftDelete
}

// TableName returns the database table name.
func (PermissionDefinitionModel) TableName() string {
	return "authorization_permission_definitions"
}

// PermissionRuleModel is the GORM model for permission policy rules.
type PermissionRuleModel struct {
	orm.ID
	Permission     string `gorm:"size:120;not null;index"`
	ObjectType     string `gorm:"size:64;not null;index"`
	Relation       string `gorm:"size:64;not null;index"`
	ConditionsJSON string `gorm:"type:text;not null;default:'[]'"`
	Priority       int    `gorm:"not null;default:0;index"`
	Enabled        bool   `gorm:"not null;default:true;index"`
	orm.Timestamps
	orm.SoftDelete
}

// TableName returns the database table name.
func (PermissionRuleModel) TableName() string {
	return "authorization_policy_rules"
}
