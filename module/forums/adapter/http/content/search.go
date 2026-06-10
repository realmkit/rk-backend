package content

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/niflaot/gamehub-go/module/forums/adapter/http/shared"
	"github.com/niflaot/gamehub-go/module/forums/port"
)

// searchForums searches all visible forums.
func (handler handler) searchForums(ctx *fiber.Ctx) error {
	return handler.search(ctx, uuid.Nil)
}

// searchForum searches one visible forum.
func (handler handler) searchForum(ctx *fiber.Ctx) error {
	forumID, err := shared.IDFromParam(ctx, "forum_id")
	if err != nil {
		return err
	}
	return handler.search(ctx, forumID)
}

// search searches forum content.
func (handler handler) search(ctx *fiber.Ctx, forumID uuid.UUID) error {
	actor, err := shared.OptionalUserID(ctx)
	if err != nil {
		return err
	}
	page, err := shared.PageFromQuery(ctx)
	if err != nil {
		return err
	}
	query := ctx.Query("query")
	if query == "" {
		query = ctx.Query("q")
	}
	command := port.SearchCommand{
		ActorUserID: actor,
		ForumID:     forumID,
		Query:       query,
	}
	result, err := handler.services.Operations.Search(ctx.Context(), command, page)
	if err != nil {
		return shared.HandleError(ctx, err)
	}
	return shared.WriteJSON(ctx, fiber.StatusOK, searchListResponse{
		Items:         result.Items,
		NextPageToken: result.NextCursor,
	})
}
