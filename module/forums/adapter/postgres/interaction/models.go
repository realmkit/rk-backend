package interaction

import (
	"time"

	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/module/forums/domain"
	"github.com/realmkit/rk-backend/pkg/pagination"
	"gorm.io/gorm"
)

// postLikeModel is the GORM model for post likes.
type postLikeModel struct {
	ID        uuid.UUID `gorm:"column:id;type:uuid;primaryKey"`
	PostID    uuid.UUID `gorm:"column:post_id;type:uuid;not null;index"`
	ThreadID  uuid.UUID `gorm:"column:thread_id;type:uuid;not null;index"`
	ForumID   uuid.UUID `gorm:"column:forum_id;type:uuid;not null;index"`
	UserID    uuid.UUID `gorm:"column:user_id;type:uuid;not null;index"`
	CreatedAt time.Time `gorm:"column:created_at;not null"`
	DeletedAt gorm.DeletedAt
}

// TableName returns the database table name.
func (postLikeModel) TableName() string {
	return "forum_post_likes"
}

// threadReadStateModel is the GORM model for thread read states.
type threadReadStateModel struct {
	ID                   uuid.UUID `gorm:"column:id;type:uuid;primaryKey"`
	UserID               uuid.UUID `gorm:"column:user_id;type:uuid;not null;index"`
	ForumID              uuid.UUID `gorm:"column:forum_id;type:uuid;not null;index"`
	ThreadID             uuid.UUID `gorm:"column:thread_id;type:uuid;not null;index"`
	LastReadPostSequence int64     `gorm:"column:last_read_post_sequence;not null"`
	LastReadAt           time.Time `gorm:"column:last_read_at;not null;index"`
	CreatedAt            time.Time `gorm:"column:created_at"`
	UpdatedAt            time.Time `gorm:"column:updated_at"`
}

// TableName returns the database table name.
func (threadReadStateModel) TableName() string {
	return "forum_thread_read_states"
}

// latestPostRow is a compact latest-post query row.
type latestPostRow struct {
	ForumID      uuid.UUID
	ThreadID     uuid.UUID
	PostID       uuid.UUID
	AuthorUserID uuid.UUID
	Sequence     int64
	ThreadTitle  string
	ThreadSlug   string
	Excerpt      string
	CreatedAt    time.Time
}

// mostLikedPostRow is a compact most-liked query row.
type mostLikedPostRow struct {
	latestPostRow
	LikeCount int64
}

// forumReadTargetRow is a thread read-state target row.
type forumReadTargetRow struct {
	ThreadID             uuid.UUID
	ForumID              uuid.UUID
	LastReadPostSequence int64
}

// unreadForumRow is an unread count query row.
type unreadForumRow struct {
	ForumID           uuid.UUID
	UnreadThreadCount int64
}

// likeModelFromDomain maps like to persistence.
func likeModelFromDomain(like domain.PostLike) postLikeModel {
	return postLikeModel{
		ID:        like.ID,
		PostID:    like.PostID,
		ThreadID:  like.ThreadID,
		ForumID:   like.ForumID,
		UserID:    like.UserID,
		CreatedAt: like.CreatedAt,
	}
}

// readStateModelFromDomain maps read state to persistence.
func readStateModelFromDomain(state domain.ThreadReadState) threadReadStateModel {
	return threadReadStateModel{
		ID:                   state.ID,
		UserID:               state.UserID,
		ForumID:              state.ForumID,
		ThreadID:             state.ThreadID,
		LastReadPostSequence: state.LastReadPostSequence,
		LastReadAt:           state.LastReadAt,
		CreatedAt:            state.CreatedAt,
		UpdatedAt:            state.UpdatedAt,
	}
}

// latestPostPage maps latest-post rows into a page.
func latestPostPage(rows []latestPostRow, limit int) pagination.Result[domain.LatestPostSummary] {
	next := ""
	if len(rows) > limit {
		next = rows[limit-1].PostID.String()
		rows = rows[:limit]
	}
	items := make([]domain.LatestPostSummary, 0, len(rows))
	for _, row := range rows {
		items = append(items, domain.LatestPostSummary{
			ForumID:      row.ForumID,
			ThreadID:     row.ThreadID,
			PostID:       row.PostID,
			AuthorUserID: row.AuthorUserID,
			Sequence:     row.Sequence,
			ThreadTitle:  row.ThreadTitle,
			ThreadSlug:   domain.Slug(row.ThreadSlug),
			Excerpt:      row.Excerpt,
			CreatedAt:    row.CreatedAt,
		})
	}
	return pagination.Result[domain.LatestPostSummary]{Items: items, NextCursor: next}
}

// mostLikedPostPage maps most-liked rows into a page.
func mostLikedPostPage(rows []mostLikedPostRow, limit int) pagination.Result[domain.MostLikedPost] {
	next := ""
	if len(rows) > limit {
		next = rows[limit-1].PostID.String()
		rows = rows[:limit]
	}
	items := make([]domain.MostLikedPost, 0, len(rows))
	for _, row := range rows {
		items = append(items, domain.MostLikedPost{
			ForumID:      row.ForumID,
			ThreadID:     row.ThreadID,
			PostID:       row.PostID,
			AuthorUserID: row.AuthorUserID,
			Sequence:     row.Sequence,
			ThreadTitle:  row.ThreadTitle,
			ThreadSlug:   domain.Slug(row.ThreadSlug),
			Excerpt:      row.Excerpt,
			LikeCount:    row.LikeCount,
			CreatedAt:    row.CreatedAt,
		})
	}
	return pagination.Result[domain.MostLikedPost]{Items: items, NextCursor: next}
}

// visiblePostStatuses returns statuses normal widget readers may see.
func visiblePostStatuses() []domain.PostStatus {
	return []domain.PostStatus{domain.PostStatusVisible, domain.PostStatusSystem}
}

// visibleThreadStatuses returns statuses normal widget readers may see.
func visibleThreadStatuses() []domain.ThreadStatus {
	return []domain.ThreadStatus{
		domain.ThreadStatusOpen,
		domain.ThreadStatusClosed,
		domain.ThreadStatusLocked,
	}
}
