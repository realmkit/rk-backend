package operations

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/module/forums/domain"
	"github.com/realmkit/rk-backend/module/forums/port"
	"gorm.io/gorm"
)

// VerifyStats reports counter drift without mutating rows.
func (repository Repository) VerifyStats(ctx context.Context) (domain.CounterDriftReport, error) {
	threadExpected, err := repository.expectedThreadStats(ctx)
	if err != nil {
		return domain.CounterDriftReport{}, err
	}
	forumExpected, err := repository.expectedForumStats(ctx)
	if err != nil {
		return domain.CounterDriftReport{}, err
	}
	report := domain.CounterDriftReport{Mismatches: []domain.CounterDrift{}}
	if err := repository.appendThreadDrift(ctx, &report, threadExpected); err != nil {
		return domain.CounterDriftReport{}, err
	}
	if err := repository.appendForumDrift(ctx, &report, forumExpected); err != nil {
		return domain.CounterDriftReport{}, err
	}
	return report, nil
}

// RebuildStats repairs stats and post/thread counters from source rows.
func (repository Repository) RebuildStats(ctx context.Context) (domain.CounterDriftReport, error) {
	report, err := repository.VerifyStats(ctx)
	if err != nil {
		return domain.CounterDriftReport{}, err
	}
	if err := repository.rebuildThreadStats(ctx); err != nil {
		return domain.CounterDriftReport{}, err
	}
	if err := repository.rebuildForumStats(ctx); err != nil {
		return domain.CounterDriftReport{}, err
	}
	report.Repaired = true
	return report, nil
}

// VerifyLikes reports like counter drift without mutating rows.
func (repository Repository) VerifyLikes(ctx context.Context) (domain.CounterDriftReport, error) {
	postLikes, threadLikes, err := repository.expectedLikeStats(ctx)
	if err != nil {
		return domain.CounterDriftReport{}, err
	}
	report := domain.CounterDriftReport{Mismatches: []domain.CounterDrift{}}
	var posts []postCounterRow
	if err := repository.activeRows(ctx, "forum_posts").Find(&posts).Error; err != nil {
		return domain.CounterDriftReport{}, err
	}
	for _, post := range posts {
		report.Mismatches = appendDrift(report.Mismatches, "forum_post", post.ID, "like_count", postLikes[post.ID], post.LikeCount)
	}
	var threads []threadCounterRow
	if err := repository.activeRows(ctx, "forum_threads").Find(&threads).Error; err != nil {
		return domain.CounterDriftReport{}, err
	}
	for _, thread := range threads {
		report.Mismatches = appendDrift(
			report.Mismatches,
			"forum_thread",
			thread.ID,
			"like_count",
			threadLikes[thread.ID],
			thread.LikeCount,
		)
	}
	return report, nil
}

// RebuildLikes repairs like counters from active like rows.
func (repository Repository) RebuildLikes(ctx context.Context) (domain.CounterDriftReport, error) {
	report, err := repository.VerifyLikes(ctx)
	if err != nil {
		return domain.CounterDriftReport{}, err
	}
	postLikes, threadLikes, err := repository.expectedLikeStats(ctx)
	if err != nil {
		return domain.CounterDriftReport{}, err
	}
	if err := repository.resetAndApplyLikes(ctx, "forum_posts", postLikes); err != nil {
		return domain.CounterDriftReport{}, err
	}
	if err := repository.resetAndApplyLikes(ctx, "forum_threads", threadLikes); err != nil {
		return domain.CounterDriftReport{}, err
	}
	report.Repaired = true
	return report, nil
}

// ApplyThreadViews flushes buffered view increments into threads.
func (repository Repository) ApplyThreadViews(ctx context.Context, increments map[uuid.UUID]int64) error {
	for threadID, increment := range increments {
		if increment <= 0 {
			continue
		}
		err := repository.store.DB(ctx).
			Table("forum_threads").
			Where("id = ?", threadID).
			Update("view_count", gorm.Expr("view_count + ?", increment)).Error
		if err != nil {
			return err
		}
	}
	return nil
}

// rebuildThreadStats writes expected thread counters.
func (repository Repository) rebuildThreadStats(ctx context.Context) error {
	threadExpected, err := repository.expectedThreadStats(ctx)
	if err != nil {
		return err
	}
	for threadID, expected := range threadExpected {
		updates := map[string]any{
			"post_count":          expected.PostCount,
			"visible_post_count":  expected.VisiblePostCount,
			"reply_count":         expected.ReplyCount,
			"visible_reply_count": expected.VisibleReplyCount,
		}
		if err := repository.store.DB(ctx).Table("forum_threads").Where("id = ?", threadID).Updates(updates).Error; err != nil {
			return err
		}
	}
	return nil
}

// rebuildForumStats writes expected forum counters.
func (repository Repository) rebuildForumStats(ctx context.Context) error {
	forumExpected, err := repository.expectedForumStats(ctx)
	if err != nil {
		return err
	}
	for forumID, expected := range forumExpected {
		updates := map[string]any{
			"thread_count":         expected.ThreadCount,
			"visible_thread_count": expected.VisibleThreadCount,
			"post_count":           expected.PostCount,
			"visible_post_count":   expected.VisiblePostCount,
			"updated_at":           time.Now().UTC(),
		}
		if err := repository.store.DB(ctx).Table("forum_stats").Where("forum_id = ?", forumID).Updates(updates).Error; err != nil {
			return err
		}
	}
	return nil
}

// resetAndApplyLikes resets and reapplies like counts for one table.
func (repository Repository) resetAndApplyLikes(
	ctx context.Context,
	table string,
	counts map[uuid.UUID]int64,
) error {
	if err := repository.activeRows(ctx, table).Update("like_count", 0).Error; err != nil {
		return err
	}
	for id, count := range counts {
		if err := repository.store.DB(ctx).Table(table).Where("id = ?", id).Update("like_count", count).Error; err != nil {
			return err
		}
	}
	return nil
}

// Ensure Repository implements port.OperationsRepository.
var _ port.OperationsRepository = Repository{}
