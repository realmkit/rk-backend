package operations

import (
	"context"

	"github.com/google/uuid"
	"github.com/niflaot/gamehub-go/module/forums/domain"
	"gorm.io/gorm"
)

// expectedThreadStats calculates thread counters from posts.
func (repository Repository) expectedThreadStats(ctx context.Context) (map[uuid.UUID]threadExpectation, error) {
	result, err := repository.emptyThreadExpectations(ctx)
	if err != nil {
		return nil, err
	}
	var rows []threadPostCounterRow
	err = repository.store.DB(ctx).
		Table("forum_posts").
		Select(threadPostCounterSelect, visiblePostStatuses()).
		Where("deleted_at IS NULL").
		Group("thread_id").
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	for _, row := range rows {
		expected := result[row.ThreadID]
		expected.PostCount = row.PostCount
		expected.VisiblePostCount = row.VisiblePostCount
		if row.PostCount > 0 {
			expected.ReplyCount = row.PostCount - 1
		}
		if row.VisiblePostCount > 0 {
			expected.VisibleReplyCount = row.VisiblePostCount - 1
		}
		result[row.ThreadID] = expected
	}
	return result, nil
}

// expectedForumStats calculates forum counters from threads and posts.
func (repository Repository) expectedForumStats(ctx context.Context) (map[uuid.UUID]forumExpectation, error) {
	result, err := repository.emptyForumExpectations(ctx)
	if err != nil {
		return nil, err
	}
	if err := repository.applyForumThreadCounts(ctx, result); err != nil {
		return nil, err
	}
	if err := repository.applyForumPostCounts(ctx, result); err != nil {
		return nil, err
	}
	return result, nil
}

// expectedLikeStats calculates like counters from active likes.
func (repository Repository) expectedLikeStats(ctx context.Context) (map[uuid.UUID]int64, map[uuid.UUID]int64, error) {
	postLikes, err := repository.likeCounts(ctx, "post_id")
	if err != nil {
		return nil, nil, err
	}
	threadLikes, err := repository.likeCounts(ctx, "thread_id")
	if err != nil {
		return nil, nil, err
	}
	return postLikes, threadLikes, nil
}

// appendThreadDrift compares stored and expected thread counters.
func (repository Repository) appendThreadDrift(
	ctx context.Context,
	report *domain.CounterDriftReport,
	expected map[uuid.UUID]threadExpectation,
) error {
	var threads []threadCounterRow
	if err := repository.activeRows(ctx, "forum_threads").Find(&threads).Error; err != nil {
		return err
	}
	for _, thread := range threads {
		want := expected[thread.ID]
		report.Mismatches = appendDrift(report.Mismatches, "forum_thread", thread.ID, "post_count", want.PostCount, thread.PostCount)
		report.Mismatches = appendDrift(report.Mismatches, "forum_thread", thread.ID, "visible_post_count", want.VisiblePostCount, thread.VisiblePostCount)
		report.Mismatches = appendDrift(report.Mismatches, "forum_thread", thread.ID, "reply_count", want.ReplyCount, thread.ReplyCount)
		report.Mismatches = appendDrift(report.Mismatches, "forum_thread", thread.ID, "visible_reply_count", want.VisibleReplyCount, thread.VisibleReplyCount)
	}
	return nil
}

// appendForumDrift compares stored and expected forum counters.
func (repository Repository) appendForumDrift(
	ctx context.Context,
	report *domain.CounterDriftReport,
	expected map[uuid.UUID]forumExpectation,
) error {
	var stats []forumCounterRow
	if err := repository.store.DB(ctx).Table("forum_stats").Find(&stats).Error; err != nil {
		return err
	}
	for _, stat := range stats {
		want := expected[stat.ForumID]
		report.Mismatches = appendDrift(report.Mismatches, "forum_stats", stat.ForumID, "thread_count", want.ThreadCount, stat.ThreadCount)
		report.Mismatches = appendDrift(report.Mismatches, "forum_stats", stat.ForumID, "visible_thread_count", want.VisibleThreadCount, stat.VisibleThreadCount)
		report.Mismatches = appendDrift(report.Mismatches, "forum_stats", stat.ForumID, "post_count", want.PostCount, stat.PostCount)
		report.Mismatches = appendDrift(report.Mismatches, "forum_stats", stat.ForumID, "visible_post_count", want.VisiblePostCount, stat.VisiblePostCount)
	}
	return nil
}

// activeRows scopes a table to active source rows.
func (repository Repository) activeRows(ctx context.Context, table string) *gorm.DB {
	return repository.store.DB(ctx).Table(table).Where("deleted_at IS NULL")
}

// appendDrift appends one mismatch when expected and actual differ.
func appendDrift(
	items []domain.CounterDrift,
	objectType string,
	objectID uuid.UUID,
	field string,
	expected int64,
	actual int64,
) []domain.CounterDrift {
	if expected == actual {
		return items
	}
	return append(items, domain.CounterDrift{
		ObjectType: objectType,
		ObjectID:   objectID,
		Field:      field,
		Expected:   expected,
		Actual:     actual,
	})
}
