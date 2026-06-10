// Package operations adapts forum operational repositories to PostgreSQL.
package operations

import (
	"context"
	"sort"
	"strings"

	"github.com/google/uuid"
	"github.com/niflaot/gamehub-go/module/forums/domain"
	"github.com/niflaot/gamehub-go/module/forums/port"
	"github.com/niflaot/gamehub-go/pkg/orm"
	"github.com/niflaot/gamehub-go/pkg/pagination"
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
	queryText := strings.TrimSpace(filter.Query)
	searchArgument := "%" + strings.ToLower(queryText) + "%"
	threadCondition := "LOWER(title) LIKE ?"
	postCondition := "LOWER(p.content_text) LIKE ?"
	if repository.store.DB(ctx).Dialector.Name() == "postgres" {
		searchArgument = queryText
		threadCondition = "to_tsvector('simple', title) @@ plainto_tsquery('simple', ?)"
		postCondition = "to_tsvector('simple', p.content_text) @@ plainto_tsquery('simple', ?)"
	}
	threads, err := repository.searchThreads(ctx, filter.ForumIDs, threadCondition, searchArgument, page)
	if err != nil {
		return pagination.Result[domain.SearchResult]{}, err
	}
	posts, err := repository.searchPosts(ctx, filter.ForumIDs, postCondition, searchArgument, page)
	if err != nil {
		return pagination.Result[domain.SearchResult]{}, err
	}
	return searchPage(threads, posts, page.Limit), nil
}

// searchThreads returns matching thread title rows.
func (repository Repository) searchThreads(
	ctx context.Context,
	forumIDs []uuid.UUID,
	condition string,
	argument string,
	page pagination.Page,
) ([]threadSearchRow, error) {
	var rows []threadSearchRow
	err := repository.store.DB(ctx).
		Table("forum_threads").
		Select("id, forum_id, author_user_id, title, slug, created_at").
		Where("forum_id IN ? AND deleted_at IS NULL AND status IN ?", forumIDs, visibleThreadStatuses()).
		Where(condition, argument).
		Limit(page.Limit + 1).
		Find(&rows).Error
	return rows, err
}

// searchPosts returns matching post content rows.
func (repository Repository) searchPosts(
	ctx context.Context,
	forumIDs []uuid.UUID,
	condition string,
	argument string,
	page pagination.Page,
) ([]postSearchRow, error) {
	var rows []postSearchRow
	err := repository.store.DB(ctx).
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
		Limit(page.Limit + 1).
		Find(&rows).Error
	return rows, err
}

// searchPage merges thread and post search rows into one deterministic page.
func searchPage(
	threads []threadSearchRow,
	posts []postSearchRow,
	limit int,
) pagination.Result[domain.SearchResult] {
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
		next = results[limit-1].ThreadID.String()
		results = results[:limit]
	}
	return pagination.Result[domain.SearchResult]{Items: results, NextCursor: next}
}
