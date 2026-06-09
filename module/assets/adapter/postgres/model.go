package postgres

import (
	"github.com/google/uuid"
	"github.com/niflaot/gamehub-go/pkg/orm"
)

// AssetModel is the GORM model for assets.
type AssetModel struct {
	orm.ID
	Namespace       string     `gorm:"size:64;not null;index"`
	Path            string     `gorm:"size:700;not null;index"`
	Filename        string     `gorm:"size:160;not null"`
	DisplayName     string     `gorm:"size:160;not null"`
	Visibility      string     `gorm:"size:64;not null;index"`
	Status          string     `gorm:"size:64;not null;index"`
	StorageKey      string     `gorm:"size:1024;not null;uniqueIndex"`
	Bucket          string     `gorm:"size:128;not null"`
	ContentType     string     `gorm:"size:160;not null"`
	SizeBytes       int64      `gorm:"not null"`
	ETag            string     `gorm:"size:256;column:etag"`
	CreatedByUserID *uuid.UUID `gorm:"type:uuid;index"`
	Version         uint64     `gorm:"not null;default:1"`
	orm.Timestamps
	orm.SoftDelete
}

// TableName returns the database table name.
func (AssetModel) TableName() string {
	return "assets"
}
