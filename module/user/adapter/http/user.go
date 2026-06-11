package http

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/niflaot/gamehub-go/module/groups/port"
	"github.com/niflaot/gamehub-go/module/user/domain"
	userport "github.com/niflaot/gamehub-go/module/user/port"
	"github.com/niflaot/gamehub-go/pkg/api/principal"
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
	user, err := handler.services.Users.Current(ctx.Context(), current.UserID)
	if err != nil {
		return handleError(ctx, err)
	}
	response := currentUserResponse{User: user.User}
	if user.Claims != nil {
		response.ProviderClaims = claimsSummary(*user.Claims)
	}
	if handler.services.Groups != nil {
		groups, err := handler.services.Groups.ListUserGroups(ctx.Context(), current.UserID)
		if err != nil {
			return handleError(ctx, err)
		}
		response.Groups = &groups
	}
	setETag(ctx, user.User.Version)
	return writeJSON(ctx, fiber.StatusOK, response)
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
		ctx.Context(),
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
