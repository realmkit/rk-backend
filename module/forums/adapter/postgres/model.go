package postgres

import (
	"time"

	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/pkg/orm"
)

// CategoryModel is the GORM model for forum categories.
type CategoryModel struct {
	orm.ID                // ID embeds shared fields.
	Key            string `gorm:"size:64;not null;uniqueIndex"`  // Key stores the key value.
	Name           string `gorm:"size:120;not null"`             // Name stores the name value.
	Description    string `gorm:"size:1000;not null;default:''"` // Description stores the description value.
	DisplayOrder   int    `gorm:"not null;default:0;index"`      // DisplayOrder stores the display order value.
	Status         string `gorm:"size:64;not null;index"`        // Status stores the status value.
	Version        uint64 `gorm:"not null;default:1"`            // Version stores the version value.
	orm.Timestamps        // Timestamps embeds shared fields.
	orm.SoftDelete        // SoftDelete embeds shared fields.
}

// TableName returns the database table name.
func (CategoryModel) TableName() string {
	return "forum_categories"
}

// ForumModel is the GORM model for forums.
type ForumModel struct {
	orm.ID                                   // ID embeds shared fields.
	CategoryID                    uuid.UUID  `gorm:"type:uuid;not null;index"`      // CategoryID stores the category i d value.
	ParentForumID                 *uuid.UUID `gorm:"type:uuid;index"`               // ParentForumID stores the parent forum i d value.
	Kind                          string     `gorm:"size:64;not null;index"`        // Kind stores the kind value.
	Key                           string     `gorm:"size:64;not null;uniqueIndex"`  // Key stores the key value.
	Slug                          string     `gorm:"size:120;not null;index"`       // Slug stores the slug value.
	Name                          string     `gorm:"size:120;not null"`             // Name stores the name value.
	Description                   string     `gorm:"size:1000;not null;default:''"` // Description stores the description value.
	DisplayOrder                  int        `gorm:"not null;default:0;index"`      // DisplayOrder stores the display order value.
	Path                          string     `gorm:"size:700;not null;index"`       // Path stores the path value.
	Depth                         int        `gorm:"not null;default:0"`            // Depth stores the depth value.
	ExternalURL                   string     `gorm:"size:2048;not null;default:''"` // ExternalURL stores the external u r l value.
	IconAssetID                   *uuid.UUID `gorm:"type:uuid;index"`               // IconAssetID stores the icon asset i d value.
	ThreadVisibilityMode          string     `gorm:"size:64;not null"`              // ThreadVisibilityMode stores the thread visibility mode value.
	MaxStickyThreads              int        `gorm:"not null;default:0"`            // MaxStickyThreads stores the max sticky threads value.
	DefaultThreadStatus           string     `gorm:"size:64;not null"`              // DefaultThreadStatus stores the default thread status value.
	AuthorPostEditWindowSeconds   int        `gorm:"not null;default:600"`          // AuthorPostEditWindowSeconds stores the author post edit window seconds value.
	AuthorPostDeleteWindowSeconds int        `gorm:"not null;default:300"`          // AuthorPostDeleteWindowSeconds stores the author post delete window seconds value.
	Status                        string     `gorm:"size:64;not null;index"`        // Status stores the status value.
	Version                       uint64     `gorm:"not null;default:1"`            // Version stores the version value.
	orm.Timestamps                           // Timestamps embeds shared fields.
	orm.SoftDelete                           // SoftDelete embeds shared fields.
}

// TableName returns the database table name.
func (ForumModel) TableName() string {
	return "forums"
}

// StatsModel is the GORM model for forum stats.
type StatsModel struct {
	ForumID                uuid.UUID  `gorm:"type:uuid;primaryKey"` // ForumID stores the forum i d value.
	ThreadCount            int64      `gorm:"not null;default:0"`   // ThreadCount stores the thread count value.
	VisibleThreadCount     int64      `gorm:"not null;default:0"`   // VisibleThreadCount stores the visible thread count value.
	PostCount              int64      `gorm:"not null;default:0"`   // PostCount stores the post count value.
	VisiblePostCount       int64      `gorm:"not null;default:0"`   // VisiblePostCount stores the visible post count value.
	LatestThreadID         *uuid.UUID `gorm:"type:uuid"`            // LatestThreadID stores the latest thread i d value.
	LatestPostID           *uuid.UUID `gorm:"type:uuid"`            // LatestPostID stores the latest post i d value.
	LatestPostAuthorUserID *uuid.UUID `gorm:"type:uuid"`            // LatestPostAuthorUserID stores the latest post author user i d value.
	LatestPostAt           *time.Time // LatestPostAt stores the latest post at value.
	UpdatedAt              time.Time  `gorm:"not null"` // UpdatedAt stores the updated at value.
}

// TableName returns the database table name.
func (StatsModel) TableName() string {
	return "forum_stats"
}

// ThreadModel is the GORM model for forum threads.
type ThreadModel struct {
	orm.ID                            // ID embeds shared fields.
	ForumID                uuid.UUID  `gorm:"type:uuid;not null;index"` // ForumID stores the forum i d value.
	AuthorUserID           uuid.UUID  `gorm:"type:uuid;not null;index"` // AuthorUserID stores the author user i d value.
	OpenerPostID           uuid.UUID  `gorm:"type:uuid"`                // OpenerPostID stores the opener post i d value.
	LatestPostID           uuid.UUID  `gorm:"type:uuid"`                // LatestPostID stores the latest post i d value.
	LatestPostAuthorUserID uuid.UUID  `gorm:"type:uuid"`                // LatestPostAuthorUserID stores the latest post author user i d value.
	LatestPostAt           time.Time  `gorm:"not null;index"`           // LatestPostAt stores the latest post at value.
	Title                  string     `gorm:"size:160;not null"`        // Title stores the title value.
	Slug                   string     `gorm:"size:160;not null;index"`  // Slug stores the slug value.
	Status                 string     `gorm:"size:64;not null;index"`   // Status stores the status value.
	StickyState            string     `gorm:"size:64;not null;index"`   // StickyState stores the sticky state value.
	StickyOrder            int        `gorm:"not null;default:0"`       // StickyOrder stores the sticky order value.
	StickyUntil            *time.Time // StickyUntil stores the sticky until value.
	LockedReason           string     `gorm:"size:500;not null;default:''"` // LockedReason stores the locked reason value.
	ReplyCount             int64      `gorm:"not null;default:0"`           // ReplyCount stores the reply count value.
	VisibleReplyCount      int64      `gorm:"not null;default:0"`           // VisibleReplyCount stores the visible reply count value.
	PostCount              int64      `gorm:"not null;default:0"`           // PostCount stores the post count value.
	VisiblePostCount       int64      `gorm:"not null;default:0"`           // VisiblePostCount stores the visible post count value.
	LikeCount              int64      `gorm:"not null;default:0"`           // LikeCount stores the like count value.
	ViewCount              int64      `gorm:"not null;default:0"`           // ViewCount stores the view count value.
	Version                uint64     `gorm:"not null;default:1"`           // Version stores the version value.
	orm.Timestamps                    // Timestamps embeds shared fields.
	orm.SoftDelete                    // SoftDelete embeds shared fields.
}

// TableName returns the database table name.
func (ThreadModel) TableName() string {
	return "forum_threads"
}

// PostModel is the GORM model for forum posts.
type PostModel struct {
	orm.ID                         // ID embeds shared fields.
	ThreadID            uuid.UUID  `gorm:"type:uuid;not null;index"`     // ThreadID stores the thread i d value.
	ForumID             uuid.UUID  `gorm:"type:uuid;not null;index"`     // ForumID stores the forum i d value.
	AuthorUserID        uuid.UUID  `gorm:"type:uuid;not null;index"`     // AuthorUserID stores the author user i d value.
	Sequence            int64      `gorm:"not null;index"`               // Sequence stores the sequence value.
	Status              string     `gorm:"size:64;not null;index"`       // Status stores the status value.
	ContentFormat       string     `gorm:"size:64;not null"`             // ContentFormat stores the content format value.
	ContentDocumentJSON string     `gorm:"type:jsonb;not null"`          // ContentDocumentJSON stores the content document j s o n value.
	ContentText         string     `gorm:"type:text;not null"`           // ContentText stores the content text value.
	ContentChecksum     string     `gorm:"size:128;not null;default:''"` // ContentChecksum stores the content checksum value.
	EditedAt            *time.Time // EditedAt stores the edited at value.
	EditedByUserID      *uuid.UUID `gorm:"type:uuid"`          // EditedByUserID stores the edited by user i d value.
	EditCount           int64      `gorm:"not null;default:0"` // EditCount stores the edit count value.
	LikeCount           int64      `gorm:"not null;default:0"` // LikeCount stores the like count value.
	ReplyReferenceCount int64      `gorm:"not null;default:0"` // ReplyReferenceCount stores the reply reference count value.
	Version             uint64     `gorm:"not null;default:1"` // Version stores the version value.
	orm.Timestamps                 // Timestamps embeds shared fields.
	orm.SoftDelete                 // SoftDelete embeds shared fields.
}

// TableName returns the database table name.
func (PostModel) TableName() string {
	return "forum_posts"
}

// PostRevisionModel is the GORM model for post revisions.
type PostRevisionModel struct {
	orm.ID                                // ID embeds shared fields.
	PostID                      uuid.UUID `gorm:"type:uuid;not null;index"`     // PostID stores the post i d value.
	EditedByUserID              uuid.UUID `gorm:"type:uuid;not null"`           // EditedByUserID stores the edited by user i d value.
	PreviousContentDocumentJSON string    `gorm:"type:jsonb;not null"`          // PreviousContentDocumentJSON stores the previous content document j s o n value.
	PreviousContentText         string    `gorm:"type:text;not null"`           // PreviousContentText stores the previous content text value.
	EditReason                  string    `gorm:"size:500;not null;default:''"` // EditReason stores the edit reason value.
	CreatedAt                   time.Time `gorm:"not null"`                     // CreatedAt stores the created at value.
}

// TableName returns the database table name.
func (PostRevisionModel) TableName() string {
	return "forum_post_revisions"
}

// PostReferenceModel is the GORM model for post references.
type PostReferenceModel struct {
	orm.ID                   // ID embeds shared fields.
	SourcePostID  uuid.UUID  `gorm:"type:uuid;not null;index"`      // SourcePostID stores the source post i d value.
	TargetPostID  *uuid.UUID `gorm:"type:uuid;index"`               // TargetPostID stores the target post i d value.
	TargetUserID  *uuid.UUID `gorm:"type:uuid;index"`               // TargetUserID stores the target user i d value.
	TargetAssetID *uuid.UUID `gorm:"type:uuid;index"`               // TargetAssetID stores the target asset i d value.
	ReferenceType string     `gorm:"size:64;not null;index"`        // ReferenceType stores the reference type value.
	QuoteExcerpt  string     `gorm:"size:500;not null;default:''"`  // QuoteExcerpt stores the quote excerpt value.
	LinkURL       string     `gorm:"size:2048;not null;default:''"` // LinkURL stores the link u r l value.
	CreatedAt     time.Time  `gorm:"not null"`                      // CreatedAt stores the created at value.
}

// TableName returns the database table name.
func (PostReferenceModel) TableName() string {
	return "forum_post_references"
}

// PostLikeModel is the GORM model for post likes.
type PostLikeModel struct {
	orm.ID                   // ID embeds shared fields.
	PostID         uuid.UUID `gorm:"type:uuid;not null;index"` // PostID stores the post i d value.
	ThreadID       uuid.UUID `gorm:"type:uuid;not null;index"` // ThreadID stores the thread i d value.
	ForumID        uuid.UUID `gorm:"type:uuid;not null;index"` // ForumID stores the forum i d value.
	UserID         uuid.UUID `gorm:"type:uuid;not null;index"` // UserID stores the user i d value.
	CreatedAt      time.Time `gorm:"not null"`                 // CreatedAt stores the created at value.
	orm.SoftDelete           // SoftDelete embeds shared fields.
}

// TableName returns the database table name.
func (PostLikeModel) TableName() string {
	return "forum_post_likes"
}

// ThreadReadStateModel is the GORM model for thread read states.
type ThreadReadStateModel struct {
	orm.ID                         // ID embeds shared fields.
	UserID               uuid.UUID `gorm:"type:uuid;not null;index"` // UserID stores the user i d value.
	ForumID              uuid.UUID `gorm:"type:uuid;not null;index"` // ForumID stores the forum i d value.
	ThreadID             uuid.UUID `gorm:"type:uuid;not null;index"` // ThreadID stores the thread i d value.
	LastReadPostSequence int64     `gorm:"not null"`                 // LastReadPostSequence stores the last read post sequence value.
	LastReadAt           time.Time `gorm:"not null;index"`           // LastReadAt stores the last read at value.
	CreatedAt            time.Time // CreatedAt stores the created at value.
	UpdatedAt            time.Time // UpdatedAt stores the updated at value.
}

// TableName returns the database table name.
func (ThreadReadStateModel) TableName() string {
	return "forum_thread_read_states"
}
