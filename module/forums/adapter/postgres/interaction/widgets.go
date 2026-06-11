package interaction

import (
	"context"

	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/module/forums/domain"
	"github.com/realmkit/rk-backend/module/forums/port"
	"github.com/realmkit/rk-backend/pkg/pagination"
)

// ListLatestPosts returns latest visible post summaries.
func (repository Repository) ListLatestPosts(
	ctx context.Context,
	filter port.LatestPostFilter,
	page pagination.Page,
) (pagination.Result[domain.LatestPostSummary], error) {
	if len(filter.ForumIDs) == 0 {
		return pagination.Result[domain.LatestPostSummary]{
			Items: []domain.LatestPostSummary{},
		}, nil
	}
	var rows []latestPostRow
	err := repository.store.DB(ctx).
		Table("forum_posts AS p").
		Select(latestPostSelect).
		Joins("JOIN forum_threads AS t ON t.id = p.thread_id AND t.deleted_at IS NULL").
		Where(
			"p.forum_id IN ? AND p.deleted_at IS NULL AND p.status IN ? AND t.status IN ?",
			filter.ForumIDs,
			visiblePostStatuses(),
			visibleThreadStatuses(),
		).
		Order("p.created_at DESC, p.id ASC").
		Limit(page.Limit + 1).
		Find(&rows).Error
	if err != nil {
		return pagination.Result[domain.LatestPostSummary]{}, err
	}
	return latestPostPage(rows, page.Limit), nil
}

// ListMostLikedPosts returns most-liked visible posts.
func (repository Repository) ListMostLikedPosts(
	ctx context.Context,
	filter port.MostLikedFilter,
	page pagination.Page,
) (pagination.Result[domain.MostLikedPost], error) {
	var rows []mostLikedPostRow
	err := repository.store.DB(ctx).
		Table("forum_posts AS p").
		Select(mostLikedPostSelect).
		Joins("JOIN forum_threads AS t ON t.id = p.thread_id AND t.deleted_at IS NULL").
		Where(
			"p.forum_id = ? AND p.deleted_at IS NULL AND p.status IN ? AND t.status IN ?",
			filter.ForumID,
			visiblePostStatuses(),
			visibleThreadStatuses(),
		).
		Order("p.like_count DESC, p.created_at DESC, p.id ASC").
		Limit(page.Limit + 1).
		Find(&rows).Error
	if err != nil {
		return pagination.Result[domain.MostLikedPost]{}, err
	}
	return mostLikedPostPage(rows, page.Limit), nil
}

// UnreadSummary returns unread counts for visible forums.
func (repository Repository) UnreadSummary(
	ctx context.Context,
	userID uuid.UUID,
	forumIDs []uuid.UUID,
) (domain.UnreadSummary, error) {
	summary := domain.UnreadSummary{UserID: userID, Forums: []domain.ForumUnreadSummary{}}
	if len(forumIDs) == 0 {
		return summary, nil
	}
	var rows []unreadForumRow
	err := repository.store.DB(ctx).
		Table("forum_threads AS t").
		Select("t.forum_id, COUNT(*) AS unread_thread_count").
		Joins("LEFT JOIN forum_thread_read_states AS rs ON rs.thread_id = t.id AND rs.user_id = ?", userID).
		Where(
			"t.forum_id IN ? AND t.deleted_at IS NULL AND t.status IN ? "+
				"AND COALESCE(rs.last_read_post_sequence, 0) < t.visible_post_count",
			forumIDs,
			visibleThreadStatuses(),
		).
		Group("t.forum_id").
		Find(&rows).Error
	if err != nil {
		return domain.UnreadSummary{}, err
	}
	for _, row := range rows {
		summary.UnreadThreadCount += row.UnreadThreadCount
		summary.Forums = append(summary.Forums, domain.ForumUnreadSummary{
			ForumID:           row.ForumID,
			UnreadThreadCount: row.UnreadThreadCount,
		})
	}
	return summary, nil
}

// Widget query projections.
const (
	latestPostSelect = "p.forum_id, p.thread_id, p.id AS post_id, p.author_user_id, " +
		"p.sequence, t.title AS thread_title, t.slug AS thread_slug, " +
		"p.content_text AS excerpt, p.created_at"
	mostLikedPostSelect = latestPostSelect + ", p.like_count"
)
