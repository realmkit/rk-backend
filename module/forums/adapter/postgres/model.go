package postgres

import (
	"time"

	"github.com/google/uuid"
	"github.com/niflaot/gamehub-go/pkg/orm"
)

// CategoryModel is the GORM model for forum categories.
type CategoryModel struct {
	orm.ID
	Key          string `gorm:"size:64;not null;uniqueIndex"`
	Name         string `gorm:"size:120;not null"`
	Description  string `gorm:"size:1000;not null;default:''"`
	DisplayOrder int    `gorm:"not null;default:0;index"`
	Status       string `gorm:"size:64;not null;index"`
	Version      uint64 `gorm:"not null;default:1"`
	orm.Timestamps
	orm.SoftDelete
}

// TableName returns the database table name.
func (CategoryModel) TableName() string {
	return "forum_categories"
}

// ForumModel is the GORM model for forums.
type ForumModel struct {
	orm.ID
	CategoryID           uuid.UUID  `gorm:"type:uuid;not null;index"`
	ParentForumID        *uuid.UUID `gorm:"type:uuid;index"`
	Kind                 string     `gorm:"size:64;not null;index"`
	Key                  string     `gorm:"size:64;not null;uniqueIndex"`
	Slug                 string     `gorm:"size:120;not null;index"`
	Name                 string     `gorm:"size:120;not null"`
	Description          string     `gorm:"size:1000;not null;default:''"`
	DisplayOrder         int        `gorm:"not null;default:0;index"`
	Path                 string     `gorm:"size:700;not null;index"`
	Depth                int        `gorm:"not null;default:0"`
	ExternalURL          string     `gorm:"size:2048;not null;default:''"`
	IconAssetID          *uuid.UUID `gorm:"type:uuid;index"`
	ThreadVisibilityMode string     `gorm:"size:64;not null"`
	MaxStickyThreads     int        `gorm:"not null;default:0"`
	DefaultThreadStatus  string     `gorm:"size:64;not null"`
	Status               string     `gorm:"size:64;not null;index"`
	Version              uint64     `gorm:"not null;default:1"`
	orm.Timestamps
	orm.SoftDelete
}

// TableName returns the database table name.
func (ForumModel) TableName() string {
	return "forums"
}

// StatsModel is the GORM model for forum stats.
type StatsModel struct {
	ForumID                uuid.UUID  `gorm:"type:uuid;primaryKey"`
	ThreadCount            int64      `gorm:"not null;default:0"`
	VisibleThreadCount     int64      `gorm:"not null;default:0"`
	PostCount              int64      `gorm:"not null;default:0"`
	VisiblePostCount       int64      `gorm:"not null;default:0"`
	LatestThreadID         *uuid.UUID `gorm:"type:uuid"`
	LatestPostID           *uuid.UUID `gorm:"type:uuid"`
	LatestPostAuthorUserID *uuid.UUID `gorm:"type:uuid"`
	LatestPostAt           *time.Time
	UpdatedAt              time.Time `gorm:"not null"`
}

// TableName returns the database table name.
func (StatsModel) TableName() string {
	return "forum_stats"
}
