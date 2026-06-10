package structure

import (
	"context"

	"github.com/google/uuid"
	"github.com/niflaot/gamehub-go/module/forums/domain"
	"github.com/niflaot/gamehub-go/module/forums/port"
	"github.com/niflaot/gamehub-go/pkg/pagination"
)

// CreateForum creates a forum.
func (service Service) CreateForum(
	ctx context.Context,
	command port.CreateForumCommand,
) (domain.Forum, error) {
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
func (service Service) UpdateForum(
	ctx context.Context,
	command port.UpdateForumCommand,
) (domain.Forum, error) {
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

// MoveForum moves a forum.
func (service Service) MoveForum(
	ctx context.Context,
	command port.MoveForumCommand,
) (domain.Forum, error) {
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
func (service Service) ListForums(
	ctx context.Context,
	filter port.ForumFilter,
	page pagination.Page,
) (pagination.Result[domain.Forum], error) {
	return service.forums.List(ctx, filter, page)
}

// DeleteForum deletes a forum.
func (service Service) DeleteForum(
	ctx context.Context,
	command port.DeleteForumCommand,
) error {
	if err := service.requireManage(ctx, command.ActorUserID, command.ID); err != nil {
		return err
	}
	if err := service.forums.Delete(ctx, command.ID, command.ExpectedVersion); err != nil {
		return err
	}
	return service.clearTree(ctx)
}

// ReorderForums reorders forums.
func (service Service) ReorderForums(
	ctx context.Context,
	command port.ReorderForumsCommand,
) error {
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
