package http

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/module/groups/domain"
	"github.com/realmkit/rk-backend/module/groups/port"
	userdomain "github.com/realmkit/rk-backend/module/user/domain"
	userport "github.com/realmkit/rk-backend/module/user/port"
)

// membershipRequest is the membership assign body.
type membershipRequest struct {
	Status           domain.MembershipStatus `json:"status"`
	AssignedByUserID *uuid.UUID              `json:"assigned_by_user_id"`
	AssignedReason   string                  `json:"assigned_reason"`
	StartsAt         *time.Time              `json:"starts_at"`
	ExpiresAt        *time.Time              `json:"expires_at"`
}

// membershipListResponse contains one membership page.
type membershipListResponse struct {
	Items         []membershipListItem `json:"items"`
	NextPageToken string               `json:"next_page_token,omitempty"`
}

// membershipListItem contains a membership with optional local user display data.
type membershipListItem struct {
	Membership domain.Membership       `json:"membership"`
	User       *membershipUserResponse `json:"user,omitempty"`
}

// membershipUserResponse contains one display-safe local user summary.
type membershipUserResponse struct {
	User           userdomain.User       `json:"user"`
	ProviderClaims *membershipClaimCache `json:"provider_claims,omitempty"`
}

// membershipClaimCache exposes provider-owned display claims without raw subject.
type membershipClaimCache struct {
	Username        string    `json:"username"`
	Email           string    `json:"email"`
	EmailVerified   bool      `json:"email_verified"`
	DisplayName     string    `json:"display_name"`
	PictureURL      string    `json:"picture_url"`
	PreferredLocale string    `json:"preferred_locale"`
	ClaimsHash      string    `json:"claims_hash"`
	SyncedAt        time.Time `json:"synced_at"`
}

// listGroupMembers lists memberships for a group.
func (handler handler) listGroupMembers(ctx *fiber.Ctx) error {
	groupID, err := idFromParam(ctx, "group_id")
	if err != nil {
		return err
	}
	page, err := pageFromQuery(ctx)
	if err != nil {
		return err
	}
	result, err := handler.services.Memberships.ListGroupMembers(ctx.UserContext(), groupID, page)
	if err != nil {
		return handleError(ctx, err)
	}
	items, err := handler.membershipItems(ctx.UserContext(), result.Items)
	if err != nil {
		return handleError(ctx, err)
	}
	return writeJSON(ctx, fiber.StatusOK, membershipListResponse{Items: items, NextPageToken: result.NextCursor})
}

// assignMembership assigns a membership.
func (handler handler) assignMembership(ctx *fiber.Ctx) error {
	if err := requireIdempotency(ctx); err != nil {
		return err
	}
	groupID, userID, err := membershipIDs(ctx)
	if err != nil {
		return err
	}
	var request membershipRequest
	if err := decodeJSON(ctx, &request); err != nil {
		return err
	}
	membership, err := handler.services.Memberships.Assign(
		ctx.UserContext(),
		port.AssignMembershipCommand{Membership: membershipFromRequest(groupID, userID, request)},
	)
	if err != nil {
		return handleError(ctx, err)
	}
	setETag(ctx, membership.Version)
	return writeJSON(ctx, fiber.StatusOK, membership)
}

// removeMembership removes a membership.
func (handler handler) removeMembership(ctx *fiber.Ctx) error {
	if err := requireIdempotency(ctx); err != nil {
		return err
	}
	groupID, userID, err := membershipIDs(ctx)
	if err != nil {
		return err
	}
	version, err := expectedVersion(ctx)
	if err != nil {
		return err
	}
	command := port.RemoveMembershipCommand{
		GroupID:         groupID,
		UserID:          userID,
		ExpectedVersion: &version,
	}
	if err := handler.services.Memberships.Remove(ctx.UserContext(), command); err != nil {
		return handleError(ctx, err)
	}
	return writeNoContent(ctx)
}

// listUserGroups lists groups for user.
func (handler handler) listUserGroups(ctx *fiber.Ctx) error {
	userID, err := idFromParam(ctx, "user_id")
	if err != nil {
		return err
	}
	return handler.writeUserGroups(ctx, userID)
}

// listCurrentUserGroups lists groups for current user.
func (handler handler) listCurrentUserGroups(ctx *fiber.Ctx) error {
	userID, err := currentUserID(ctx)
	if err != nil {
		return err
	}
	return handler.writeUserGroups(ctx, userID)
}

// writeUserGroups writes user group response.
func (handler handler) writeUserGroups(ctx *fiber.Ctx, userID uuid.UUID) error {
	groups, err := handler.services.Memberships.ListUserGroups(ctx.UserContext(), userID)
	if err != nil {
		return handleError(ctx, err)
	}
	return writeJSON(ctx, fiber.StatusOK, groups)
}

// membershipIDs parses membership path IDs.
func membershipIDs(ctx *fiber.Ctx) (uuid.UUID, uuid.UUID, error) {
	groupID, err := idFromParam(ctx, "group_id")
	if err != nil {
		return uuid.Nil, uuid.Nil, err
	}
	userID, err := idFromParam(ctx, "user_id")
	if err != nil {
		return uuid.Nil, uuid.Nil, err
	}
	return groupID, userID, nil
}

// membershipFromRequest maps HTTP request to membership.
func membershipFromRequest(groupID uuid.UUID, userID uuid.UUID, request membershipRequest) domain.Membership {
	return domain.Membership{
		GroupID:          groupID,
		UserID:           userID,
		Status:           request.Status,
		AssignedByUserID: request.AssignedByUserID,
		AssignedReason:   request.AssignedReason,
		StartsAt:         request.StartsAt,
		ExpiresAt:        request.ExpiresAt,
	}
}

// membershipItems maps memberships and enriches local user summaries when available.
func (handler handler) membershipItems(
	ctx context.Context,
	memberships []domain.Membership,
) ([]membershipListItem, error) {
	items := make([]membershipListItem, 0, len(memberships))
	summaries, err := handler.membershipUserSummaries(ctx, memberships)
	if err != nil {
		return nil, err
	}
	for _, membership := range memberships {
		item := membershipListItem{Membership: membership}
		if summary, ok := summaries[membership.UserID]; ok {
			response := membershipUserSummary(summary)
			item.User = &response
		}
		items = append(items, item)
	}
	return items, nil
}

// membershipUserSummaries loads user summaries for one membership page.
func (handler handler) membershipUserSummaries(
	ctx context.Context,
	memberships []domain.Membership,
) (map[uuid.UUID]userport.UserSummary, error) {
	if handler.services.Users == nil || len(memberships) == 0 {
		return map[uuid.UUID]userport.UserSummary{}, nil
	}
	ids := make([]uuid.UUID, 0, len(memberships))
	for _, membership := range memberships {
		ids = append(ids, membership.UserID)
	}
	return handler.services.Users.FindSummariesByIDs(ctx, ids)
}

// membershipUserSummary maps a user summary into the membership response shape.
func membershipUserSummary(summary userport.UserSummary) membershipUserResponse {
	result := membershipUserResponse{User: summary.User}
	if summary.Claims != nil {
		result.ProviderClaims = &membershipClaimCache{
			Username:        summary.Claims.Username,
			Email:           summary.Claims.Email,
			EmailVerified:   summary.Claims.EmailVerified,
			DisplayName:     summary.Claims.DisplayName,
			PictureURL:      summary.Claims.PictureURL,
			PreferredLocale: summary.Claims.PreferredLocale,
			ClaimsHash:      summary.Claims.ClaimsHash,
			SyncedAt:        summary.Claims.SyncedAt,
		}
	}
	return result
}
