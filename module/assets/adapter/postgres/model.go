package postgres

import (
	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/pkg/orm"
)

// AssetModel is the GORM model for assets.
type AssetModel struct {
	orm.ID                     // ID embeds shared fields.
	Namespace       string     `gorm:"size:64;not null;index"`         // Namespace stores the namespace value.
	Path            string     `gorm:"size:700;not null;index"`        // Path stores the path value.
	Filename        string     `gorm:"size:160;not null"`              // Filename stores the filename value.
	DisplayName     string     `gorm:"size:160;not null"`              // DisplayName stores the display name value.
	Visibility      string     `gorm:"size:64;not null;index"`         // Visibility stores the visibility value.
	Status          string     `gorm:"size:64;not null;index"`         // Status stores the status value.
	StorageKey      string     `gorm:"size:1024;not null;uniqueIndex"` // StorageKey stores the storage key value.
	Bucket          string     `gorm:"size:128;not null"`              // Bucket stores the bucket value.
	ContentType     string     `gorm:"size:160;not null"`              // ContentType stores the content type value.
	SizeBytes       int64      `gorm:"not null"`                       // SizeBytes stores the size bytes value.
	ETag            string     `gorm:"size:256;column:etag"`           // ETag stores the e tag value.
	CreatedByUserID *uuid.UUID `gorm:"type:uuid;index"`                // CreatedByUserID stores the created by user i d value.
	Version         uint64     `gorm:"not null;default:1"`             // Version stores the version value.
	orm.Timestamps             // Timestamps embeds shared fields.
	orm.SoftDelete             // SoftDelete embeds shared fields.
}

// TableName returns the database table name.
func (AssetModel) TableName() string {
	return "assets"
}
