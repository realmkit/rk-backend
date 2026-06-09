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

// ThreadModel is the GORM model for forum threads.
type ThreadModel struct {
	orm.ID
	ForumID                uuid.UUID `gorm:"type:uuid;not null;index"`
	AuthorUserID           uuid.UUID `gorm:"type:uuid;not null;index"`
	OpenerPostID           uuid.UUID `gorm:"type:uuid"`
	LatestPostID           uuid.UUID `gorm:"type:uuid"`
	LatestPostAuthorUserID uuid.UUID `gorm:"type:uuid"`
	LatestPostAt           time.Time `gorm:"not null;index"`
	Title                  string    `gorm:"size:160;not null"`
	Slug                   string    `gorm:"size:160;not null;index"`
	Status                 string    `gorm:"size:64;not null;index"`
	StickyState            string    `gorm:"size:64;not null;index"`
	StickyOrder            int       `gorm:"not null;default:0"`
	StickyUntil            *time.Time
	LockedReason           string `gorm:"size:500;not null;default:''"`
	ReplyCount             int64  `gorm:"not null;default:0"`
	VisibleReplyCount      int64  `gorm:"not null;default:0"`
	PostCount              int64  `gorm:"not null;default:0"`
	VisiblePostCount       int64  `gorm:"not null;default:0"`
	LikeCount              int64  `gorm:"not null;default:0"`
	ViewCount              int64  `gorm:"not null;default:0"`
	Version                uint64 `gorm:"not null;default:1"`
	orm.Timestamps
	orm.SoftDelete
}

// TableName returns the database table name.
func (ThreadModel) TableName() string {
	return "forum_threads"
}

// PostModel is the GORM model for forum posts.
type PostModel struct {
	orm.ID
	ThreadID            uuid.UUID `gorm:"type:uuid;not null;index"`
	ForumID             uuid.UUID `gorm:"type:uuid;not null;index"`
	AuthorUserID        uuid.UUID `gorm:"type:uuid;not null;index"`
	Sequence            int64     `gorm:"not null;index"`
	Status              string    `gorm:"size:64;not null;index"`
	ContentFormat       string    `gorm:"size:64;not null"`
	ContentDocumentJSON string    `gorm:"type:jsonb;not null"`
	ContentText         string    `gorm:"type:text;not null"`
	ContentChecksum     string    `gorm:"size:128;not null;default:''"`
	EditedAt            *time.Time
	EditedByUserID      *uuid.UUID `gorm:"type:uuid"`
	EditCount           int64      `gorm:"not null;default:0"`
	LikeCount           int64      `gorm:"not null;default:0"`
	ReplyReferenceCount int64      `gorm:"not null;default:0"`
	Version             uint64     `gorm:"not null;default:1"`
	orm.Timestamps
	orm.SoftDelete
}

// TableName returns the database table name.
func (PostModel) TableName() string {
	return "forum_posts"
}

// PostRevisionModel is the GORM model for post revisions.
type PostRevisionModel struct {
	orm.ID
	PostID                      uuid.UUID `gorm:"type:uuid;not null;index"`
	EditedByUserID              uuid.UUID `gorm:"type:uuid;not null"`
	PreviousContentDocumentJSON string    `gorm:"type:jsonb;not null"`
	PreviousContentText         string    `gorm:"type:text;not null"`
	EditReason                  string    `gorm:"size:500;not null;default:''"`
	CreatedAt                   time.Time `gorm:"not null"`
}

// TableName returns the database table name.
func (PostRevisionModel) TableName() string {
	return "forum_post_revisions"
}

// PostReferenceModel is the GORM model for post references.
type PostReferenceModel struct {
	orm.ID
	SourcePostID  uuid.UUID  `gorm:"type:uuid;not null;index"`
	TargetPostID  *uuid.UUID `gorm:"type:uuid;index"`
	TargetUserID  *uuid.UUID `gorm:"type:uuid;index"`
	TargetAssetID *uuid.UUID `gorm:"type:uuid;index"`
	ReferenceType string     `gorm:"size:64;not null;index"`
	QuoteExcerpt  string     `gorm:"size:500;not null;default:''"`
	LinkURL       string     `gorm:"size:2048;not null;default:''"`
	CreatedAt     time.Time  `gorm:"not null"`
}

// TableName returns the database table name.
func (PostReferenceModel) TableName() string {
	return "forum_post_references"
}

// PostLikeModel is the GORM model for post likes.
type PostLikeModel struct {
	orm.ID
	PostID    uuid.UUID `gorm:"type:uuid;not null;index"`
	ThreadID  uuid.UUID `gorm:"type:uuid;not null;index"`
	ForumID   uuid.UUID `gorm:"type:uuid;not null;index"`
	UserID    uuid.UUID `gorm:"type:uuid;not null;index"`
	CreatedAt time.Time `gorm:"not null"`
	orm.SoftDelete
}

// TableName returns the database table name.
func (PostLikeModel) TableName() string {
	return "forum_post_likes"
}

// ThreadReadStateModel is the GORM model for thread read states.
type ThreadReadStateModel struct {
	orm.ID
	UserID               uuid.UUID `gorm:"type:uuid;not null;index"`
	ForumID              uuid.UUID `gorm:"type:uuid;not null;index"`
	ThreadID             uuid.UUID `gorm:"type:uuid;not null;index"`
	LastReadPostSequence int64     `gorm:"not null"`
	LastReadAt           time.Time `gorm:"not null;index"`
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

// TableName returns the database table name.
func (ThreadReadStateModel) TableName() string {
	return "forum_thread_read_states"
}
