package application

import (
	"context"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/niflaot/gamehub-go/module/forums/domain"
	"github.com/niflaot/gamehub-go/module/forums/port"
	"github.com/niflaot/gamehub-go/pkg/pagination"
	"github.com/niflaot/gamehub-go/pkg/transaction"
)

// treeCacheTTL is the visible forum tree cache lifetime.
const treeCacheTTL = 30 * time.Second

// widgetCacheTTL is the forum widget cache lifetime.
const widgetCacheTTL = 20 * time.Second

// Service manages forum structure use cases.
type Service struct {
	categories   port.CategoryRepository
	forums       port.ForumRepository
	threads      port.ThreadRepository
	posts        port.PostRepository
	interactions port.InteractionRepository
	operations   port.OperationsRepository
	assets       port.AssetResolver
	authorizer   port.VisibilityAuthorizer
	permissions  port.PermissionAdmin
	cache        port.ReadCache
	transactions transaction.Runner
}

// Dependencies contains forum service dependencies.
type Dependencies struct {
	// Categories stores categories.
	Categories port.CategoryRepository

	// Forums stores forums.
	Forums port.ForumRepository

	// Threads stores threads.
	Threads port.ThreadRepository

	// Posts stores posts.
	Posts port.PostRepository

	// Interactions stores likes, widgets, and read state.
	Interactions port.InteractionRepository

	// Operations runs search, repair, and counter flushes.
	Operations port.OperationsRepository

	// Assets resolves attachment assets.
	Assets port.AssetResolver

	// Authorizer checks forum permissions.
	Authorizer port.VisibilityAuthorizer

	// Permissions manages forum permission configuration.
	Permissions port.PermissionAdmin

	// Cache caches visible trees.
	Cache port.ReadCache

	// Transactions runs transactional use cases.
	Transactions transaction.Runner
}

// NewService creates a forum service.
func NewService(deps Dependencies) Service {
	permissions := deps.Permissions
	if permissions == nil {
		if admin, ok := deps.Authorizer.(port.PermissionAdmin); ok {
			permissions = admin
		}
	}
	return Service{
		categories:   deps.Categories,
		forums:       deps.Forums,
		threads:      deps.Threads,
		posts:        deps.Posts,
		interactions: deps.Interactions,
		operations:   deps.Operations,
		assets:       deps.Assets,
		authorizer:   deps.Authorizer,
		permissions:  permissions,
		cache:        deps.Cache,
		transactions: deps.Transactions,
	}
}

// CreateCategory creates a category.
func (service Service) CreateCategory(ctx context.Context, command port.CreateCategoryCommand) (domain.ForumCategory, error) {
	if err := service.requireManage(ctx, command.ActorUserID, domain.RootForumObjectID()); err != nil {
		return domain.ForumCategory{}, err
	}
	category := command.Category.Normalize()
	if category.ID == uuid.Nil {
		category.ID = uuid.New()
	}
	if err := category.Validate(); err != nil {
		return domain.ForumCategory{}, err
	}
	created, err := service.categories.Create(ctx, category)
	if err != nil {
		return domain.ForumCategory{}, err
	}
	return created, service.clearTree(ctx)
}

// UpdateCategory updates a category.
func (service Service) UpdateCategory(ctx context.Context, command port.UpdateCategoryCommand) (domain.ForumCategory, error) {
	if err := service.requireManage(ctx, command.ActorUserID, domain.RootForumObjectID()); err != nil {
		return domain.ForumCategory{}, err
	}
	category := command.Category.Normalize()
	if err := category.Validate(); err != nil {
		return domain.ForumCategory{}, err
	}
	updated, err := service.categories.Update(ctx, category, command.ExpectedVersion)
	if err != nil {
		return domain.ForumCategory{}, err
	}
	return updated, service.clearTree(ctx)
}

// GetCategory returns one category.
func (service Service) GetCategory(ctx context.Context, id uuid.UUID) (domain.ForumCategory, error) {
	return service.categories.FindByID(ctx, id)
}

// ListCategories lists categories.
func (service Service) ListCategories(ctx context.Context, filter port.CategoryFilter, page pagination.Page) (pagination.Result[domain.ForumCategory], error) {
	return service.categories.List(ctx, filter, page)
}

// DeleteCategory deletes a category.
func (service Service) DeleteCategory(ctx context.Context, command port.DeleteCategoryCommand) error {
	if err := service.requireManage(ctx, command.ActorUserID, domain.RootForumObjectID()); err != nil {
		return err
	}
	if err := service.categories.Delete(ctx, command.ID, command.ExpectedVersion); err != nil {
		return err
	}
	return service.clearTree(ctx)
}

// ReorderCategories reorders categories.
func (service Service) ReorderCategories(ctx context.Context, command port.ReorderCategoriesCommand) error {
	if err := service.requireManage(ctx, command.ActorUserID, domain.RootForumObjectID()); err != nil {
		return err
	}
	if err := service.validateReorder(command.Items); err != nil {
		return err
	}
	if err := service.categories.Reorder(ctx, command.Items); err != nil {
		return err
	}
	return service.clearTree(ctx)
}

// CreateForum creates a forum.
func (service Service) CreateForum(ctx context.Context, command port.CreateForumCommand) (domain.Forum, error) {
	if err := service.requireManage(ctx, command.ActorUserID, domain.RootForumObjectID()); err != nil {
		return domain.Forum{}, err
	}
	var created domain.Forum
	err := service.transactions.WithinTx(ctx, func(ctx context.Context) error {
		forum, err := service.prepareForum(ctx, command.Forum.Normalize())
		if err != nil {
			return err
		}
		stored, err := service.forums.Create(ctx, forum)
		if err != nil {
			return err
		}
		created = stored
		return service.clearTree(ctx)
	})
	return created, err
}

// UpdateForum updates a forum.
func (service Service) UpdateForum(ctx context.Context, command port.UpdateForumCommand) (domain.Forum, error) {
	if err := service.requireManage(ctx, command.ActorUserID, command.Forum.ID); err != nil {
		return domain.Forum{}, err
	}
	current, err := service.forums.FindByID(ctx, command.Forum.ID)
	if err != nil {
		return domain.Forum{}, err
	}
	forum := command.Forum.Normalize()
	forum.CategoryID = current.CategoryID
	forum.ParentForumID = current.ParentForumID
	forum.Path = current.Path
	forum.Depth = current.Depth
	if err := forum.Validate(); err != nil {
		return domain.Forum{}, err
	}
	updated, err := service.forums.Update(ctx, forum, command.ExpectedVersion)
	if err != nil {
		return domain.Forum{}, err
	}
	return updated, service.clearTree(ctx)
}

// GetForumSettings returns admin forum settings.
func (service Service) GetForumSettings(ctx context.Context, actorUserID uuid.UUID, forumID uuid.UUID) (domain.ForumSettings, error) {
	if err := service.requireManage(ctx, actorUserID, forumID); err != nil {
		return domain.ForumSettings{}, err
	}
	forum, err := service.forums.FindByID(ctx, forumID)
	if err != nil {
		return domain.ForumSettings{}, err
	}
	return forum.Settings(), nil
}

// UpdateForumSettings updates admin forum settings.
func (service Service) UpdateForumSettings(ctx context.Context, command port.UpdateForumSettingsCommand) (domain.ForumSettings, error) {
	if err := service.requireManage(ctx, command.ActorUserID, command.Settings.ForumID); err != nil {
		return domain.ForumSettings{}, err
	}
	current, err := service.forums.FindByID(ctx, command.Settings.ForumID)
	if err != nil {
		return domain.ForumSettings{}, err
	}
	settings := command.Settings.Normalize()
	if err := settings.Validate(); err != nil {
		return domain.ForumSettings{}, err
	}
	updatedForum := current
	updatedForum.Kind = settings.Kind
	updatedForum.ExternalURL = settings.ExternalURL
	updatedForum.ThreadVisibilityMode = settings.ThreadVisibilityMode
	updatedForum.MaxStickyThreads = settings.MaxStickyThreads
	updatedForum.DefaultThreadStatus = settings.DefaultThreadStatus
	updatedForum.AuthorPostEditWindowSeconds = settings.AuthorPostEditWindowSeconds
	updatedForum.AuthorPostDeleteWindowSeconds = settings.AuthorPostDeleteWindowSeconds
	if err := updatedForum.Validate(); err != nil {
		return domain.ForumSettings{}, err
	}
	updated, err := service.forums.Update(ctx, updatedForum, command.ExpectedVersion)
	if err != nil {
		return domain.ForumSettings{}, err
	}
	return updated.Settings(), service.ClearReadCache(ctx)
}

// GetForumPermissionSettings returns forum permission grants.
func (service Service) GetForumPermissionSettings(ctx context.Context, actorUserID uuid.UUID, forumID uuid.UUID) (domain.ForumPermissionSettings, error) {
	if err := service.requireManage(ctx, actorUserID, forumID); err != nil {
		return domain.ForumPermissionSettings{}, err
	}
	if _, err := service.forums.FindByID(ctx, forumID); err != nil {
		return domain.ForumPermissionSettings{}, err
	}
	if service.permissions == nil {
		return domain.ForumPermissionSettings{}, port.ErrForbidden
	}
	return service.permissions.ForumPermissionSettings(ctx, forumID)
}

// UpdateForumPermissionSettings replaces forum permission grants.
func (service Service) UpdateForumPermissionSettings(ctx context.Context, command port.UpdateForumPermissionSettingsCommand) error {
	settings := command.Settings.Normalize()
	if err := service.requireManage(ctx, command.ActorUserID, settings.ForumID); err != nil {
		return err
	}
	if _, err := service.forums.FindByID(ctx, settings.ForumID); err != nil {
		return err
	}
	if err := settings.Validate(); err != nil {
		return err
	}
	if service.permissions == nil {
		return port.ErrForbidden
	}
	err := service.transactions.WithinTx(ctx, func(ctx context.Context) error {
		if err := service.permissions.UpdateForumPermissionSettings(ctx, command.ActorUserID, settings); err != nil {
			return err
		}
		return service.ClearReadCache(ctx)
	})
	return err
}

// SimulateForumPermission simulates one forum permission.
func (service Service) SimulateForumPermission(ctx context.Context, command port.SimulateForumPermissionCommand) (domain.ForumPermissionSimulationResult, error) {
	if err := service.requireManage(ctx, command.ActorUserID, command.ForumID); err != nil {
		return domain.ForumPermissionSimulationResult{}, err
	}
	if _, err := service.forums.FindByID(ctx, command.ForumID); err != nil {
		return domain.ForumPermissionSimulationResult{}, err
	}
	request := command.Request.Normalize(command.ForumID)
	if err := request.Validate(); err != nil {
		return domain.ForumPermissionSimulationResult{}, err
	}
	if service.permissions == nil {
		return domain.ForumPermissionSimulationResult{}, port.ErrForbidden
	}
	return service.permissions.SimulateForumPermission(ctx, command.ForumID, request)
}

// MoveForum moves a forum.
func (service Service) MoveForum(ctx context.Context, command port.MoveForumCommand) (domain.Forum, error) {
	if err := service.requireManage(ctx, command.ActorUserID, command.ID); err != nil {
		return domain.Forum{}, err
	}
	var moved domain.Forum
	err := service.transactions.WithinTx(ctx, func(ctx context.Context) error {
		current, err := service.forums.FindByID(ctx, command.ID)
		if err != nil {
			return err
		}
		target, err := service.moveTarget(ctx, current, command)
		if err != nil {
			return err
		}
		stored, err := service.forums.Move(ctx, target, current.Path, command.ExpectedVersion)
		if err != nil {
			return err
		}
		moved = stored
		return service.clearTree(ctx)
	})
	return moved, err
}

// GetForum returns one forum.
func (service Service) GetForum(ctx context.Context, id uuid.UUID) (domain.Forum, error) {
	return service.forums.FindByID(ctx, id)
}

// ListForums lists forums.
func (service Service) ListForums(ctx context.Context, filter port.ForumFilter, page pagination.Page) (pagination.Result[domain.Forum], error) {
	return service.forums.List(ctx, filter, page)
}

// DeleteForum deletes a forum.
func (service Service) DeleteForum(ctx context.Context, command port.DeleteForumCommand) error {
	if err := service.requireManage(ctx, command.ActorUserID, command.ID); err != nil {
		return err
	}
	if err := service.forums.Delete(ctx, command.ID, command.ExpectedVersion); err != nil {
		return err
	}
	return service.clearTree(ctx)
}

// ReorderForums reorders forums.
func (service Service) ReorderForums(ctx context.Context, command port.ReorderForumsCommand) error {
	if err := service.requireManage(ctx, command.ActorUserID, domain.RootForumObjectID()); err != nil {
		return err
	}
	if err := service.validateReorder(command.Items); err != nil {
		return err
	}
	if err := service.forums.Reorder(ctx, command.Items); err != nil {
		return err
	}
	return service.clearTree(ctx)
}

// Tree returns the visible forum tree.
func (service Service) Tree(ctx context.Context, actorUserID uuid.UUID) (domain.ForumTree, error) {
	key := treeCacheKey(actorUserID)
	if service.cache != nil {
		if cached, ok, err := service.cache.GetTree(ctx, key); err == nil && ok {
			return cached, nil
		}
	}
	tree, err := service.loadTree(ctx, actorUserID)
	if err != nil {
		return domain.ForumTree{}, err
	}
	if service.cache != nil {
		if err := service.cache.SetTree(ctx, key, tree, treeCacheTTL); err != nil {
			return tree, nil
		}
	}
	return tree, nil
}

// requireManage verifies structure-management permission.
func (service Service) requireManage(ctx context.Context, actorUserID uuid.UUID, forumID uuid.UUID) error {
	if service.authorizer == nil {
		return nil
	}
	allowed, err := service.authorizer.CanManageForum(ctx, actorUserID, forumID)
	if err != nil {
		return err
	}
	if !allowed {
		return port.ErrForbidden
	}
	return nil
}

// clearTree clears cached trees when a cache is configured.
func (service Service) clearTree(ctx context.Context) error {
	if service.cache == nil {
		return nil
	}
	return service.cache.ClearTree(ctx)
}

// validateReorder validates reorder items.
func (service Service) validateReorder(items []port.ReorderItem) error {
	var violations []domain.Violation
	if len(items) == 0 {
		violations = domain.AppendViolation(violations, "items", "must contain at least one item")
	}
	for index, item := range items {
		if item.ID == uuid.Nil {
			violations = domain.AppendViolation(violations, "items["+strconv.Itoa(index)+"].id", "is required")
		}
		violations = append(violations, domain.ValidateDisplayOrder("items["+strconv.Itoa(index)+"].display_order", item.DisplayOrder)...)
	}
	return domain.NewValidationError(violations)
}

// Ensure Service implements port.Service.
var _ port.Service = Service{}
