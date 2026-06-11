package structure

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/module/forums/adapter/http/shared"
	"github.com/realmkit/rk-backend/module/forums/domain"
)

// writeActorObjectVersion reads the common actor, object ID, and expected version values.
func writeActorObjectVersion(
	ctx *fiber.Ctx,
	idParam string,
) (uuid.UUID, uuid.UUID, uint64, error) {
	actor, err := shared.CurrentUserID(ctx)
	if err != nil {
		return uuid.Nil, uuid.Nil, 0, err
	}
	id, err := shared.IDFromParam(ctx, idParam)
	if err != nil {
		return uuid.Nil, uuid.Nil, 0, err
	}
	version, err := shared.ExpectedVersion(ctx)
	if err != nil {
		return uuid.Nil, uuid.Nil, 0, err
	}
	return actor, id, version, nil
}

// forumFromRequest maps request data to domain.
func forumFromRequest(id uuid.UUID, request forumRequest) domain.Forum {
	return domain.Forum{
		ID:                            id,
		CategoryID:                    request.CategoryID,
		ParentForumID:                 request.ParentForumID,
		Kind:                          request.Kind,
		Key:                           request.Key,
		Slug:                          request.Slug,
		Name:                          request.Name,
		Description:                   request.Description,
		DisplayOrder:                  request.DisplayOrder,
		ExternalURL:                   request.ExternalURL,
		IconAssetID:                   request.IconAssetID,
		ThreadVisibilityMode:          request.ThreadVisibilityMode,
		MaxStickyThreads:              request.MaxStickyThreads,
		DefaultThreadStatus:           request.DefaultThreadStatus,
		AuthorPostEditWindowSeconds:   request.AuthorPostEditWindowSeconds,
		AuthorPostDeleteWindowSeconds: request.AuthorPostDeleteWindowSeconds,
		Status:                        request.Status,
	}
}
