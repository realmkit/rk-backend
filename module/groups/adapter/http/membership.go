package http

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/module/groups/adapter/httpguard"
	"github.com/realmkit/rk-backend/module/groups/domain"
	"github.com/realmkit/rk-backend/module/groups/port"
	userdomain "github.com/realmkit/rk-backend/module/user/domain"
	userport "github.com/realmkit/rk-backend/module/user/port"
)

// membershipRequest is the membership assign body.
type membershipRequest struct {
	Status           domain.MembershipStatus `json:"status"`              // Status stores the status value.
	AssignedByUserID *uuid.UUID              `json:"assigned_by_user_id"` // AssignedByUserID stores the assigned by user i d value.
	AssignedReason   string                  `json:"assigned_reason"`     // AssignedReason stores the assigned reason value.
	StartsAt         *time.Time              `json:"starts_at"`           // StartsAt stores the starts at value.
	ExpiresAt        *time.Time              `json:"expires_at"`          // ExpiresAt stores the expires at value.
}

// membershipListResponse contains one membership page.
type membershipListResponse struct {
	Items         []membershipListItem `json:"items"`                     // Items stores the items value.
	NextPageToken string               `json:"next_page_token,omitempty"` // NextPageToken stores the next page token value.
}

// membershipListItem contains a membership with optional local user display data.
type membershipListItem struct {
	Membership domain.Membership       `json:"membership"`     // Membership stores the membership value.
	User       *membershipUserResponse `json:"user,omitempty"` // User stores the user value.
}

// membershipUserResponse contains one display-safe local user summary.
type membershipUserResponse struct {
	User           userdomain.User       `json:"user"`                      // User stores the user value.
	ProviderClaims *membershipClaimCache `json:"provider_claims,omitempty"` // ProviderClaims stores the provider claims value.
}

// membershipClaimCache exposes provider-owned display claims without raw subject.
type membershipClaimCache struct {
	Username        string    `json:"username"`         // Username stores the username value.
	Email           string    `json:"email"`            // Email stores the email value.
	EmailVerified   bool      `json:"email_verified"`   // EmailVerified stores the email verified value.
	DisplayName     string    `json:"display_name"`     // DisplayName stores the display name value.
	PictureURL      string    `json:"picture_url"`      // PictureURL stores the picture u r l value.
	PreferredLocale string    `json:"preferred_locale"` // PreferredLocale stores the preferred locale value.
	ClaimsHash      string    `json:"claims_hash"`      // ClaimsHash stores the claims hash value.
	SyncedAt        time.Time `json:"synced_at"`        // SyncedAt stores the synced at value.
}

// listGroupMembers lists memberships for a group.
func (handler handler) listGroupMembers(ctx *fiber.Ctx) error {
	groupID, err := idFromParam(ctx, "group_id")
	if err != nil {
		return err
	}
	if _, err := httpguard.Require(
		ctx,
		handler.services.Checker,
		httpguard.Object(domain.PermissionGroupsReadMembers, domain.ObjectGroup, groupID),
	); err != nil {
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
	if _, err := httpguard.Require(
		ctx,
		handler.services.Checker,
		httpguard.Object(domain.PermissionGroupsAssignMember, domain.ObjectGroup, groupID),
	); err != nil {
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
	if _, err := httpguard.Require(
		ctx,
		handler.services.Checker,
		httpguard.Object(domain.PermissionGroupsAssignMember, domain.ObjectGroup, groupID),
	); err != nil {
		return err
	}
	version, err := expectedVersion(ctx)
	if err != nil {
		return err
	}
	if err := handler.services.Memberships.Remove(ctx.UserContext(), port.RemoveMembershipCommand{
		GroupID:         groupID,
		UserID:          userID,
		ExpectedVersion: &version,
	}); err != nil {
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
	if _, err := httpguard.RequireSelfOr(
		ctx,
		handler.services.Checker,
		userID,
		httpguard.All(domain.PermissionGroupsReadMembers, domain.ObjectGroup),
	); err != nil {
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
func (handler handler) membershipItems(ctx context.Context, memberships []domain.Membership) ([]membershipListItem, error) {
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
func (handler handler) membershipUserSummaries(ctx context.Context, memberships []domain.Membership) (map[uuid.UUID]userport.UserSummary, error) {
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
