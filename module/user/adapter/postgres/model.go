// Package postgres stores users in PostgreSQL through GORM.
package postgres

import (
	"time"

	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/pkg/orm"
)

// UserModel is the GORM model for local users.
type UserModel struct {
	orm.ID                    // ID embeds shared fields.
	Status         string     `gorm:"size:64;not null;index"` // Status stores the status value.
	AvatarAssetID  *uuid.UUID `gorm:"type:uuid;index"`        // AvatarAssetID stores the avatar asset i d value.
	FirstSeenAt    time.Time  `gorm:"not null"`               // FirstSeenAt stores the first seen at value.
	LastSeenAt     *time.Time `gorm:"index"`                  // LastSeenAt stores the last seen at value.
	Version        uint64     `gorm:"not null;default:1"`     // Version stores the version value.
	orm.Timestamps            // Timestamps embeds shared fields.
	orm.SoftDelete            // SoftDelete embeds shared fields.
}

// TableName returns the database table name.
func (UserModel) TableName() string {
	return "users"
}

// IdentityLinkModel is the GORM model for provider identity links.
type IdentityLinkModel struct {
	orm.ID                    // ID embeds shared fields.
	UserID         uuid.UUID  `gorm:"type:uuid;not null;index"`    // UserID stores the user i d value.
	Provider       string     `gorm:"size:64;not null;index"`      // Provider stores the provider value.
	Issuer         string     `gorm:"size:512;not null;index"`     // Issuer stores the issuer value.
	Subject        string     `gorm:"size:512;not null;index"`     // Subject stores the subject value.
	SubjectHash    string     `gorm:"size:64;not null;index"`      // SubjectHash stores the subject hash value.
	ClaimsHash     string     `gorm:"size:64;not null;default:''"` // ClaimsHash stores the claims hash value.
	LinkedAt       time.Time  `gorm:"not null"`                    // LinkedAt stores the linked at value.
	LastSeenAt     *time.Time `gorm:"index"`                       // LastSeenAt stores the last seen at value.
	LastSyncedAt   *time.Time `gorm:"index"`                       // LastSyncedAt stores the last synced at value.
	orm.Timestamps            // Timestamps embeds shared fields.
	orm.SoftDelete            // SoftDelete embeds shared fields.
}

// TableName returns the database table name.
func (IdentityLinkModel) TableName() string {
	return "user_identity_links"
}

// ClaimCacheModel is the GORM model for provider claim cache rows.
type ClaimCacheModel struct {
	orm.ID                    // ID embeds shared fields.
	UserID          uuid.UUID `gorm:"type:uuid;not null;index"`          // UserID stores the user i d value.
	Issuer          string    `gorm:"size:512;not null;index"`           // Issuer stores the issuer value.
	Subject         string    `gorm:"size:512;not null;index"`           // Subject stores the subject value.
	Username        string    `gorm:"size:256;not null;default:''"`      // Username stores the username value.
	Email           string    `gorm:"size:320;not null;default:''"`      // Email stores the email value.
	EmailVerified   bool      `gorm:"not null;default:false"`            // EmailVerified stores the email verified value.
	DisplayName     string    `gorm:"size:256;not null;default:''"`      // DisplayName stores the display name value.
	PictureURL      string    `gorm:"size:1024;not null;default:''"`     // PictureURL stores the picture u r l value.
	PreferredLocale string    `gorm:"size:64;not null;default:''"`       // PreferredLocale stores the preferred locale value.
	ClaimsHash      string    `gorm:"size:64;not null;default:'';index"` // ClaimsHash stores the claims hash value.
	SyncedAt        time.Time `gorm:"not null;index"`                    // SyncedAt stores the synced at value.
	orm.Timestamps            // Timestamps embeds shared fields.
	orm.SoftDelete            // SoftDelete embeds shared fields.
}

// TableName returns the database table name.
func (ClaimCacheModel) TableName() string {
	return "user_provider_claim_cache"
}
