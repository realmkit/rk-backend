// Package interaction adapts forum interaction repositories to PostgreSQL.
package interaction

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/niflaot/gamehub-go/module/forums/domain"
	"github.com/niflaot/gamehub-go/module/forums/port"
	"github.com/niflaot/gamehub-go/pkg/orm"
	"github.com/niflaot/gamehub-go/pkg/pagination"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Repository stores forum interactions in PostgreSQL.
type Repository struct {
	store orm.Store
}

// NewRepository creates an interaction repository.
func NewRepository(store orm.Store) Repository {
	return Repository{store: store}
}

// LikePost creates or restores an active like and returns whether counters changed.
func (repository Repository) LikePost(ctx context.Context, like domain.PostLike) (bool, error) {
	var active postLikeModel
	err := repository.likeQuery(ctx, like.PostID, like.UserID).First(&active).Error
	if err == nil {
		return false, nil
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return false, err
	}
	var existing postLikeModel
	err = repository.likeQuery(ctx, like.PostID, like.UserID).Unscoped().First(&existing).Error
	if err == nil {
		return true, repository.restoreLike(ctx, existing.ID, like)
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return false, err
	}
	if err := repository.store.DB(ctx).Create(likeModelFromDomain(like)).Error; err != nil {
		return false, port.ErrConflict
	}
	return true, repository.updateLikeCounts(ctx, like.PostID, like.ThreadID, 1)
}

// UnlikePost removes an active like and returns whether counters changed.
func (repository Repository) UnlikePost(ctx context.Context, postID uuid.UUID, userID uuid.UUID) (bool, error) {
	var active postLikeModel
	err := repository.likeQuery(ctx, postID, userID).First(&active).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if err := repository.store.DB(ctx).Delete(&active).Error; err != nil {
		return false, err
	}
	return true, repository.updateLikeCounts(ctx, active.PostID, active.ThreadID, -1)
}

// LikedByUser reports whether user currently likes post.
func (repository Repository) LikedByUser(ctx context.Context, postID uuid.UUID, userID uuid.UUID) (bool, error) {
	if userID == uuid.Nil {
		return false, nil
	}
	var count int64
	err := repository.likeQuery(ctx, postID, userID).Count(&count).Error
	return count > 0, err
}

// MarkThreadRead stores one thread read state.
func (repository Repository) MarkThreadRead(ctx context.Context, state domain.ThreadReadState) error {
	now := state.LastReadAt
	if state.CreatedAt.IsZero() {
		state.CreatedAt = now
	}
	state.UpdatedAt = now
	model := readStateModelFromDomain(state)
	updates := map[string]any{
		"forum_id":                state.ForumID,
		"last_read_post_sequence": state.LastReadPostSequence,
		"last_read_at":            state.LastReadAt,
		"updated_at":              state.UpdatedAt,
	}
	return repository.store.DB(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "user_id"}, {Name: "thread_id"}},
			DoUpdates: clause.Assignments(updates),
		}).
		Create(&model).Error
}

// MarkForumRead stores read states for every visible thread in a forum.
func (repository Repository) MarkForumRead(
	ctx context.Context,
	userID uuid.UUID,
	forumID uuid.UUID,
	readAt time.Time,
) error {
	var rows []forumReadTargetRow
	err := repository.store.DB(ctx).
		Table("forum_threads AS t").
		Select("t.id AS thread_id, t.forum_id, COALESCE(MAX(p.sequence), 1) AS last_read_post_sequence").
		Joins("JOIN forum_posts AS p ON p.thread_id = t.id AND p.deleted_at IS NULL AND p.status IN ?", visiblePostStatuses()).
		Where("t.forum_id = ? AND t.deleted_at IS NULL AND t.status IN ?", forumID, visibleThreadStatuses()).
		Group("t.id, t.forum_id").
		Find(&rows).Error
	if err != nil {
		return err
	}
	for _, row := range rows {
		state := domain.ThreadReadState{
			ID:                   uuid.New(),
			UserID:               userID,
			ForumID:              row.ForumID,
			ThreadID:             row.ThreadID,
			LastReadPostSequence: row.LastReadPostSequence,
			LastReadAt:           readAt,
		}
		if err := repository.MarkThreadRead(ctx, state); err != nil {
			return err
		}
	}
	return nil
}

// likeQuery scopes a like lookup by post and user.
func (repository Repository) likeQuery(ctx context.Context, postID uuid.UUID, userID uuid.UUID) *gorm.DB {
	return repository.store.DB(ctx).
		Where("post_id = ? AND user_id = ?", postID, userID).
		Model(&postLikeModel{})
}

// restoreLike restores a previously soft-deleted like.
func (repository Repository) restoreLike(ctx context.Context, id uuid.UUID, like domain.PostLike) error {
	updates := map[string]any{
		"deleted_at": nil,
		"thread_id":  like.ThreadID,
		"forum_id":   like.ForumID,
		"created_at": like.CreatedAt,
	}
	err := repository.store.DB(ctx).Unscoped().
		Model(&postLikeModel{}).
		Where("id = ?", id).
		Updates(updates).Error
	if err != nil {
		return err
	}
	return repository.updateLikeCounts(ctx, like.PostID, like.ThreadID, 1)
}

// updateLikeCounts changes post and thread like counters.
func (repository Repository) updateLikeCounts(ctx context.Context, postID uuid.UUID, threadID uuid.UUID, delta int64) error {
	expr := gorm.Expr("like_count + ?", delta)
	if delta < 0 {
		expr = gorm.Expr("CASE WHEN like_count > 0 THEN like_count - 1 ELSE 0 END")
	}
	if err := repository.store.DB(ctx).Table("forum_posts").Where("id = ?", postID).Update("like_count", expr).Error; err != nil {
		return err
	}
	return repository.store.DB(ctx).Table("forum_threads").Where("id = ?", threadID).Update("like_count", expr).Error
}

// Ensure Repository implements port.InteractionRepository.
var _ port.InteractionRepository = Repository{}

// keepPaginationImport keeps pagination referenced by sibling files in package docs.
var keepPaginationImport pagination.Page
