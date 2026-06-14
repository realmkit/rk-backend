// Package operations adapts forum operational repositories to PostgreSQL.
package operations

import (
	"context"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/module/forums/domain"
	"github.com/realmkit/rk-backend/module/forums/port"
	"github.com/realmkit/rk-backend/pkg/orm"
	"github.com/realmkit/rk-backend/pkg/pagination"
	"github.com/realmkit/rk-backend/pkg/search"
	"gorm.io/gorm"
)

// Repository runs forum search, repair, and counter flushes in PostgreSQL.
type Repository struct {
	store orm.Store
}

// NewRepository creates an operations repository.
func NewRepository(store orm.Store) Repository {
	return Repository{store: store}
}

// Search returns visible search results from PostgreSQL.
func (repository Repository) Search(
	ctx context.Context,
	filter port.SearchFilter,
	page pagination.Page,
) (pagination.Result[domain.SearchResult], error) {
	if len(filter.ForumIDs) == 0 {
		return pagination.Result[domain.SearchResult]{Items: []domain.SearchResult{}}, nil
	}
	resultSort := search.Sort{Key: "created_at", Direction: search.DirectionDesc}
	filterHash := search.HashFilter(filter.ForumIDs, filter.Query)
	cursor, hasCursor, err := search.RequireCursor(page.Cursor, filterHash, resultSort)
	if err != nil {
		return pagination.Result[domain.SearchResult]{}, err
	}
	queryText := strings.TrimSpace(filter.Query)
	searchArgument := "%" + strings.ToLower(queryText) + "%"
	threadCondition := "LOWER(title) LIKE ?"
	postCondition := "LOWER(p.content_text) LIKE ?"
	if repository.store.DB(ctx).Dialector.Name() == "postgres" {
		searchArgument = queryText
		threadCondition = "to_tsvector('simple', title) @@ plainto_tsquery('simple', ?)"
		postCondition = "to_tsvector('simple', p.content_text) @@ plainto_tsquery('simple', ?)"
	}
	threads, err := repository.searchThreads(ctx, filter.ForumIDs, threadCondition, searchArgument, page, cursor, hasCursor)
	if err != nil {
		return pagination.Result[domain.SearchResult]{}, err
	}
	posts, err := repository.searchPosts(ctx, filter.ForumIDs, postCondition, searchArgument, page, cursor, hasCursor)
	if err != nil {
		return pagination.Result[domain.SearchResult]{}, err
	}
	return searchPage(threads, posts, page.Limit, filterHash, resultSort)
}

// searchThreads returns matching thread title rows.
func (repository Repository) searchThreads(
	ctx context.Context,
	forumIDs []uuid.UUID,
	condition string,
	argument string,
	page pagination.Page,
	cursor search.Cursor,
	hasCursor bool,
) ([]threadSearchRow, error) {
	var rows []threadSearchRow
	query := repository.store.DB(ctx).
		Table("forum_threads").
		Select("id, forum_id, author_user_id, title, slug, created_at").
		Where("forum_id IN ? AND deleted_at IS NULL AND status IN ?", forumIDs, visibleThreadStatuses()).
		Where(condition, argument).
		Limit(page.Limit + 1)
	query = applySearchCursor(query, cursor, hasCursor, "created_at", "id")
	err := query.Find(&rows).Error
	return rows, err
}

// searchPosts returns matching post content rows.
func (repository Repository) searchPosts(
	ctx context.Context,
	forumIDs []uuid.UUID,
	condition string,
	argument string,
	page pagination.Page,
	cursor search.Cursor,
	hasCursor bool,
) ([]postSearchRow, error) {
	var rows []postSearchRow
	query := repository.store.DB(ctx).
		Table("forum_posts AS p").
		Select(searchPostSelect).
		Joins("JOIN forum_threads AS t ON t.id = p.thread_id AND t.deleted_at IS NULL").
		Where(
			"p.forum_id IN ? AND p.deleted_at IS NULL AND p.status IN ? AND t.status IN ?",
			forumIDs,
			visiblePostStatuses(),
			visibleThreadStatuses(),
		).
		Where(condition, argument).
		Limit(page.Limit + 1)
	query = applySearchCursor(query, cursor, hasCursor, "p.created_at", "p.id")
	err := query.Find(&rows).Error
	return rows, err
}

// searchPage merges thread and post search rows into one deterministic page.
func searchPage(
	threads []threadSearchRow,
	posts []postSearchRow,
	limit int,
	filterHash string,
	resultSort search.Sort,
) (pagination.Result[domain.SearchResult], error) {
	results := make([]domain.SearchResult, 0, len(threads)+len(posts))
	for _, thread := range threads {
		results = append(results, thread.searchResult())
	}
	for _, post := range posts {
		results = append(results, post.searchResult())
	}
	sort.Slice(results, func(i int, j int) bool {
		if results[i].CreatedAt.Equal(results[j].CreatedAt) {
			return results[i].ThreadID.String() < results[j].ThreadID.String()
		}
		return results[i].CreatedAt.After(results[j].CreatedAt)
	})
	next := ""
	if len(results) > limit {
		cursor, err := searchResultCursor(results[limit-1], filterHash, resultSort)
		if err != nil {
			return pagination.Result[domain.SearchResult]{}, err
		}
		next = cursor
		results = results[:limit]
	}
	return pagination.Result[domain.SearchResult]{Items: results, NextCursor: next}, nil
}

// applySearchCursor applies a shared created-at search cursor.
func applySearchCursor(query *gorm.DB, cursor search.Cursor, ok bool, createdColumn string, idColumn string) *gorm.DB {
	if !ok || len(cursor.Values) == 0 {
		return query
	}
	createdAt, err := time.Parse(time.RFC3339Nano, cursor.Values[0])
	id, parseErr := uuid.Parse(cursor.ID)
	if err != nil || parseErr != nil {
		return query.Where("1 = 0")
	}
	return query.Where(createdColumn+" < ? OR ("+createdColumn+" = ? AND "+idColumn+" > ?)", createdAt, createdAt, id)
}

// searchResultCursor returns the next cursor for a merged search result.
func searchResultCursor(result domain.SearchResult, filterHash string, sort search.Sort) (string, error) {
	id := result.ThreadID
	if result.PostID != nil {
		id = *result.PostID
	}
	return search.EncodeCursor(search.Cursor{
		FilterHash: filterHash,
		Sort:       sort.Key,
		Direction:  sort.Direction,
		Values:     []string{result.CreatedAt.Format(time.RFC3339Nano)},
		ID:         id.String(),
	})
}
