package operations

import (
	"context"

	"github.com/google/uuid"
)

// emptyThreadExpectations initializes expectation rows for active threads.
func (repository Repository) emptyThreadExpectations(ctx context.Context) (map[uuid.UUID]threadExpectation, error) {
	var threads []threadIDRow
	if err := repository.activeRows(ctx, "forum_threads").Find(&threads).Error; err != nil {
		return nil, err
	}
	result := make(map[uuid.UUID]threadExpectation, len(threads))
	for _, thread := range threads {
		result[thread.ID] = threadExpectation{}
	}
	return result, nil
}

// emptyForumExpectations initializes expectation rows for stats rows.
func (repository Repository) emptyForumExpectations(ctx context.Context) (map[uuid.UUID]forumExpectation, error) {
	var stats []forumIDRow
	if err := repository.store.DB(ctx).Table("forum_stats").Find(&stats).Error; err != nil {
		return nil, err
	}
	result := make(map[uuid.UUID]forumExpectation, len(stats))
	for _, stat := range stats {
		result[stat.ForumID] = forumExpectation{}
	}
	return result, nil
}

// applyForumThreadCounts applies source thread counts to forum expectations.
func (repository Repository) applyForumThreadCounts(
	ctx context.Context,
	result map[uuid.UUID]forumExpectation,
) error {
	var rows []forumThreadCounterRow
	err := repository.store.DB(ctx).
		Table("forum_threads").
		Select(forumThreadCounterSelect, visibleThreadStatuses()).
		Where("deleted_at IS NULL").
		Group("forum_id").
		Find(&rows).Error
	if err != nil {
		return err
	}
	for _, row := range rows {
		expected := result[row.ForumID]
		expected.ThreadCount = row.ThreadCount
		expected.VisibleThreadCount = row.VisibleThreadCount
		result[row.ForumID] = expected
	}
	return nil
}

// applyForumPostCounts applies source post counts to forum expectations.
func (repository Repository) applyForumPostCounts(
	ctx context.Context,
	result map[uuid.UUID]forumExpectation,
) error {
	var rows []forumPostCounterRow
	err := repository.store.DB(ctx).
		Table("forum_posts").
		Select(forumPostCounterSelect, visiblePostStatuses()).
		Where("deleted_at IS NULL").
		Group("forum_id").
		Find(&rows).Error
	if err != nil {
		return err
	}
	for _, row := range rows {
		expected := result[row.ForumID]
		expected.PostCount = row.PostCount
		expected.VisiblePostCount = row.VisiblePostCount
		result[row.ForumID] = expected
	}
	return nil
}

// likeCounts groups active likes by target column.
func (repository Repository) likeCounts(ctx context.Context, column string) (map[uuid.UUID]int64, error) {
	var rows []likeCounterRow
	err := repository.store.DB(ctx).
		Table("forum_post_likes").
		Select(column + " AS id, COUNT(*) AS count").
		Where("deleted_at IS NULL").
		Group(column).
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	result := map[uuid.UUID]int64{}
	for _, row := range rows {
		result[row.ID] = row.Count
	}
	return result, nil
}
