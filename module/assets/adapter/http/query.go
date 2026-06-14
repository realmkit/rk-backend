package http

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/realmkit/rk-backend/module/assets/domain"
	"github.com/realmkit/rk-backend/module/assets/port"
	"github.com/realmkit/rk-backend/pkg/api/problem"
	"github.com/realmkit/rk-backend/pkg/search"
)

// assetFilterFromQuery returns asset filters from query params.
func assetFilterFromQuery(ctx *fiber.Ctx) (port.AssetFilter, error) {
	query, err := search.NewTextQuery(ctx.Query("q"), search.QueryOptions{})
	if err != nil {
		return port.AssetFilter{}, searchProblem(err)
	}
	sort, err := search.NewSort(
		ctx.Query("sort"),
		ctx.Query("direction"),
		port.DefaultAssetSort(),
		port.AllowedAssetSorts(),
	)
	if err != nil {
		return port.AssetFilter{}, searchProblem(err)
	}
	return port.AssetFilter{
		Namespace:  domain.Namespace(ctx.Query("namespace")),
		Path:       domain.VirtualPath(ctx.Query("path")),
		PathPrefix: domain.VirtualPath(ctx.Query("path_prefix")),
		Status:     domain.Status(ctx.Query("status")),
		Visibility: domain.Visibility(ctx.Query("visibility")),
		Query:      query,
		Sort:       sort,
	}, nil
}

// searchProblem maps invalid search parameters to a problem response.
func searchProblem(err error) error {
	code := "invalid_search"
	if errors.Is(err, search.ErrInvalidCursor) {
		code = "invalid_page_token"
	}
	return problem.Error{
		Problem: problem.New(fiber.StatusBadRequest, code, "Search parameters are invalid."),
	}
}
