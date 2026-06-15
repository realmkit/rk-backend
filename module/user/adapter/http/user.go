package http

import (
	"errors"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/module/groups/adapter/httpguard"
	groupsdomain "github.com/realmkit/rk-backend/module/groups/domain"
	"github.com/realmkit/rk-backend/module/groups/port"
	"github.com/realmkit/rk-backend/module/user/domain"
	userport "github.com/realmkit/rk-backend/module/user/port"
	"github.com/realmkit/rk-backend/pkg/api/principal"
	"github.com/realmkit/rk-backend/pkg/api/problem"
	"github.com/realmkit/rk-backend/pkg/search"
)

// updateCurrentRequest is the local current-user update body.
type updateCurrentRequest struct {
	AvatarAssetID *uuid.UUID `json:"avatar_asset_id"`
}

// currentUserResponse is the current user response body.
type currentUserResponse struct {
	User           domain.User        `json:"user"`
	ProviderClaims *claimCacheSummary `json:"provider_claims,omitempty"`
	Groups         *port.UserGroups   `json:"groups,omitempty"`
}

// userListResponse contains one user list page.
type userListResponse struct {
	Items         []userSummaryResponse `json:"items"`
	NextPageToken string                `json:"next_page_token,omitempty"`
	Query         string                `json:"query,omitempty"`
	Sort          string                `json:"sort,omitempty"`
	Direction     string                `json:"direction,omitempty"`
}

// userSummaryResponse contains one admin user list item.
type userSummaryResponse struct {
	User           domain.User        `json:"user"`
	ProviderClaims *claimCacheSummary `json:"provider_claims,omitempty"`
}

// claimCacheSummary exposes provider-owned cache data without raw subject.
type claimCacheSummary struct {
	Username        string    `json:"username"`
	Email           string    `json:"email"`
	EmailVerified   bool      `json:"email_verified"`
	DisplayName     string    `json:"display_name"`
	PictureURL      string    `json:"picture_url"`
	PreferredLocale string    `json:"preferred_locale"`
	ClaimsHash      string    `json:"claims_hash"`
	SyncedAt        time.Time `json:"synced_at"`
}

// accountURLResponse contains a provider account URL.
type accountURLResponse struct {
	URL string `json:"url"`
}

// currentUser returns the authenticated current user.
func (handler handler) currentUser(ctx *fiber.Ctx) error {
	current, err := principal.Require(ctx)
	if err != nil {
		return handleError(ctx, err)
	}
	user, err := handler.services.Users.Current(ctx.UserContext(), current.UserID)
	if err != nil {
		return handleError(ctx, err)
	}
	response := currentUserResponse{User: user.User}
	if user.Claims != nil {
		response.ProviderClaims = claimsSummary(*user.Claims)
	}
	if handler.services.Groups != nil {
		groups, err := handler.services.Groups.ListUserGroups(ctx.UserContext(), current.UserID)
		if err != nil {
			return handleError(ctx, err)
		}
		response.Groups = &groups
	}
	setETag(ctx, user.User.Version)
	return writeJSON(ctx, fiber.StatusOK, response)
}

// listUsers returns searchable users.
func (handler handler) listUsers(ctx *fiber.Ctx) error {
	current, err := principal.Require(ctx)
	if err != nil {
		return handleError(ctx, err)
	}
	if err := httpguard.Check(
		ctx,
		handler.services.Checker,
		current.UserID,
		httpguard.All(groupsdomain.PermissionUsersRead, groupsdomain.ObjectUser),
	); err != nil {
		return err
	}
	page, err := pageFromQuery(ctx)
	if err != nil {
		return err
	}
	filter, err := userFilterFromQuery(ctx)
	if err != nil {
		return err
	}
	result, err := handler.services.Users.List(ctx.UserContext(), filter, page)
	if err != nil {
		return handleError(ctx, err)
	}
	return writeJSON(ctx, fiber.StatusOK, userListResponse{
		Items:         userSummaries(result.Items),
		NextPageToken: result.NextCursor,
		Query:         filter.Query.String(),
		Sort:          filter.Sort.Key,
		Direction:     string(filter.Sort.Direction),
	})
}

// updateCurrentUser updates local current-user settings.
func (handler handler) updateCurrentUser(ctx *fiber.Ctx) error {
	if err := requireIdempotency(ctx); err != nil {
		return err
	}
	current, err := principal.Require(ctx)
	if err != nil {
		return handleError(ctx, err)
	}
	version, err := expectedVersion(ctx)
	if err != nil {
		return err
	}
	var request updateCurrentRequest
	if err := decodeJSON(ctx, &request); err != nil {
		return err
	}
	user, err := handler.services.Users.UpdateCurrent(
		ctx.UserContext(),
		userport.UpdateCurrentCommand{UserID: current.UserID, AvatarAssetID: request.AvatarAssetID, ExpectedVersion: version},
	)
	if err != nil {
		return handleError(ctx, err)
	}
	setETag(ctx, user.Version)
	return writeJSON(ctx, fiber.StatusOK, user)
}

// accountURL returns provider account URL when configured.
func (handler handler) accountURL(ctx *fiber.Ctx) error {
	return handleError(ctx, errAccountURLUnavailable)
}

// claimsSummary maps claim cache to response summary.
func claimsSummary(cache domain.ClaimCache) *claimCacheSummary {
	return &claimCacheSummary{
		Username:        cache.Username,
		Email:           cache.Email,
		EmailVerified:   cache.EmailVerified,
		DisplayName:     cache.DisplayName,
		PictureURL:      cache.PictureURL,
		PreferredLocale: cache.PreferredLocale,
		ClaimsHash:      cache.ClaimsHash,
		SyncedAt:        cache.SyncedAt,
	}
}

// userFilterFromQuery maps query params into a user filter.
func userFilterFromQuery(ctx *fiber.Ctx) (userport.UserFilter, error) {
	query, err := search.NewTextQuery(ctx.Query("q"), search.QueryOptions{})
	if err != nil {
		return userport.UserFilter{}, searchProblem(err)
	}
	sort, err := search.NewSort(ctx.Query("sort"), ctx.Query("direction"), userport.DefaultUserSort(), userport.AllowedUserSorts())
	if err != nil {
		return userport.UserFilter{}, searchProblem(err)
	}
	return userport.UserFilter{
		Status: domain.Status(ctx.Query("status")),
		Query:  query,
		Sort:   sort,
	}, nil
}

// userSummaries maps port user summaries into HTTP responses.
func userSummaries(items []userport.UserSummary) []userSummaryResponse {
	result := make([]userSummaryResponse, 0, len(items))
	for _, item := range items {
		summary := userSummaryResponse{User: item.User}
		if item.Claims != nil {
			summary.ProviderClaims = claimsSummary(*item.Claims)
		}
		result = append(result, summary)
	}
	return result
}

// searchProblem maps invalid search parameters to a problem response.
func searchProblem(err error) error {
	code := "invalid_search"
	if errors.Is(err, search.ErrInvalidCursor) {
		code = "invalid_page_token"
	}
	return problem.Error{Problem: problem.New(fiber.StatusBadRequest, code, "Search parameters are invalid.")}
}
