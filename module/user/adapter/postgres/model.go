package postgres

import (
	"time"

	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/pkg/orm"
)

// UserModel is the GORM model for local users.
type UserModel struct {
	orm.ID
	Status        string     `gorm:"size:64;not null;index"`
	AvatarAssetID *uuid.UUID `gorm:"type:uuid;index"`
	FirstSeenAt   time.Time  `gorm:"not null"`
	LastSeenAt    *time.Time `gorm:"index"`
	Version       uint64     `gorm:"not null;default:1"`
	orm.Timestamps
	orm.SoftDelete
}

// TableName returns the database table name.
func (UserModel) TableName() string {
	return "users"
}

// IdentityLinkModel is the GORM model for provider identity links.
type IdentityLinkModel struct {
	orm.ID
	UserID       uuid.UUID  `gorm:"type:uuid;not null;index"`
	Provider     string     `gorm:"size:64;not null;index"`
	Issuer       string     `gorm:"size:512;not null;index"`
	Subject      string     `gorm:"size:512;not null;index"`
	SubjectHash  string     `gorm:"size:64;not null;index"`
	ClaimsHash   string     `gorm:"size:64;not null;default:''"`
	LinkedAt     time.Time  `gorm:"not null"`
	LastSeenAt   *time.Time `gorm:"index"`
	LastSyncedAt *time.Time `gorm:"index"`
	orm.Timestamps
	orm.SoftDelete
}

// TableName returns the database table name.
func (IdentityLinkModel) TableName() string {
	return "user_identity_links"
}

// ClaimCacheModel is the GORM model for provider claim cache rows.
type ClaimCacheModel struct {
	orm.ID
	UserID          uuid.UUID `gorm:"type:uuid;not null;index"`
	Issuer          string    `gorm:"size:512;not null;index"`
	Subject         string    `gorm:"size:512;not null;index"`
	Username        string    `gorm:"size:256;not null;default:''"`
	Email           string    `gorm:"size:320;not null;default:''"`
	EmailVerified   bool      `gorm:"not null;default:false"`
	DisplayName     string    `gorm:"size:256;not null;default:''"`
	PictureURL      string    `gorm:"size:1024;not null;default:''"`
	PreferredLocale string    `gorm:"size:64;not null;default:''"`
	ClaimsHash      string    `gorm:"size:64;not null;default:'';index"`
	SyncedAt        time.Time `gorm:"not null;index"`
	orm.Timestamps
	orm.SoftDelete
}

// TableName returns the database table name.
func (ClaimCacheModel) TableName() string {
	return "user_provider_claim_cache"
}
