// Package postgres adapts forum repositories to PostgreSQL through GORM.
package postgres

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

// ThreadRepository stores forum threads in PostgreSQL.
type ThreadRepository struct {
	store orm.Store
}

// NewThreadRepository creates a thread repository.
func NewThreadRepository(store orm.Store) ThreadRepository {
	return ThreadRepository{store: store}
}

// Create stores a thread.
func (repository ThreadRepository) Create(ctx context.Context, thread domain.Thread) (domain.Thread, error) {
	model := threadModelFromDomain(thread)
	if err := repository.store.DB(ctx).Create(&model).Error; err != nil {
		return domain.Thread{}, port.ErrConflict
	}
	if err := repository.incrementForumStats(ctx, thread); err != nil {
		return domain.Thread{}, err
	}
	return threadFromModel(model), nil
}

// FindByID returns one thread.
func (repository ThreadRepository) FindByID(ctx context.Context, id uuid.UUID) (domain.Thread, error) {
	var model ThreadModel
	if err := repository.store.DB(ctx).First(&model, "id = ?", id).Error; err != nil {
		return domain.Thread{}, mapError(err)
	}
	return threadFromModel(model), nil
}

// List returns matching threads.
func (repository ThreadRepository) List(ctx context.Context, filter port.ThreadFilter, page pagination.Page) (pagination.Result[domain.Thread], error) {
	query := repository.store.DB(ctx).Model(&ThreadModel{}).Where("forum_id = ?", filter.ForumID).Order("sticky_state desc, sticky_order asc, latest_post_at desc, id asc").Limit(page.Limit + 1)
	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}
	if filter.Section == "sticky" {
		query = query.Where("sticky_state <> ?", domain.StickyStateNormal)
	}
	if filter.Section == "normal" {
		query = query.Where("sticky_state = ?", domain.StickyStateNormal)
	}
	var models []ThreadModel
	if err := query.Find(&models).Error; err != nil {
		return pagination.Result[domain.Thread]{}, err
	}
	return threadPage(models, page.Limit), nil
}

// UpdateTitle updates thread title fields.
func (repository ThreadRepository) UpdateTitle(ctx context.Context, thread domain.Thread, expectedVersion uint64) (domain.Thread, error) {
	result := repository.store.DB(ctx).Model(&ThreadModel{}).Where("id = ? AND version = ?", thread.ID, expectedVersion).Updates(map[string]any{"title": thread.Title, "slug": string(thread.Slug), "version": expectedVersion + 1})
	if result.Error != nil {
		return domain.Thread{}, result.Error
	}
	if result.RowsAffected == 0 {
		return domain.Thread{}, port.ErrPreconditionFailed
	}
	return repository.FindByID(ctx, thread.ID)
}

// Delete soft deletes a thread.
func (repository ThreadRepository) Delete(ctx context.Context, id uuid.UUID, expectedVersion uint64) error {
	result := repository.store.DB(ctx).Where("id = ? AND version = ?", id, expectedVersion).Delete(&ThreadModel{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return port.ErrPreconditionFailed
	}
	return nil
}

// incrementForumStats updates forum stats after thread creation.
func (repository ThreadRepository) incrementForumStats(ctx context.Context, thread domain.Thread) error {
	updates := map[string]any{"thread_count": gorm.Expr("thread_count + ?", 1), "post_count": gorm.Expr("post_count + ?", thread.PostCount), "latest_thread_id": thread.ID, "latest_post_id": thread.LatestPostID, "latest_post_author_user_id": thread.LatestPostAuthorUserID, "latest_post_at": thread.LatestPostAt, "updated_at": time.Now().UTC()}
	if thread.Visible() {
		updates["visible_thread_count"] = gorm.Expr("visible_thread_count + ?", 1)
		updates["visible_post_count"] = gorm.Expr("visible_post_count + ?", thread.VisiblePostCount)
	}
	return repository.store.DB(ctx).Model(&StatsModel{}).Where("forum_id = ?", thread.ForumID).Updates(updates).Error
}

// PostRepository stores forum posts in PostgreSQL.
type PostRepository struct {
	store orm.Store
}

// NewPostRepository creates a post repository.
func NewPostRepository(store orm.Store) PostRepository {
	return PostRepository{store: store}
}

// Create stores a post with references.
func (repository PostRepository) Create(ctx context.Context, post domain.Post, references []domain.PostReference) (domain.Post, error) {
	model := postModelFromDomain(post)
	if err := repository.store.DB(ctx).Create(&model).Error; err != nil {
		return domain.Post{}, port.ErrConflict
	}
	created := postFromModel(model)
	if created.Sequence > 1 {
		if err := repository.incrementThreadAndForum(ctx, created); err != nil {
			return domain.Post{}, err
		}
	}
	for _, reference := range references {
		if reference.CreatedAt.IsZero() {
			reference.CreatedAt = time.Now().UTC()
		}
		model := referenceModelFromDomain(reference)
		if err := repository.store.DB(ctx).Create(&model).Error; err != nil {
			return domain.Post{}, err
		}
	}
	return created, nil
}

// FindByID returns one post.
func (repository PostRepository) FindByID(ctx context.Context, id uuid.UUID) (domain.Post, error) {
	var model PostModel
	if err := repository.store.DB(ctx).First(&model, "id = ?", id).Error; err != nil {
		return domain.Post{}, mapError(err)
	}
	return postFromModel(model), nil
}

// List returns matching posts.
func (repository PostRepository) List(ctx context.Context, filter port.PostFilter, page pagination.Page) (pagination.Result[domain.Post], error) {
	query := repository.store.DB(ctx).Model(&PostModel{}).Where("thread_id = ?", filter.ThreadID).Order("sequence asc").Limit(page.Limit + 1)
	if !filter.IncludeHidden {
		query = query.Where("status IN ?", []domain.PostStatus{domain.PostStatusVisible, domain.PostStatusSystem})
	}
	var models []PostModel
	if err := query.Find(&models).Error; err != nil {
		return pagination.Result[domain.Post]{}, err
	}
	return postPage(models, page.Limit), nil
}

// NextSequence returns the next post sequence for a thread.
func (repository PostRepository) NextSequence(ctx context.Context, threadID uuid.UUID) (int64, error) {
	var maxSequence int64
	err := repository.store.DB(ctx).Model(&PostModel{}).Where("thread_id = ?", threadID).Select("COALESCE(MAX(sequence), 0)").Scan(&maxSequence).Error
	return maxSequence + 1, err
}

// UpdateWithRevision updates a post and writes a revision.
func (repository PostRepository) UpdateWithRevision(ctx context.Context, post domain.Post, revision domain.PostRevision, expectedVersion uint64) (domain.Post, error) {
	if revision.CreatedAt.IsZero() {
		revision.CreatedAt = time.Now().UTC()
	}
	model := revisionModelFromDomain(revision)
	if err := repository.store.DB(ctx).Create(&model).Error; err != nil {
		return domain.Post{}, err
	}
	result := repository.store.DB(ctx).Model(&PostModel{}).Where("id = ? AND version = ?", post.ID, expectedVersion).Updates(map[string]any{"content_document_json": string(post.ContentDocumentJSON), "content_text": post.ContentText, "content_checksum": post.ContentChecksum, "edited_at": post.EditedAt, "edited_by_user_id": post.EditedByUserID, "edit_count": post.EditCount, "version": expectedVersion + 1})
	if result.Error != nil {
		return domain.Post{}, result.Error
	}
	if result.RowsAffected == 0 {
		return domain.Post{}, port.ErrPreconditionFailed
	}
	return repository.FindByID(ctx, post.ID)
}

// Delete soft deletes one post.
func (repository PostRepository) Delete(ctx context.Context, id uuid.UUID, expectedVersion uint64) error {
	result := repository.store.DB(ctx).Where("id = ? AND version = ?", id, expectedVersion).Delete(&PostModel{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return port.ErrPreconditionFailed
	}
	return nil
}

// ListRevisions returns post revisions.
func (repository PostRepository) ListRevisions(ctx context.Context, postID uuid.UUID, page pagination.Page) (pagination.Result[domain.PostRevision], error) {
	var models []PostRevisionModel
	err := repository.store.DB(ctx).Where("post_id = ?", postID).Order("created_at desc, id asc").Limit(page.Limit + 1).Find(&models).Error
	if err != nil {
		return pagination.Result[domain.PostRevision]{}, err
	}
	return revisionPage(models, page.Limit), nil
}

// ListReferences returns references for posts.
func (repository PostRepository) ListReferences(ctx context.Context, postIDs []uuid.UUID) (map[uuid.UUID][]domain.PostReference, error) {
	result := map[uuid.UUID][]domain.PostReference{}
	if len(postIDs) == 0 {
		return result, nil
	}
	var models []PostReferenceModel
	if err := repository.store.DB(ctx).Where("source_post_id IN ?", postIDs).Find(&models).Error; err != nil {
		return nil, err
	}
	for _, model := range models {
		reference := referenceFromModel(model)
		result[reference.SourcePostID] = append(result[reference.SourcePostID], reference)
	}
	return result, nil
}

// incrementThreadAndForum updates counters after reply creation.
func (repository PostRepository) incrementThreadAndForum(ctx context.Context, post domain.Post) error {
	visible := int64(0)
	if post.Visible() {
		visible = 1
	}
	if err := repository.store.DB(ctx).Model(&ThreadModel{}).Where("id = ?", post.ThreadID).Updates(map[string]any{"reply_count": gorm.Expr("reply_count + ?", 1), "visible_reply_count": gorm.Expr("visible_reply_count + ?", visible), "post_count": gorm.Expr("post_count + ?", 1), "visible_post_count": gorm.Expr("visible_post_count + ?", visible), "latest_post_id": post.ID, "latest_post_author_user_id": post.AuthorUserID, "latest_post_at": post.CreatedAt}).Error; err != nil {
		return err
	}
	return repository.store.DB(ctx).Model(&StatsModel{}).Where("forum_id = ?", post.ForumID).Updates(map[string]any{"post_count": gorm.Expr("post_count + ?", 1), "visible_post_count": gorm.Expr("visible_post_count + ?", visible), "latest_post_id": post.ID, "latest_post_author_user_id": post.AuthorUserID, "latest_post_at": post.CreatedAt, "updated_at": time.Now().UTC()}).Error
}

// InteractionRepository stores forum interactions in PostgreSQL.
type InteractionRepository struct {
	store orm.Store
}

// NewInteractionRepository creates an interaction repository.
func NewInteractionRepository(store orm.Store) InteractionRepository {
	return InteractionRepository{store: store}
}

// LikePost creates or restores an active like and returns whether counters changed.
func (repository InteractionRepository) LikePost(ctx context.Context, like domain.PostLike) (bool, error) {
	var active PostLikeModel
	err := repository.store.DB(ctx).Where("post_id = ? AND user_id = ?", like.PostID, like.UserID).First(&active).Error
	if err == nil {
		return false, nil
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return false, err
	}
	var existing PostLikeModel
	err = repository.store.DB(ctx).Unscoped().Where("post_id = ? AND user_id = ?", like.PostID, like.UserID).First(&existing).Error
	if err == nil {
		if err := repository.store.DB(ctx).Unscoped().Model(&PostLikeModel{}).Where("id = ?", existing.ID.ID).Updates(map[string]any{"deleted_at": nil, "thread_id": like.ThreadID, "forum_id": like.ForumID, "created_at": like.CreatedAt}).Error; err != nil {
			return false, err
		}
		return true, repository.updateLikeCounts(ctx, like.PostID, like.ThreadID, 1)
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return false, err
	}
	model := likeModelFromDomain(like)
	if err := repository.store.DB(ctx).Create(&model).Error; err != nil {
		return false, port.ErrConflict
	}
	return true, repository.updateLikeCounts(ctx, like.PostID, like.ThreadID, 1)
}

// UnlikePost removes an active like and returns whether counters changed.
func (repository InteractionRepository) UnlikePost(ctx context.Context, postID uuid.UUID, userID uuid.UUID) (bool, error) {
	var active PostLikeModel
	err := repository.store.DB(ctx).Where("post_id = ? AND user_id = ?", postID, userID).First(&active).Error
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
func (repository InteractionRepository) LikedByUser(ctx context.Context, postID uuid.UUID, userID uuid.UUID) (bool, error) {
	if userID == uuid.Nil {
		return false, nil
	}
	var count int64
	err := repository.store.DB(ctx).Model(&PostLikeModel{}).Where("post_id = ? AND user_id = ?", postID, userID).Count(&count).Error
	return count > 0, err
}

// ListLatestPosts returns latest visible post summaries.
func (repository InteractionRepository) ListLatestPosts(ctx context.Context, filter port.LatestPostFilter, page pagination.Page) (pagination.Result[domain.LatestPostSummary], error) {
	if len(filter.ForumIDs) == 0 {
		return pagination.Result[domain.LatestPostSummary]{Items: []domain.LatestPostSummary{}}, nil
	}
	var rows []latestPostRow
	err := repository.store.DB(ctx).Table("forum_posts AS p").Select("p.forum_id, p.thread_id, p.id AS post_id, p.author_user_id, p.sequence, t.title AS thread_title, t.slug AS thread_slug, p.content_text AS excerpt, p.created_at").Joins("JOIN forum_threads AS t ON t.id = p.thread_id AND t.deleted_at IS NULL").Where("p.forum_id IN ? AND p.deleted_at IS NULL AND p.status IN ? AND t.status IN ?", filter.ForumIDs, visiblePostStatuses(), visibleThreadStatuses()).Order("p.created_at DESC, p.id ASC").Limit(page.Limit + 1).Find(&rows).Error
	if err != nil {
		return pagination.Result[domain.LatestPostSummary]{}, err
	}
	return latestPostPage(rows, page.Limit), nil
}

// ListMostLikedPosts returns most-liked visible posts.
func (repository InteractionRepository) ListMostLikedPosts(ctx context.Context, filter port.MostLikedFilter, page pagination.Page) (pagination.Result[domain.MostLikedPost], error) {
	var rows []mostLikedPostRow
	err := repository.store.DB(ctx).Table("forum_posts AS p").Select("p.forum_id, p.thread_id, p.id AS post_id, p.author_user_id, p.sequence, t.title AS thread_title, t.slug AS thread_slug, p.content_text AS excerpt, p.like_count, p.created_at").Joins("JOIN forum_threads AS t ON t.id = p.thread_id AND t.deleted_at IS NULL").Where("p.forum_id = ? AND p.deleted_at IS NULL AND p.status IN ? AND t.status IN ?", filter.ForumID, visiblePostStatuses(), visibleThreadStatuses()).Order("p.like_count DESC, p.created_at DESC, p.id ASC").Limit(page.Limit + 1).Find(&rows).Error
	if err != nil {
		return pagination.Result[domain.MostLikedPost]{}, err
	}
	return mostLikedPostPage(rows, page.Limit), nil
}

// MarkThreadRead stores one thread read state.
func (repository InteractionRepository) MarkThreadRead(ctx context.Context, state domain.ThreadReadState) error {
	now := state.LastReadAt
	if state.CreatedAt.IsZero() {
		state.CreatedAt = now
	}
	state.UpdatedAt = now
	model := readStateModelFromDomain(state)
	return repository.store.DB(ctx).Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "user_id"}, {Name: "thread_id"}}, DoUpdates: clause.Assignments(map[string]any{"forum_id": state.ForumID, "last_read_post_sequence": state.LastReadPostSequence, "last_read_at": state.LastReadAt, "updated_at": state.UpdatedAt})}).Create(&model).Error
}

// MarkForumRead stores read states for every visible thread in a forum.
func (repository InteractionRepository) MarkForumRead(ctx context.Context, userID uuid.UUID, forumID uuid.UUID, readAt time.Time) error {
	var rows []forumReadTargetRow
	err := repository.store.DB(ctx).Table("forum_threads AS t").Select("t.id AS thread_id, t.forum_id, COALESCE(MAX(p.sequence), 1) AS last_read_post_sequence").Joins("JOIN forum_posts AS p ON p.thread_id = t.id AND p.deleted_at IS NULL AND p.status IN ?", visiblePostStatuses()).Where("t.forum_id = ? AND t.deleted_at IS NULL AND t.status IN ?", forumID, visibleThreadStatuses()).Group("t.id, t.forum_id").Find(&rows).Error
	if err != nil {
		return err
	}
	for _, row := range rows {
		state := domain.ThreadReadState{ID: uuid.New(), UserID: userID, ForumID: row.ForumID, ThreadID: row.ThreadID, LastReadPostSequence: row.LastReadPostSequence, LastReadAt: readAt}
		if err := repository.MarkThreadRead(ctx, state); err != nil {
			return err
		}
	}
	return nil
}

// UnreadSummary returns unread counts for visible forums.
func (repository InteractionRepository) UnreadSummary(ctx context.Context, userID uuid.UUID, forumIDs []uuid.UUID) (domain.UnreadSummary, error) {
	summary := domain.UnreadSummary{UserID: userID, Forums: []domain.ForumUnreadSummary{}}
	if len(forumIDs) == 0 {
		return summary, nil
	}
	var rows []unreadForumRow
	err := repository.store.DB(ctx).Table("forum_threads AS t").Select("t.forum_id, COUNT(*) AS unread_thread_count").Joins("LEFT JOIN forum_thread_read_states AS rs ON rs.thread_id = t.id AND rs.user_id = ?", userID).Where("t.forum_id IN ? AND t.deleted_at IS NULL AND t.status IN ? AND COALESCE(rs.last_read_post_sequence, 0) < t.visible_post_count", forumIDs, visibleThreadStatuses()).Group("t.forum_id").Find(&rows).Error
	if err != nil {
		return domain.UnreadSummary{}, err
	}
	for _, row := range rows {
		summary.UnreadThreadCount += row.UnreadThreadCount
		summary.Forums = append(summary.Forums, domain.ForumUnreadSummary{ForumID: row.ForumID, UnreadThreadCount: row.UnreadThreadCount})
	}
	return summary, nil
}

// updateLikeCounts changes post and thread like counters.
func (repository InteractionRepository) updateLikeCounts(ctx context.Context, postID uuid.UUID, threadID uuid.UUID, delta int64) error {
	expr := gorm.Expr("like_count + ?", delta)
	if delta < 0 {
		expr = gorm.Expr("CASE WHEN like_count > 0 THEN like_count - 1 ELSE 0 END")
	}
	if err := repository.store.DB(ctx).Model(&PostModel{}).Where("id = ?", postID).Update("like_count", expr).Error; err != nil {
		return err
	}
	return repository.store.DB(ctx).Model(&ThreadModel{}).Where("id = ?", threadID).Update("like_count", expr).Error
}

// visiblePostStatuses returns statuses normal widget readers may see.
func visiblePostStatuses() []domain.PostStatus {
	return []domain.PostStatus{domain.PostStatusVisible, domain.PostStatusSystem}
}

// visibleThreadStatuses returns statuses normal widget readers may see.
func visibleThreadStatuses() []domain.ThreadStatus {
	return []domain.ThreadStatus{domain.ThreadStatusOpen, domain.ThreadStatusClosed, domain.ThreadStatusLocked}
}

// threadPage maps models into a page.
func threadPage(models []ThreadModel, limit int) pagination.Result[domain.Thread] {
	next := ""
	if len(models) > limit {
		next = models[limit-1].ID.ID.String()
		models = models[:limit]
	}
	items := make([]domain.Thread, 0, len(models))
	for _, model := range models {
		items = append(items, threadFromModel(model))
	}
	return pagination.Result[domain.Thread]{Items: items, NextCursor: next}
}

// postPage maps models into a page.
func postPage(models []PostModel, limit int) pagination.Result[domain.Post] {
	next := ""
	if len(models) > limit {
		next = models[limit-1].ID.ID.String()
		models = models[:limit]
	}
	items := make([]domain.Post, 0, len(models))
	for _, model := range models {
		items = append(items, postFromModel(model))
	}
	return pagination.Result[domain.Post]{Items: items, NextCursor: next}
}

// revisionPage maps models into a page.
func revisionPage(models []PostRevisionModel, limit int) pagination.Result[domain.PostRevision] {
	next := ""
	if len(models) > limit {
		next = models[limit-1].ID.ID.String()
		models = models[:limit]
	}
	items := make([]domain.PostRevision, 0, len(models))
	for _, model := range models {
		items = append(items, revisionFromModel(model))
	}
	return pagination.Result[domain.PostRevision]{Items: items, NextCursor: next}
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
	ForumID      uuid.UUID
	ThreadID     uuid.UUID
	PostID       uuid.UUID
	AuthorUserID uuid.UUID
	Sequence     int64
	ThreadTitle  string
	ThreadSlug   string
	Excerpt      string
	LikeCount    int64
	CreatedAt    time.Time
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

// latestPostPage maps latest-post rows into a page.
func latestPostPage(rows []latestPostRow, limit int) pagination.Result[domain.LatestPostSummary] {
	next := ""
	if len(rows) > limit {
		next = rows[limit-1].PostID.String()
		rows = rows[:limit]
	}
	items := make([]domain.LatestPostSummary, 0, len(rows))
	for _, row := range rows {
		items = append(items, domain.LatestPostSummary{ForumID: row.ForumID, ThreadID: row.ThreadID, PostID: row.PostID, AuthorUserID: row.AuthorUserID, Sequence: row.Sequence, ThreadTitle: row.ThreadTitle, ThreadSlug: domain.Slug(row.ThreadSlug), Excerpt: row.Excerpt, CreatedAt: row.CreatedAt})
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
		items = append(items, domain.MostLikedPost{ForumID: row.ForumID, ThreadID: row.ThreadID, PostID: row.PostID, AuthorUserID: row.AuthorUserID, Sequence: row.Sequence, ThreadTitle: row.ThreadTitle, ThreadSlug: domain.Slug(row.ThreadSlug), Excerpt: row.Excerpt, LikeCount: row.LikeCount, CreatedAt: row.CreatedAt})
	}
	return pagination.Result[domain.MostLikedPost]{Items: items, NextCursor: next}
}

// Ensure ThreadRepository implements port.ThreadRepository.
var _ port.ThreadRepository = ThreadRepository{}

// Ensure PostRepository implements port.PostRepository.
var _ port.PostRepository = PostRepository{}

// Ensure InteractionRepository implements port.InteractionRepository.
var _ port.InteractionRepository = InteractionRepository{}
