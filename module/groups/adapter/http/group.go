package http

import (
	"errors"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/module/groups/domain"
	"github.com/realmkit/rk-backend/module/groups/port"
	"github.com/realmkit/rk-backend/pkg/api/problem"
	"github.com/realmkit/rk-backend/pkg/search"
)

// groupRequest is the group create or update body.
type groupRequest struct {
	Key         domain.Key         `json:"key"`
	Name        string             `json:"name"`
	Description string             `json:"description"`
	Color       domain.Color       `json:"color"`
	Weight      int                `json:"weight"`
	Status      domain.GroupStatus `json:"status"`
	IconAssetID *uuid.UUID         `json:"icon_asset_id"`
}

// groupListResponse contains one group page.
type groupListResponse struct {
	Items         []domain.Group `json:"items"`
	NextPageToken string         `json:"next_page_token,omitempty"`
	Query         string         `json:"query,omitempty"`
	Sort          string         `json:"sort,omitempty"`
	Direction     string         `json:"direction,omitempty"`
}

// createGroup creates a group.
func (handler handler) createGroup(ctx *fiber.Ctx) error {
	if err := requireIdempotency(ctx); err != nil {
		return err
	}
	var request groupRequest
	if err := decodeJSON(ctx, &request); err != nil {
		return err
	}
	group, err := handler.services.Groups.Create(ctx.UserContext(), port.CreateGroupCommand{Group: groupFromRequest(request)})
	if err != nil {
		return handleError(ctx, err)
	}
	setETag(ctx, group.Version)
	return writeJSON(ctx, fiber.StatusCreated, group)
}

// listGroups lists groups.
func (handler handler) listGroups(ctx *fiber.Ctx) error {
	page, err := pageFromQuery(ctx)
	if err != nil {
		return err
	}
	filter, err := groupFilterFromQuery(ctx)
	if err != nil {
		return err
	}
	result, err := handler.services.Groups.List(ctx.UserContext(), filter, page)
	if err != nil {
		return handleError(ctx, err)
	}
	return writeJSON(ctx, fiber.StatusOK, groupListResponse{
		Items:         result.Items,
		NextPageToken: result.NextCursor,
		Query:         filter.Query.String(),
		Sort:          filter.Sort.Key,
		Direction:     string(filter.Sort.Direction),
	})
}

// getGroup returns one group.
func (handler handler) getGroup(ctx *fiber.Ctx) error {
	id, err := idFromParam(ctx, "group_id")
	if err != nil {
		return err
	}
	group, err := handler.services.Groups.Get(ctx.UserContext(), id)
	if err != nil {
		return handleError(ctx, err)
	}
	setETag(ctx, group.Version)
	return writeJSON(ctx, fiber.StatusOK, group)
}

// updateGroup updates a group.
func (handler handler) updateGroup(ctx *fiber.Ctx) error {
	if err := requireIdempotency(ctx); err != nil {
		return err
	}
	id, err := idFromParam(ctx, "group_id")
	if err != nil {
		return err
	}
	version, err := expectedVersion(ctx)
	if err != nil {
		return err
	}
	var request groupRequest
	if err := decodeJSON(ctx, &request); err != nil {
		return err
	}
	group := groupFromRequest(request)
	group.ID = id
	updated, err := handler.services.Groups.Update(ctx.UserContext(), port.UpdateGroupCommand{Group: group, ExpectedVersion: version})
	if err != nil {
		return handleError(ctx, err)
	}
	setETag(ctx, updated.Version)
	return writeJSON(ctx, fiber.StatusOK, updated)
}

// deleteGroup deletes a group.
func (handler handler) deleteGroup(ctx *fiber.Ctx) error {
	if err := requireIdempotency(ctx); err != nil {
		return err
	}
	id, err := idFromParam(ctx, "group_id")
	if err != nil {
		return err
	}
	version, err := expectedVersion(ctx)
	if err != nil {
		return err
	}
	if err := handler.services.Groups.Delete(ctx.UserContext(), port.DeleteGroupCommand{ID: id, ExpectedVersion: version}); err != nil {
		return handleError(ctx, err)
	}
	return writeNoContent(ctx)
}

// groupFromRequest maps HTTP request to group.
func groupFromRequest(request groupRequest) domain.Group {
	return domain.Group{
		Key:         request.Key,
		Name:        request.Name,
		Description: request.Description,
		Color:       request.Color,
		Weight:      request.Weight,
		Status:      request.Status,
		IconAssetID: request.IconAssetID,
	}
}

// groupFilterFromQuery maps query params into a group filter.
func groupFilterFromQuery(ctx *fiber.Ctx) (port.GroupFilter, error) {
	query, err := search.NewTextQuery(ctx.Query("q"), search.QueryOptions{})
	if err != nil {
		return port.GroupFilter{}, searchProblem(err)
	}
	sort, err := search.NewSort(ctx.Query("sort"), ctx.Query("direction"), port.DefaultGroupSort(), port.AllowedGroupSorts())
	if err != nil {
		return port.GroupFilter{}, searchProblem(err)
	}
	hasIcon, err := optionalBool(ctx.Query("has_icon"))
	if err != nil {
		return port.GroupFilter{}, searchProblem(err)
	}
	minWeight, err := optionalInt(ctx.Query("min_weight"))
	if err != nil {
		return port.GroupFilter{}, searchProblem(err)
	}
	maxWeight, err := optionalInt(ctx.Query("max_weight"))
	if err != nil {
		return port.GroupFilter{}, searchProblem(err)
	}
	return port.GroupFilter{
		Status:    domain.GroupStatus(ctx.Query("status")),
		Query:     query,
		HasIcon:   hasIcon,
		MinWeight: minWeight,
		MaxWeight: maxWeight,
		Sort:      sort,
	}, nil
}

// optionalBool parses an optional boolean query value.
func optionalBool(value string) (*bool, error) {
	if value == "" {
		return nil, nil
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}

// optionalInt parses an optional integer query value.
func optionalInt(value string) (*int, error) {
	if value == "" {
		return nil, nil
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}

// searchProblem maps invalid search parameters to a problem response.
func searchProblem(err error) error {
	code := "invalid_search"
	if errors.Is(err, search.ErrInvalidCursor) {
		code = "invalid_page_token"
	}
	return problem.Error{Problem: problem.New(fiber.StatusBadRequest, code, "Search parameters are invalid.")}
}
