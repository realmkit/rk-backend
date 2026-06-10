// Package application implements forum use cases.
package application

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/niflaot/gamehub-go/module/forums/domain"
	"github.com/niflaot/gamehub-go/module/forums/port"
	"github.com/niflaot/gamehub-go/pkg/pagination"
)

// authorPostEditWindow is the default author self-edit window before admin configuration exists.
const authorPostEditWindow = 10 * time.Minute

// authorPostDeleteWindow is the default author self-delete window before admin configuration exists.
const authorPostDeleteWindow = 5 * time.Minute

// CreateThread creates a thread and opener post.
func (service Service) CreateThread(ctx context.Context, command port.CreateThreadCommand) (domain.Thread, domain.Post, error) {
	forum, err := service.forums.FindByID(ctx, command.ForumID)
	if err != nil {
		return domain.Thread{}, domain.Post{}, err
	}
	if !forum.Discussion() {
		return domain.Thread{}, domain.Post{}, port.ErrConflict
	}
	if err := service.requireThreadCreate(ctx, command.ActorUserID, forum.ID); err != nil {
		return domain.Thread{}, domain.Post{}, err
	}
	now := time.Now().UTC()
	threadID := uuid.New()
	postID := uuid.New()
	thread := domain.Thread{ID: threadID, ForumID: forum.ID, AuthorUserID: command.ActorUserID, OpenerPostID: postID, LatestPostID: postID, LatestPostAuthorUserID: command.ActorUserID, LatestPostAt: now, Title: command.Title, Slug: command.Slug, Status: forum.DefaultThreadStatus, StickyState: domain.StickyStateNormal, PostCount: 1, VisiblePostCount: 1, Version: 1}.Normalize()
	post := domain.Post{ID: postID, ThreadID: threadID, ForumID: forum.ID, AuthorUserID: command.ActorUserID, Sequence: 1, ContentDocumentJSON: command.ContentDocumentJSON, ContentText: contentText(command.ContentText, command.ContentDocumentJSON), ContentChecksum: checksum(command.ContentChecksum, command.ContentDocumentJSON), Version: 1}.Normalize()
	if err := thread.Validate(); err != nil {
		return domain.Thread{}, domain.Post{}, err
	}
	if err := post.Validate(); err != nil {
		return domain.Thread{}, domain.Post{}, err
	}
	var createdThread domain.Thread
	var createdPost domain.Post
	err = service.transactions.WithinTx(ctx, func(ctx context.Context) error {
		storedThread, err := service.threads.Create(ctx, thread)
		if err != nil {
			return err
		}
		storedPost, err := service.posts.Create(ctx, post, nil)
		if err != nil {
			return err
		}
		createdThread = storedThread
		createdPost = storedPost
		return service.clearTree(ctx)
	})
	return createdThread, createdPost, err
}

// GetThread returns one visible thread.
func (service Service) GetThread(ctx context.Context, actorUserID uuid.UUID, id uuid.UUID) (domain.Thread, error) {
	thread, err := service.threads.FindByID(ctx, id)
	if err != nil {
		return domain.Thread{}, err
	}
	if err := service.requireThreadView(ctx, actorUserID, thread); err != nil {
		return domain.Thread{}, err
	}
	if service.cache != nil {
		_ = service.cache.IncrementThreadView(ctx, thread.ID.String())
	}
	return thread, nil
}

// ListThreads lists visible threads.
func (service Service) ListThreads(ctx context.Context, actorUserID uuid.UUID, filter port.ThreadFilter, page pagination.Page) (pagination.Result[domain.Thread], error) {
	forum, err := service.forums.FindByID(ctx, filter.ForumID)
	if err != nil {
		return pagination.Result[domain.Thread]{}, err
	}
	visible, err := service.authorizer.VisibleForums(ctx, actorUserID, []uuid.UUID{forum.ID})
	if err != nil {
		return pagination.Result[domain.Thread]{}, err
	}
	if !visible[forum.ID] {
		return pagination.Result[domain.Thread]{}, port.ErrForbidden
	}
	return service.threads.List(ctx, filter, page)
}

// UpdateThreadTitle updates one thread title.
func (service Service) UpdateThreadTitle(ctx context.Context, command port.UpdateThreadTitleCommand) (domain.Thread, error) {
	thread, err := service.threads.FindByID(ctx, command.ThreadID)
	if err != nil {
		return domain.Thread{}, err
	}
	if command.ActorUserID != thread.AuthorUserID {
		if err := service.requireManageThreads(ctx, command.ActorUserID, thread.ForumID); err != nil {
			return domain.Thread{}, err
		}
	}
	thread.Title = command.Title
	thread.Slug = command.Slug
	thread = thread.Normalize()
	if err := thread.Validate(); err != nil {
		return domain.Thread{}, err
	}
	return service.threads.UpdateTitle(ctx, thread, command.ExpectedVersion)
}

// DeleteThread deletes one thread.
func (service Service) DeleteThread(ctx context.Context, command port.DeleteThreadCommand) error {
	thread, err := service.threads.FindByID(ctx, command.ThreadID)
	if err != nil {
		return err
	}
	if command.ActorUserID != thread.AuthorUserID {
		if err := service.requireManageThreads(ctx, command.ActorUserID, thread.ForumID); err != nil {
			return err
		}
	}
	if err := service.threads.Delete(ctx, command.ThreadID, command.ExpectedVersion); err != nil {
		return err
	}
	return service.clearTree(ctx)
}

// CreateReply creates a reply post.
func (service Service) CreateReply(ctx context.Context, command port.CreateReplyCommand) (domain.Post, error) {
	thread, err := service.threads.FindByID(ctx, command.ThreadID)
	if err != nil {
		return domain.Post{}, err
	}
	if !thread.Replyable() {
		return domain.Post{}, port.ErrConflict
	}
	if err := service.requireReply(ctx, command.ActorUserID, thread.ForumID); err != nil {
		return domain.Post{}, err
	}
	sequence, err := service.posts.NextSequence(ctx, thread.ID)
	if err != nil {
		return domain.Post{}, err
	}
	post := domain.Post{ID: uuid.New(), ThreadID: thread.ID, ForumID: thread.ForumID, AuthorUserID: command.ActorUserID, Sequence: sequence, ContentDocumentJSON: command.ContentDocumentJSON, ContentText: contentText(command.ContentText, command.ContentDocumentJSON), ContentChecksum: checksum(command.ContentChecksum, command.ContentDocumentJSON), Version: 1}.Normalize()
	if err := post.Validate(); err != nil {
		return domain.Post{}, err
	}
	references := prepareReferences(post.ID, append(extractReferences(command.ContentDocumentJSON), command.References...))
	for _, reference := range references {
		if err := reference.Validate(); err != nil {
			return domain.Post{}, err
		}
	}
	if err := service.validateReferences(ctx, command.ActorUserID, references); err != nil {
		return domain.Post{}, err
	}
	var created domain.Post
	err = service.transactions.WithinTx(ctx, func(ctx context.Context) error {
		stored, err := service.posts.Create(ctx, post, references)
		if err != nil {
			return err
		}
		created = stored
		return service.clearTree(ctx)
	})
	return created, err
}

// ListPosts lists posts for a visible thread.
func (service Service) ListPosts(ctx context.Context, actorUserID uuid.UUID, filter port.PostFilter, page pagination.Page) (pagination.Result[domain.Post], error) {
	thread, err := service.threads.FindByID(ctx, filter.ThreadID)
	if err != nil {
		return pagination.Result[domain.Post]{}, err
	}
	if filter.IncludeHidden {
		if err := service.requireManagePosts(ctx, actorUserID, thread.ForumID); err != nil {
			return pagination.Result[domain.Post]{}, err
		}
	} else if err := service.requireThreadView(ctx, actorUserID, thread); err != nil {
		return pagination.Result[domain.Post]{}, err
	}
	return service.posts.List(ctx, filter, page)
}

// GetPost returns one post.
func (service Service) GetPost(ctx context.Context, actorUserID uuid.UUID, id uuid.UUID) (domain.Post, error) {
	post, err := service.posts.FindByID(ctx, id)
	if err != nil {
		return domain.Post{}, err
	}
	if !post.Visible() {
		if err := service.requireManagePosts(ctx, actorUserID, post.ForumID); err != nil {
			return domain.Post{}, err
		}
		return post, nil
	}
	thread, err := service.threads.FindByID(ctx, post.ThreadID)
	if err != nil {
		return domain.Post{}, err
	}
	if err := service.requireThreadView(ctx, actorUserID, thread); err != nil {
		return domain.Post{}, err
	}
	return post, nil
}

// UpdatePost updates one post and records a revision.
func (service Service) UpdatePost(ctx context.Context, command port.UpdatePostCommand) (domain.Post, error) {
	current, err := service.posts.FindByID(ctx, command.PostID)
	if err != nil {
		return domain.Post{}, err
	}
	if command.ActorUserID == current.AuthorUserID {
		allowed, err := service.authorCanUpdatePost(ctx, current)
		if err != nil {
			return domain.Post{}, err
		}
		if !allowed {
			return domain.Post{}, port.ErrForbidden
		}
	} else {
		if err := service.requireManagePosts(ctx, command.ActorUserID, current.ForumID); err != nil {
			return domain.Post{}, err
		}
	}
	updated := current
	updated.ContentDocumentJSON = command.ContentDocumentJSON
	updated.ContentText = contentText(command.ContentText, command.ContentDocumentJSON)
	updated.ContentChecksum = checksum(command.ContentChecksum, command.ContentDocumentJSON)
	updated.EditCount++
	now := time.Now().UTC()
	updated.EditedAt = &now
	updated.EditedByUserID = &command.ActorUserID
	updated = updated.Normalize()
	if err := updated.Validate(); err != nil {
		return domain.Post{}, err
	}
	revision := domain.PostRevision{ID: uuid.New(), PostID: current.ID, EditedByUserID: command.ActorUserID, PreviousContentDocumentJSON: current.ContentDocumentJSON, PreviousContentText: current.ContentText, EditReason: strings.TrimSpace(command.EditReason)}
	return service.posts.UpdateWithRevision(ctx, updated, revision, command.ExpectedVersion)
}

// DeletePost deletes one post.
func (service Service) DeletePost(ctx context.Context, command port.DeletePostCommand) error {
	post, err := service.posts.FindByID(ctx, command.PostID)
	if err != nil {
		return err
	}
	if command.ActorUserID == post.AuthorUserID {
		allowed, err := service.authorCanDeletePost(ctx, post)
		if err != nil {
			return err
		}
		if !allowed {
			return port.ErrForbidden
		}
	} else {
		if err := service.requireManagePosts(ctx, command.ActorUserID, post.ForumID); err != nil {
			return err
		}
	}
	if err := service.posts.Delete(ctx, command.PostID, command.ExpectedVersion); err != nil {
		return err
	}
	return service.clearTree(ctx)
}

// ListPostRevisions lists post revisions for moderators.
func (service Service) ListPostRevisions(ctx context.Context, actorUserID uuid.UUID, postID uuid.UUID, page pagination.Page) (pagination.Result[domain.PostRevision], error) {
	post, err := service.posts.FindByID(ctx, postID)
	if err != nil {
		return pagination.Result[domain.PostRevision]{}, err
	}
	if err := service.requireManagePosts(ctx, actorUserID, post.ForumID); err != nil {
		return pagination.Result[domain.PostRevision]{}, err
	}
	return service.posts.ListRevisions(ctx, postID, page)
}

// LikePost likes one post idempotently.
func (service Service) LikePost(ctx context.Context, command port.LikePostCommand) (domain.PostLikeSummary, error) {
	if command.ActorUserID == uuid.Nil {
		return domain.PostLikeSummary{}, port.ErrForbidden
	}
	post, err := service.posts.FindByID(ctx, command.PostID)
	if err != nil {
		return domain.PostLikeSummary{}, err
	}
	if err := service.requireLikePosts(ctx, command.ActorUserID, post.ForumID); err != nil {
		return domain.PostLikeSummary{}, err
	}
	like := domain.PostLike{ID: uuid.New(), PostID: post.ID, ThreadID: post.ThreadID, ForumID: post.ForumID, UserID: command.ActorUserID, CreatedAt: time.Now().UTC()}
	if err := like.Validate(); err != nil {
		return domain.PostLikeSummary{}, err
	}
	changed, err := service.interactions.LikePost(ctx, like)
	if err != nil {
		return domain.PostLikeSummary{}, err
	}
	if changed {
		_ = service.clearInteractionCaches(ctx)
	}
	updated, err := service.posts.FindByID(ctx, post.ID)
	if err != nil {
		return domain.PostLikeSummary{}, err
	}
	return domain.PostLikeSummary{PostID: updated.ID, LikeCount: updated.LikeCount, LikedByActor: true}, nil
}

// UnlikePost unlikes one post idempotently.
func (service Service) UnlikePost(ctx context.Context, command port.UnlikePostCommand) (domain.PostLikeSummary, error) {
	if command.ActorUserID == uuid.Nil {
		return domain.PostLikeSummary{}, port.ErrForbidden
	}
	post, err := service.posts.FindByID(ctx, command.PostID)
	if err != nil {
		return domain.PostLikeSummary{}, err
	}
	if err := service.requireLikePosts(ctx, command.ActorUserID, post.ForumID); err != nil {
		return domain.PostLikeSummary{}, err
	}
	changed, err := service.interactions.UnlikePost(ctx, post.ID, command.ActorUserID)
	if err != nil {
		return domain.PostLikeSummary{}, err
	}
	if changed {
		_ = service.clearInteractionCaches(ctx)
	}
	updated, err := service.posts.FindByID(ctx, post.ID)
	if err != nil {
		return domain.PostLikeSummary{}, err
	}
	return domain.PostLikeSummary{PostID: updated.ID, LikeCount: updated.LikeCount, LikedByActor: false}, nil
}

// ListLatestPosts lists latest posts across visible forums.
func (service Service) ListLatestPosts(ctx context.Context, actorUserID uuid.UUID, forumID uuid.UUID, page pagination.Page) (pagination.Result[domain.LatestPostSummary], error) {
	forumIDs, err := service.visibleForumIDs(ctx, actorUserID, forumID)
	if err != nil {
		return pagination.Result[domain.LatestPostSummary]{}, err
	}
	key := latestPostsCacheKey(actorUserID, forumID, page)
	if service.cache != nil {
		if cached, ok, err := service.cache.GetLatestPosts(ctx, key); err == nil && ok {
			return cached, nil
		}
	}
	result, err := service.interactions.ListLatestPosts(ctx, port.LatestPostFilter{ForumIDs: forumIDs}, page)
	if err != nil {
		return pagination.Result[domain.LatestPostSummary]{}, err
	}
	if service.cache != nil {
		_ = service.cache.SetLatestPosts(ctx, key, result, widgetCacheTTL)
	}
	return result, nil
}

// ListMostLikedPosts lists most-liked posts for one visible forum.
func (service Service) ListMostLikedPosts(ctx context.Context, actorUserID uuid.UUID, forumID uuid.UUID, page pagination.Page) (pagination.Result[domain.MostLikedPost], error) {
	forumIDs, err := service.visibleForumIDs(ctx, actorUserID, forumID)
	if err != nil {
		return pagination.Result[domain.MostLikedPost]{}, err
	}
	if len(forumIDs) == 0 {
		return pagination.Result[domain.MostLikedPost]{}, port.ErrForbidden
	}
	key := mostLikedCacheKey(actorUserID, forumID, page)
	if service.cache != nil {
		if cached, ok, err := service.cache.GetMostLikedPosts(ctx, key); err == nil && ok {
			return cached, nil
		}
	}
	result, err := service.interactions.ListMostLikedPosts(ctx, port.MostLikedFilter{ForumID: forumIDs[0]}, page)
	if err != nil {
		return pagination.Result[domain.MostLikedPost]{}, err
	}
	if service.cache != nil {
		_ = service.cache.SetMostLikedPosts(ctx, key, result, widgetCacheTTL)
	}
	return result, nil
}

// MarkThreadRead stores read state for one thread.
func (service Service) MarkThreadRead(ctx context.Context, command port.MarkThreadReadCommand) (domain.ThreadReadState, error) {
	if command.ActorUserID == uuid.Nil {
		return domain.ThreadReadState{}, port.ErrForbidden
	}
	thread, err := service.threads.FindByID(ctx, command.ThreadID)
	if err != nil {
		return domain.ThreadReadState{}, err
	}
	if err := service.requireThreadView(ctx, command.ActorUserID, thread); err != nil {
		return domain.ThreadReadState{}, err
	}
	sequence := command.LastReadPostSequence
	if sequence < 1 {
		sequence = thread.VisiblePostCount
	}
	state := domain.ThreadReadState{ID: uuid.New(), UserID: command.ActorUserID, ForumID: thread.ForumID, ThreadID: thread.ID, LastReadPostSequence: sequence, LastReadAt: time.Now().UTC()}
	if err := state.Validate(); err != nil {
		return domain.ThreadReadState{}, err
	}
	return state, service.interactions.MarkThreadRead(ctx, state)
}

// MarkForumRead stores read state for visible threads in one forum.
func (service Service) MarkForumRead(ctx context.Context, command port.MarkForumReadCommand) error {
	if command.ActorUserID == uuid.Nil {
		return port.ErrForbidden
	}
	forumIDs, err := service.visibleForumIDs(ctx, command.ActorUserID, command.ForumID)
	if err != nil {
		return err
	}
	if len(forumIDs) == 0 {
		return port.ErrForbidden
	}
	return service.interactions.MarkForumRead(ctx, command.ActorUserID, forumIDs[0], time.Now().UTC())
}

// GetUnreadSummary returns unread totals for visible forums.
func (service Service) GetUnreadSummary(ctx context.Context, actorUserID uuid.UUID) (domain.UnreadSummary, error) {
	if actorUserID == uuid.Nil {
		return domain.UnreadSummary{}, port.ErrForbidden
	}
	forumIDs, err := service.visibleForumIDs(ctx, actorUserID, uuid.Nil)
	if err != nil {
		return domain.UnreadSummary{}, err
	}
	return service.interactions.UnreadSummary(ctx, actorUserID, forumIDs)
}

// Search searches visible forum content.
func (service Service) Search(ctx context.Context, command port.SearchCommand, page pagination.Page) (pagination.Result[domain.SearchResult], error) {
	query := strings.TrimSpace(command.Query)
	if len(query) < 2 || len(query) > 120 {
		return pagination.Result[domain.SearchResult]{}, domain.NewValidationError([]domain.Violation{{Field: "query", Message: "must be between 2 and 120 characters"}})
	}
	forumIDs, err := service.visibleForumIDs(ctx, command.ActorUserID, command.ForumID)
	if err != nil {
		return pagination.Result[domain.SearchResult]{}, err
	}
	if len(forumIDs) == 0 {
		return pagination.Result[domain.SearchResult]{Items: []domain.SearchResult{}}, nil
	}
	return service.operations.Search(ctx, port.SearchFilter{ForumIDs: forumIDs, Query: query}, page)
}

// VerifyStats reports stats counter drift.
func (service Service) VerifyStats(ctx context.Context) (domain.CounterDriftReport, error) {
	return service.operations.VerifyStats(ctx)
}

// RebuildStats repairs stats counters.
func (service Service) RebuildStats(ctx context.Context) (domain.CounterDriftReport, error) {
	return service.operations.RebuildStats(ctx)
}

// VerifyLikes reports like counter drift.
func (service Service) VerifyLikes(ctx context.Context) (domain.CounterDriftReport, error) {
	return service.operations.VerifyLikes(ctx)
}

// RebuildLikes repairs like counters.
func (service Service) RebuildLikes(ctx context.Context) (domain.CounterDriftReport, error) {
	return service.operations.RebuildLikes(ctx)
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
	increments := map[uuid.UUID]int64{}
	var total int64
	for key, value := range raw {
		id, err := uuid.Parse(key)
		if err != nil {
			continue
		}
		if value <= 0 {
			continue
		}
		increments[id] += value
		total += value
	}
	if len(increments) == 0 {
		return 0, nil
	}
	return total, service.operations.ApplyThreadViews(ctx, increments)
}

// ClearReadCache clears forum read caches.
func (service Service) ClearReadCache(ctx context.Context) error {
	if service.cache == nil {
		return nil
	}
	return service.cache.ClearAll(ctx)
}

// validateReferences verifies post and attachment references.
func (service Service) validateReferences(ctx context.Context, actorUserID uuid.UUID, references []domain.PostReference) error {
	for _, reference := range references {
		if reference.TargetPostID != nil {
			if _, err := service.GetPost(ctx, actorUserID, *reference.TargetPostID); err != nil {
				return err
			}
		}
		if reference.TargetAssetID != nil && service.assets != nil {
			exists, err := service.assets.AssetExists(ctx, *reference.TargetAssetID)
			if err != nil {
				return err
			}
			if !exists {
				return port.ErrNotFound
			}
		}
	}
	return nil
}

// authorCanUpdatePost reports whether an author can still edit a post.
func (service Service) authorCanUpdatePost(ctx context.Context, post domain.Post) (bool, error) {
	forum, err := service.forums.FindByID(ctx, post.ForumID)
	if err != nil {
		return false, err
	}
	if !insideAuthorWindow(post.CreatedAt, forum.AuthorPostEditWindowSeconds, authorPostEditWindow) {
		return false, nil
	}
	thread, err := service.threads.FindByID(ctx, post.ThreadID)
	if err != nil {
		return false, err
	}
	return thread.Replyable(), nil
}

// authorCanDeletePost reports whether an author can still delete a post.
func (service Service) authorCanDeletePost(ctx context.Context, post domain.Post) (bool, error) {
	forum, err := service.forums.FindByID(ctx, post.ForumID)
	if err != nil {
		return false, err
	}
	return insideAuthorWindow(post.CreatedAt, forum.AuthorPostDeleteWindowSeconds, authorPostDeleteWindow), nil
}

// insideAuthorWindow reports whether createdAt is inside configured author window.
func insideAuthorWindow(createdAt time.Time, configuredSeconds int, fallback time.Duration) bool {
	if createdAt.IsZero() || configuredSeconds < 0 {
		return false
	}
	window := time.Duration(configuredSeconds) * time.Second
	if configuredSeconds == 0 {
		window = fallback
	}
	return time.Since(createdAt) <= window
}

// requireThreadCreate verifies thread creation permission.
func (service Service) requireThreadCreate(ctx context.Context, actorUserID uuid.UUID, forumID uuid.UUID) error {
	allowed, err := service.authorizer.CanCreateThread(ctx, actorUserID, forumID)
	return decisionError(allowed, err)
}

// requireReply verifies reply permission.
func (service Service) requireReply(ctx context.Context, actorUserID uuid.UUID, forumID uuid.UUID) error {
	allowed, err := service.authorizer.CanReply(ctx, actorUserID, forumID)
	return decisionError(allowed, err)
}

// requireLikePosts verifies like permission.
func (service Service) requireLikePosts(ctx context.Context, actorUserID uuid.UUID, forumID uuid.UUID) error {
	allowed, err := service.authorizer.CanLikePosts(ctx, actorUserID, forumID)
	return decisionError(allowed, err)
}

// requireManageThreads verifies thread management permission.
func (service Service) requireManageThreads(ctx context.Context, actorUserID uuid.UUID, forumID uuid.UUID) error {
	allowed, err := service.authorizer.CanManageThreads(ctx, actorUserID, forumID)
	return decisionError(allowed, err)
}

// requireManagePosts verifies post management permission.
func (service Service) requireManagePosts(ctx context.Context, actorUserID uuid.UUID, forumID uuid.UUID) error {
	allowed, err := service.authorizer.CanManagePosts(ctx, actorUserID, forumID)
	return decisionError(allowed, err)
}

// requireThreadView verifies thread visibility.
func (service Service) requireThreadView(ctx context.Context, actorUserID uuid.UUID, thread domain.Thread) error {
	visible, err := service.authorizer.VisibleForums(ctx, actorUserID, []uuid.UUID{thread.ForumID})
	if err != nil {
		return err
	}
	if visible[thread.ForumID] && thread.Visible() {
		return nil
	}
	if actorUserID == thread.AuthorUserID {
		return nil
	}
	return service.requireManageThreads(ctx, actorUserID, thread.ForumID)
}

// visibleForumIDs returns visible forum IDs for widget reads.
func (service Service) visibleForumIDs(ctx context.Context, actorUserID uuid.UUID, forumID uuid.UUID) ([]uuid.UUID, error) {
	forumIDs := []uuid.UUID{}
	if forumID != uuid.Nil {
		forumIDs = append(forumIDs, forumID)
	} else {
		forums, err := service.forums.List(ctx, port.ForumFilter{Status: domain.ForumStatusActive}, port.Page{Limit: 1000})
		if err != nil {
			return nil, err
		}
		for _, forum := range forums.Items {
			forumIDs = append(forumIDs, forum.ID)
		}
	}
	visible, err := service.authorizer.VisibleForums(ctx, actorUserID, forumIDs)
	if err != nil {
		return nil, err
	}
	result := make([]uuid.UUID, 0, len(forumIDs))
	for _, id := range forumIDs {
		if visible[id] {
			result = append(result, id)
		}
	}
	return result, nil
}

// clearInteractionCaches clears caches affected by interaction writes.
func (service Service) clearInteractionCaches(ctx context.Context) error {
	if service.cache == nil {
		return nil
	}
	if err := service.cache.ClearLatestPosts(ctx); err != nil {
		return err
	}
	return service.cache.ClearMostLikedPosts(ctx)
}

// decisionError maps authorization decision to error.
func decisionError(allowed bool, err error) error {
	if err != nil {
		return err
	}
	if !allowed {
		return port.ErrForbidden
	}
	return nil
}

// checksum returns provided checksum or computes one from content.
func checksum(provided string, content []byte) string {
	if strings.TrimSpace(provided) != "" {
		return strings.TrimSpace(provided)
	}
	sum := sha256.Sum256(content)
	return hex.EncodeToString(sum[:])
}

// latestPostsCacheKey returns a cache key for latest-post widgets.
func latestPostsCacheKey(actorUserID uuid.UUID, forumID uuid.UUID, page pagination.Page) string {
	return "forums:latest:v1:" + widgetScope(forumID) + ":" + actorScope(actorUserID) + ":" + page.Cursor + ":" + strconv.Itoa(page.Limit)
}

// mostLikedCacheKey returns a cache key for most-liked widgets.
func mostLikedCacheKey(actorUserID uuid.UUID, forumID uuid.UUID, page pagination.Page) string {
	return "forums:most-liked:v1:" + forumID.String() + ":all:" + actorScope(actorUserID) + ":" + page.Cursor + ":" + strconv.Itoa(page.Limit)
}

// widgetScope returns cache scope for global or forum-specific widgets.
func widgetScope(forumID uuid.UUID) string {
	if forumID == uuid.Nil {
		return "global:all"
	}
	return "forum:" + forumID.String()
}

// actorScope returns cache scope for actor visibility.
func actorScope(actorUserID uuid.UUID) string {
	if actorUserID == uuid.Nil {
		return "anonymous"
	}
	return "user:" + actorUserID.String()
}

// prepareReferences fills source IDs on references.
func prepareReferences(sourcePostID uuid.UUID, references []domain.PostReference) []domain.PostReference {
	prepared := make([]domain.PostReference, 0, len(references))
	for _, reference := range references {
		reference.ID = uuid.New()
		reference.SourcePostID = sourcePostID
		prepared = append(prepared, reference)
	}
	return prepared
}

// contentText returns explicit text or extracts text from document.
func contentText(explicit string, document []byte) string {
	if strings.TrimSpace(explicit) != "" {
		return strings.TrimSpace(explicit)
	}
	var payload any
	if err := json.Unmarshal(document, &payload); err != nil {
		return ""
	}
	var parts []string
	collectText(payload, &parts)
	return strings.TrimSpace(strings.Join(parts, " "))
}

// collectText recursively collects ProseMirror text nodes.
func collectText(value any, parts *[]string) {
	switch typed := value.(type) {
	case map[string]any:
		if text, ok := typed["text"].(string); ok && strings.TrimSpace(text) != "" {
			*parts = append(*parts, strings.TrimSpace(text))
		}
		for _, nested := range typed {
			collectText(nested, parts)
		}
	case []any:
		for _, nested := range typed {
			collectText(nested, parts)
		}
	}
}

// extractReferences returns references from known ProseMirror node shapes.
func extractReferences(document []byte) []domain.PostReference {
	var payload any
	if err := json.Unmarshal(document, &payload); err != nil {
		return nil
	}
	var references []domain.PostReference
	collectReferences(payload, &references)
	return references
}

// collectReferences recursively collects supported references.
func collectReferences(value any, references *[]domain.PostReference) {
	switch typed := value.(type) {
	case map[string]any:
		appendNodeReference(typed, references)
		for _, nested := range typed {
			collectReferences(nested, references)
		}
	case []any:
		for _, nested := range typed {
			collectReferences(nested, references)
		}
	}
}

// appendNodeReference appends references for one node or mark object.
func appendNodeReference(node map[string]any, references *[]domain.PostReference) {
	nodeType, _ := node["type"].(string)
	attrs, _ := node["attrs"].(map[string]any)
	switch nodeType {
	case "mention":
		if id := uuidFromAttr(attrs, "id", "user_id"); id != uuid.Nil {
			*references = append(*references, domain.PostReference{TargetUserID: &id, ReferenceType: domain.ReferenceMention})
		}
	case "attachment":
		if id := uuidFromAttr(attrs, "asset_id", "id"); id != uuid.Nil {
			*references = append(*references, domain.PostReference{TargetAssetID: &id, ReferenceType: domain.ReferenceAttachment})
		}
	case "quote":
		if id := uuidFromAttr(attrs, "post_id", "id"); id != uuid.Nil {
			excerpt, _ := attrs["excerpt"].(string)
			*references = append(*references, domain.PostReference{TargetPostID: &id, ReferenceType: domain.ReferenceQuote, QuoteExcerpt: excerpt})
		}
	case "reply_to":
		if id := uuidFromAttr(attrs, "post_id", "id"); id != uuid.Nil {
			*references = append(*references, domain.PostReference{TargetPostID: &id, ReferenceType: domain.ReferenceReplyTo})
		}
	case "link":
		if href, _ := attrs["href"].(string); strings.TrimSpace(href) != "" {
			*references = append(*references, domain.PostReference{ReferenceType: domain.ReferenceLink, LinkURL: strings.TrimSpace(href)})
		}
	}
}

// uuidFromAttr returns the first UUID found in attrs.
func uuidFromAttr(attrs map[string]any, keys ...string) uuid.UUID {
	for _, key := range keys {
		if raw, ok := attrs[key].(string); ok {
			parsed, err := uuid.Parse(raw)
			if err == nil {
				return parsed
			}
		}
	}
	return uuid.Nil
}
