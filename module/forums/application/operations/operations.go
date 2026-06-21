package operations

import (
	"context"

	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/module/forums/domain"
	"github.com/realmkit/rk-backend/module/forums/port"
	"github.com/realmkit/rk-backend/pkg/pagination"
	"github.com/realmkit/rk-backend/pkg/search"
)

// Search searches visible forum content.
func (service Service) Search(
	ctx context.Context,
	command port.SearchCommand,
	page pagination.Page,
) (pagination.Result[domain.SearchResult], error) {
	query, err := search.NewTextQuery(command.Query, search.QueryOptions{})
	if err != nil || query.Empty() {
		return pagination.Result[domain.SearchResult]{}, domain.NewValidationError([]domain.Violation{
			{Field: "query", Message: "must be between 2 and 120 characters"},
		})
	}
	forumIDs, err := service.visibleForumIDs(ctx, command.ActorUserID, command.ForumID)
	if err != nil {
		return pagination.Result[domain.SearchResult]{}, err
	}
	if len(forumIDs) == 0 {
		return pagination.Result[domain.SearchResult]{Items: []domain.SearchResult{}}, nil
	}
	filter := port.SearchFilter{
		ForumIDs: forumIDs,
		Query:    query.String(),
	}
	return service.operations.Search(ctx, filter, page)
}

// VerifyStats reports stats counter drift.
func (service Service) VerifyStats(ctx context.Context) (domain.CounterDriftReport, error) {
	return service.operations.VerifyStats(ctx)
}

// RebuildStats repairs stats counters.
func (service Service) RebuildStats(ctx context.Context) (domain.CounterDriftReport, error) {
	report, err := service.operations.RebuildStats(ctx)
	if err != nil {
		return domain.CounterDriftReport{}, err
	}
	return report, service.publishOperationEvent(ctx, "forums.stats.rebuilt", driftPayload(report))
}

// VerifyLikes reports like counter drift.
func (service Service) VerifyLikes(ctx context.Context) (domain.CounterDriftReport, error) {
	return service.operations.VerifyLikes(ctx)
}

// RebuildLikes repairs like counters.
func (service Service) RebuildLikes(ctx context.Context) (domain.CounterDriftReport, error) {
	report, err := service.operations.RebuildLikes(ctx)
	if err != nil {
		return domain.CounterDriftReport{}, err
	}
	return report, service.publishOperationEvent(ctx, "forums.likes.rebuilt", driftPayload(report))
}

// FlushThreadViews persists buffered view counters.
func (service Service) FlushThreadViews(ctx context.Context) (int64, error) {
	if service.cache == nil {
		return 0, nil
	}
	raw, err := service.cache.DrainThreadViews(ctx)
	if err != nil {
		return 0, err
	}
	increments, total := threadViewIncrements(raw)
	if len(increments) == 0 {
		return 0, nil
	}
	if err := service.operations.ApplyThreadViews(ctx, increments); err != nil {
		return 0, err
	}
	return total, service.publishOperationEvent(
		ctx,
		"forums.views.flushed",
		map[string]any{"thread_count": len(increments), "view_count": total},
	)
}

// ClearReadCache clears forum read caches.
func (service Service) ClearReadCache(ctx context.Context) error {
	if service.cache == nil {
		return nil
	}
	return service.cache.ClearAll(ctx)
}

// threadViewIncrements supports package behavior.
func threadViewIncrements(raw map[string]int64) (map[uuid.UUID]int64, int64) {
	increments := map[uuid.UUID]int64{}
	var total int64
	for key, value := range raw {
		id, err := uuid.Parse(key)
		if err != nil || value <= 0 {
			continue
		}
		increments[id] += value
		total += value
	}
	return increments, total
}
