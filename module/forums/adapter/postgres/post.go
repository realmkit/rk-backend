package postgres

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/module/forums/domain"
	"github.com/realmkit/rk-backend/module/forums/port"
	"github.com/realmkit/rk-backend/pkg/orm"
	"github.com/realmkit/rk-backend/pkg/pagination"
	"gorm.io/gorm"
)

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
	return created, repository.createReferences(ctx, references)
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
func (repository PostRepository) List(
	ctx context.Context,
	filter port.PostFilter,
	page pagination.Page,
) (pagination.Result[domain.Post], error) {
	query := repository.store.DB(ctx).
		Model(&PostModel{}).
		Where("thread_id = ?", filter.ThreadID).
		Order("sequence asc").
		Limit(page.Limit + 1)
	if !filter.IncludeHidden {
		query = query.Where(
			"status IN ?",
			[]domain.PostStatus{domain.PostStatusVisible, domain.PostStatusSystem},
		)
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
	err := repository.store.DB(ctx).
		Model(&PostModel{}).
		Where("thread_id = ?", threadID).
		Select("COALESCE(MAX(sequence), 0)").
		Scan(&maxSequence).Error
	return maxSequence + 1, err
}

// UpdateWithRevision updates a post and writes a revision.
func (repository PostRepository) UpdateWithRevision(
	ctx context.Context,
	post domain.Post,
	revision domain.PostRevision,
	expectedVersion uint64,
) (domain.Post, error) {
	if revision.CreatedAt.IsZero() {
		revision.CreatedAt = time.Now().UTC()
	}
	model := revisionModelFromDomain(revision)
	if err := repository.store.DB(ctx).Create(&model).Error; err != nil {
		return domain.Post{}, err
	}
	result := repository.store.DB(ctx).
		Model(&PostModel{}).
		Where("id = ? AND version = ?", post.ID, expectedVersion).
		Updates(map[string]any{
			"content_document_json": string(post.ContentDocumentJSON),
			"content_text":          post.ContentText,
			"content_checksum":      post.ContentChecksum,
			"edited_at":             post.EditedAt,
			"edited_by_user_id":     post.EditedByUserID,
			"edit_count":            post.EditCount,
			"version":               expectedVersion + 1,
		})
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
	result := repository.store.DB(ctx).
		Where("id = ? AND version = ?", id, expectedVersion).
		Delete(&PostModel{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return port.ErrPreconditionFailed
	}
	return nil
}

// ListRevisions returns post revisions.
func (repository PostRepository) ListRevisions(
	ctx context.Context,
	postID uuid.UUID,
	page pagination.Page,
) (pagination.Result[domain.PostRevision], error) {
	var models []PostRevisionModel
	err := repository.store.DB(ctx).
		Where("post_id = ?", postID).
		Order("created_at desc, id asc").
		Limit(page.Limit + 1).
		Find(&models).Error
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

// createReferences stores structured post references.
func (repository PostRepository) createReferences(ctx context.Context, references []domain.PostReference) error {
	for _, reference := range references {
		if reference.CreatedAt.IsZero() {
			reference.CreatedAt = time.Now().UTC()
		}
		model := referenceModelFromDomain(reference)
		if err := repository.store.DB(ctx).Create(&model).Error; err != nil {
			return err
		}
	}
	return nil
}

// incrementThreadAndForum updates counters after reply creation.
func (repository PostRepository) incrementThreadAndForum(ctx context.Context, post domain.Post) error {
	visible := int64(0)
	if post.Visible() {
		visible = 1
	}
	threadUpdates := map[string]any{
		"reply_count":                gorm.Expr("reply_count + ?", 1),
		"visible_reply_count":        gorm.Expr("visible_reply_count + ?", visible),
		"post_count":                 gorm.Expr("post_count + ?", 1),
		"visible_post_count":         gorm.Expr("visible_post_count + ?", visible),
		"latest_post_id":             post.ID,
		"latest_post_author_user_id": post.AuthorUserID,
		"latest_post_at":             post.CreatedAt,
	}
	if err := repository.store.DB(ctx).
		Model(&ThreadModel{}).
		Where("id = ?", post.ThreadID).
		Updates(threadUpdates).Error; err != nil {
		return err
	}
	forumUpdates := map[string]any{
		"post_count":                 gorm.Expr("post_count + ?", 1),
		"visible_post_count":         gorm.Expr("visible_post_count + ?", visible),
		"latest_post_id":             post.ID,
		"latest_post_author_user_id": post.AuthorUserID,
		"latest_post_at":             post.CreatedAt,
		"updated_at":                 time.Now().UTC(),
	}
	return repository.store.DB(ctx).
		Model(&StatsModel{}).
		Where("forum_id = ?", post.ForumID).
		Updates(forumUpdates).Error
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

// Ensure PostRepository implements port.PostRepository.
var _ port.PostRepository = PostRepository{}
